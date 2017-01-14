// Author: Antoine Mercadal
// See LICENSE file for full LICENSE
// Copyright 2016 Aporeto.

package bahamut

import (
	"fmt"
	"net/http"
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
		request, _ := elemental.NewRequestFromHTTPRequest(req)

		c := NewContext(elemental.OperationRetrieveMany)
		c.ReadElementalRequest(request)

		Convey("Then it should be correctly initialized", func() {

			So(c.Request.Parameters.Get("page"), ShouldEqual, "1")
			So(c.Request.Parameters.Get("per_page"), ShouldEqual, "10")
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
		request, _ := elemental.NewRequestFromHTTPRequest(req)

		c := NewContext(elemental.OperationRetrieveMany)
		c.ReadElementalRequest(request)

		Convey("Then it should be correctly initialized", func() {

			So(c.Page.Current, ShouldEqual, 0)
			So(c.Page.Size, ShouldEqual, 0)
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

func TestContext_String(t *testing.T) {

	Convey("Given I have a Context, Info, Count, and Page", t, func() {

		count := &Count{
			Total:   10,
			Current: 1,
		}

		req := &elemental.Request{
			Parameters:     url.Values{"hello": []string{"world"}},
			Headers:        http.Header{"header": []string{"h1"}},
			Identity:       elemental.EmptyIdentity,
			ParentID:       "xxxx",
			ParentIdentity: elemental.EmptyIdentity,
		}

		page := &Page{
			Current: 1,
			First:   1,
			Last:    1,
			Next:    2,
			Prev:    0,
			Size:    5,
		}

		ctx := NewContext(elemental.OperationCreate)
		ctx.Request = req
		ctx.Count = count
		ctx.Page = page

		Convey("When I call the String method", func() {

			s := ctx.String()

			Convey("Then the string should be correct", func() {
				So(s, ShouldEqual, fmt.Sprintf("<context id:%s operation: create request: <request id: operation: namespace: recursive:false identity:<Identity |> objectid: parentidentity:<Identity |> parentid:xxxx> page: <page current:1 size:5> count: <count total:10 current:1>>", ctx.Identifier()))
			})
		})
	})
}
