// Author: Antoine Mercadal
// See LICENSE file for full LICENSE
// Copyright 2016 Aporeto.

package bahamut

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"go.aporeto.io/elemental"
	"go.aporeto.io/elemental/test/model"
)

func TestContext_MakeContext(t *testing.T) {

	Convey("Given I create Context from a request with pagination info", t, func() {

		url, _ := url.Parse("http://link.com/path?page=1&per_page=10")
		req := &http.Request{
			Host:   "link.com",
			URL:    url,
			Method: http.MethodGet,
		}
		request, _ := elemental.NewRequestFromHTTPRequest(req, testmodel.Manager())

		c := newContext(context.TODO(), request)

		Convey("Then it should be correctly initialized", func() {

			So(c.request.Parameters.Get("page"), ShouldEqual, "1")
			So(c.request.Parameters.Get("per_page"), ShouldEqual, "10")
		})
	})
}

func TestContext_Identifier(t *testing.T) {

	Convey("Given I have a context", t, func() {

		ctx := newContext(context.TODO(), elemental.NewRequest())
		ctx.request.Operation = elemental.OperationCreate

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

		c := newContext(context.TODO(), elemental.NewRequest())
		c.request.Operation = elemental.OperationCreate

		Convey("When I enqueue 2 events", func() {

			c.EnqueueEvents(
				elemental.NewEvent(elemental.EventCreate, testmodel.NewList()),
				elemental.NewEvent(elemental.EventCreate, testmodel.NewList()))

			Convey("Then I should have 2 events in the queue", func() {
				So(len(c.events), ShouldEqual, 2)
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

		ctx := newContext(context.TODO(), elemental.NewRequest())
		ctx.request = req
		ctx.count = 10

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

		ctx := newContext(context.TODO(), elemental.NewRequest())
		ctx.request = req
		ctx.count = 10
		ctx.inputData = "input"
		ctx.outputData = "output"
		ctx.statusCode = 42
		ctx.AddMessage("a")
		ctx.AddMessage("b")
		ctx.SetMetadata("hello", "world")
		ctx.SetClaims([]string{"ouais=yes"})

		Convey("When I call the Duplicate method", func() {

			ctx2 := ctx.Duplicate()

			Convey("Then the duplicated context should be correct", func() {
				So(ctx.count, ShouldEqual, ctx2.(*bcontext).count)
				So(ctx.Metadata("hello").(string), ShouldEqual, "world")
				So(ctx.inputData, ShouldEqual, ctx2.InputData())
				So(ctx.outputData, ShouldEqual, ctx2.(*bcontext).outputData)
				So(ctx.request.Namespace, ShouldEqual, ctx2.Request().Namespace)
				So(ctx.request.ParentID, ShouldEqual, ctx2.Request().ParentID)
				So(ctx.statusCode, ShouldEqual, ctx2.(*bcontext).statusCode)
				So(ctx.claims, ShouldResemble, ctx2.Claims())
				So(ctx.claimsMap, ShouldResemble, ctx2.ClaimsMap())
				So(ctx.messages, ShouldResemble, ctx2.(*bcontext).messages)
			})
		})
	})
}

func TestContext_GetClaims(t *testing.T) {

	Convey("Given I have a Context with claims", t, func() {

		ctx := newContext(context.TODO(), elemental.NewRequest())
		ctx.SetClaims([]string{"ouais=yes"})

		Convey("When I call GetClaims", func() {

			claims := ctx.Claims()

			Convey("Then claims should be correct", func() {
				So(claims, ShouldResemble, []string{"ouais=yes"})
			})
		})

		Convey("When I call GetClaimsMap", func() {

			claimsMap := ctx.ClaimsMap()

			Convey("Then claims should be correct", func() {
				So(claimsMap, ShouldResemble, map[string]string{"ouais": "yes"})
			})
		})
	})

	Convey("Given I have a Context nil claims", t, func() {

		ctx := newContext(context.TODO(), elemental.NewRequest())
		ctx.SetClaims(nil)

		Convey("When I call GetClaims", func() {

			claims := ctx.Claims()

			Convey("Then claims should be correct", func() {
				So(len(claims), ShouldEqual, 0)
			})
		})

		Convey("When I call GetClaimsMap", func() {

			claimsMap := ctx.ClaimsMap()

			Convey("Then claims should be correct", func() {
				So(claimsMap, ShouldResemble, map[string]string{})
			})
		})
	})
}
