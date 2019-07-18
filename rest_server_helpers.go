// Copyright 2019 Aporeto Inc.
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//     http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
	if encoding == elemental.EncodingTypeJSON {
		w.Header().Set("Content-Type", string(encoding)+"; charset=UTF-8")
	} else {
		w.Header().Set("Content-Type", string(encoding))
	}

	w.Header().Set("X-Frame-Options", "SAMEORIGIN")
	w.Header().Set("Cache-control", "private, no-transform")
	w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
	w.Header().Set("Access-Control-Allow-Origin", origin)
	w.Header().Set("Access-Control-Expose-Headers", "X-Requested-With, X-Count-Total, X-Namespace, X-Messages, X-Fields")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, HEAD, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Authorization, Accept, Content-Type, Cache-Control, If-Modified-Since, X-Requested-With, X-Count-Total, X-Namespace, X-External-Tracking-Type, X-External-Tracking-ID, X-TLS-Client-Certificate, Accept-Encoding, X-Fields, X-Read-Consistency, X-Write-Consistency")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
}

func makeCORSHandler(origin string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		o := r.Header.Get("Origin")
		if origin != "" {
			o = origin
		}
		_, writeEncoding, _ := elemental.EncodingFromHeaders(r.Header)
		setCommonHeader(w, o, writeEncoding)
		w.WriteHeader(http.StatusOK)
	}
}

func makeNotFoundHandler(origin string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		writeHTTPResponse(origin, w, makeErrorResponse(r.Context(), elemental.NewResponse(elemental.NewRequest()), ErrNotFound, nil))
	}
}

// writeHTTPResponse writes the response into the given http.ResponseWriter.
func writeHTTPResponse(CORSOrigin string, w http.ResponseWriter, r *elemental.Response) int {

	// If r is nil, we simply stop.
	// It mostly means the client closed the connection and
	// no response is needed.
	if r == nil {
		return 0
	}

	origin := r.Request.Headers.Get("Origin")
	if CORSOrigin != "" {
		origin = CORSOrigin
	}

	setCommonHeader(w, origin, r.Request.Accept)

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
