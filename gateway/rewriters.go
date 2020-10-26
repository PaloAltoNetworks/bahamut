package gateway

import (
	"encoding/pem"
	"fmt"
	"net/http"
	"strings"

	"go.aporeto.io/tg/tglib"
	"go.uber.org/zap"
)

type requestRewriter struct {
	customRewriter     RequestRewriter
	blockOpenTracing   bool
	private            bool
	trustForwardHeader bool
}

func (s *requestRewriter) Rewrite(r *http.Request) {

	if s.customRewriter != nil {
		if err := s.customRewriter(r, s.private); err != nil {
			zap.L().Error("Unable rewrite request with custom rewriter", zap.Error(err))
			panic(fmt.Sprintf("unable to rewrite request with custom rewriter: %s", err)) // panic are recovered from oxy
		}
	}

	if s.blockOpenTracing {
		r.Header.Del("X-B3-TraceID")
		r.Header.Del("X-B3-SpanID")
		r.Header.Del("X-B3-ParentSpanID")
		r.Header.Del("X-B3-Sampled")
		r.Header.Del("Uber-Trace-ID")
		r.Header.Del("Jaeger-Baggage")
		r.Header.Del("TraceParent")
		r.Header.Del("TraceState")
	}

	// Will be rewritten by the forwarder,
	// based on proxy protocol if enabled.
	// unless trustForwardHeader is set.
	if !s.trustForwardHeader {
		r.Header.Del("X-Forwarded-For")
		r.Header.Del("X-Real-IP")
	}

	if r.TLS != nil && len(r.TLS.PeerCertificates) == 1 {

		block, err := tglib.CertToPEM(r.TLS.PeerCertificates[0])
		if err != nil {
			zap.L().Error("Unable to handle client TLS certificate", zap.Error(err))
			panic(fmt.Sprintf("unable to handle client TLS certificate: %s", err)) // panic are recovered from oxy
		}

		r.Header.Add("X-TLS-Client-Certificate", strings.ReplaceAll(string(pem.EncodeToMemory(block)), "\n", " "))
	}
}

type circuitBreakerHandler struct{}

func (h *circuitBreakerHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	writeError(w, r, makeError(http.StatusServiceUnavailable, "Service Unavailable", "The service is busy handling requests. Please retry in a moment"))
}
