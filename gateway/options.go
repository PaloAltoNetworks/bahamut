package gateway

import (
	"crypto/tls"
	"net/http"
	"net/http/httputil"
	"time"

	"go.aporeto.io/bahamut"
	"go.aporeto.io/elemental"
	"golang.org/x/time/rate"
)

// A RequestRewriter can be used to rewrite the request
// before it is sent to the upstream.
// The private parameter tells if the gateway is configured or not
// to serve the private APIs.
type RequestRewriter func(req *httputil.ProxyRequest, private bool) error

// A ResponseRewriter can be used to rewrite the response
// before it is sent back to the client
type ResponseRewriter func(*http.Response) error

// An InterceptorFunc is a function that can be used to intercept and request
// based on its prefix and apply custom operation and returns an InterceptorAction
// to tell the gateway it should proceed from there.
// If it returns an error, the error is returned to the client as an internal server error.
//
// The given corsInjector function can be called if you wish your response to contain the CORS information
// the gateway would normally add. This is mandatory if you add your own headers in the interceptor.
// Otherwise, the gateway will add the CORS information for you.
//
// NOTE: It is not possible to rewrite the request. To do so, you can use a RequestRewriter.
type InterceptorFunc func(w http.ResponseWriter, req *http.Request, ew ErrorWriter, corsInjector func()) (action InterceptorAction, upstream string, err error)

// ErrorWriter is a function that can be used to return a standard formatted error to the client.
type ErrorWriter func(w http.ResponseWriter, r *http.Request, eerr elemental.Error)

// A InterceptorAction represents the decision
// on how to continue handling the request
type InterceptorAction int

const (
	// InterceptorActionForward means the Gateway will continue forwarding the request.
	// In that case the Interceptor must only modify the request, and MUST NOT use
	// the HTTP response writer.
	InterceptorActionForward InterceptorAction = iota + 1

	// InterceptorActionForwardWS means the Gateway will continue forwarding the request as a websocket.
	// In that case the Interceptor must only modify the request, and MUST NOT use
	// the HTTP response writer.
	InterceptorActionForwardWS

	// InterceptorActionForwardDirect means the Gateway will continue forwarding the request directly.
	// In that case the Interceptor must only modify the request, and MUST NOT use
	// the HTTP response writer.
	InterceptorActionForwardDirect

	// InterceptorActionStop means the interceptor handled the request
	// and the gateway will not do anything more.
	InterceptorActionStop
)

type gwconfig struct {
	sourceExtractor                    SourceExtractor
	metricsManager                     bahamut.MetricsManager
	sourceRateLimitingMetricManager    LimiterMetricManager
	tcpClientSourceExtractor           SourceExtractor
	sourceRateExtractor                RateExtractor
	tcpGlobalRateLimitingMetricManager LimiterMetricManager
	exactInterceptors                  map[string]InterceptorFunc
	requestRewriter                    RequestRewriter
	upstreamTLSConfig                  *tls.Config
	serverTLSConfig                    *tls.Config
	responseRewriter                   ResponseRewriter
	prefixInterceptors                 map[string]InterceptorFunc
	suffixInterceptors                 map[string]InterceptorFunc
	corsOrigin                         string
	proxyProtocolSubnet                string
	upstreamCircuitBreakerCond         string
	upstreamURLScheme                  string
	additionalCorsOrigin               []string
	tcpClientMaxConnections            int
	upstreamIdleConnTimeout            time.Duration
	sourceRateLimitingRPS              rate.Limit
	httpIdleTimeout                    time.Duration
	sourceRateLimitingBurst            int
	upstreamTLSHandshakeTimeout        time.Duration
	tcpGlobalRateLimitingBurst         int
	tcpGlobalRateLimitingCPS           rate.Limit
	httpReadTimeout                    time.Duration
	upstreamMaxIdleConnsPerHost        int
	upstreamMaxIdleConns               int
	upstreamMaxConnsPerHost            int
	httpWriteTimeout                   time.Duration
	tcpClientMaxConnectionsEnabled     bool
	upstreamUseHTTP2                   bool
	trace                              bool
	maintenance                        bool
	tcpGlobalRateLimitingEnabled       bool
	proxyProtocolEnabled               bool
	sourceRateLimitingEnabled          bool
	upstreamEnableCompression          bool
	httpDisableKeepAlive               bool
	exposePrivateAPIs                  bool
	corsAllowCredentials               bool
	blockOpenTracingHeaders            bool
	trustForwardHeader                 bool
}

func newGatewayConfig() *gwconfig {
	return &gwconfig{
		corsOrigin:                  bahamut.CORSOriginMirror,
		corsAllowCredentials:        true,
		prefixInterceptors:          map[string]InterceptorFunc{},
		suffixInterceptors:          map[string]InterceptorFunc{},
		exactInterceptors:           map[string]InterceptorFunc{},
		tcpGlobalRateLimitingBurst:  200,
		tcpGlobalRateLimitingCPS:    100.0,
		upstreamIdleConnTimeout:     time.Hour,
		upstreamMaxConnsPerHost:     64,
		upstreamMaxIdleConns:        32000,
		upstreamMaxIdleConnsPerHost: 64,
		upstreamTLSHandshakeTimeout: 10 * time.Second,
		upstreamURLScheme:           "https",
		upstreamUseHTTP2:            false,
		httpIdleTimeout:             240 * time.Second,
		httpReadTimeout:             120 * time.Second,
		httpWriteTimeout:            240 * time.Second,
		sourceExtractor:             &defaultSourceExtractor{},
		tcpClientSourceExtractor:    &defaultTCPSourceExtractor{},
	}
}

// A Option represents possible options for the Gateway.
type Option func(*gwconfig)

// OptionEnableProxyProtocol enables and configure the support
// for ProxyProtocol.
func OptionEnableProxyProtocol(enabled bool, subnet string) Option {
	return func(cfg *gwconfig) {
		cfg.proxyProtocolEnabled = enabled
		cfg.proxyProtocolSubnet = subnet
	}
}

// OptionTCPGlobalRateLimiting enables and configures the TCP rate limiter to
// the rate of the total number of TCP connection the gateway handle.
func OptionTCPGlobalRateLimiting(cps rate.Limit, burst int) Option {
	return func(cfg *gwconfig) {
		cfg.tcpGlobalRateLimitingEnabled = true
		cfg.tcpGlobalRateLimitingCPS = cps
		cfg.tcpGlobalRateLimitingBurst = burst
	}
}

// OptionTCPClientMaxConnections sets the maximum number of TCP connections
// a client can do at the same time. 0 means no limit.
// If the sourceExtractor is nil, the default one will be used, which uses
// the request's RemoteAddr as token.
func OptionTCPClientMaxConnections(maxConnections int) Option {
	return func(cfg *gwconfig) {
		cfg.tcpClientMaxConnectionsEnabled = maxConnections > 0
		cfg.tcpClientMaxConnections = maxConnections
	}
}

// OptionTCPClientMaxConnectionsSourceExtractor sets the source extractor
// to use to uniquely identify a client TCP connection.
// The default one uses the http.Request RemoteAddr property.
// Passing nil will reset to the default source extractor.
func OptionTCPClientMaxConnectionsSourceExtractor(sourceExtractor SourceExtractor) Option {
	return func(cfg *gwconfig) {
		if sourceExtractor == nil {
			sourceExtractor = &defaultTCPSourceExtractor{}
		}
		cfg.tcpClientSourceExtractor = sourceExtractor
	}
}

// OptionSourceRateLimiting sets the rate limit for a single source.
// If OptionSourceRateLimiting option is used, this option has no effect.
func OptionSourceRateLimiting(rps rate.Limit, burst int) Option {
	return func(cfg *gwconfig) {
		cfg.sourceRateLimitingEnabled = true
		cfg.sourceRateLimitingRPS = rps
		cfg.sourceRateLimitingBurst = burst
	}
}

// OptionSourceRateLimitingDynamic sets the RateExtractor to use to dynamically
// set the rates for a uniquely identified client.
// If this option is used, OptionSourceRateLimiting has no effect.
func OptionSourceRateLimitingDynamic(rateExtractor RateExtractor) Option {
	return func(cfg *gwconfig) {
		cfg.sourceRateLimitingEnabled = true
		cfg.sourceRateExtractor = rateExtractor
	}
}

// OptionSourceRateLimitingSourceExtractor configures a custom SourceExtractor
// to decide how to uniquely identify a client.
// The default one uses a hash of the authorization header.
// Passing nil will reset to the default source extractor.
func OptionSourceRateLimitingSourceExtractor(sourceExtractor SourceExtractor) Option {
	return func(cfg *gwconfig) {
		if sourceExtractor == nil {
			sourceExtractor = &defaultSourceExtractor{}
		}
		cfg.sourceExtractor = sourceExtractor
	}
}

// OptionEnableTrace enables deep oxy logging.
func OptionEnableTrace(enabled bool) Option {
	return func(cfg *gwconfig) {
		cfg.trace = enabled
	}
}

// OptionEnableMaintenance enables the maintenance mode.
func OptionEnableMaintenance(enabled bool) Option {
	return func(cfg *gwconfig) {
		cfg.maintenance = enabled
	}
}

// OptionHTTPTimeouts configures the HTTP timeouts.
func OptionHTTPTimeouts(read, write, idle time.Duration, disableKeepAlive bool) Option {
	return func(cfg *gwconfig) {
		cfg.httpReadTimeout = read
		cfg.httpWriteTimeout = write
		cfg.httpIdleTimeout = idle
		cfg.httpDisableKeepAlive = disableKeepAlive
	}
}

// OptionExposePrivateAPIs configures if the gateway should expose the private apis.
func OptionExposePrivateAPIs(enabled bool) Option {
	return func(cfg *gwconfig) {
		cfg.exposePrivateAPIs = enabled
	}
}

// OptionBlockOpenTracingHeaders configures if the gateway should strip
// any open tracing related header coming from the clients.
func OptionBlockOpenTracingHeaders(block bool) Option {
	return func(cfg *gwconfig) {
		cfg.blockOpenTracingHeaders = block
	}
}

// OptionUpstreamConfig configures the connections
// to the upstream backends.
func OptionUpstreamConfig(
	upstreamMaxConnsPerHost int,
	upstreamMaxIdleConns int,
	upstreamMaxIdleConnsPerHost int,
	upstreamTLSHandshakeTimeout time.Duration,
	upstreamIdleConnTimeout time.Duration,
	upstreamCircuitBreakerCond string,
	useHTTP2 bool,
) Option {
	return func(cfg *gwconfig) {
		cfg.upstreamMaxConnsPerHost = upstreamMaxConnsPerHost
		cfg.upstreamMaxIdleConns = upstreamMaxIdleConns
		cfg.upstreamMaxIdleConnsPerHost = upstreamMaxIdleConnsPerHost
		cfg.upstreamTLSHandshakeTimeout = upstreamTLSHandshakeTimeout
		cfg.upstreamIdleConnTimeout = upstreamIdleConnTimeout
		cfg.upstreamCircuitBreakerCond = upstreamCircuitBreakerCond
		cfg.upstreamUseHTTP2 = useHTTP2
	}
}

// OptionUpstreamURLScheme sets the URL scheme to use
// to connect to the upstreams. default is https.
func OptionUpstreamURLScheme(scheme string) Option {
	return func(cfg *gwconfig) {
		cfg.upstreamURLScheme = scheme
	}
}

// OptionUpstreamEnableCompression enables using compression between
// the gateway and the upstreams. This can lead to performance issues.
func OptionUpstreamEnableCompression(enable bool) Option {
	return func(cfg *gwconfig) {
		cfg.upstreamEnableCompression = enable
	}
}

// OptionMetricsManager registers set the MetricsManager to use.
// This will enable response time load balancing of endpoints.
func OptionMetricsManager(metricsManager bahamut.MetricsManager) Option {
	return func(cfg *gwconfig) {
		cfg.metricsManager = metricsManager
	}
}

// OptionRegisterPrefixInterceptor registers a given InterceptorFunc for the given path prefix.
func OptionRegisterPrefixInterceptor(prefix string, f InterceptorFunc) Option {
	return func(cfg *gwconfig) {
		cfg.prefixInterceptors[prefix] = f
	}
}

// OptionRegisterSuffixInterceptor registers a given InterceptorFunc for the given path suffix.
func OptionRegisterSuffixInterceptor(prefix string, f InterceptorFunc) Option {
	return func(cfg *gwconfig) {
		cfg.suffixInterceptors[prefix] = f
	}
}

// OptionRegisterExactInterceptor registers a given InterceptorFunc for the given path.
func OptionRegisterExactInterceptor(path string, f InterceptorFunc) Option {
	return func(cfg *gwconfig) {
		cfg.exactInterceptors[path] = f
	}
}

// OptionSetCustomRequestRewriter sets a custom RequestRewriter.
func OptionSetCustomRequestRewriter(r RequestRewriter) Option {
	return func(cfg *gwconfig) {
		cfg.requestRewriter = r
	}
}

// OptionSetCustomResponseRewriter sets a custom ResponseRewriter.
func OptionSetCustomResponseRewriter(r ResponseRewriter) Option {
	return func(cfg *gwconfig) {
		cfg.responseRewriter = r
	}
}

// OptionServerTLSConfig sets the tls.Config to use for the
// front end server.
func OptionServerTLSConfig(tlsConfig *tls.Config) Option {
	return func(cfg *gwconfig) {
		cfg.serverTLSConfig = tlsConfig
	}
}

// OptionUpstreamTLSConfig sets the tls.Config to use for the
// upstream servers.
func OptionUpstreamTLSConfig(tlsConfig *tls.Config) Option {
	return func(cfg *gwconfig) {
		cfg.upstreamTLSConfig = tlsConfig
	}
}

// OptionAllowedCORSOrigin sets allowed CORS origin.
// If set to CORSOriginMirror the gateway will mirror
// whatever is set in the upcoming request Origin header.
// This is not secure to be used in production
// when a browser is calling the gateway.
//
// By default, it is set to CORSOriginMirror.
func OptionAllowedCORSOrigin(origin string) Option {
	return func(cfg *gwconfig) {
		cfg.corsOrigin = origin
	}
}

// OptionAdditionnalAllowedCORSOrigin sets allowed CORS origin.
// If set, the gateway will mirror whatever is in the upcoming
// request Origin header as long as there is a match.
func OptionAdditionnalAllowedCORSOrigin(origins []string) Option {
	return func(cfg *gwconfig) {
		cfg.additionalCorsOrigin = origins
	}
}

// OptionCORSAllowCredentials sets if the header Access-Control-Allow-Credentials
// should be set to true.
//
// By default, it is set to true.
func OptionCORSAllowCredentials(allow bool) Option {
	return func(cfg *gwconfig) {
		cfg.corsAllowCredentials = allow
	}
}

// OptionTrustForwardHeader configures if the gateway should strip
// the X-Forwarded-For header or not.
func OptionTrustForwardHeader(trust bool) Option {
	return func(cfg *gwconfig) {
		cfg.trustForwardHeader = trust
	}
}

// OptionTCPGlobalRateLimitingManager sets the LimiterMetricManager to
// use to get metrics on the TCP global rate limiter.
func OptionTCPGlobalRateLimitingManager(m LimiterMetricManager) Option {
	return func(cfg *gwconfig) {
		cfg.tcpGlobalRateLimitingMetricManager = m
	}
}

// OptionSourceRateLimitingManager sets the LimiterMetricManager to
// use to get metrics on the source rate limiter.
func OptionSourceRateLimitingManager(m LimiterMetricManager) Option {
	return func(cfg *gwconfig) {
		cfg.sourceRateLimitingMetricManager = m
	}
}
