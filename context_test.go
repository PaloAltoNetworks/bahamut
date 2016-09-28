// Author: Antoine Mercadal
// See LICENSE file for full LICENSE
// Copyright 2016 Aporeto.

package bahamut

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/aporeto-inc/elemental"
	. "github.com/smartystreets/goconvey/convey"
)

func TestContext_MakeContext(t *testing.T) {

	Convey("Given I create Context from a request with pagination info", t, func() {

		url, _ := url.Parse("http://link.com/path?page=1&per_page=10")
		req := &http.Request{
			Host:   "link.com",
			URL:    url,
			Method: http.MethodGet,
		}

		c := NewContext(elemental.OperationRetrieveMany)
		c.ReadRequest(req)

		Convey("Then it should be correctly initialized", func() {

			So(c.Info.Parameters.Get("page"), ShouldEqual, "1")
			So(c.Info.Parameters.Get("per_page"), ShouldEqual, "10")
			So(c.Page.Current, ShouldEqual, 1)
			So(c.Page.Size, ShouldEqual, 10)
		})
	})

	Convey("Given I create Context from a request with no pagination info", t, func() {

		url, _ := url.Parse("http://link.com/path")
		req := &http.Request{
			Host:   "link.com",
			URL:    url,
			Method: http.MethodGet,
		}

		c := NewContext(elemental.OperationRetrieveMany)
		c.ReadRequest(req)

		Convey("Then it should be correctly initialized", func() {

			So(c.Page.Current, ShouldEqual, 1)
			So(c.Page.Size, ShouldEqual, 100)
		})
	})
}

func TestContext_Identifier(t *testing.T) {

	Convey("Given I have a context", t, func() {

		ctx := NewContext(elemental.OperationCreate)

		Convey("When I get its Identifier", func() {

			id := ctx.Identifier()

			Convey("Then the identifier should not be empty", func() {
				So(id, ShouldNotBeEmpty)
			})
		})
	})
}

func TestContext_WriteResponse(t *testing.T) {

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

		c := NewContext(elemental.OperationRetrieveMany)
		c.ReadRequest(req)

		c.Count.Total = 40
		c.OutputData = []*Entity{e1, e2}

		Convey("When I write the response from a context with no error for a retrieve", func() {

			w := httptest.NewRecorder()
			c.WriteResponse(w)

			Convey("Then the status code should be default to 200", func() {
				So(w.Code, ShouldEqual, 200)
			})

			Convey("Then the pagination headers should be correct", func() {
				So(w.Header().Get("X-Page-First"), ShouldEqual, "http://link.com/path?page=1&per_page=10")
				So(w.Header().Get("X-Page-Prev"), ShouldEqual, "http://link.com/path?page=1&per_page=10")
				So(w.Header().Get("X-Page-Next"), ShouldEqual, "http://link.com/path?page=3&per_page=10")
				So(w.Header().Get("X-Page-Last"), ShouldEqual, "http://link.com/path?page=4&per_page=10")
			})

			Convey("Then the status should be 200", func() {
				So(string(w.Body.Bytes()), ShouldEqual, "[{\"name\":\"e1\"},{\"name\":\"e2\"}]\n")
			})

		})

		Convey("When I write the response from a context with no error for a create", func() {

			w := httptest.NewRecorder()
			c.Operation = elemental.OperationCreate
			c.WriteResponse(w)

			Convey("Then the status code should be default to 201", func() {
				So(w.Code, ShouldEqual, 201)
			})
		})

		Convey("When I write the response from a context with errors", func() {

			w := httptest.NewRecorder()
			c.AddErrors(elemental.NewError("error", "description", "subject", 4042))
			c.WriteResponse(w)

			Convey("Then the status code should be correct", func() {
				So(w.Code, ShouldEqual, 4042)
			})

			Convey("Then the body should be correct", func() {
				So(string(w.Body.Bytes()), ShouldEqual, "[{\"code\":4042,\"description\":\"description\",\"subject\":\"subject\",\"title\":\"error\",\"data\":null}]\n")
			})
		})

		Convey("When I write the response from a context with unmarshallable errors", func() {

			w := httptest.NewRecorder()
			e := elemental.NewError("error", "description", "subject", 42)
			e.Data = NewUnmarshalableList()
			c.AddErrors(e)
			err := c.WriteResponse(w)

			Convey("Then err should not be nil", func() {
				So(err, ShouldNotBeNil)
			})
		})

		Convey("When I try write the response with an unmarshallable object", func() {

			w := httptest.NewRecorder()
			c.OutputData = NewUnmarshalableList()
			err := c.WriteResponse(w)

			Convey("Then err should not be nil", func() {
				So(err, ShouldNotBeNil)
			})
		})
	})
}

func TestContext_Errors(t *testing.T) {

	Convey("Given I create a Context", t, func() {

		c := NewContext(elemental.OperationRetrieveMany)

		Convey("Then the context should not have any Error", func() {
			So(c.HasErrors(), ShouldBeFalse)
		})

		Convey("When I add an Error", func() {

			c.AddErrors(elemental.NewError("title", "description", "subject", 42))

			Convey("Then the context should have some Errors", func() {
				So(c.HasErrors(), ShouldBeTrue)
				So(len(c.Errors()), ShouldEqual, 1)
			})
		})

		Convey("When I set the Errors", func() {

			c.SetErrors(elemental.NewErrors(
				elemental.NewError("title", "description", "subject", 42),
				elemental.NewError("title", "description", "subject", 42),
			))

			Convey("Then the context should have some Errors", func() {
				So(c.HasErrors(), ShouldBeTrue)
				So(len(c.Errors()), ShouldEqual, 2)
			})
		})
	})
}

func TestContext_Events(t *testing.T) {

	Convey("Given I create a Context", t, func() {

		c := NewContext(elemental.OperationRetrieveMany)

		Convey("When I enqueue 2 events", func() {

			c.EnqueueEvents(
				elemental.NewEvent(elemental.EventCreate, NewList()),
				elemental.NewEvent(elemental.EventCreate, NewList()))

			Convey("Then I should have 2 events in the queue", func() {
				So(c.HasEvents(), ShouldBeTrue)
				So(len(c.Events()), ShouldEqual, 2)
			})
		})

		Convey("When I set the Events", func() {

			c.EnqueueEvents(
				elemental.NewEvent(elemental.EventCreate, NewList()),
				elemental.NewEvent(elemental.EventCreate, NewList()),
			)

			c.SetEvents(
				elemental.NewEvents(
					elemental.NewEvent(elemental.EventCreate, NewList()),
				),
			)

			Convey("Then the context should have some Event", func() {
				So(c.HasEvents(), ShouldBeTrue)
				So(len(c.Events()), ShouldEqual, 1)
			})
		})
	})
}

func TestError_WriteHTTPError(t *testing.T) {

	Convey("Given I create a http.ResponseWriter", t, func() {

		w := httptest.NewRecorder()

		Convey("When I use WriteHTTPError", func() {

			WriteHTTPError(w, 43, elemental.NewError("title", "description", "subject", 42))

			Convey("Then the status should be 42", func() {
				So(w.Code, ShouldEqual, 43)
			})

			Convey("Then the body should be correct", func() {
				So(string(w.Body.Bytes()), ShouldEqual, "[{\"code\":42,\"description\":\"description\",\"subject\":\"subject\",\"title\":\"title\",\"data\":null}]\n")
			})
		})
	})
}

func TestError_commonHeaders(t *testing.T) {

	Convey("Given I create a http.ResponseWriter", t, func() {

		w := httptest.NewRecorder()

		Convey("When I use setCommonHeader with a referer", func() {

			setCommonHeader(w, "http://toto.com:8443/something")

			Convey("Then the common headers should be set", func() {
				So(w.Header().Get("Content-Type"), ShouldEqual, "application/json; charset=UTF-8")
				So(w.Header().Get("Access-Control-Allow-Origin"), ShouldEqual, "http://toto.com:8443")
				So(w.Header().Get("Access-Control-Expose-Headers"), ShouldEqual, "X-Requested-With, X-Count-Local, X-Count-Total, X-PageCurrent, X-Page-Size, X-Page-Prev, X-Page-Next, X-Page-First, X-Page-Last, X-Namespace")
				So(w.Header().Get("Access-Control-Allow-Methods"), ShouldEqual, "GET, POST, PUT, DELETE, PATCH, HEAD, OPTIONS")
				So(w.Header().Get("Access-Control-Allow-Headers"), ShouldEqual, "Authorization, Content-Type, Cache-Control, If-Modified-Since, X-Requested-With, X-Count-Local, X-Count-Total, X-PageCurrent, X-Page-Size, X-Page-Prev, X-Page-Next, X-Page-First, X-Page-Last, X-Namespace")
				So(w.Header().Get("Access-Control-Allow-Credentials"), ShouldEqual, "true")
			})
		})

		Convey("When I use setCommonHeader without a referer", func() {

			setCommonHeader(w, "")

			Convey("Then the common headers should be set", func() {
				So(w.Header().Get("Content-Type"), ShouldEqual, "application/json; charset=UTF-8")
				So(w.Header().Get("Access-Control-Allow-Origin"), ShouldEqual, "*")
				So(w.Header().Get("Access-Control-Expose-Headers"), ShouldEqual, "X-Requested-With, X-Count-Local, X-Count-Total, X-PageCurrent, X-Page-Size, X-Page-Prev, X-Page-Next, X-Page-First, X-Page-Last, X-Namespace")
				So(w.Header().Get("Access-Control-Allow-Methods"), ShouldEqual, "GET, POST, PUT, DELETE, PATCH, HEAD, OPTIONS")
				So(w.Header().Get("Access-Control-Allow-Headers"), ShouldEqual, "Authorization, Content-Type, Cache-Control, If-Modified-Since, X-Requested-With, X-Count-Local, X-Count-Total, X-PageCurrent, X-Page-Size, X-Page-Prev, X-Page-Next, X-Page-First, X-Page-Last, X-Namespace")
				So(w.Header().Get("Access-Control-Allow-Credentials"), ShouldEqual, "true")
			})
		})
	})
}
