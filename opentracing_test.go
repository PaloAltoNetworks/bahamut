package bahamut

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"github.com/aporeto-inc/elemental"
	"github.com/aporeto-inc/elemental/test/model"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
	. "github.com/smartystreets/goconvey/convey"
)

type testSpanContext struct {
}

func (t *testSpanContext) ForeachBaggageItem(handler func(k, v string) bool) {}

type testTracer struct {
	currentSpan *testSpan
}

func (t *testTracer) StartSpan(string, ...opentracing.StartSpanOption) opentracing.Span {
	if t.currentSpan == nil {
		t.currentSpan = newTestSpan(t)
	}
	return t.currentSpan
}
func (t *testTracer) Inject(opentracing.SpanContext, interface{}, interface{}) error { return nil }
func (t *testTracer) Extract(interface{}, interface{}) (opentracing.SpanContext, error) {
	return &testSpanContext{}, nil
}

type testSpan struct {
	finished bool
	tracer   opentracing.Tracer
	tags     map[string]interface{}
	fields   []log.Field
}

func newTestSpan(tracer opentracing.Tracer) *testSpan {
	return &testSpan{
		tracer: tracer,
		tags:   map[string]interface{}{},
		fields: []log.Field{},
	}
}
func (s *testSpan) Finish()                                                { s.finished = true }
func (s *testSpan) FinishWithOptions(opts opentracing.FinishOptions)       { s.finished = true }
func (s *testSpan) Context() opentracing.SpanContext                       { return &testSpanContext{} }
func (s *testSpan) SetOperationName(operationName string) opentracing.Span { return s }
func (s *testSpan) SetTag(key string, value interface{}) opentracing.Span {
	s.tags[key] = value
	return s
}
func (s *testSpan) LogFields(fields ...log.Field) {
	s.fields = append(s.fields, fields...)
}
func (s *testSpan) LogKV(alternatingKeyValues ...interface{})                   {}
func (s *testSpan) SetBaggageItem(restrictedKey, value string) opentracing.Span { return s }
func (s *testSpan) BaggageItem(restrictedKey string) string                     { return "" }
func (s *testSpan) Tracer() opentracing.Tracer                                  { return s.tracer }
func (s *testSpan) LogEvent(event string)                                       {}
func (s *testSpan) LogEventWithPayload(event string, payload interface{})       {}
func (s *testSpan) Log(data opentracing.LogData)                                {}

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

		req, _ := http.NewRequest(http.MethodGet, "http://server/lists/xx/tasks?p=v", nil)
		req.Header.Add("X-Namespace", "ns")
		req.Header.Add("Authorization", "Bearer "+token)
		r, _ := elemental.NewRequestFromHTTPRequest(req)

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
				So(claims, ShouldEqual, `{}`)
			})
		})
	})

	Convey("Given I have a Request with almost invalid token in Password", t, func() {

		req := elemental.NewRequest()
		req.Password = "a.b.c"

		Convey("When I extract the claims", func() {

			claims := extractClaims(req)

			Convey("Then claims should be correct", func() {
				So(claims, ShouldEqual, `{}`)
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

		req := elemental.NewRequest()
		req.Parameters = url.Values{
			"token": {"1", "2"},
		}
		req.Headers = http.Header{
			"authorization": {"3", "4"},
		}
		req.Version = 2
		req.Identity = testmodel.UserIdentity
		req.Recursive = true
		req.Operation = elemental.OperationCreate
		req.OverrideProtection = true
		req.ExternalTrackingID = "wee"
		req.ExternalTrackingType = "yeah"
		req.Namespace = "/a"
		req.ObjectID = "id"
		req.ParentID = "pid"
		req.ParentIdentity = testmodel.TaskIdentity
		req.Page = 3
		req.PageSize = 30
		req.ClientIP = "127.0.0.1"
		req.Order = []string{"a", "b"}
		req.Data = []byte("the data")

		tracer := &testTracer{}
		ts := newTestSpan(tracer)

		ctx := opentracing.ContextWithSpan(context.Background(), ts)

		Convey("When I call traceRequest with no tracer", func() {
			tctx := traceRequest(ctx, req, nil)

			Convey("Then the returned context should should be the same", func() {
				So(tctx, ShouldEqual, ctx)
			})
		})

		Convey("When I call traceRequest", func() {

			tctx := traceRequest(ctx, req, tracer)

			span := opentracing.SpanFromContext(tctx).(*testSpan)

			Convey("Then the new context should be spanned", func() {
				So(span, ShouldNotBeNil)
			})

			Convey("Then the span fields should be correct", func() {
				So(len(span.fields), ShouldEqual, 8)
				So(span.fields[0].String(), ShouldEqual, "req.page.number:3")
				So(span.fields[1].String(), ShouldEqual, "req.page.size:30")
				So(span.fields[2].String(), ShouldEqual, fmt.Sprintf("req.headers:map[authorization:%s]", snipSlice))
				So(span.fields[3].String(), ShouldEqual, "req.claims:{}")
				So(span.fields[4].String(), ShouldEqual, "req.client_ip:127.0.0.1")
				So(span.fields[5].String(), ShouldEqual, fmt.Sprintf("req.parameters:map[token:%s]", snipSlice))
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
				So(span.tags["req.object.id"], ShouldEqual, "id")
				So(span.tags["req.api_version"], ShouldEqual, 2)
				So(span.tags["req.identity"], ShouldEqual, "user")
				So(span.tags["req.operation"], ShouldEqual, "create")
				So(span.tags["req.override_protection"], ShouldBeTrue)
				So(span.tags["req.parent.id"], ShouldEqual, "pid")
			})
		})
	})
}

func TestTracing_finishTracing(t *testing.T) {

	Convey("Given I have a context with a span", t, func() {

		tracer := &testTracer{}
		ts := newTestSpan(tracer)

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
