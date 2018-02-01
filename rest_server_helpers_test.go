package bahamut

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aporeto-inc/elemental/test/model"

	"github.com/aporeto-inc/elemental"

	. "github.com/smartystreets/goconvey/convey"
)

func TestRestServerHelpers_commonHeaders(t *testing.T) {

	Convey("Given I create a http.ResponseWriter", t, func() {

		w := httptest.NewRecorder()

		Convey("When I use setCommonHeader with a referer", func() {

			setCommonHeader(w, "http://toto.com:8443")

			Convey("Then the common headers should be set", func() {
				So(w.Header().Get("Content-Type"), ShouldEqual, "application/json; charset=UTF-8")
				So(w.Header().Get("Access-Control-Allow-Origin"), ShouldEqual, "http://toto.com:8443")
				So(w.Header().Get("Access-Control-Expose-Headers"), ShouldEqual, "X-Requested-With, X-Count-Total, X-Namespace, X-Messages")
				So(w.Header().Get("Access-Control-Allow-Methods"), ShouldEqual, "GET, POST, PUT, DELETE, PATCH, HEAD, OPTIONS")
				So(w.Header().Get("Access-Control-Allow-Headers"), ShouldEqual, "Authorization, Content-Type, Cache-Control, If-Modified-Since, X-Requested-With, X-Count-Total, X-Namespace, X-External-Tracking-Type, X-External-Tracking-ID")
				So(w.Header().Get("Access-Control-Allow-Credentials"), ShouldEqual, "true")
			})
		})

		Convey("When I use setCommonHeader without a referer", func() {

			setCommonHeader(w, "")

			Convey("Then the common headers should be set", func() {
				So(w.Header().Get("Content-Type"), ShouldEqual, "application/json; charset=UTF-8")
				So(w.Header().Get("Access-Control-Allow-Origin"), ShouldEqual, "*")
				So(w.Header().Get("Access-Control-Expose-Headers"), ShouldEqual, "X-Requested-With, X-Count-Total, X-Namespace, X-Messages")
				So(w.Header().Get("Access-Control-Allow-Methods"), ShouldEqual, "GET, POST, PUT, DELETE, PATCH, HEAD, OPTIONS")
				So(w.Header().Get("Access-Control-Allow-Headers"), ShouldEqual, "Authorization, Content-Type, Cache-Control, If-Modified-Since, X-Requested-With, X-Count-Total, X-Namespace, X-External-Tracking-Type, X-External-Tracking-ID")
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
		corsHandler(w, &http.Request{Header: h})

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
		notFoundHandler(w, &http.Request{Header: h})

		Convey("Then the response should be correct", func() {
			So(w.Code, ShouldEqual, http.StatusNotFound)
		})
	})
}

func TestRestServerHelper_writeHTTPResponse(t *testing.T) {

	Convey("Given I have a response with a redirect", t, func() {

		w := httptest.NewRecorder()
		r := elemental.NewResponse(elemental.NewRequest())
		r.Redirect = "https://la.bas"

		Convey("When I call writeHTTPResponse", func() {

			writeHTTPResponse(w, r)

			Convey("Then the should header Location should be set", func() {
				So(w.Header().Get("location"), ShouldEqual, "https://la.bas")
			})
		})
	})

	Convey("Given I have a response with no data", t, func() {

		w := httptest.NewRecorder()
		r := elemental.NewResponse(elemental.NewRequest())

		r.StatusCode = http.StatusNoContent

		Convey("When I call writeHTTPResponse", func() {

			writeHTTPResponse(w, r)

			Convey("Then the should headers should be correct", func() {
				So(w.Header().Get("X-Count-Total"), ShouldEqual, "0")
				So(w.Header().Get("X-Messages"), ShouldEqual, "")
			})

			Convey("Then the code should correct", func() {
				So(w.Code, ShouldEqual, http.StatusNoContent)
			})
		})
	})

	Convey("Given I have a response messages", t, func() {

		w := httptest.NewRecorder()
		r := elemental.NewResponse(elemental.NewRequest())

		r.Messages = []string{"msg1", "msg2"}

		Convey("When I call writeHTTPResponse", func() {

			writeHTTPResponse(w, r)

			Convey("Then the should header message should be set", func() {
				So(w.Header().Get("X-Messages"), ShouldEqual, "msg1;msg2")
			})
		})
	})

	Convey("Given I have a response with data", t, func() {

		w := httptest.NewRecorder()
		r := elemental.NewResponse(elemental.NewRequest())

		l1 := testmodel.NewList()
		l1.ID = "id"
		l1.Name = "toto"

		_ = r.Encode(l1)

		Convey("When I call writeHTTPResponse", func() {

			writeHTTPResponse(w, r)

			Convey("Then the body should be correct", func() {
				So(w.Header().Get("X-Count-Total"), ShouldEqual, "0")
				So(w.Header().Get("X-Messages"), ShouldEqual, "")
				So(w.Body.String(), ShouldEqual, string(r.Data))
			})
		})
	})
}
