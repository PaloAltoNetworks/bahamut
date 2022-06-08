package gateway

import (
	"fmt"
	"net/http"

	"github.com/cespare/xxhash"
)

type defaultSourceExtractor struct {
	authCookieName string
}

// NewDefaultSourceExtractor returns a default SourceExtractor.
// A source extractor will discriminate the source of a request
// based on a hash of its authentication string.
// It will first use an eventual cookie with the given name,
// then use then use the Authorization header.
// If both are empty, the bucket key will be 'default'.
// If authCookieName is empty, only the value of the Authorization
// header will be taken into account.
func NewDefaultSourceExtractor(authCookieName string) SourceExtractor {
	return defaultSourceExtractor{
		authCookieName: authCookieName,
	}
}

func (f defaultSourceExtractor) ExtractSource(r *http.Request) (string, error) {

	var v string
	authHeader := r.Header.Get("Authorization")

	var authCookie *http.Cookie
	if f.authCookieName != "" {
		authCookie, _ = r.Cookie(f.authCookieName)
	}

	switch {
	case authCookie != nil && authCookie.Value != "":
		v = authCookie.Value
	case authHeader != "":
		v = authHeader
	default:
		return "default", nil
	}

	return fmt.Sprintf("%d", xxhash.Sum64([]byte(v))), nil
}

type defaultTCPSourceExtractor struct{}

func (f defaultTCPSourceExtractor) ExtractSource(r *http.Request) (string, error) {

	return r.RemoteAddr, nil
}
