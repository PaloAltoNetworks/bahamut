package bahamut

import (
	"net/http"
	"strconv"
	"strings"

	"go.aporeto.io/elemental"
	"go.uber.org/zap"
)

// Various common errors
var (
	ErrNotFound  = elemental.NewError("Not Found", "Unable to find the requested resource", "bahamut", http.StatusNotFound)
	ErrRateLimit = elemental.NewError("Rate Limit", "You have exceeded your rate limit", "bahamut", http.StatusTooManyRequests)
)

func setCommonHeader(w http.ResponseWriter, origin string, encoding elemental.EncodingType) {

	if origin == "" {
		origin = "*"
	}

	w.Header().Set("Accept", "application/msgpack,application/json")
	w.Header().Set("Content-Type", string(encoding))
	w.Header().Set("X-Frame-Options", "SAMEORIGIN")
	w.Header().Set("Cache-control", "private, no-transform")
	w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
	w.Header().Set("Access-Control-Allow-Origin", origin)
	w.Header().Set("Access-Control-Expose-Headers", "X-Requested-With, X-Count-Total, X-Namespace, X-Messages, X-Fields")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, HEAD, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Authorization, Accept, Content-Type, Cache-Control, If-Modified-Since, X-Requested-With, X-Count-Total, X-Namespace, X-External-Tracking-Type, X-External-Tracking-ID, X-TLS-Client-Certificate, Accept-Encoding, X-Fields, X-Read-Consistency, X-Write-Consistency")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
}

func corsHandler(w http.ResponseWriter, r *http.Request) {
	_, writeEncoding, _ := elemental.EncodingFromHeaders(r.Header)
	setCommonHeader(w, r.Header.Get("Origin"), writeEncoding)
	w.WriteHeader(http.StatusOK)
}

func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	_, writeEncoding, _ := elemental.EncodingFromHeaders(r.Header)
	setCommonHeader(w, r.Header.Get("Origin"), writeEncoding)
	writeHTTPResponse(w, makeErrorResponse(r.Context(), elemental.NewResponse(elemental.NewRequest()), ErrNotFound))
}

// writeHTTPResponse writes the response into the given http.ResponseWriter.
func writeHTTPResponse(w http.ResponseWriter, r *elemental.Response) int {

	// If r is nil, we simply stop.
	// It mostly means the client closed the connection and
	// no response is needed.
	if r == nil {
		return 0
	}

	if r.Redirect != "" {
		w.Header().Set("Location", r.Redirect)
		w.WriteHeader(http.StatusFound)
		return http.StatusFound
	}

	w.Header().Set("X-Count-Total", strconv.Itoa(r.Total))

	if len(r.Messages) > 0 {
		w.Header().Set("X-Messages", strings.Join(r.Messages, ";"))
	}

	w.WriteHeader(r.StatusCode)

	if r.Data != nil {

		if _, err := w.Write(r.Data); err != nil {
			zap.L().Debug("Unable to send http response to client", zap.Error(err))
		}
	}

	return r.StatusCode
}
