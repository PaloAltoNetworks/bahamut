package gateway

import (
	"crypto/tls"
	"net/http"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
	"go.aporeto.io/bahamut"
	"golang.org/x/time/rate"
)

func Test_Options(t *testing.T) {

	Convey("Calling OptionEnableProxyProtocol should work", t, func() {
		c := newGatewayConfig()
		OptionEnableProxyProtocol(true, "10.0.0.0/0")(c)
		So(c.proxyProtocolEnabled, ShouldEqual, true)
		So(c.proxyProtocolSubnet, ShouldEqual, "10.0.0.0/0")
		So(c.corsAllowCredentials, ShouldEqual, true)
		So(c.corsOrigin, ShouldEqual, bahamut.CORSOriginMirror)
	})

	Convey("Calling OptionTCPGobalRateLimiting should work", t, func() {
		c := newGatewayConfig()
		OptionTCPGlobalRateLimiting(1.0, 2)(c)
		So(c.tcpGlobalRateLimitingEnabled, ShouldEqual, true)
		So(c.tcpGlobalRateLimitingCPS, ShouldEqual, 1.0)
		So(c.tcpGlobalRateLimitingBurst, ShouldEqual, 2)
	})

	Convey("Calling OptionTCPClientMaxConnections should work", t, func() {
		c := newGatewayConfig()
		OptionTCPClientMaxConnections(3)(c)
		So(c.tcpClientMaxConnectionsEnabled, ShouldEqual, true)
		So(c.tcpClientMaxConnections, ShouldEqual, 3)

		OptionTCPClientMaxConnections(0)(c)
		So(c.tcpClientMaxConnectionsEnabled, ShouldEqual, false)
		So(c.tcpClientMaxConnections, ShouldEqual, 0)
	})

	Convey("Calling OptionTCPClientMaxConnectionsSourceExtractor should work", t, func() {
		c := newGatewayConfig()
		OptionTCPClientMaxConnectionsSourceExtractor(nil)(c)
		So(c.tcpClientSourceExtractor, ShouldHaveSameTypeAs, &defaultTCPSourceExtractor{})
		se := &defaultTCPSourceExtractor{}
		OptionTCPClientMaxConnectionsSourceExtractor(se)(c)
		So(c.tcpClientSourceExtractor, ShouldEqual, se)
	})

	Convey("Calling OptionSourceRateLimiting should work", t, func() {
		c := newGatewayConfig()
		OptionSourceRateLimiting(rate.Limit(10), 20)(c)
		So(c.sourceRateLimitingEnabled, ShouldEqual, true)
		So(c.sourceRateLimitingRPS, ShouldEqual, rate.Limit(10))
		So(c.sourceRateLimitingBurst, ShouldEqual, 20)
	})

	Convey("Calling OptionSourceRateLimitingDynamic should work", t, func() {
		c := newGatewayConfig()
		re := RateExtractor(nil)
		OptionSourceRateLimitingDynamic(re)(c)
		So(c.sourceRateExtractor, ShouldEqual, re)
	})

	Convey("Calling OptionSourceRateLimitingSourceExtractor should work", t, func() {
		c := newGatewayConfig()
		OptionSourceRateLimitingSourceExtractor(nil)(c)
		So(c.sourceExtractor, ShouldHaveSameTypeAs, &defaultSourceExtractor{})
		se := &defaultTCPSourceExtractor{}
		OptionSourceRateLimitingSourceExtractor(se)(c)
		So(c.sourceExtractor, ShouldEqual, se)
	})

	Convey("Calling OptionEnableTrace should work", t, func() {
		c := newGatewayConfig()
		OptionEnableTrace(true)(c)
		So(c.trace, ShouldEqual, true)
	})

	Convey("Calling OptionEnableMaintenance should work", t, func() {
		c := newGatewayConfig()
		OptionEnableMaintenance(true)(c)
		So(c.maintenance, ShouldEqual, true)
	})

	Convey("Calling OptionHTTPTimeouts should work", t, func() {
		c := newGatewayConfig()
		OptionHTTPTimeouts(time.Second, time.Minute, time.Hour, true)(c)
		So(c.httpReadTimeout, ShouldEqual, time.Second)
		So(c.httpWriteTimeout, ShouldEqual, time.Minute)
		So(c.httpIdleTimeout, ShouldEqual, time.Hour)
		So(c.httpDisableKeepAlive, ShouldEqual, true)
	})

	Convey("Calling OptionExposePrivateAPIs should work", t, func() {
		c := newGatewayConfig()
		OptionExposePrivateAPIs(true)(c)
		So(c.exposePrivateAPIs, ShouldEqual, true)
	})

	Convey("Calling OptionBlockOpenTracingHeaders should work", t, func() {
		c := newGatewayConfig()
		OptionBlockOpenTracingHeaders(true)(c)
		So(c.blockOpenTracingHeaders, ShouldEqual, true)
	})

	Convey("Calling OptionUpstreamConfig should work", t, func() {
		c := newGatewayConfig()
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
		c := newGatewayConfig()
		f := func(http.ResponseWriter, *http.Request, ErrorWriter, func()) (InterceptorAction, string, error) {
			return InterceptorActionForward, "", nil
		}
		OptionRegisterPrefixInterceptor("/prefix", f)(c)
		So(c.prefixInterceptors["/prefix"], ShouldEqual, f)
	})

	Convey("Calling OptionRegisterSuffixInterceptor should work", t, func() {
		c := newGatewayConfig()
		f := func(http.ResponseWriter, *http.Request, ErrorWriter, func()) (InterceptorAction, string, error) {
			return InterceptorActionForward, "", nil
		}
		OptionRegisterSuffixInterceptor("/suffix", f)(c)
		So(c.suffixInterceptors["/suffix"], ShouldEqual, f)
	})

	Convey("Calling OptionRegisterExactInterceptor should work", t, func() {
		c := newGatewayConfig()
		f := func(http.ResponseWriter, *http.Request, ErrorWriter, func()) (InterceptorAction, string, error) {
			return InterceptorActionForward, "", nil
		}
		OptionRegisterExactInterceptor("/exact", f)(c)
		So(c.exactInterceptors["/exact"], ShouldEqual, f)
	})

	Convey("Calling OptionSetCustomRequestRewriter should work", t, func() {
		c := newGatewayConfig()
		f := func(*http.Request, bool) error {
			return nil
		}
		OptionSetCustomRequestRewriter(f)(c)
		So(c.requestRewriter, ShouldEqual, f)
	})

	Convey("Calling OptionSetCustomResponseRewriter should work", t, func() {
		c := newGatewayConfig()
		f := func(*http.Response) error {
			return nil
		}
		OptionSetCustomResponseRewriter(f)(c)
		So(c.responseRewriter, ShouldEqual, f)
	})

	Convey("Calling OptionMetricsManager should work", t, func() {
		c := newGatewayConfig()
		mm := bahamut.MetricsManager(nil)
		OptionMetricsManager(mm)(c)
		So(c.metricsManager, ShouldEqual, mm)
	})

	Convey("Calling OptionUpstreamTLSConfig should work", t, func() {
		c := newGatewayConfig()
		tlscfg := &tls.Config{}
		OptionUpstreamTLSConfig(tlscfg)(c)
		So(c.upstreamTLSConfig, ShouldEqual, tlscfg)
	})

	Convey("Calling OptionServerTLSConfig should work", t, func() {
		c := newGatewayConfig()
		tlscfg := &tls.Config{}
		OptionServerTLSConfig(tlscfg)(c)
		So(c.serverTLSConfig, ShouldEqual, tlscfg)
	})

	Convey("Calling OptionAllowedCORSOrigin should work", t, func() {
		c := newGatewayConfig()
		OptionAllowedCORSOrigin("dog")(c)
		So(c.corsOrigin, ShouldEqual, "dog")
	})

	Convey("Calling OptionAdditionnalAllowedCORSOrigin should work", t, func() {
		c := newGatewayConfig()
		OptionAdditionnalAllowedCORSOrigin([]string{"dog"})(c)
		So(c.additionalCorsOrigin, ShouldResemble, []string{"dog"})
	})

	Convey("Calling OptionUpstreamURLScheme should work", t, func() {
		c := newGatewayConfig()
		OptionUpstreamURLScheme("http")(c)
		So(c.upstreamURLScheme, ShouldEqual, "http")
	})

	Convey("Calling OptionTrustForwardHeader should work", t, func() {
		c := newGatewayConfig()
		OptionTrustForwardHeader(true)(c)
		So(c.trustForwardHeader, ShouldBeTrue)
	})

	Convey("Calling OptionUpstreamEnableCompression should work", t, func() {
		c := newGatewayConfig()
		OptionUpstreamEnableCompression(true)(c)
		So(c.upstreamEnableCompression, ShouldBeTrue)
	})

	Convey("Calling OptionCORSAllowCredentials should work", t, func() {
		c := newGatewayConfig()
		OptionCORSAllowCredentials(false)(c)
		So(c.corsAllowCredentials, ShouldBeFalse)
	})

	Convey("Calling OptionTCPGlobalRateLimitingManager should work", t, func() {
		c := newGatewayConfig()
		m := &fakeListenerLimiterMetricManager{}
		OptionTCPGlobalRateLimitingManager(m)(c)
		So(c.tcpGlobalRateLimitingMetricManager, ShouldEqual, m)
	})

	Convey("Calling OptionSourceRateLimitingManager should work", t, func() {
		c := newGatewayConfig()
		m := &fakeListenerLimiterMetricManager{}
		OptionSourceRateLimitingManager(m)(c)
		So(c.sourceRateLimitingMetricManager, ShouldEqual, m)
	})
}
