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
	"testing"

	// nolint:revive // Allow dot imports for readability in tests
	. "github.com/smartystreets/goconvey/convey"
	"go.aporeto.io/elemental"
	testmodel "go.aporeto.io/elemental/test/model"
)

func TestMockContext_NewMockContext(t *testing.T) {

	Convey("Given I call NewMockContext", t, func() {

		c := NewMockContext(context.Background())

		Convey("Then it should be correctly initialized", func() {
			So(c.MockCtx, ShouldResemble, context.Background())
			So(c.Metadata("hello"), ShouldBeNil)
			So(c, ShouldImplement, (*Context)(nil))
		})
	})
}

func TestMockContext_Identifier(t *testing.T) {

	Convey("Identifier should work", t, func() {
		ctx := NewMockContext(context.Background())
		So(ctx.Identifier(), ShouldNotBeEmpty)
	})
}

func TestMockContext_Events(t *testing.T) {

	Convey("Given I create a Context", t, func() {

		c := NewMockContext(context.Background())

		Convey("When I enqueue 2 events", func() {

			c.EnqueueEvents(
				elemental.NewEvent(elemental.EventCreate, testmodel.NewList()),
				elemental.NewEvent(elemental.EventCreate, testmodel.NewList()),
			)

			Convey("Then I should have 2 events in the queue", func() {
				So(len(c.MockEvents), ShouldEqual, 2)
			})
		})
	})
}

func TestMockContext_Duplicate(t *testing.T) {

	Convey("Given I have a Context, Info, Count, and Page", t, func() {

		req := &elemental.Request{
			Namespace:      "/thens",
			Headers:        http.Header{"header": []string{"h1"}},
			Identity:       elemental.EmptyIdentity,
			ParentID:       "xxxx",
			ParentIdentity: elemental.EmptyIdentity,
			Operation:      elemental.OperationCreate,
		}

		cookies := []*http.Cookie{{}, {}}
		rwriter := func(http.ResponseWriter) int { return 0 }

		ctx := NewMockContext(context.Background())
		ctx.MockRequest = req
		ctx.SetCount(10)
		ctx.SetInputData("input")
		ctx.SetOutputData("output")
		ctx.SetStatusCode(42)
		ctx.AddMessage("a")
		ctx.SetRedirect("laba")
		ctx.AddMessage("b")
		ctx.SetMetadata("hello", "world")
		ctx.SetClaims([]string{"ouais=yes"})
		ctx.SetNext("next")
		ctx.AddOutputCookies(cookies[0], cookies[1])
		ctx.SetResponseWriter(rwriter)
		ctx.SetDisableOutputDataPush(true)

		Convey("When I call the Duplicate method", func() {

			ctx2 := ctx.Duplicate()

			Convey("Then the duplicated context should be correct", func() {
				So(ctx.MockCtx, ShouldResemble, ctx2.Context())
				So(ctx.MockCount, ShouldEqual, ctx2.Count())
				So(ctx.Metadata("hello").(string), ShouldEqual, "world")
				So(ctx.MockInputData, ShouldEqual, ctx2.InputData())
				So(ctx.MockOutputData, ShouldEqual, ctx2.OutputData())
				So(ctx.MockRequest.Namespace, ShouldEqual, ctx2.Request().Namespace)
				So(ctx.MockRequest.ParentID, ShouldEqual, ctx2.Request().ParentID)
				So(ctx.MockStatusCode, ShouldEqual, ctx2.StatusCode())
				So(ctx.MockClaims, ShouldResemble, ctx2.Claims())
				So(ctx.MockClaimsMap, ShouldResemble, ctx2.ClaimsMap())
				So(ctx.MockRedirect, ShouldEqual, ctx2.Redirect())
				So(ctx.MockNext, ShouldEqual, ctx2.(*MockContext).MockNext)
				So(ctx.MockMessages, ShouldResemble, ctx2.(*MockContext).MockMessages)
				So(ctx.MockOutputCookies, ShouldResemble, ctx2.(*MockContext).MockOutputCookies)
				So(ctx.MockOutputCookies, ShouldResemble, cookies)
				So(ctx.MockResponseWriter, ShouldEqual, rwriter)
				So(ctx.MockDisableOutputDataPush, ShouldEqual, ctx.MockDisableOutputDataPush)
			})
		})
	})
}

func TestMockContext_GetClaims(t *testing.T) {

	Convey("Given I have a Context with claims", t, func() {

		oc := []string{"ouais=yes"}

		ctx := NewMockContext(context.Background())
		ctx.SetClaims(oc)

		Convey("When I call GetClaims", func() {

			claims := ctx.Claims()

			Convey("Then claims should be correct", func() {
				So(claims, ShouldResemble, oc)
				So(claims, ShouldNotEqual, oc)
			})
		})

		Convey("When I call GetClaimsMap", func() {

			claimsMap := ctx.ClaimsMap()

			Convey("Then claims should be correct", func() {
				So(claimsMap, ShouldResemble, map[string]string{"ouais": "yes"})
				So(claimsMap, ShouldNotEqual, ctx.MockClaimsMap)
			})
		})
	})

	Convey("Given I have a Context nil claims", t, func() {

		ctx := NewMockContext(context.Background())
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
