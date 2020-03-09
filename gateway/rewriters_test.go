package gateway

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"go.aporeto.io/tg/tglib"
)

func Test_requestRewriter(t *testing.T) {

	Convey("Given I have a requestRewriter", t, func() {

		rw := &requestRewriter{}

		customReqriter := func(req *http.Request, private bool) error {
			switch private {
			case true:
				req.Header.Add("X-Aporeto-Gateway", "private")
			case false:
				req.Header.Add("X-Aporeto-Gateway", "public")
			}

			return nil
		}

		Convey("When I call Rewrite with a custom rewriter handling private mode", func() {

			rw.private = true
			rw.customRewriter = customReqriter

			r, _ := http.NewRequest(http.MethodGet, "http://127.0.0.1", nil)
			r.TLS = &tls.ConnectionState{}

			rw.Rewrite(r)

			Convey("Then the response should be correct", func() {
				So(r.Header.Get("X-Aporeto-Gateway"), ShouldEqual, "private")
			})
		})

		Convey("When I call Rewrite with a custom rewriter handling public mode", func() {

			rw.private = false
			rw.customRewriter = customReqriter

			r, _ := http.NewRequest(http.MethodGet, "http://127.0.0.1", nil)
			r.TLS = &tls.ConnectionState{}

			rw.Rewrite(r)

			Convey("Then the response should be correct", func() {
				So(r.Header.Get("X-Aporeto-Gateway"), ShouldEqual, "public")
			})
		})

		Convey("When I call Rewrite it with custom X-Forwarded-For and X-Real-IP", func() {

			r, _ := http.NewRequest(http.MethodGet, "http://127.0.0.1", nil)
			r.TLS = &tls.ConnectionState{}

			r.Header.Set("X-Forwarded-For ", "A")
			r.Header.Set("X-Real-IP", "B")

			rw.Rewrite(r)

			Convey("Then the response should be correct", func() {
				So(r.Header.Get("X-Forwarded-For"), ShouldEqual, "")
				So(r.Header.Get("X-Real-IP"), ShouldEqual, "")
			})
		})

		Convey("When I call Rewrite it with a valid TLS client certificate", func() {

			certData := []byte(`-----BEGIN CERTIFICATE-----
MIIBKjCB0qADAgECAhBLliCl1URppVpoHheDFLdKMAoGCCqGSM49BAMCMA8xDTAL
BgNVBAMTBHRvdG8wHhcNMjAwMjI4MTgzMTA2WhcNMzAwMTA2MTgzMTA2WjAPMQ0w
CwYDVQQDEwR0b3RvMFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEy6WbbGdtOM0b
DXryjj9EZiFKphTurtQTeo5whOoTTodNHJwMxAj2pSl+lAYDxokdk6PdUd/jZ5s2
+LXXqIhT0aMQMA4wDAYDVR0TAQH/BAIwADAKBggqhkjOPQQDAgNHADBEAiAxmm8w
Ag5DuK1V5vjCqZeuXWyVrfoL3rbMHdrpsYVqSAIgeY9F6wqMEpqIPjAtbCkSC+DG
f6eiTREm5FRLzNkfhxQ=
-----END CERTIFICATE-----
`,
			)

			cert, err := tglib.ParseCertificate(certData)
			if err != nil {
				panic(err)
			}

			r, _ := http.NewRequest(http.MethodGet, "http://127.0.0.1", nil)
			r.TLS = &tls.ConnectionState{
				PeerCertificates: []*x509.Certificate{cert},
			}

			rw.Rewrite(r)

			Convey("Then the response should be correct", func() {
				So(r.Header.Get("X-TLS-Client-Certificate"), ShouldEqual, strings.ReplaceAll(string(certData), "\n", " "))
			})
		})

		Convey("When I call Rewrite it with a invalid TLS client certificate", func() {

			r, _ := http.NewRequest(http.MethodGet, "http://127.0.0.1", nil)
			r.TLS = &tls.ConnectionState{
				PeerCertificates: []*x509.Certificate{nil},
			}

			Convey("Then the response should be correct", func() {
				So(func() { rw.Rewrite(r) }, ShouldPanicWith, "unable to handle client TLS certificate: nil certificate provided")
			})
		})

		Convey("When I call Rewrite without blocking opentracing", func() {

			r, _ := http.NewRequest(http.MethodGet, "http://127.0.0.1", nil)
			r.TLS = &tls.ConnectionState{}

			r.Header.Set("X-B3-TraceID", "X-B3-TraceID")
			r.Header.Set("X-B3-SpanID", "X-B3-SpanID")
			r.Header.Set("X-B3-ParentSpanID", "X-B3-ParentSpanID")
			r.Header.Set("X-B3-Sampled", "X-B3-Sampled")
			r.Header.Set("Uber-Trace-ID", "Uber-Trace-ID")
			r.Header.Set("Jaeger-Baggage", "Jaeger-Baggage")
			r.Header.Set("TraceParent", "TraceParent")
			r.Header.Set("TraceState", "TraceState")

			rw.Rewrite(r)

			Convey("Then the response should be correct", func() {
				So(r.Header.Get("X-B3-TraceID"), ShouldEqual, "X-B3-TraceID")
				So(r.Header.Get("X-B3-SpanID"), ShouldEqual, "X-B3-SpanID")
				So(r.Header.Get("X-B3-ParentSpanID"), ShouldEqual, "X-B3-ParentSpanID")
				So(r.Header.Get("X-B3-Sampled"), ShouldEqual, "X-B3-Sampled")
				So(r.Header.Get("Uber-Trace-ID"), ShouldEqual, "Uber-Trace-ID")
				So(r.Header.Get("Jaeger-Baggage"), ShouldEqual, "Jaeger-Baggage")
				So(r.Header.Get("TraceParent"), ShouldEqual, "TraceParent")
				So(r.Header.Get("TraceState"), ShouldEqual, "TraceState")

			})
		})

		Convey("When I call Rewrite blocking opentracing", func() {

			rw.blockOpenTracing = true

			r, _ := http.NewRequest(http.MethodGet, "http://127.0.0.1", nil)
			r.TLS = &tls.ConnectionState{}

			r.Header.Set("X-B3-TraceID", "X-B3-TraceID")
			r.Header.Set("X-B3-SpanID", "X-B3-SpanID")
			r.Header.Set("X-B3-ParentSpanID", "X-B3-ParentSpanID")
			r.Header.Set("X-B3-Sampled", "X-B3-Sampled")
			r.Header.Set("Uber-Trace-ID", "Uber-Trace-ID")
			r.Header.Set("Jaeger-Baggage", "Jaeger-Baggage")
			r.Header.Set("TraceParent", "TraceParent")
			r.Header.Set("TraceState", "TraceState")

			rw.Rewrite(r)

			Convey("Then the response should be correct", func() {
				So(r.Header.Get("X-B3-TraceID"), ShouldEqual, "")
				So(r.Header.Get("X-B3-SpanID"), ShouldEqual, "")
				So(r.Header.Get("X-B3-ParentSpanID"), ShouldEqual, "")
				So(r.Header.Get("X-B3-Sampled"), ShouldEqual, "")
				So(r.Header.Get("Uber-Trace-ID"), ShouldEqual, "")
				So(r.Header.Get("Jaeger-Baggage"), ShouldEqual, "")
				So(r.Header.Get("TraceParent"), ShouldEqual, "")
				So(r.Header.Get("TraceState"), ShouldEqual, "")

			})
		})

		Convey("When I call Rewrite with a custom rewriter that works", func() {

			rw.customRewriter = func(resp *http.Request, private bool) error {
				resp.Header.Set("yey", "ouais")
				return nil
			}

			r, _ := http.NewRequest(http.MethodGet, "http://127.0.0.1", nil)
			r.TLS = &tls.ConnectionState{}

			rw.Rewrite(r)

			Convey("Then the response should be correct", func() {
				So(r.Header.Get("yey"), ShouldEqual, "ouais")
			})
		})

		Convey("When I call Rewrite with a custom rewriter that returns an error", func() {

			rw.customRewriter = func(resp *http.Request, private bool) error {
				resp.Header.Set("yey", "ouais")
				return fmt.Errorf("boom")
			}

			r, _ := http.NewRequest(http.MethodGet, "http://127.0.0.1", nil)
			r.TLS = &tls.ConnectionState{}

			Convey("Then the response should be correct", func() {
				So(func() { rw.Rewrite(r) }, ShouldPanicWith, "unable to rewrite request with custom rewriter: boom")
			})
		})
	})
}

func Test_sourceExtractor_Extract(t *testing.T) {
	type args struct {
		req *http.Request
	}
	tests := []struct {
		name       string
		s          *sourceExtractor
		args       args
		wantToken  string
		wantAmount int64
		wantErr    bool
	}{
		{
			"authenticated",
			&sourceExtractor{},
			args{
				&http.Request{
					URL: &url.URL{Path: "/toto"},
					Header: http.Header{
						"Authorization": []string{"Bearer X"},
					},
				},
			},
			"Bearer X",
			1,
			false,
		},
		{
			"unauthenticated",
			&sourceExtractor{},
			args{
				&http.Request{
					URL:    &url.URL{Path: "/toto"},
					Header: http.Header{},
				},
			},
			"default",
			1,
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &sourceExtractor{}
			gotToken, gotAmount, err := s.Extract(tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("sourceExtractor.Extract() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotToken != tt.wantToken {
				t.Errorf("sourceExtractor.Extract() gotToken = %v, want %v", gotToken, tt.wantToken)
			}
			if gotAmount != tt.wantAmount {
				t.Errorf("sourceExtractor.Extract() gotAmount = %v, want %v", gotAmount, tt.wantAmount)
			}
		})
	}
}

func Test_circuitBreakerHandler(t *testing.T) {

	Convey("Given I have a circuitBreakerHandler", t, func() {
		cb := &circuitBreakerHandler{}

		Convey("When I call ServeHTTP", func() {

			recorder := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodGet, "http://127.0.0.1", nil)
			cb.ServeHTTP(recorder, req)

			Convey("Then the response should be correct", func() {
				data, _ := ioutil.ReadAll(recorder.Body)
				So(string(data), ShouldEqual, `[{"code":503,"description":"The service is busy handling requests. Please retry in a moment","subject":"gateway","title":"Service Unavailable"}]`)
			})
		})
	})
}
