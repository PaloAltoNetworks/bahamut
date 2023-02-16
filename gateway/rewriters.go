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
		r.In.Header.Del("X-B3-TraceID")
		r.In.Header.Del("X-B3-SpanID")
		r.In.Header.Del("X-B3-ParentSpanID")
		r.In.Header.Del("X-B3-Sampled")
		r.In.Header.Del("Uber-Trace-ID")
		r.In.Header.Del("Jaeger-Baggage")
		r.In.Header.Del("TraceParent")
		r.In.Header.Del("TraceState")
	}

	// Will be rewritten by the forwarder,
	// based on proxy protocol if enabled.
	// unless trustForwardHeader is set.
	if !s.trustForwardHeader {
		r.In.Header.Del("X-Forwarded-For")
		r.In.Header.Del("X-Real-IP")
	}

	// If the request has been marked as a ws proxy, we set
	// the X-Forwarded header ourselves, since oxy does not
	// do it (for some reasons).
	if r.In.Header.Get(internalWSMarkingHeader) != "" {
		r.In.Header.Del(internalWSMarkingHeader)
		if clientIP, _, err := net.SplitHostPort(r.In.RemoteAddr); err == nil {
			r.In.Header.Set("X-Forwarded-For", clientIP)
		}
	}

	if r.In.TLS != nil && len(r.In.TLS.PeerCertificates) == 1 {

		block, err := tglib.CertToPEM(r.In.TLS.PeerCertificates[0])
		if err != nil {
			zap.L().Error("Unable to handle client TLS certificate", zap.Error(err))
			panic(fmt.Sprintf("unable to handle client TLS certificate: %s", err)) // panic are recovered from oxy
		}

		r.In.Header.Add("X-TLS-Client-Certificate", strings.ReplaceAll(string(pem.EncodeToMemory(block)), "\n", " "))
	}
}

type circuitBreakerHandler struct{}

func (h *circuitBreakerHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	writeError(w, r, makeError(http.StatusServiceUnavailable, "Service Unavailable", "The service is busy handling requests. Please retry in a moment"))
}
