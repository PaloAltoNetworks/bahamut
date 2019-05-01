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
	"net/http/httptest"
	"net/url"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"go.aporeto.io/elemental"
)

func TestRestServerHelpers_commonHeaders(t *testing.T) {

	Convey("Given I create a http.ResponseWriter", t, func() {

		w := httptest.NewRecorder()

		Convey("When I use setCommonHeader with a referer", func() {

			setCommonHeader(w, "http://toto.com:8443", elemental.EncodingTypeJSON)

			Convey("Then the common headers should be set", func() {
				So(w.Header().Get("Accept"), ShouldEqual, "application/msgpack,application/json")
				So(w.Header().Get("Content-Type"), ShouldEqual, "application/json")
				So(w.Header().Get("Access-Control-Allow-Origin"), ShouldEqual, "http://toto.com:8443")
				So(w.Header().Get("Access-Control-Expose-Headers"), ShouldEqual, "X-Requested-With, X-Count-Total, X-Namespace, X-Messages, X-Fields")
				So(w.Header().Get("Access-Control-Allow-Methods"), ShouldEqual, "GET, POST, PUT, DELETE, PATCH, HEAD, OPTIONS")
				So(w.Header().Get("Access-Control-Allow-Headers"), ShouldEqual, "Authorization, Accept, Content-Type, Cache-Control, If-Modified-Since, X-Requested-With, X-Count-Total, X-Namespace, X-External-Tracking-Type, X-External-Tracking-ID, X-TLS-Client-Certificate, Accept-Encoding, X-Fields, X-Read-Consistency, X-Write-Consistency")
				So(w.Header().Get("Access-Control-Allow-Credentials"), ShouldEqual, "true")
			})
		})

		Convey("When I use setCommonHeader without a referer", func() {

			setCommonHeader(w, "", elemental.EncodingTypeMSGPACK)

			Convey("Then the common headers should be set", func() {
				So(w.Header().Get("Accept"), ShouldEqual, "application/msgpack,application/json")
				So(w.Header().Get("Content-Type"), ShouldEqual, "application/msgpack")
				So(w.Header().Get("Access-Control-Allow-Origin"), ShouldEqual, "*")
				So(w.Header().Get("Access-Control-Expose-Headers"), ShouldEqual, "X-Requested-With, X-Count-Total, X-Namespace, X-Messages, X-Fields")
				So(w.Header().Get("Access-Control-Allow-Methods"), ShouldEqual, "GET, POST, PUT, DELETE, PATCH, HEAD, OPTIONS")
				So(w.Header().Get("Access-Control-Allow-Headers"), ShouldEqual, "Authorization, Accept, Content-Type, Cache-Control, If-Modified-Since, X-Requested-With, X-Count-Total, X-Namespace, X-External-Tracking-Type, X-External-Tracking-ID, X-TLS-Client-Certificate, Accept-Encoding, X-Fields, X-Read-Consistency, X-Write-Consistency")
				So(w.Header().Get("Access-Control-Allow-Credentials"), ShouldEqual, "true")
			})
		})
	})
}

func TestRestServerHelper_corsHandler(t *testing.T) {

	Convey("Given I call the corsHandler", t, func() {

		h := http.Header{}
		h.Add("Origin", "toto")

		w := httptest.NewRecorder()
		corsHandler(w, &http.Request{Header: h, URL: &url.URL{Path: "/path"}})

		Convey("Then the response should be correct", func() {
			So(w.Code, ShouldEqual, http.StatusOK)
		})
	})
}

func TestRestServerHelper_notFoundHandler(t *testing.T) {

	Convey("Given I call the notFoundHandler", t, func() {

		h := http.Header{}
		h.Add("Origin", "toto")

		w := httptest.NewRecorder()
		notFoundHandler(w, &http.Request{Header: h, URL: &url.URL{Path: "/path"}})

		Convey("Then the response should be correct", func() {
			So(w.Code, ShouldEqual, http.StatusNotFound)
		})
	})
}

func TestRestServerHelper_writeHTTPResponse(t *testing.T) {

	Convey("Given I have a response with a nil response", t, func() {

		w := httptest.NewRecorder()

		Convey("When I call writeHTTPResponse", func() {

			code := writeHTTPResponse(w, nil)

			Convey("Then the code should be 0", func() {
				So(code, ShouldEqual, 0)
			})
		})
	})

	Convey("Given I have a response with a redirect", t, func() {

		w := httptest.NewRecorder()
		r := elemental.NewResponse(elemental.NewRequest())
		r.Redirect = "https://la.bas"

		Convey("When I call writeHTTPResponse", func() {

			code := writeHTTPResponse(w, r)

			Convey("Then the should header Location should be set", func() {
				So(w.Header().Get("location"), ShouldEqual, "https://la.bas")
			})

			Convey("Then the code should be 302", func() {
				So(code, ShouldEqual, 302)
			})
		})
	})

	Convey("Given I have a response with no data", t, func() {

		w := httptest.NewRecorder()
		r := elemental.NewResponse(elemental.NewRequest())

		r.StatusCode = http.StatusNoContent

		Convey("When I call writeHTTPResponse", func() {

			code := writeHTTPResponse(w, r)

			Convey("Then the should headers should be correct", func() {
				So(w.Header().Get("X-Count-Total"), ShouldEqual, "0")
				So(w.Header().Get("X-Messages"), ShouldEqual, "")
			})

			Convey("Then the code should correct", func() {
				So(w.Code, ShouldEqual, http.StatusNoContent)
			})

			Convey("Then the code should be http.StatusNoContent", func() {
				So(code, ShouldEqual, http.StatusNoContent)
			})
		})
	})

	Convey("Given I have a response messages", t, func() {

		w := httptest.NewRecorder()
		r := elemental.NewResponse(elemental.NewRequest())

		r.Messages = []string{"msg1", "msg2"}
		r.StatusCode = 200

		Convey("When I call writeHTTPResponse", func() {

			code := writeHTTPResponse(w, r)

			Convey("Then the should header message should be set", func() {
				So(w.Header().Get("X-Messages"), ShouldEqual, "msg1;msg2")
			})

			Convey("Then the code should be http.StatusNoContent", func() {
				So(code, ShouldEqual, http.StatusOK)
			})
		})
	})

	Convey("Given I have a response with data", t, func() {

		w := httptest.NewRecorder()
		r := elemental.NewResponse(elemental.NewRequest())

		r.StatusCode = http.StatusCreated
		r.Data = []byte("hello")

		Convey("When I call writeHTTPResponse", func() {

			code := writeHTTPResponse(w, r)

			Convey("Then the body should be correct", func() {
				So(w.Header().Get("X-Count-Total"), ShouldEqual, "0")
				So(w.Header().Get("X-Messages"), ShouldEqual, "")
				So(w.Body.String(), ShouldEqual, string(r.Data))
			})

			Convey("Then the code should be http.StatusNoContent", func() {
				So(code, ShouldEqual, http.StatusCreated)
			})
		})
	})
}
