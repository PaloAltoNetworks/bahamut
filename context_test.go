// Author: Antoine Mercadal
// See LICENSE file for full LICENSE
// Copyright 2016 Aporeto.

package bahamut

import (
	"context"
	"net/http"
	"net/url"
	"testing"

	"go.aporeto.io/elemental"
	"go.aporeto.io/elemental/test/model"

	. "github.com/smartystreets/goconvey/convey"
)

func TestContext_NewContext(t *testing.T) {

	Convey("Given I call newContext", t, func() {

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
			So(c.ctx, ShouldEqual, context.TODO())
			So(c, ShouldImplement, (*Context)(nil))
		})
	})

	Convey("Given I call newContext with a nil context", t, func() {

		Convey("Then it should panic", func() {
			So(func() { newContext(nil, nil) }, ShouldPanicWith, "nil context") // nolint
		})
	})

	Convey("Given I call NewContext", t, func() {

		url, _ := url.Parse("http://link.com/path?page=1&per_page=10")
		req := &http.Request{
			Host:   "link.com",
			URL:    url,
			Method: http.MethodGet,
		}
		request, _ := elemental.NewRequestFromHTTPRequest(req, testmodel.Manager())

		c := NewContext(context.TODO(), request)

		Convey("Then it should be correctly initialized", func() {

			So(c.Request().Parameters.Get("page"), ShouldEqual, "1")
			So(c.Request().Parameters.Get("per_page"), ShouldEqual, "10")
			So(c.Context(), ShouldEqual, context.TODO())
			So(c.Metadata("hello"), ShouldBeNil)
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

		ctx := newContext(context.TODO(), req)
		ctx.SetCount(10)
		ctx.SetInputData("input")
		ctx.SetInputData("output")
		ctx.SetStatusCode(42)
		ctx.AddMessage("a")
		ctx.SetRedirect("laba")
		ctx.AddMessage("b")
		ctx.SetMetadata("hello", "world")
		ctx.SetClaims([]string{"ouais=yes"})

		Convey("When I call the Duplicate method", func() {

			ctx2 := ctx.Duplicate()

			Convey("Then the duplicated context should be correct", func() {
				So(ctx.count, ShouldEqual, ctx2.Count())
				So(ctx.Metadata("hello").(string), ShouldEqual, "world")
				So(ctx.inputData, ShouldEqual, ctx2.InputData())
				So(ctx.outputData, ShouldEqual, ctx2.OutputData())
				So(ctx.request.Namespace, ShouldEqual, ctx2.Request().Namespace)
				So(ctx.request.ParentID, ShouldEqual, ctx2.Request().ParentID)
				So(ctx.statusCode, ShouldEqual, ctx2.StatusCode())
				So(ctx.claims, ShouldResemble, ctx2.Claims())
				So(ctx.claimsMap, ShouldResemble, ctx2.ClaimsMap())
				So(ctx.redirect, ShouldResemble, ctx2.Redirect())
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
