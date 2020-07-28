package gateway

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/armon/go-proxyproto"
	"github.com/sirupsen/logrus"
	"github.com/valyala/tcplisten"
	"github.com/vulcand/oxy/buffer"
	"github.com/vulcand/oxy/cbreaker"
	"github.com/vulcand/oxy/connlimit"
	"github.com/vulcand/oxy/forward"
	"github.com/vulcand/oxy/ratelimit"
	"go.aporeto.io/bahamut"
	"go.uber.org/zap"
)

// An Upstreamer is the interface that can compute upstreams.
type Upstreamer interface {
	Upstream(req *http.Request) (upstream string, load float64)
}

// A LatencyBasedUpstreamer is the interface that can circle back
// response time as an input for Upstreamer decision.
type LatencyBasedUpstreamer interface {
	CollectLatency(address string, responseTime time.Duration)
	Upstreamer
}

// A Gateway can be used as an api gateway.
type Gateway interface {
	Start(ctx context.Context)
	Stop()
}

// An gateway is cool
type gateway struct {
	server            *http.Server
	upstreamer        Upstreamer
	upstreamerLatency LatencyBasedUpstreamer
	forwarder         *forward.Forwarder
	proxyHandler      http.Handler
	listener          net.Listener
	goodbyeServer     *http.Server
	gatewayConfig     *gwconfig
}

// New returns a new Gateway.
func New(listenAddr string, upstreamer Upstreamer, options ...Option) (Gateway, error) {

	cfg := newGatewayConfig()
	for _, o := range options {
		o(cfg)
	}

	var listener net.Listener

	rootListener, err := (&tcplisten.Config{
		ReusePort:   true,
		DeferAccept: true,
		FastOpen:    true,
	}).NewListener("tcp4", listenAddr)
	if err != nil {
		return nil, fmt.Errorf("unable build fast tcp listener: %s", err)
	}

	if cfg.proxyProtocolEnabled {

		sc, err := makeProxyProtocolSourceChecker(cfg.proxyProtocolSubnet)
		if err != nil {
			return nil, fmt.Errorf("unable build proxy protocol source checker: %s", err)
		}

		if cfg.serverTLSConfig != nil {
			listener = tls.NewListener(
				&proxyproto.Listener{
					Listener:    rootListener,
					SourceCheck: sc,
				},
				cfg.serverTLSConfig,
			)
		} else {
			listener = &proxyproto.Listener{
				Listener:    rootListener,
				SourceCheck: sc,
			}
		}

	} else {
		if cfg.serverTLSConfig != nil {
			listener = tls.NewListener(rootListener, cfg.serverTLSConfig)
		} else {
			listener = rootListener
		}
	}

	if cfg.tcpRateLimitingEnabled {
		listener = newLimitedListener(listener, cfg.tcpRateLimitingCPS, cfg.tcpRateLimitingBurst)
	}

	var serverLogger *log.Logger
	if !cfg.trace {
		serverLogger, err = zap.NewStdLogAt(zap.L(), zap.DebugLevel)
		if err != nil {
			return nil, fmt.Errorf("unable create zap std logger: %s", err)
		}
	}

	if !cfg.trace {
		logrus.SetLevel(logrus.PanicLevel)
	} else {
		logrus.SetLevel(logrus.TraceLevel)
	}

	s := &gateway{
		goodbyeServer: makeGoodbyeServer(listenAddr, cfg.serverTLSConfig),
		listener:      listener,
		upstreamer:    upstreamer,
		gatewayConfig: cfg,
	}

	if u, ok := s.upstreamer.(LatencyBasedUpstreamer); ok {
		s.upstreamerLatency = u
	}

	s.server = &http.Server{
		ReadTimeout:  cfg.httpReadTimeout,
		WriteTimeout: cfg.httpWriteTimeout,
		IdleTimeout:  cfg.httpIdleTimeout,
		ErrorLog:     serverLogger,
		Handler:      s,
		ConnState: func(conn net.Conn, state http.ConnState) {
			switch state {
			case http.StateNew:
				if mm := cfg.metricsManager; mm != nil {
					mm.RegisterTCPConnection()
				}
			case http.StateClosed, http.StateHijacked:
				if mm := cfg.metricsManager; mm != nil {
					mm.UnregisterTCPConnection()
				}
			}
		},
	}

	var topProxyHandler http.Handler

	if s.forwarder, err = forward.New(
		forward.BufferPool(newPool(1024*1024)),
		forward.WebsocketTLSClientConfig(cfg.upstreamTLSConfig),
		forward.ErrorHandler(&errorHandler{}),
		forward.Rewriter(
			&requestRewriter{
				blockOpenTracing: (!cfg.exposePrivateAPIs && cfg.blockOpenTracingHeaders),
				private:          cfg.exposePrivateAPIs,
				customRewriter:   cfg.requestRewriter,
			},
		),
		forward.ResponseModifier(
			func(resp *http.Response) error {

				injectGeneralHeader(resp.Header)
				injectCORSHeader(resp.Header, cfg.corsOrigin, resp.Request.Header.Get("origin"), resp.Request.Method)

				if s.gatewayConfig.responseRewriter != nil {
					if err := s.gatewayConfig.responseRewriter(resp); err != nil {
						return fmt.Errorf("unable to execute response rewriter: %s", err)
					}
				}
				return nil
			},
		),
		forward.RoundTripper(
			&http.Transport{
				DialContext: (&net.Dialer{
					Timeout:   30 * time.Second,
					KeepAlive: 30 * time.Second,
					DualStack: true,
				}).DialContext,
				ForceAttemptHTTP2:   cfg.upstreamUseHTTP2,
				TLSClientConfig:     cfg.upstreamTLSConfig,
				DisableCompression:  cfg.exposePrivateAPIs,
				MaxConnsPerHost:     cfg.upstreamMaxConnsPerHost,
				MaxIdleConns:        cfg.upstreamMaxIdleConns,
				MaxIdleConnsPerHost: cfg.upstreamMaxIdleConnsPerHost,
				TLSHandshakeTimeout: cfg.upstreamTLSHandshakeTimeout,
				IdleConnTimeout:     cfg.upstreamIdleConnTimeout,
			},
		),
	); err != nil {
		return nil, fmt.Errorf("unable to initialize forwarder: %s", err)
	}

	if topProxyHandler, err = buffer.New(
		s.forwarder,
		buffer.MaxRequestBodyBytes(1024*1024),
		buffer.MemRequestBodyBytes(1024*1024*1024),
		buffer.ErrorHandler(&errorHandler{}),
	); err != nil {
		return nil, fmt.Errorf("unable to initialize request buffer: %s", err)
	}

	if cfg.tcpMaxConnections > 0 {

		if topProxyHandler, err = connlimit.New(
			topProxyHandler,
			&sourceExtractor{},
			int64(cfg.tcpMaxConnections),
			connlimit.ErrorHandler(&errorHandler{}),
		); err != nil {
			return nil, fmt.Errorf("unable to initialize connection limiter: %s", err)
		}
	}

	if cfg.rateLimitingEnabled {

		rates := ratelimit.NewRateSet()
		if err := rates.Add(time.Second, cfg.rateLimitingRPS, cfg.rateLimitingBurst); err != nil {
			return nil, fmt.Errorf("unable to make rate set: %s", err)
		}

		if topProxyHandler, err = ratelimit.New(
			topProxyHandler,
			&sourceExtractor{},
			rates,
			ratelimit.ErrorHandler(&errorHandler{}),
		); err != nil {
			return nil, fmt.Errorf("unable to initialize ratelimiter: %s", err)
		}
	}

	if cfg.upstreamCircuitBreakerCond != "" {
		if topProxyHandler, err = cbreaker.New(
			topProxyHandler,
			cfg.upstreamCircuitBreakerCond,
			cbreaker.Fallback(&circuitBreakerHandler{}),
		); err != nil {
			return nil, fmt.Errorf("unable to initialize circuit breaker: %s", err)
		}
	}

	s.proxyHandler = topProxyHandler

	return s, nil
}

// Start starts the http server
func (s *gateway) Start(ctx context.Context) {

	go func() {

		if err := s.server.Serve(s.listener); err != nil {
			if err == http.ErrServerClosed {
				return
			}
			zap.L().Fatal("Unable to start internal API server", zap.Error(err))
		}
	}()
}

func (s *gateway) Stop() {

	stopCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)

	// Stopping main server
	go func() {
		defer cancel()
		if err := s.server.Shutdown(stopCtx); err != nil {
			zap.L().Error("Could not gracefully stop internal API server", zap.Error(err))
		} else {
			zap.L().Debug("Internet API server stopped")
		}
	}()

	// We start a temporary server to tell the world we are not serving requests anymore
	// We due this due to kubernetes continuing service traffic to the terminating pod.
	// As nobody responds anymore while nginx finishes treating the requests, this leads
	// to connection timeout, with mostly no chance of retrying.
	// This server makes sure we return immediately with a retryable error.
	go func() {
		zap.L().Info("Starting temporary redirect server...")
		for {
			if err := s.goodbyeServer.ListenAndServeTLS("", ""); err != nil {
				if strings.Contains(err.Error(), "address already in use") {
					continue
				}
				if err == http.ErrServerClosed {
					return
				}
				zap.L().Error("Unable to start temporary redirect server", zap.Error(err))
				return
			}
		}
	}()

	<-stopCtx.Done()

	stopCtx, cancel = context.WithTimeout(context.Background(), 1*time.Second)
	go func() {
		defer cancel()
		zap.L().Info("Stopping temporary redirect server...")
		if err := s.goodbyeServer.Shutdown(stopCtx); err != nil {
			zap.L().Error("Could not gracefully stop temp server", zap.Error(err))
		}
	}()

	<-stopCtx.Done()
}

func (s *gateway) checkInterceptor(
	registry map[string]InterceptorFunc,
	checker func(string, string) bool,
	w http.ResponseWriter,
	r *http.Request,
	path string,
) (InterceptorAction, string, error) {

	for key, interceptor := range registry {

		if !checker(path, key) {
			continue
		}

		return interceptor(w, r, writeError)
	}

	return 0, "", nil
}

func (s *gateway) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	if r.Method == http.MethodOptions {
		h := w.Header()
		injectCORSHeader(h, s.gatewayConfig.corsOrigin, r.Header.Get("Origin"), r.Method)
		w.WriteHeader(http.StatusOK) // nolint: errcheck
		return
	}

	if s.gatewayConfig.maintenance {
		h := w.Header()
		h.Set("Content-Type", "application/msgpack, application/json")
		injectCORSHeader(h, s.gatewayConfig.corsOrigin, r.Header.Get("Origin"), r.Method)
		writeError(w, r, errLocked)
		return
	}

	path := r.URL.Path

	var upstream string
	var load float64
	var interceptAction InterceptorAction
	var err error

	// First we look for the exact match
	if interceptAction, upstream, err = s.checkInterceptor(
		s.gatewayConfig.exactInterceptors,
		func(path string, key string) bool { return path == key },
		w, r, path,
	); interceptAction != 0 {
		goto HANDLE_INTERCEPTION
	}

	// If we reach here, we check for prefix match
	if interceptAction, upstream, err = s.checkInterceptor(
		s.gatewayConfig.prefixInterceptors,
		func(path string, key string) bool { return strings.HasPrefix(path, key) },
		w, r, path,
	); interceptAction != 0 {
		goto HANDLE_INTERCEPTION
	}

	// If we reach here, we check for suffix match
	if interceptAction, upstream, err = s.checkInterceptor(
		s.gatewayConfig.suffixInterceptors,
		func(path string, key string) bool { return strings.HasSuffix(path, key) },
		w, r, path,
	); interceptAction != 0 {
		goto HANDLE_INTERCEPTION
	}

HANDLE_INTERCEPTION:
	if err != nil {
		writeError(w, r, makeError(http.StatusInternalServerError, "Internal Server Error", fmt.Sprintf("unable to run interceptor: %s", err)))
		return
	}
	if interceptAction == InterceptorActionStop {
		injectCORSHeader(w.Header(), s.gatewayConfig.corsOrigin, r.Header.Get("Origin"), r.Method)
		return
	}

	// If we don't have an upstream returned by an interceptor,
	// we find it as usual.
	if upstream == "" {

		upstream, load = s.upstreamer.Upstream(r)
		if upstream == "" {
			writeError(w, r, errServiceUnavailable)
			return
		}
	}

	zap.L().Debug("request",
		zap.String("ip", r.RemoteAddr),
		zap.String("method", r.Method),
		zap.String("proto", r.Proto),
		zap.String("path", r.URL.Path),
		zap.String("ns", r.Header.Get("X-Namespace")),
		zap.String("routed", upstream),
		zap.Float64("load", load),
		zap.String("scheme", s.gatewayConfig.upstreamURLScheme),
	)

	r.URL.Host = upstream
	r.URL.Scheme = s.gatewayConfig.upstreamURLScheme

	switch interceptAction {
	case InterceptorActionForwardWS:
		if mm := s.gatewayConfig.metricsManager; mm != nil {
			mm.RegisterWSConnection()
		}

		s.forwarder.ServeHTTP(w, r)

		if mm := s.gatewayConfig.metricsManager; mm != nil {
			mm.UnregisterWSConnection()
		}

	case InterceptorActionForwardDirect:

		var finish bahamut.FinishMeasurementFunc

		if mm := s.gatewayConfig.metricsManager; mm != nil {
			finish = mm.MeasureRequest(r.Method, path)
		}

		s.forwarder.ServeHTTP(w, r)

		if finish != nil {
			rt := finish(0, nil)
			if s.upstreamerLatency != nil {
				s.upstreamerLatency.CollectLatency(upstream, rt)
			}
		}

	default:

		var finish bahamut.FinishMeasurementFunc

		if mm := s.gatewayConfig.metricsManager; mm != nil {
			finish = mm.MeasureRequest(r.Method, path)
		}

		s.proxyHandler.ServeHTTP(w, r)

		if finish != nil {
			rt := finish(0, nil)
			if s.upstreamerLatency != nil {
				s.upstreamerLatency.CollectLatency(upstream, rt)
			}
		}
	}
}
