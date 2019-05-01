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
	"context"
	"net/http"
	"net/url"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"go.aporeto.io/elemental"
	testmodel "go.aporeto.io/elemental/test/model"
)

func TestContext_NewContext(t *testing.T) {

	Convey("Given I call newContext", t, func() {

		url, _ := url.Parse("http://link.com/path?page=1&pagesize=10")
		req := &http.Request{
			Host:   "link.com",
			URL:    url,
			Method: http.MethodGet,
		}
		request, _ := elemental.NewRequestFromHTTPRequest(req, testmodel.Manager())

		c := newContext(context.Background(), request)

		Convey("Then it should be correctly initialized", func() {

			So(c.request.Page, ShouldEqual, 1)
			So(c.request.PageSize, ShouldEqual, 10)
			So(c.ctx, ShouldEqual, context.Background())
			So(c, ShouldImplement, (*Context)(nil))
		})
	})

	Convey("Given I call newContext with a nil context", t, func() {

		Convey("Then it should panic", func() {
			So(func() { newContext(nil, nil) }, ShouldPanicWith, "nil context") // nolint
		})
	})

	Convey("Given I call NewContext", t, func() {

		url, _ := url.Parse("http://link.com/path?page=1&pagesize=10")
		req := &http.Request{
			Host:   "link.com",
			URL:    url,
			Method: http.MethodGet,
		}
		request, _ := elemental.NewRequestFromHTTPRequest(req, testmodel.Manager())

		c := NewContext(context.TODO(), request)

		Convey("Then it should be correctly initialized", func() {

			So(c.Request().Page, ShouldEqual, 1)
			So(c.Request().PageSize, ShouldEqual, 10)
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
