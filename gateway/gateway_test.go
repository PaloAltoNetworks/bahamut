package gateway

import (
	"crypto/tls"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/opentracing/opentracing-go"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/vulcand/oxy/ratelimit"
	"go.aporeto.io/bahamut"
	"go.aporeto.io/tg/tglib"
)

type simpleUpstreamer struct {
	ups1    *httptest.Server
	ups2    *httptest.Server
	nextErr error
}

func (u *simpleUpstreamer) Upstream(req *http.Request) (upstream string, err error) {

	if u.nextErr != nil {
		e := u.nextErr
		u.nextErr = nil
		return "", e
	}
	switch req.URL.Path {
	case "/ups1":
		return strings.Replace(u.ups1.URL, "https://", "", 1), nil
	case "/ups2":
		return strings.Replace(u.ups2.URL, "https://", "", 1), nil
	default:
		return "", nil
	}
}

// Implement LatencyBasedUpstreamer interface
func (u *simpleUpstreamer) CollectLatency(address string, rt time.Duration) {
	// noop
}

type fakeMetricManager struct {
	registerWSConnectionCalled    int64
	unregisterWSConnectionCalled  int64
	registerTCPConnectionCalled   int64
	unregisterTCPConnectionCalled int64
}

func (m *fakeMetricManager) MeasureRequest(method string, url string) bahamut.FinishMeasurementFunc {
	return func(code int, span opentracing.Span) time.Duration { return 0 }
}

func (m *fakeMetricManager) RegisterWSConnection() {
	atomic.AddInt64(&m.registerWSConnectionCalled, 1)
}
func (m *fakeMetricManager) UnregisterWSConnection() {
	atomic.AddInt64(&m.unregisterWSConnectionCalled, 1)
}
func (m *fakeMetricManager) RegisterTCPConnection() {
	atomic.AddInt64(&m.registerTCPConnectionCalled, 1)
}
func (m *fakeMetricManager) UnregisterTCPConnection() {
	atomic.AddInt64(&m.unregisterTCPConnectionCalled, 1)
}
func (m *fakeMetricManager) Write(w http.ResponseWriter, r *http.Request) {}

func makeServerCert() tls.Certificate {
	certPem, keyPem, err := tglib.Issue(pkix.Name{}, tglib.OptIssueTypeServerAuth())
	if err != nil {
		panic(err)
	}

	cert, key, err := tglib.ReadCertificate(pem.EncodeToMemory(certPem), pem.EncodeToMemory(keyPem), "")
	if err != nil {
		panic(err)
	}

	tlsCert, err := tglib.ToTLSCertificate(cert, key)
	if err != nil {
		panic(err)
	}

	return tlsCert
}

type simpleLimiter struct{}

func (l *simpleLimiter) DefaultRates() *ratelimit.RateSet {
	rl := ratelimit.NewRateSet()
	rl.Add(time.Second, 100, 1000)
	return rl
}

func (l *simpleLimiter) ExtractRates(r *http.Request) (*ratelimit.RateSet, error) {
	rl := ratelimit.NewRateSet()
	rl.Add(time.Second, 100, 1000)
	return rl, nil
}

func (l *simpleLimiter) ExtractSource(req *http.Request) (token string, amount int64, err error) {
	return "default", 1, nil
}

func TestGateway(t *testing.T) {

	Convey("Given I have 2 tls upstreams and an Upstreamer", t, func() {

		ups1 := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			if r.Header.Get("inject") != "" {
				w.Header().Add("inject", r.Header.Get("inject"))
			}

			w.WriteHeader(601)
		}))
		defer ups1.Close()

		ups2 := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(602)
		}))
		defer ups2.Close()

		u := &simpleUpstreamer{
			ups1: ups1,
			ups2: ups2,
		}

		Convey("When I start the gateway with no front end TLS config", func() {

			mm := &fakeMetricManager{}

			gw, err := New(
				"127.0.0.1:7765",
				u,
				OptionUpstreamTLSConfig(&tls.Config{InsecureSkipVerify: true}),
				OptionEnableProxyProtocol(true, "0.0.0.0/0"),
				OptionRateLimiter(&simpleLimiter{}),
				OptionTCPRateLimiting(true, 200.0, 200.0, 100),
				OptionUpstreamConfig(0, 0, 0, 0, 0, "NetworkErrorRatio() > 0.5", false),
				OptionEnableTrace(true),
				OptionMetricsManager(mm),
			)
			defer gw.Stop()

			So(err, ShouldBeNil)
			So(gw, ShouldNotBeNil)

			testclient := &http.Client{
				Transport: &http.Transport{
					TLSClientConfig: &tls.Config{
						InsecureSkipVerify: true,
					},
				},
			}

			gw.Start()

			Convey("Then we I call existing ep 1", func() {
				req, _ := http.NewRequest(http.MethodGet, "http://127.0.0.1:7765/ups1", nil)
				req.Close = true
				resp, err := testclient.Do(req)
				So(err, ShouldBeNil)
				So(resp.StatusCode, ShouldEqual, 601)
			})

			Convey("Then we I call existing ep2", func() {
				req, _ := http.NewRequest(http.MethodGet, "http://127.0.0.1:7765/ups2", nil)
				req.Close = true
				resp, err := testclient.Do(req)
				So(err, ShouldBeNil)
				So(resp.StatusCode, ShouldEqual, 602)
			})

			Convey("Then we I call existing ep3", func() {
				req, _ := http.NewRequest(http.MethodGet, "http://127.0.0.1:7765/ups3", nil)
				req.Close = true
				resp, err := testclient.Do(req)
				So(err, ShouldBeNil)
				So(resp.StatusCode, ShouldEqual, 503)
			})

			Convey("Then we I call and get a ErrUpstreamerTooManyRequests", func() {
				u.nextErr = ErrUpstreamerTooManyRequests
				req, _ := http.NewRequest(http.MethodGet, "http://127.0.0.1:7765/ups3", nil)
				req.Close = true
				resp, err := testclient.Do(req)
				So(err, ShouldBeNil)
				So(resp.StatusCode, ShouldEqual, http.StatusTooManyRequests)
			})

			Convey("Then we I call and get an unknown error", func() {
				u.nextErr = fmt.Errorf("oh no")
				req, _ := http.NewRequest(http.MethodGet, "http://127.0.0.1:7765/ups3", nil)
				req.Close = true
				resp, err := testclient.Do(req)
				So(err, ShouldBeNil)
				So(resp.StatusCode, ShouldEqual, http.StatusInternalServerError)
			})

			// Convey("Then the metric manager should have been called", func() {
			// 	So(atomic.AddInt64(&mm.registerTCPConnectionCalled, 0), ShouldBeGreaterThan, 0)
			// 	So(atomic.AddInt64(&mm.unregisterTCPConnectionCalled, 0), ShouldBeGreaterThan, 0)
			// })
		})

		Convey("When I start the gateway in maintenance", func() {

			gw, err := New(
				"127.0.0.1:7765",
				u,
				OptionUpstreamTLSConfig(&tls.Config{InsecureSkipVerify: true}),
				OptionEnableProxyProtocol(true, "0.0.0.0/0"),
				OptionRateLimiter(&simpleLimiter{}),
				OptionEnableMaintenance(true),
			)
			defer gw.Stop()

			So(err, ShouldBeNil)
			So(gw, ShouldNotBeNil)

			testclient := &http.Client{
				Transport: &http.Transport{
					TLSClientConfig: &tls.Config{
						InsecureSkipVerify: true,
					},
				},
			}

			gw.Start()

			Convey("Then calling a GET any api will return 423", func() {
				req, _ := http.NewRequest(http.MethodGet, "http://127.0.0.1:7765/ups1", nil)
				resp, err := testclient.Do(req)
				So(err, ShouldBeNil)
				So(resp.StatusCode, ShouldEqual, 423)
			})

			Convey("Then calling a OPTION any api will return 200", func() {
				req, _ := http.NewRequest(http.MethodOptions, "http://127.0.0.1:7765/ups1", nil)
				resp, err := testclient.Do(req)
				So(err, ShouldBeNil)
				So(resp.StatusCode, ShouldEqual, 200)
			})
		})

		Convey("When I start the gateway with a custom request rewriter and response rewriter", func() {

			gw, err := New(
				"127.0.0.1:7765",
				u,
				OptionUpstreamTLSConfig(&tls.Config{InsecureSkipVerify: true}),
				OptionEnableProxyProtocol(true, "0.0.0.0/0"),
				OptionRateLimiter(&simpleLimiter{}),
				OptionSetCustomRequestRewriter(func(req *http.Request, private bool) error {
					req.Header.Add("inject", "hello")
					return nil
				}),
				OptionSetCustomResponseRewriter(func(req *http.Response) error {
					req.Header.Add("from-response", "hello")
					return nil
				}),
			)
			defer gw.Stop()

			So(err, ShouldBeNil)
			So(gw, ShouldNotBeNil)

			testclient := &http.Client{
				Transport: &http.Transport{
					TLSClientConfig: &tls.Config{
						InsecureSkipVerify: true,
					},
				},
			}

			gw.Start()

			Convey("Then we I call existing ep 1", func() {
				req, _ := http.NewRequest(http.MethodGet, "http://127.0.0.1:7765/ups1", nil)
				resp, err := testclient.Do(req)
				So(err, ShouldBeNil)
				So(resp.StatusCode, ShouldEqual, 601)
				So(resp.Header.Get("inject"), ShouldEqual, "hello")
				So(resp.Header.Get("from-response"), ShouldEqual, "hello")
			})
		})

		Convey("When I start the gateway with a custom exact handler that handles the request", func() {

			gw, err := New(
				"127.0.0.1:7765",
				u,
				OptionUpstreamTLSConfig(&tls.Config{InsecureSkipVerify: true}),
				OptionRegisterExactInterceptor("/hello", func(w http.ResponseWriter, req *http.Request, ew ErrorWriter) (InterceptorAction, string, error) {
					w.WriteHeader(604)
					return InterceptorActionStop, "", nil
				}),
			)
			defer gw.Stop()

			So(err, ShouldBeNil)
			So(gw, ShouldNotBeNil)

			testclient := &http.Client{
				Transport: &http.Transport{
					TLSClientConfig: &tls.Config{
						InsecureSkipVerify: true,
					},
				},
			}

			gw.Start()

			Convey("Then we I call existing ep 1", func() {
				req, _ := http.NewRequest(http.MethodGet, "http://127.0.0.1:7765/hello", nil)
				resp, err := testclient.Do(req)
				So(err, ShouldBeNil)
				So(resp.StatusCode, ShouldEqual, 604)
			})
		})

		Convey("When I start the gateway with a custom exact handler that modifies the request", func() {

			gw, err := New(
				"127.0.0.1:7765",
				u,
				OptionMetricsManager(&fakeMetricManager{}),
				OptionUpstreamTLSConfig(&tls.Config{InsecureSkipVerify: true}),
				OptionRegisterExactInterceptor("/ups1", func(w http.ResponseWriter, req *http.Request, ew ErrorWriter) (InterceptorAction, string, error) {
					return InterceptorActionForwardWS, strings.Replace(u.ups2.URL, "https://", "", 1), nil
				}),
			)
			defer gw.Stop()

			So(err, ShouldBeNil)
			So(gw, ShouldNotBeNil)

			testclient := &http.Client{
				Transport: &http.Transport{
					TLSClientConfig: &tls.Config{
						InsecureSkipVerify: true,
					},
				},
			}

			gw.Start()

			Convey("Then we I call existing ep 1", func() {
				req, _ := http.NewRequest(http.MethodGet, "http://127.0.0.1:7765/ups1", nil)
				resp, err := testclient.Do(req)
				So(err, ShouldBeNil)
				So(resp.StatusCode, ShouldEqual, 602)
			})
		})

		Convey("When I start the gateway with a custom suffix handler that handles the request", func() {

			gw, err := New(
				"127.0.0.1:7765",
				u,
				OptionUpstreamTLSConfig(&tls.Config{InsecureSkipVerify: true}),
				OptionRegisterSuffixInterceptor("/hello", func(w http.ResponseWriter, req *http.Request, ew ErrorWriter) (InterceptorAction, string, error) {
					w.WriteHeader(604)
					return InterceptorActionStop, "", nil
				}),
			)
			defer gw.Stop()

			So(err, ShouldBeNil)
			So(gw, ShouldNotBeNil)

			testclient := &http.Client{
				Transport: &http.Transport{
					TLSClientConfig: &tls.Config{
						InsecureSkipVerify: true,
					},
				},
			}

			gw.Start()

			Convey("Then we I call existing ep 1", func() {
				req, _ := http.NewRequest(http.MethodGet, "http://127.0.0.1:7765/chien/hello", nil)
				resp, err := testclient.Do(req)
				So(err, ShouldBeNil)
				So(resp.StatusCode, ShouldEqual, 604)
			})
		})

		Convey("When I start the gateway with a custom suffix handler that modifies the request", func() {

			gw, err := New(
				"127.0.0.1:7765",
				u,
				OptionMetricsManager(&fakeMetricManager{}),
				OptionUpstreamTLSConfig(&tls.Config{InsecureSkipVerify: true}),
				OptionRegisterSuffixInterceptor("/ups1", func(w http.ResponseWriter, req *http.Request, ew ErrorWriter) (InterceptorAction, string, error) {
					return InterceptorActionForwardDirect, strings.Replace(u.ups2.URL, "https://", "", 1), nil
				}),
			)
			defer gw.Stop()

			So(err, ShouldBeNil)
			So(gw, ShouldNotBeNil)

			testclient := &http.Client{
				Transport: &http.Transport{
					TLSClientConfig: &tls.Config{
						InsecureSkipVerify: true,
					},
				},
			}

			gw.Start()

			Convey("Then we I call existing ep 1", func() {
				req, _ := http.NewRequest(http.MethodGet, "http://127.0.0.1:7765/chien/ups1", nil)
				resp, err := testclient.Do(req)
				So(err, ShouldBeNil)
				So(resp.StatusCode, ShouldEqual, 602)
			})
		})

		Convey("When I start the gateway with a custom prefix handler that returns an error", func() {

			gw, err := New(
				"127.0.0.1:7765",
				u,
				OptionUpstreamTLSConfig(&tls.Config{InsecureSkipVerify: true}),
				OptionRegisterPrefixInterceptor("/ohnows", func(w http.ResponseWriter, req *http.Request, ew ErrorWriter) (InterceptorAction, string, error) {
					return InterceptorActionForward, "", fmt.Errorf("boom")
				}),
				OptionRegisterPrefixInterceptor("/ups1", func(w http.ResponseWriter, req *http.Request, ew ErrorWriter) (InterceptorAction, string, error) {
					return InterceptorActionForward, "", fmt.Errorf("boom")
				}),
			)
			defer gw.Stop()

			So(err, ShouldBeNil)
			So(gw, ShouldNotBeNil)

			testclient := &http.Client{
				Transport: &http.Transport{
					TLSClientConfig: &tls.Config{
						InsecureSkipVerify: true,
					},
				},
			}

			gw.Start()

			Convey("Then we I call existing ep 1", func() {
				req, _ := http.NewRequest(http.MethodGet, "http://127.0.0.1:7765/ups1/chien", nil)
				resp, err := testclient.Do(req)
				So(err, ShouldBeNil)
				So(resp.StatusCode, ShouldEqual, 500)
			})
		})

		Convey("When I start the gateway with front end TLS config with proxyprotocol enabled", func() {

			gw, err := New(
				"127.0.0.1:7765",
				u,
				OptionServerTLSConfig(&tls.Config{Certificates: []tls.Certificate{makeServerCert()}}),
				OptionUpstreamTLSConfig(&tls.Config{InsecureSkipVerify: true}),
				OptionTCPRateLimiting(true, 200.0, 200.0, 100),
				OptionEnableProxyProtocol(true, "0.0.0.0/0"),
			)
			defer gw.Stop()

			So(err, ShouldBeNil)
			So(gw, ShouldNotBeNil)

			testclient := &http.Client{
				Transport: &http.Transport{
					TLSClientConfig: &tls.Config{
						InsecureSkipVerify: true,
					},
				},
			}

			gw.Start()

			Convey("Then we I call existing ep 1", func() {
				req, _ := http.NewRequest(http.MethodGet, "https://127.0.0.1:7765/ups1", nil)
				resp, err := testclient.Do(req)
				So(err, ShouldBeNil)
				So(resp.StatusCode, ShouldEqual, 601)
			})

			Convey("Then we I call existing ep2", func() {
				req, _ := http.NewRequest(http.MethodGet, "https://127.0.0.1:7765/ups2", nil)
				resp, err := testclient.Do(req)
				So(err, ShouldBeNil)
				So(resp.StatusCode, ShouldEqual, 602)
			})
		})

		Convey("When I start the gateway with front end TLS config without proxyprotocol enabled", func() {

			gw, err := New(
				"127.0.0.1:7765",
				u,
				OptionServerTLSConfig(&tls.Config{Certificates: []tls.Certificate{makeServerCert()}}),
				OptionUpstreamTLSConfig(&tls.Config{InsecureSkipVerify: true}),
				OptionTCPRateLimiting(true, 200.0, 200.0, 100),
			)
			defer gw.Stop()

			So(err, ShouldBeNil)
			So(gw, ShouldNotBeNil)

			testclient := &http.Client{
				Transport: &http.Transport{
					TLSClientConfig: &tls.Config{
						InsecureSkipVerify: true,
					},
				},
			}

			gw.Start()

			Convey("Then we I call existing ep 1", func() {
				req, _ := http.NewRequest(http.MethodGet, "https://127.0.0.1:7765/ups1", nil)
				resp, err := testclient.Do(req)
				So(err, ShouldBeNil)
				So(resp.StatusCode, ShouldEqual, 601)
			})

			Convey("Then we I call existing ep2", func() {
				req, _ := http.NewRequest(http.MethodGet, "https://127.0.0.1:7765/ups2", nil)
				resp, err := testclient.Do(req)
				So(err, ShouldBeNil)
				So(resp.StatusCode, ShouldEqual, 602)
			})
		})

		Convey("When I start the gateway with front end TLS config with proxyprotocol enabled and a bad subnet", func() {

			gw, err := New(
				"127.0.0.1:7765",
				u,
				OptionServerTLSConfig(&tls.Config{Certificates: []tls.Certificate{makeServerCert()}}),
				OptionUpstreamTLSConfig(&tls.Config{InsecureSkipVerify: true}),
				OptionTCPRateLimiting(true, 200.0, 200.0, 100),
				OptionEnableProxyProtocol(true, "oopsy"),
			)

			So(gw, ShouldBeNil)
			So(err, ShouldNotBeNil)
		})

	})
}
