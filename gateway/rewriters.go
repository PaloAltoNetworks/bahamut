package gateway

import (
	"encoding/pem"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"strings"

	"go.aporeto.io/tg/tglib"
	"go.uber.org/zap"
)

const internalWSMarkingHeader = "__internal_ws__"

type requestRewriter struct {
	customRewriter     RequestRewriter
	blockOpenTracing   bool
	private            bool
	trustForwardHeader bool
}

func (s *requestRewriter) Rewrite(r *httputil.ProxyRequest) {

	if s.customRewriter != nil {
		if err := s.customRewriter(r, s.private); err != nil {
			zap.L().Error("Unable rewrite request with custom rewriter", zap.Error(err))
			panic(fmt.Sprintf("unable to rewrite request with custom rewriter: %s", err)) // panic are recovered from oxy
		}
	}

	if s.blockOpenTracing {
		r.Out.Header.Del("X-B3-TraceID")
		r.Out.Header.Del("X-B3-SpanID")
		r.Out.Header.Del("X-B3-ParentSpanID")
		r.Out.Header.Del("X-B3-Sampled")
		r.Out.Header.Del("Uber-Trace-ID")
		r.Out.Header.Del("Jaeger-Baggage")
		r.Out.Header.Del("TraceParent")
		r.Out.Header.Del("TraceState")
	}

	// If we trust the forward headers, we backport the ones from
	// the inbound request to the outbound request.
	// Otherwise, per documentation, they have already been removed
	// from the outbound request.
	if s.trustForwardHeader {
		r.Out.Header["X-Forwarded-For"] = r.In.Header["X-Forwarded-For"]
		r.Out.Header["X-Forwarded-Proto"] = r.In.Header["X-Forwarded-Proto"]
		r.Out.Header["X-Forwarded-Host"] = r.In.Header["X-Forwarded-Host"]
	}

	// Now, if we reach here, and still have no X-Forwarded-For, we set them
	// using the inbound request client IP.
	if r.Out.Header.Get("X-Forwarded-For") == "" {
		if clientIP, _, err := net.SplitHostPort(r.In.RemoteAddr); err == nil {
			r.Out.Header.Set("X-Forwarded-For", clientIP)
			r.Out.Header.Set("X-Forwarded-Host", r.In.Host)
			r.Out.Header.Set("X-Forwarded-Proto", r.In.Proto)
		}
	}

	// Here we delete the internalWSMarkingHeader if it has
	// been set.
	if r.In.Header.Get(internalWSMarkingHeader) != "" {
		r.Out.Header.Del(internalWSMarkingHeader)
	}

	if r.In.TLS != nil && len(r.In.TLS.PeerCertificates) == 1 {

		block, err := tglib.CertToPEM(r.In.TLS.PeerCertificates[0])
		if err != nil {
			zap.L().Error("Unable to handle client TLS certificate", zap.Error(err))
			panic(fmt.Sprintf("unable to handle client TLS certificate: %s", err)) // panic are recovered from oxy
		}

		r.Out.Header.Add("X-TLS-Client-Certificate", strings.ReplaceAll(string(pem.EncodeToMemory(block)), "\n", " "))
	}
}

type circuitBreakerHandler struct{}

func (h *circuitBreakerHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	writeError(w, r, makeError(http.StatusServiceUnavailable, "Service Unavailable", "The service is busy handling requests. Please retry in a moment"))
}
