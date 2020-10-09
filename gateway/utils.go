package gateway

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/http"

	"github.com/go-zoo/bone"
)

func injectGeneralHeader(h http.Header) http.Header {

	h.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
	h.Set("X-Frame-Options", "DENY")
	h.Set("X-Content-Type-Options", "nosniff")
	h.Set("X-Xss-Protection", "1; mode=block")
	h.Set("Cache-Control", "private, no-transform")

	return h
}

func injectCORSHeader(h http.Header, corsOrigin string, additionalCorsOrigin map[string]struct{}, origin string, method string) http.Header {

	if corsOrigin == "*" && origin != "" {
		corsOrigin = origin
	} else if _, ok := additionalCorsOrigin[origin]; ok {
		corsOrigin = origin
	}

	if method == http.MethodOptions {
		h.Set("Access-Control-Allow-Headers", "Authorization, Accept, Content-Type, Cache-Control, Cookie, If-Modified-Since, X-Requested-With, X-Count-Total, X-Namespace, X-External-Tracking-Type, X-External-Tracking-ID, X-TLS-Client-Certificate, Accept-Encoding, X-Fields, X-Read-Consistency, X-Write-Consistency, Idempotency-Key")
		h.Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, HEAD, OPTIONS")
		h.Set("Access-Control-Max-Age", "1500")
	}

	h.Set("Access-Control-Allow-Origin", corsOrigin)
	h.Set("Access-Control-Expose-Headers", "X-Requested-With, X-Count-Total, X-Namespace, X-Messages, X-Fields, X-Next")
	h.Set("Access-Control-Allow-Credentials", "true")

	return h
}

func makeProxyProtocolSourceChecker(allowed string) (func(net.Addr) (bool, error), error) {

	_, allowedSubnet, err := net.ParseCIDR(allowed)
	if err != nil {
		return nil, fmt.Errorf("unable to parse CIDR: %s", err)
	}

	return func(addr net.Addr) (bool, error) {

		ipstr, _, err := net.SplitHostPort(addr.String())
		if err != nil {
			return false, fmt.Errorf("unable to parse net.Addr: %s", err)
		}

		return allowedSubnet.Contains(net.ParseIP(ipstr)), nil
	}, nil
}

func makeGoodbyeServer(listen string, serverTLSConfig *tls.Config) *http.Server {

	mux := bone.New()
	mux.NotFound(
		http.HandlerFunc(
			func(w http.ResponseWriter, req *http.Request) {
				w.WriteHeader(503)
				_, _ = w.Write([]byte(`[{"code":503,"title":"Service Not Available","description":"Shutting down. Please retry your request","subject":"gateway"}]`))
			},
		),
	)

	return &http.Server{
		TLSConfig: serverTLSConfig,
		Addr:      listen,
		Handler:   mux,
	}
}
