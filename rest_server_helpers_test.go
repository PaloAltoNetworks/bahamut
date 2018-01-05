package bahamut

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/aporeto-inc/elemental"
	. "github.com/smartystreets/goconvey/convey"
)

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

	type Entity struct {
		Name string `json:"name"`
	}

	e1 := &Entity{Name: "e1"}
	e2 := &Entity{Name: "e2"}

	Convey("Given I create Context from a request with pagination info", t, func() {

		u, _ := url.Parse("http://link.com/path?page=2&per_page=10")
		req := &http.Request{
			Host:   "link.com",
			URL:    u,
			Method: http.MethodGet,
		}
		request, _ := elemental.NewRequestFromHTTPRequest(req)

		c := NewContextWithRequest(request)

		c.CountTotal = 40

		Convey("When I write the response from a context with no error for a retrieve", func() {

			w := httptest.NewRecorder()
			c.OutputData = []*Entity{e1, e2}
			req.Method = http.MethodGet
			writeHTTPResponse(w, c)

			Convey("Then the status code should be default to 200", func() {
				So(w.Code, ShouldEqual, 200)
			})

			Convey("Then the status should be 200", func() {
				So(w.Body.String(), ShouldEqual, "[{\"name\":\"e1\"},{\"name\":\"e2\"}]\n")
			})
		})

		Convey("When I write the response from a context with no error for a info", func() {

			w := httptest.NewRecorder()
			c.Request.Operation = elemental.OperationInfo
			req.Method = http.MethodHead
			writeHTTPResponse(w, c)

			Convey("Then the status code should be default to 204", func() {
				So(w.Code, ShouldEqual, 204)
			})

			Convey("Then the body should be empty", func() {
				So(len(w.Body.Bytes()), ShouldEqual, 0)
			})
		})

		Convey("When I write the response from a context with no error for a create and some data", func() {

			w := httptest.NewRecorder()
			c.Request.Operation = elemental.OperationCreate
			req.Method = http.MethodPost
			c.OutputData = struct{}{}
			writeHTTPResponse(w, c)

			Convey("Then the status code should be default to 201", func() {
				So(w.Code, ShouldEqual, 201)
			})
		})

		Convey("When I write the response from a context with no error for a create and no data", func() {

			w := httptest.NewRecorder()
			c.Request.Operation = elemental.OperationCreate
			req.Method = http.MethodPost
			writeHTTPResponse(w, c)

			Convey("Then the status code should be default to 201", func() {
				So(w.Code, ShouldEqual, 204)
			})
		})

		Convey("When I try write the response with an unmarshallable object", func() {

			w := httptest.NewRecorder()
			c.OutputData = NewUnmarshalableList()
			writeHTTPResponse(w, c)

			Convey("Then status code should not be 500", func() {
				So(w.Code, ShouldEqual, http.StatusInternalServerError)
			})
		})

		Convey("When I try write the response with an redirect object", func() {

			w := httptest.NewRecorder()
			c.Redirect = "http://toto.com"
			writeHTTPResponse(w, c)

			Convey("Then status code should not be 302", func() {
				So(w.Code, ShouldEqual, http.StatusFound)
			})

			Convey("Then redirect header should be set", func() {
				So(w.Header().Get("Location"), ShouldEqual, "http://toto.com")
			})
		})
	})
}

func TestRestServerHelpers_writeHTTPError(t *testing.T) {

	Convey("Given I create a http.ResponseWriter", t, func() {

		w := httptest.NewRecorder()
		req := elemental.NewRequest()
		req.Headers.Set("Origin", "origin")

		Convey("When I use writeHTTPError with a simple elemental.Error", func() {

			writeHTTPError(w, req, elemental.NewError("title", "description", "subject", 42))

			Convey("Then the status should be 42", func() {
				So(w.Code, ShouldEqual, 42)
			})

			Convey("Then the body should be correct", func() {
				So(w.Body.String(), ShouldResemble, "[{\"code\":42,\"description\":\"description\",\"subject\":\"subject\",\"title\":\"title\",\"data\":null,\"trace\":\""+req.RequestID+"\"}]\n")
			})
		})

		Convey("When I use writeHTTPError with an elemental.Errors", func() {

			errs := elemental.NewErrors(elemental.NewError("title", "description", "subject", 43))
			writeHTTPError(w, req, errs)

			Convey("Then the status should be 43", func() {
				So(w.Code, ShouldEqual, 43)
			})

			Convey("Then the body should be correct", func() {
				So(w.Body.String(), ShouldResemble, "[{\"code\":43,\"description\":\"description\",\"subject\":\"subject\",\"title\":\"title\",\"data\":null,\"trace\":\""+req.RequestID+"\"}]\n")
			})
		})
	})
}

func TestRestServerHelpers_commonHeaders(t *testing.T) {

	Convey("Given I create a http.ResponseWriter", t, func() {

		w := httptest.NewRecorder()

		Convey("When I use setCommonHeader with a referer", func() {

			setCommonHeader(w, "http://toto.com:8443")

			Convey("Then the common headers should be set", func() {
				So(w.Header().Get("Content-Type"), ShouldEqual, "application/json; charset=UTF-8")
				So(w.Header().Get("Access-Control-Allow-Origin"), ShouldEqual, "http://toto.com:8443")
				So(w.Header().Get("Access-Control-Expose-Headers"), ShouldEqual, "X-Requested-With, X-Count-Total, X-Namespace")
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
				So(w.Header().Get("Access-Control-Expose-Headers"), ShouldEqual, "X-Requested-With, X-Count-Total, X-Namespace")
				So(w.Header().Get("Access-Control-Allow-Methods"), ShouldEqual, "GET, POST, PUT, DELETE, PATCH, HEAD, OPTIONS")
				So(w.Header().Get("Access-Control-Allow-Headers"), ShouldEqual, "Authorization, Content-Type, Cache-Control, If-Modified-Since, X-Requested-With, X-Count-Total, X-Namespace, X-External-Tracking-Type, X-External-Tracking-ID")
				So(w.Header().Get("Access-Control-Allow-Credentials"), ShouldEqual, "true")
			})
		})
	})
}

func TestRestServerHelpers_runHTTPDispatcher(t *testing.T) {

	Convey("Given I have a fake dispatcher", t, func() {

		called := 0

		w := httptest.NewRecorder()
		gctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
		defer cancel()

		ctx := NewContext()
		ctx.Request = elemental.NewRequestWithContext(gctx)

		Convey("When I call runHTTPDispatcher", func() {

			d := func() error {
				called++
				return nil
			}

			runHTTPDispatcher(ctx, w, d)

			Convey("Then the code should be 204", func() {
				So(w.Code, ShouldEqual, 204)
			})

			Convey("Then the dispatcher should have been called once", func() {
				So(called, ShouldEqual, 1)
			})
		})

		Convey("When I call runHTTPDispatcher and it returns an error", func() {

			d := func() error {
				called++
				return elemental.NewError("nop", "nope", "test", 42)
			}

			runHTTPDispatcher(ctx, w, d)

			Convey("Then the code should be 42", func() {
				So(w.Code, ShouldEqual, 42)
			})

			Convey("Then the dispatcher should have been called once", func() {
				So(called, ShouldEqual, 1)
			})
		})

		Convey("When I call runHTTPDispatcher and cancel the context", func() {

			d := func() error {
				time.Sleep(2 * time.Second)
				called++
				return nil
			}

			go func() { runHTTPDispatcher(ctx, nil, d) }()
			time.Sleep(300 * time.Millisecond)
			cancel()

			Convey("Then the dispatcher should have been called once", func() {
				So(called, ShouldEqual, 0)
			})
		})

	})
}
