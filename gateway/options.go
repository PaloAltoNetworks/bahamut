package gateway

import (
	"crypto/tls"
	"net/http"
	"time"

	"go.aporeto.io/bahamut"
	"go.aporeto.io/elemental"
)

// A RequestRewriter can be used to rewrite the request
// before it is sent to the upstream.
// The private parameter tells if the gateway is configured or not
// to serve the private APIs.
type RequestRewriter func(req *http.Request, private bool) error

// A ResponseRewriter can be used to rewrite the response
// before it is sent back to the client
type ResponseRewriter func(*http.Response) error

// An InterceptorFunc is a function that can be used to intercept and request
// based on its prefix and apply custom operation and returns an InterceptorAction
// to tell the gateway it should proceed from there.
// If it returns an error, the error is returned to the client as an internal server error.
//
// NOTE: It is not possible to rewrite the request. To do so, you can use a RequestRewriter.
type InterceptorFunc func(w http.ResponseWriter, req *http.Request, ew ErrorWriter) (action InterceptorAction, upstream string, err error)

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

	// InterceptorActionStop means the interceptor handled the request
	// and the gateway will not do anything more.
	InterceptorActionStop
)

type gwconfig struct {
	requestRewriter             RequestRewriter
	responseRewriter            ResponseRewriter
	blockOpenTracingHeaders     bool
	exactInterceptors           map[string]InterceptorFunc
	exposePrivateAPIs           bool
	httpDisableKeepAlive        bool
	httpIdleTimeout             time.Duration
	httpReadTimeout             time.Duration
	httpWriteTimeout            time.Duration
	maintenance                 bool
	metricsManager              bahamut.MetricsManager
	prefixInterceptors          map[string]InterceptorFunc
	proxyProtocolEnabled        bool
	proxyProtocolSubnet         string
	rateLimitingBurst           int64
	rateLimitingEnabled         bool
	rateLimitingRPS             int64
	tcpMaxConnections           int
	tcpRateLimitingBurst        int
	tcpRateLimitingCPS          float64
	tcpRateLimitingEnabled      bool
	trace                       bool
	upstreamUseHTTP2            bool
	upstreamCircuitBreakerCond  string
	upstreamIdleConnTimeout     time.Duration
	upstreamMaxConnsPerHost     int
	upstreamMaxIdleConns        int
	upstreamMaxIdleConnsPerHost int
	upstreamTLSHandshakeTimeout time.Duration
	upstreamTLSConfig           *tls.Config
	serverTLSConfig             *tls.Config
	corsOrigin                  string
}

func newGatewayConfig() *gwconfig {
	return &gwconfig{
		corsOrigin:                  "*",
		prefixInterceptors:          map[string]InterceptorFunc{},
		exactInterceptors:           map[string]InterceptorFunc{},
		tcpRateLimitingBurst:        100,
		tcpRateLimitingCPS:          200.0,
		upstreamIdleConnTimeout:     time.Hour,
		upstreamMaxConnsPerHost:     64,
		upstreamMaxIdleConns:        32000,
		upstreamMaxIdleConnsPerHost: 64,
		upstreamTLSHandshakeTimeout: 10 * time.Second,
		upstreamUseHTTP2:            false,
		httpIdleTimeout:             240 * time.Second,
		httpReadTimeout:             120 * time.Second,
		httpWriteTimeout:            240 * time.Second,
		rateLimitingRPS:             2000,
		rateLimitingBurst:           500,
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

// OptionTCPRateLimiting enables and configures the TCP rate limiter.
func OptionTCPRateLimiting(enabled bool, cps float64, burst int, maxConnections int) Option {
	return func(cfg *gwconfig) {
		cfg.tcpRateLimitingEnabled = enabled
		cfg.tcpRateLimitingCPS = cps
		cfg.tcpRateLimitingBurst = burst
		cfg.tcpMaxConnections = maxConnections
	}
}

// OptionRateLimiting enables and configures the HTTP rate limiter.
func OptionRateLimiting(enabled bool, cps int, burst int) Option {
	return func(cfg *gwconfig) {
		cfg.rateLimitingEnabled = enabled
		cfg.rateLimitingRPS = int64(cps)
		cfg.rateLimitingBurst = int64(burst)
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

// OptionMetricsManager registers a given PrefixInterceptorFunc for the given path prefix.
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
// If set to empty, or "*", the gateway will mirror
// whatever is set in the upcoming request Origin header.
// This is not secure to be used in production
// when a browser is calling the gateway.
//
// By default, it is set to "*"
func OptionAllowedCORSOrigin(origin string) Option {
	return func(cfg *gwconfig) {
		cfg.corsOrigin = origin
	}
}
