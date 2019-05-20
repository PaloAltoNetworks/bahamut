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
	"bytes"
	"context"
	"net/http"
	"testing"

	opentracing "github.com/opentracing/opentracing-go"
	. "github.com/smartystreets/goconvey/convey"
	"go.aporeto.io/elemental"
	testmodel "go.aporeto.io/elemental/test/model"
)

func TestTracing_extractClaims(t *testing.T) {

	token := "x.eyJyZWFsbSI6IkNlcnRpZmljYXRlIiwiZGF0YSI6eyJjb21tb25OYW1lIjoiYWRtaW4iLCJvcmdhbml6YXRpb24iOiJzeXN0ZW0iLCJvdTpyb290IjoidHJ1ZSIsInJlYWxtIjoiY2VydGlmaWNhdGUiLCJzZXJpYWxOdW1iZXIiOiIxODY3OTg0MjcyNDEzNDMwODM2NzY2MDU2NTk0NDg1NjUxNTk4MTcifSwiYXVkIjoiYXBvcmV0by5jb20iLCJleHAiOjE1MDg1MTYxMzEsImlhdCI6MTUwODQyOTczMSwiaXNzIjoibWlkZ2FyZC5hcG9tdXguY29tIiwic3ViIjoiMTg2Nzk4NDI3MjQxMzQzMDgzNjc2NjA1NjU5NDQ4NTY1MTU5ODE3In0.y"
	tokenInavalid := "eyJyZWFsbSI6IkNlcnRpZmljYXRlIiwiZGF0YSI6eyJjb21tb25OYW1lIjoiYWRtaW4iLCJvcmdhbml6YXRpb24iOiJzeXN0ZW0iLCJvdTpyb290IjoidHJ1ZSIsInJlYWxtIjoiY2VydGlmaWNhdGUiLCJzZXJpYWxOdW1iZXIiOiIxODY3OTg0MjcyNDEzNDMwODM2NzY2MDU2NTk0NDg1NjUxNTk4MTcifSwiYXVkIjoiYXBvcmV0by5jb20iLCJleHAiOjE1MDg1MTYxMzEsImlhdCI6MTUwODQyOTczMSwiaXNzIjoibWlkZ2FyZC5hcG9tdXguY29tIiwic3ViIjoiMTg2Nzk4NDI3MjQxMzQzMDgzNjc2NjA1NjU5NDQ4NTY1MTU5ODE3In0.y"

	Convey("Given I have a Request with Password", t, func() {

		req := elemental.NewRequest()
		req.Password = token

		Convey("When I extract the claims", func() {

			claims := extractClaims(req)

			Convey("Then claims should be correct", func() {
				So(claims, ShouldEqual, `{"realm":"Certificate","data":{"commonName":"admin","organization":"system","ou:root":"true","realm":"certificate","serialNumber":"186798427241343083676605659448565159817"},"aud":"aporeto.com","exp":1508516131,"iat":1508429731,"iss":"midgard.apomux.com","sub":"186798427241343083676605659448565159817"}`)
			})
		})
	})

	Convey("Given create a request from an http request", t, func() {

		req, _ := http.NewRequest(http.MethodGet, "http://server/lists/xx/tasks", nil)
		req.Header.Add("X-Namespace", "ns")
		req.Header.Add("Authorization", "Bearer "+token)
		req.Header.Add("Content-Type", "application/json")
		req.Header.Add("Accept", "application/json")
		r, err := elemental.NewRequestFromHTTPRequest(req, testmodel.Manager())
		if err != nil {
			panic(err)
		}

		Convey("When I extract the claims", func() {

			claims := extractClaims(r)

			Convey("Then claims should be correct", func() {
				So(claims, ShouldEqual, `{"realm":"Certificate","data":{"commonName":"admin","organization":"system","ou:root":"true","realm":"certificate","serialNumber":"186798427241343083676605659448565159817"},"aud":"aporeto.com","exp":1508516131,"iat":1508429731,"iss":"midgard.apomux.com","sub":"186798427241343083676605659448565159817"}`)
			})
		})
	})

	Convey("Given I have a Request with invalid token in Password", t, func() {

		req := elemental.NewRequest()
		req.Password = tokenInavalid

		Convey("When I extract the claims", func() {

			claims := extractClaims(req)

			Convey("Then claims should be correct", func() {
				So(claims, ShouldEqual, `invalid token format: `+tokenInavalid)
			})
		})
	})

	Convey("Given I have a Request with almost invalid token in Password", t, func() {

		req := elemental.NewRequest()
		req.Password = "a.b.c"

		Convey("When I extract the claims", func() {

			claims := extractClaims(req)

			Convey("Then claims should be correct", func() {
				So(claims, ShouldEqual, `invalid token encoding: a.b.c: illegal base64 data at input byte 0`)
			})
		})
	})
}

func TestTracing_tracingName(t *testing.T) {

	Convey("Given I have a create request on some identity", t, func() {

		req := elemental.NewRequest()
		req.Identity = testmodel.ListIdentity

		Convey("When I call tracingName for operation create", func() {

			req.Operation = elemental.OperationCreate
			name := tracingName(req)

			Convey("Then name should correct", func() {
				So(name, ShouldEqual, "bahamut.handle.create.lists")
			})
		})

		Convey("When I call tracingName for operation update", func() {

			req.Operation = elemental.OperationUpdate
			name := tracingName(req)

			Convey("Then name should correct", func() {
				So(name, ShouldEqual, "bahamut.handle.update.lists")
			})
		})

		Convey("When I call tracingName for operation delete", func() {

			req.Operation = elemental.OperationDelete
			name := tracingName(req)

			Convey("Then name should correct", func() {
				So(name, ShouldEqual, "bahamut.handle.delete.lists")
			})
		})

		Convey("When I call tracingName for operation info", func() {

			req.Operation = elemental.OperationInfo
			name := tracingName(req)

			Convey("Then name should correct", func() {
				So(name, ShouldEqual, "bahamut.handle.info.lists")
			})
		})

		Convey("When I call tracingName for operation retrieve", func() {

			req.Operation = elemental.OperationRetrieve
			name := tracingName(req)

			Convey("Then name should correct", func() {
				So(name, ShouldEqual, "bahamut.handle.retrieve.lists")
			})
		})

		Convey("When I call tracingName for operation retrieve many", func() {

			req.Operation = elemental.OperationRetrieveMany
			name := tracingName(req)

			Convey("Then name should correct", func() {
				So(name, ShouldEqual, "bahamut.handle.retrieve_many.lists")
			})
		})

		Convey("When I call tracingName for operation patch", func() {

			req.Operation = elemental.OperationPatch
			name := tracingName(req)

			Convey("Then name should correct", func() {
				So(name, ShouldEqual, "bahamut.handle.patch.lists")
			})
		})

		Convey("When I call tracingName for operation unknown", func() {

			req.Operation = elemental.Operation("nope")
			name := tracingName(req)

			Convey("Then name should correct", func() {
				So(name, ShouldEqual, "Unknown operation: nope")
			})
		})
	})

}

func TestTracing_traceRequest(t *testing.T) {

	Convey("Given I have a request", t, func() {

		buf := bytes.NewBuffer([]byte("the data"))
		hreq, err := http.NewRequest("POST", "https://toto.com/v/2/tasks/pid/users?recursive=true&override=true&page=3&pagesize=30&order=a&order=b", buf)
		if err != nil {
			panic(err)
		}
		hreq.Header.Add("Authorization", "secretA")
		hreq.Header.Add("Authorization", "secretB")
		hreq.Header.Add("NotAuthorization", "notSecretA")

		req, err := elemental.NewRequestFromHTTPRequest(hreq, testmodel.Manager())
		if err != nil {
			panic(err)
		}
		req.ExternalTrackingID = "wee"
		req.ExternalTrackingType = "yeah"
		req.ClientIP = "127.0.0.1"
		req.Namespace = "/a"
		req.ObjectID = "id"
		// Add the param after calling NewRequestFromHTTPRequest as this is not valid params from specs.
		req.Parameters["token"] = elemental.NewParameter(elemental.ParameterTypeString, "1", "2")
		req.Parameters["not-token"] = elemental.NewParameter(elemental.ParameterTypeString, "notSecretB")

		tracer := &mockTracer{}
		ts := newMockSpan(tracer)

		ctx := opentracing.ContextWithSpan(context.Background(), ts)

		Convey("When I call traceRequest with no tracer", func() {
			tctx := traceRequest(ctx, req, nil, nil, nil)

			Convey("Then the returned context should should be the same", func() {
				So(tctx, ShouldEqual, ctx)
			})
		})

		Convey("When I call traceRequest on excluded identities", func() {
			tctx := traceRequest(ctx, req, tracer, map[string]struct{}{"user": struct{}{}}, nil)

			Convey("Then the returned context should should be the same", func() {
				So(tctx, ShouldEqual, ctx)
			})
		})

		Convey("When I call traceRequest", func() {

			tctx := traceRequest(ctx, req, tracer, map[string]struct{}{"not-user": struct{}{}}, nil)

			span := opentracing.SpanFromContext(tctx).(*mockSpan)

			Convey("Then the new context should be spanned", func() {
				So(span, ShouldNotBeNil)
			})

			Convey("Then the span fields should be correct", func() {
				So(len(span.fields), ShouldEqual, 8)
				So(span.fields[0].String(), ShouldEqual, "req.page.number:3")
				So(span.fields[1].String(), ShouldEqual, "req.page.size:30")
				So(span.fields[2].String(), ShouldContainSubstring, "Notauthorization:[notSecretA]")
				So(span.fields[2].String(), ShouldContainSubstring, "Authorization:[[snip]]")
				So(span.fields[3].String(), ShouldEqual, "req.claims:{}")
				So(span.fields[4].String(), ShouldEqual, "req.client_ip:127.0.0.1")
				So(span.fields[5].String(), ShouldNotContainSubstring, "secretA")
				So(span.fields[5].String(), ShouldNotContainSubstring, "secretB")
				So(span.fields[5].String(), ShouldContainSubstring, "[[snip]]")
				So(span.fields[5].String(), ShouldContainSubstring, "notSecretB")
				So(span.fields[6].String(), ShouldEqual, "req.order_by:[a b]")
				So(span.fields[7].String(), ShouldEqual, "req.payload:the data")
			})

			Convey("Then the span tags should be correct", func() {
				So(len(span.tags), ShouldEqual, 12)
				So(span.tags["req.parent.identity"], ShouldEqual, "task")
				So(span.tags["req.id"], ShouldEqual, req.RequestID)
				So(span.tags["req.recursive"], ShouldBeTrue)
				So(span.tags["req.external_tracking_id"], ShouldEqual, "wee")
				So(span.tags["req.external_tracking_type"], ShouldEqual, "yeah")
				So(span.tags["req.namespace"], ShouldEqual, "/a")
				So(span.tags["req.api_version"], ShouldEqual, 2)
				So(span.tags["req.identity"], ShouldEqual, "user")
				So(span.tags["req.operation"], ShouldEqual, "create")
				So(span.tags["req.override_protection"], ShouldBeTrue)
				So(span.tags["req.parent.id"], ShouldEqual, "pid")
				So(span.tags["req.object.id"], ShouldEqual, "id")
			})
		})

		Convey("When I call traceRequest with a cleaner", func() {

			var expectedIdentity elemental.Identity
			tctx := traceRequest(ctx, req, tracer, nil, func(i elemental.Identity, data []byte) []byte {
				expectedIdentity = i
				return []byte("modified data")
			})

			span := opentracing.SpanFromContext(tctx).(*mockSpan)

			Convey("Then the new context should be spanned", func() {
				So(span, ShouldNotBeNil)
			})

			Convey("Then the cleaner should have been called with the correct identity", func() {
				So(expectedIdentity.Name, ShouldEqual, "user")
			})

			Convey("Then the span data should have be cleaned", func() {
				So(span.fields[7].String(), ShouldEqual, "req.payload:modified data")
			})

			Convey("Then original request data should not have changed", func() {
				So(string(req.Data), ShouldEqual, "the data")
			})
		})
	})
}

func TestTracing_finishTracing(t *testing.T) {

	Convey("Given I have a context with a span", t, func() {

		tracer := &mockTracer{}
		ts := newMockSpan(tracer)

		ctx := opentracing.ContextWithSpan(context.Background(), ts)

		Convey("When I call finishTracing", func() {

			finishTracing(ctx)

			Convey("Then my span should be finished", func() {
				So(ts.finished, ShouldBeTrue)
			})
		})
	})

	Convey("Given I have a context with no span", t, func() {

		ctx := context.Background()

		Convey("When I call finishTracing", func() {

			Convey("Then it should not panic", func() {
				So(func() { finishTracing(ctx) }, ShouldNotPanic)
			})
		})
	})
}
