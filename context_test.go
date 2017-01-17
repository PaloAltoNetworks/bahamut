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

		c := NewContext()
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

		c := NewContext()
		c.ReadElementalRequest(request)

		Convey("Then it should be correctly initialized", func() {

			So(c.Page.Current, ShouldEqual, 0)
			So(c.Page.Size, ShouldEqual, 0)
		})
	})
}

func TestContext_Identifier(t *testing.T) {

	Convey("Given I have a context", t, func() {

		ctx := NewContext()
		ctx.Request.Operation = elemental.OperationCreate

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

		c := NewContext()
		c.Request.Operation = elemental.OperationCreate

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
			Namespace:      "/thens",
			Parameters:     url.Values{"hello": []string{"world"}},
			Headers:        http.Header{"header": []string{"h1"}},
			Identity:       elemental.EmptyIdentity,
			ParentID:       "xxxx",
			ParentIdentity: elemental.EmptyIdentity,
			Operation:      elemental.OperationCreate,
		}

		page := &Page{
			Current: 1,
			First:   1,
			Last:    1,
			Next:    2,
			Prev:    0,
			Size:    5,
		}

		ctx := NewContext()
		ctx.Request = req
		ctx.Count = count
		ctx.Page = page

		Convey("When I call the String method", func() {

			s := ctx.String()

			Convey("Then the string should be correct", func() {
				So(s, ShouldEqual, fmt.Sprintf("<context id:%s request:<request id: operation:create namespace:/thens recursive:false identity:<Identity |> objectid: parentidentity:<Identity |> parentid:xxxx> page:<page current:1 size:5> count:<count total:10 current:1>>", ctx.Identifier()))
			})
		})
	})
}

func TestContext_Duplicate(t *testing.T) {

	Convey("Given I have a Context, Info, Count, and Page", t, func() {

		count := &Count{
			Total:   10,
			Current: 1,
		}

		req := &elemental.Request{
			Namespace:      "/thens",
			Parameters:     url.Values{"hello": []string{"world"}},
			Headers:        http.Header{"header": []string{"h1"}},
			Identity:       elemental.EmptyIdentity,
			ParentID:       "xxxx",
			ParentIdentity: elemental.EmptyIdentity,
			Operation:      elemental.OperationCreate,
		}

		page := &Page{
			Current: 1,
			First:   2,
			Last:    3,
			Next:    4,
			Prev:    5,
			Size:    6,
		}

		ctx := NewContext()
		ctx.Request = req
		ctx.Count = count
		ctx.Page = page
		ctx.Metadata = map[string]interface{}{"hello": "world"}
		ctx.UserInfo = "ouais"
		ctx.InputData = "input"
		ctx.OutputData = "output"
		ctx.StatusCode = 42

		Convey("When I call the Duplicate method", func() {

			ctx2 := ctx.Duplicate()

			Convey("Then the duplicated context should be correct", func() {
				So(ctx.Count.Current, ShouldEqual, ctx2.Count.Current)
				So(ctx.Count.Total, ShouldEqual, ctx2.Count.Total)
				So(ctx.Metadata["hello"].(string), ShouldEqual, "world")
				So(ctx.InputData, ShouldEqual, ctx2.InputData)
				So(ctx.OutputData, ShouldEqual, ctx2.OutputData)
				So(ctx.Page.Current, ShouldEqual, ctx2.Page.Current)
				So(ctx.Page.First, ShouldEqual, ctx2.Page.First)
				So(ctx.Page.Last, ShouldEqual, ctx2.Page.Last)
				So(ctx.Page.Next, ShouldEqual, ctx2.Page.Next)
				So(ctx.Page.Prev, ShouldEqual, ctx2.Page.Prev)
				So(ctx.Page.Size, ShouldEqual, ctx2.Page.Size)
				So(ctx.Request.Namespace, ShouldEqual, ctx2.Request.Namespace)
				So(ctx.Request.ParentID, ShouldEqual, ctx2.Request.ParentID)
				So(ctx.StatusCode, ShouldEqual, ctx2.StatusCode)
				So(ctx.UserInfo, ShouldEqual, ctx2.UserInfo)
			})
		})
	})
}
