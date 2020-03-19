package gateway

import (
	"crypto/tls"
	"net/http"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
	"go.aporeto.io/bahamut"
)

func Test_Options(t *testing.T) {

	c := newGatewayConfig()

	Convey("Calling OptionEnableProxyProtocol should work", t, func() {
		OptionEnableProxyProtocol(true, "10.0.0.0/0")(c)
		So(c.proxyProtocolEnabled, ShouldEqual, true)
		So(c.proxyProtocolSubnet, ShouldEqual, "10.0.0.0/0")
	})

	Convey("Calling OptionTCPRateLimiting should work", t, func() {
		OptionTCPRateLimiting(true, 1.0, 2, 3)(c)
		So(c.tcpRateLimitingEnabled, ShouldEqual, true)
		So(c.tcpRateLimitingCPS, ShouldEqual, 1.0)
		So(c.tcpRateLimitingBurst, ShouldEqual, 2)
		So(c.tcpMaxConnections, ShouldEqual, 3)
	})

	Convey("Calling OptionRateLimiting should work", t, func() {
		OptionRateLimiting(true, 1.0, 2)(c)
		So(c.rateLimitingEnabled, ShouldEqual, true)
		So(c.rateLimitingRPS, ShouldEqual, 1.0)
		So(c.rateLimitingBurst, ShouldEqual, 2)
	})

	Convey("Calling OptionEnableTrace should work", t, func() {
		OptionEnableTrace(true)(c)
		So(c.trace, ShouldEqual, true)
	})

	Convey("Calling OptionEnableMaintenance should work", t, func() {
		OptionEnableMaintenance(true)(c)
		So(c.maintenance, ShouldEqual, true)
	})

	Convey("Calling OptionHTTPTimeouts should work", t, func() {
		OptionHTTPTimeouts(time.Second, time.Minute, time.Hour, true)(c)
		So(c.httpReadTimeout, ShouldEqual, time.Second)
		So(c.httpWriteTimeout, ShouldEqual, time.Minute)
		So(c.httpIdleTimeout, ShouldEqual, time.Hour)
		So(c.httpDisableKeepAlive, ShouldEqual, true)
	})

	Convey("Calling OptionExposePrivateAPIs should work", t, func() {
		OptionExposePrivateAPIs(true)(c)
		So(c.exposePrivateAPIs, ShouldEqual, true)
	})

	Convey("Calling OptionBlockOpenTracingHeaders should work", t, func() {
		OptionBlockOpenTracingHeaders(true)(c)
		So(c.blockOpenTracingHeaders, ShouldEqual, true)
	})

	Convey("Calling OptionUpstreamConfig should work", t, func() {
		OptionUpstreamConfig(1, 2, 3, time.Second, time.Minute, "hello = 1", true)(c)
		So(c.upstreamMaxConnsPerHost, ShouldEqual, 1)
		So(c.upstreamMaxIdleConns, ShouldEqual, 2)
		So(c.upstreamMaxIdleConnsPerHost, ShouldEqual, 3)
		So(c.upstreamTLSHandshakeTimeout, ShouldEqual, time.Second)
		So(c.upstreamIdleConnTimeout, ShouldEqual, time.Minute)
		So(c.upstreamUseHTTP2, ShouldEqual, true)
		So(c.upstreamCircuitBreakerCond, ShouldEqual, "hello = 1")
	})

	Convey("Calling OptionRegisterPrefixInterceptor should work", t, func() {
		f := func(http.ResponseWriter, *http.Request, ErrorWriter) (InterceptorAction, string, error) {
			return InterceptorActionForward, "", nil
		}
		OptionRegisterPrefixInterceptor("/prefix", f)(c)
		So(c.prefixInterceptors["/prefix"], ShouldEqual, f)
	})

	Convey("Calling OptionRegisterSuffixInterceptor should work", t, func() {
		f := func(http.ResponseWriter, *http.Request, ErrorWriter) (InterceptorAction, string, error) {
			return InterceptorActionForward, "", nil
		}
		OptionRegisterSuffixInterceptor("/suffix", f)(c)
		So(c.suffixInterceptors["/suffix"], ShouldEqual, f)
	})

	Convey("Calling OptionRegisterExactInterceptor should work", t, func() {
		f := func(http.ResponseWriter, *http.Request, ErrorWriter) (InterceptorAction, string, error) {
			return InterceptorActionForward, "", nil
		}
		OptionRegisterExactInterceptor("/exact", f)(c)
		So(c.exactInterceptors["/exact"], ShouldEqual, f)
	})

	Convey("Calling OptionSetCustomRequestRewriter should work", t, func() {
		f := func(*http.Request, bool) error {
			return nil
		}
		OptionSetCustomRequestRewriter(f)(c)
		So(c.requestRewriter, ShouldEqual, f)
	})

	Convey("Calling OptionSetCustomResponseRewriter should work", t, func() {
		f := func(*http.Response) error {
			return nil
		}
		OptionSetCustomResponseRewriter(f)(c)
		So(c.responseRewriter, ShouldEqual, f)
	})

	Convey("Calling OptionMetricsManager should work", t, func() {
		mm := bahamut.MetricsManager(nil)
		OptionMetricsManager(mm)(c)
		So(c.metricsManager, ShouldEqual, mm)
	})

	Convey("Calling OptionUpstreamTLSConfig should work", t, func() {
		tlscfg := &tls.Config{}
		OptionUpstreamTLSConfig(tlscfg)(c)
		So(c.upstreamTLSConfig, ShouldEqual, tlscfg)
	})

	Convey("Calling OptionServerTLSConfig should work", t, func() {
		tlscfg := &tls.Config{}
		OptionServerTLSConfig(tlscfg)(c)
		So(c.serverTLSConfig, ShouldEqual, tlscfg)
	})

	Convey("Calling OptionAllowedCORSOrigin should work", t, func() {
		OptionAllowedCORSOrigin("dog")(c)
		So(c.corsOrigin, ShouldEqual, "dog")
	})
}
