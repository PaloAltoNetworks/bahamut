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
	"github.com/aporeto-inc/elemental/test/model"
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

		c := NewContextWithRequest(request)

		Convey("Then it should be correctly initialized", func() {

			So(c.Request.Parameters.Get("page"), ShouldEqual, "1")
			So(c.Request.Parameters.Get("per_page"), ShouldEqual, "10")
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
				elemental.NewEvent(elemental.EventCreate, testmodel.NewList()),
				elemental.NewEvent(elemental.EventCreate, testmodel.NewList()))

			Convey("Then I should have 2 events in the queue", func() {
				So(c.HasEvents(), ShouldBeTrue)
				So(len(c.Events()), ShouldEqual, 2)
			})
		})

		Convey("When I set the Events", func() {

			c.EnqueueEvents(
				elemental.NewEvent(elemental.EventCreate, testmodel.NewList()),
				elemental.NewEvent(elemental.EventCreate, testmodel.NewList()),
			)

			c.SetEvents(
				elemental.NewEvents(
					elemental.NewEvent(elemental.EventCreate, testmodel.NewList()),
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

		req := &elemental.Request{
			Namespace:      "/thens",
			Parameters:     url.Values{"hello": []string{"world"}},
			Headers:        http.Header{"header": []string{"h1"}},
			Identity:       elemental.EmptyIdentity,
			ParentID:       "xxxx",
			ParentIdentity: elemental.EmptyIdentity,
			Operation:      elemental.OperationCreate,
			Version:        12,
		}

		ctx := NewContext()
		ctx.Request = req
		ctx.CountTotal = 10

		Convey("When I call the String method", func() {

			s := ctx.String()

			Convey("Then the string should be correct", func() {
				So(s, ShouldEqual, fmt.Sprintf("<context id:%s request:<request id: operation:create namespace:/thens recursive:false identity:<Identity |> objectid: parentidentity:<Identity |> parentid:xxxx version:12> totalcount:10>", ctx.Identifier()))
			})
		})
	})
}

func TestContext_Duplicate(t *testing.T) {

	Convey("Given I have a Context, Info, Count, and Page", t, func() {

		req := &elemental.Request{
			Namespace:      "/thens",
			Parameters:     url.Values{"hello": []string{"world"}},
			Headers:        http.Header{"header": []string{"h1"}},
			Identity:       elemental.EmptyIdentity,
			ParentID:       "xxxx",
			ParentIdentity: elemental.EmptyIdentity,
			Operation:      elemental.OperationCreate,
		}

		ctx := NewContext()
		ctx.Request = req
		ctx.CountTotal = 10
		ctx.Metadata = map[string]interface{}{"hello": "world"}
		ctx.InputData = "input"
		ctx.OutputData = "output"
		ctx.StatusCode = 42
		ctx.AddMessage("a")
		ctx.AddMessage("b")
		ctx.SetClaims([]string{"ouais=yes"})

		Convey("When I call the Duplicate method", func() {

			ctx2 := ctx.Duplicate()

			Convey("Then the duplicated context should be correct", func() {
				So(ctx.CountTotal, ShouldEqual, ctx2.CountTotal)
				So(ctx.Metadata["hello"].(string), ShouldEqual, "world")
				So(ctx.InputData, ShouldEqual, ctx2.InputData)
				So(ctx.OutputData, ShouldEqual, ctx2.OutputData)
				So(ctx.Request.Namespace, ShouldEqual, ctx2.Request.Namespace)
				So(ctx.Request.ParentID, ShouldEqual, ctx2.Request.ParentID)
				So(ctx.StatusCode, ShouldEqual, ctx2.StatusCode)
				So(ctx.claims, ShouldResemble, ctx2.claims)
				So(ctx.claimsMap, ShouldResemble, ctx2.claimsMap)
				So(ctx.messages(), ShouldResemble, ctx2.messages())
			})
		})
	})
}

func TestContext_GetClaims(t *testing.T) {

	Convey("Given I have a Context with claims", t, func() {

		ctx := NewContext()
		ctx.SetClaims([]string{"ouais=yes"})

		Convey("When I call GetClaims", func() {

			claims := ctx.GetClaims()

			Convey("Then claims should be correct", func() {
				So(claims, ShouldResemble, []string{"ouais=yes"})
			})
		})

		Convey("When I call GetClaimsMap", func() {

			claimsMap := ctx.GetClaimsMap()

			Convey("Then claims should be correct", func() {
				So(claimsMap, ShouldResemble, map[string]string{"ouais": "yes"})
			})
		})
	})

	Convey("Given I have a Context nil claims", t, func() {

		ctx := NewContext()
		ctx.SetClaims(nil)

		Convey("When I call GetClaims", func() {

			claims := ctx.GetClaims()

			Convey("Then claims should be correct", func() {
				So(claims, ShouldResemble, []string{})
			})
		})

		Convey("When I call GetClaimsMap", func() {

			claimsMap := ctx.GetClaimsMap()

			Convey("Then claims should be correct", func() {
				So(claimsMap, ShouldResemble, map[string]string{})
			})
		})
	})
}
