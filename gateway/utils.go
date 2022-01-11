package gateway

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/http"

	"github.com/go-zoo/bone"
	"go.aporeto.io/bahamut"
)

func injectGeneralHeader(h http.Header) http.Header {

	h.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
	h.Set("X-Frame-Options", "DENY")
	h.Set("X-Content-Type-Options", "nosniff")
	h.Set("X-Xss-Protection", "1; mode=block")
	h.Set("Cache-Control", "private, no-transform")

	return h
}

func injectCORSHeader(h http.Header, corsOrigin string, additionalCorsOrigin []string, allowCredentials bool, origin string, method string) http.Header {

	a := bahamut.NewDefaultCORSController(corsOrigin, additionalCorsOrigin)
	ac := a.PolicyForRequest(nil)
	ac.AllowCredentials = allowCredentials
	ac.Inject(h, origin, method == http.MethodOptions)
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
