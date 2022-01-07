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

		Convey("When I use setCommonHeader using json", func() {

			setCommonHeader(w, elemental.EncodingTypeJSON)

			Convey("Then the common headers should be set", func() {
				So(w.Header().Get("Accept"), ShouldEqual, "application/msgpack,application/json")
				So(w.Header().Get("Content-Type"), ShouldEqual, "application/json; charset=UTF-8")
			})
		})

		Convey("When I use setCommonHeader using msgpack", func() {

			setCommonHeader(w, elemental.EncodingTypeMSGPACK)

			Convey("Then the common headers should be set", func() {
				So(w.Header().Get("Accept"), ShouldEqual, "application/msgpack,application/json")
				So(w.Header().Get("Content-Type"), ShouldEqual, "application/msgpack")
			})
		})
	})
}

func TestRestServerHelper_notFoundHandler(t *testing.T) {

	Convey("Given I call the notFoundHandler", t, func() {

		h := http.Header{}
		h.Add("Origin", "toto")

		w := httptest.NewRecorder()
		makeNotFoundHandler(nil)(w, &http.Request{Header: h, URL: &url.URL{Path: "/path"}})

		So(w.Code, ShouldEqual, http.StatusNotFound)
	})
}

func TestRestServerHelper_writeHTTPResponse(t *testing.T) {

	Convey("Given I have a response with a nil response", t, func() {

		w := httptest.NewRecorder()
		a := NewDefaultCORSController("origin.com", nil)
		ac := a.PolicyForRequest(nil)

		w.Code = 200
		code := writeHTTPResponse(w, nil, "", ac)

		So(code, ShouldEqual, 0)
	})

	Convey("Given I have a response with a redirect", t, func() {

		w := httptest.NewRecorder()
		r := elemental.NewResponse(elemental.NewRequest())
		a := NewDefaultCORSController("origin.com", nil)
		ac := a.PolicyForRequest(nil)
		r.Redirect = "https://la.bas"

		code := writeHTTPResponse(w, r, "", ac)

		So(w.Header().Get("location"), ShouldEqual, "https://la.bas")
		So(w.Header().Get("Access-Control-Allow-Origin"), ShouldEqual, "origin.com")
		So(code, ShouldEqual, 302)
	})

	Convey("Given I have a response with no data", t, func() {

		w := httptest.NewRecorder()
		r := elemental.NewResponse(elemental.NewRequest())
		a := NewDefaultCORSController("origin.com", nil)
		ac := a.PolicyForRequest(nil)

		r.StatusCode = http.StatusNoContent

		code := writeHTTPResponse(w, r, "", ac)

		So(w.Header().Get("X-Count-Total"), ShouldEqual, "0")
		So(w.Header().Get("X-Messages"), ShouldEqual, "")
		So(w.Header().Get("Access-Control-Allow-Origin"), ShouldEqual, "origin.com")
		So(w.Code, ShouldEqual, http.StatusNoContent)
		So(code, ShouldEqual, http.StatusNoContent)
	})

	Convey("Given I have a response messages", t, func() {

		w := httptest.NewRecorder()
		r := elemental.NewResponse(elemental.NewRequest())
		a := NewDefaultCORSController("origin.com", nil)
		ac := a.PolicyForRequest(nil)

		r.Messages = []string{"msg1", "msg2"}
		r.StatusCode = 200

		code := writeHTTPResponse(w, r, "", ac)

		So(w.Header().Get("X-Messages"), ShouldEqual, "msg1;msg2")
		So(w.Header().Get("Access-Control-Allow-Origin"), ShouldEqual, "origin.com")
		So(code, ShouldEqual, http.StatusOK)
	})

	Convey("Given I have a response next", t, func() {

		w := httptest.NewRecorder()
		r := elemental.NewResponse(elemental.NewRequest())
		a := NewDefaultCORSController("origin.com", nil)
		ac := a.PolicyForRequest(nil)

		r.Next = "next"
		r.StatusCode = 200

		code := writeHTTPResponse(w, r, "", ac)

		So(w.Header().Get("X-Next"), ShouldEqual, "next")
		So(w.Header().Get("Access-Control-Allow-Origin"), ShouldEqual, "origin.com")
		So(code, ShouldEqual, http.StatusOK)
	})

	Convey("Given I have a response with data", t, func() {

		w := httptest.NewRecorder()
		r := elemental.NewResponse(elemental.NewRequest())
		a := NewDefaultCORSController("origin.com", nil)
		ac := a.PolicyForRequest(nil)

		r.StatusCode = http.StatusOK
		r.Data = []byte("hello")

		code := writeHTTPResponse(w, r, "", ac)

		So(w.Header().Get("X-Count-Total"), ShouldEqual, "0")
		So(w.Header().Get("X-Messages"), ShouldEqual, "")
		So(w.Header().Get("Access-Control-Allow-Origin"), ShouldEqual, "origin.com")
		So(w.Body.String(), ShouldEqual, string(r.Data))
		So(code, ShouldEqual, http.StatusOK)
	})

	Convey("Given I have a some cookies", t, func() {

		w := httptest.NewRecorder()
		r := elemental.NewResponse(elemental.NewRequest())
		a := NewDefaultCORSController("origin.com", nil)
		ac := a.PolicyForRequest(nil)
		r.StatusCode = 200
		r.Cookies = []*http.Cookie{
			{
				Name:  "ca",
				Value: "ca",
			},
			{
				Name:  "cb",
				Value: "cb",
			},
		}

		writeHTTPResponse(w, r, "", ac)

		So(w.Header()["Set-Cookie"], ShouldResemble, []string{"ca=ca", "cb=cb"})
		So(w.Header().Get("Access-Control-Allow-Origin"), ShouldEqual, "origin.com")
	})
}

func Test_extractAPIVersion(t *testing.T) {
	type args struct {
		path string
	}
	tests := []struct {
		name        string
		args        args
		wantVersion int
		wantErr     bool
	}{
		{
			"no path",
			args{
				"",
			},
			0,
			false,
		},
		{
			"valid unversionned with no heading /",
			args{
				"objects",
			},
			0,
			false,
		},
		{
			"valid unversionned with heading /",
			args{
				"/objects",
			},
			0,
			false,
		},
		{
			"valid versionned with no heading /",
			args{
				"v/4/objects",
			},
			4,
			false,
		},
		{
			"valid versionned with heading /",
			args{
				"/v/4/objects",
			},
			4,
			false,
		},
		{
			"invalid versionned with no heading /",
			args{
				"v/dog/objects",
			},
			0,
			true,
		},
		{
			"invalid versionned with heading /",
			args{
				"/v/dog/objects",
			},
			0,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotVersion, err := extractAPIVersion(tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("extractAPIVersion() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotVersion != tt.wantVersion {
				t.Errorf("extractAPIVersion() = %v, want %v", gotVersion, tt.wantVersion)
			}
		})
	}
}
