package gateway

import (
	"fmt"
	"net/http"

	"github.com/cespare/xxhash"
)

type defaultSourceExtractor struct{}

func (f defaultSourceExtractor) ExtractSource(r *http.Request) (string, error) {

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "default", nil
	}

	return fmt.Sprintf("%d", xxhash.Sum64([]byte(authHeader))), nil
}

type defaultTCPSourceExtractor struct{}

func (f defaultTCPSourceExtractor) ExtractSource(r *http.Request) (string, error) {

	return r.RemoteAddr, nil
}
