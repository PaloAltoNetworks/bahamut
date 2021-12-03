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
	"fmt"
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

func setCommonHeader(w http.ResponseWriter, encoding elemental.EncodingType) {

	w.Header().Set("Accept", "application/msgpack,application/json")
	if encoding == elemental.EncodingTypeJSON {
		w.Header().Set("Content-Type", string(encoding)+"; charset=UTF-8")
	} else {
		w.Header().Set("Content-Type", string(encoding))
	}
}

func makeNotFoundHandler(accessControl *CORSAccessControlPolicy) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		writeHTTPResponse(
			w,
			makeErrorResponse(
				r.Context(),
				elemental.NewResponse(elemental.NewRequest()),
				ErrNotFound,
				nil,
				nil,
			),
			r.Header.Get("origin"),
			accessControl,
		)
	}
}

// writeHTTPResponse writes the response into the given http.ResponseWriter.
func writeHTTPResponse(w http.ResponseWriter, r *elemental.Response, origin string, accessControl *CORSAccessControlPolicy) int {

	// If r is nil, we simply stop.
	// It mostly means the client closed the connection and
	// no response is needed.
	if r == nil {
		return 0
	}

	for _, cookie := range r.Cookies {
		http.SetCookie(w, cookie)
	}

	setCommonHeader(w, r.Request.Accept)

	if accessControl != nil {
		accessControl.Inject(w.Header(), origin, false)
	}

	if r.Redirect != "" {
		w.Header().Set("Location", r.Redirect)
		w.WriteHeader(http.StatusFound)
		return http.StatusFound
	}

	w.Header().Set("X-Count-Total", strconv.Itoa(r.Total))

	if r.Next != "" {
		w.Header().Set("X-Next", r.Next)
	}

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

// If the first one is "v" it means the next one has to be a int for the version number.
func extractAPIVersion(path string) (version int, err error) {

	components := strings.SplitN(strings.TrimPrefix(path, "/"), "/", 3)
	if components[0] == "v" {
		version, err = strconv.Atoi(components[1])
		if err != nil {
			return 0, fmt.Errorf("Invalid api version number '%s'", components[1])
		}
	}

	return version, nil
}
