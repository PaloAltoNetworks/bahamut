package bahamut

import (
	"context"
	"errors"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/aporeto-inc/elemental"

	. "github.com/smartystreets/goconvey/convey"
)

func TestWebsocketServerHelper_writeWebSocketError(t *testing.T) {

	Convey("Given I have response and an error", t, func() {

		err := errors.New("oops")
		resp := elemental.NewResponse()
		resp.Request = elemental.NewRequest()

		Convey("When I call writeWebSocketError", func() {
			r := writeWebSocketError(resp, err)

			Convey("Then status should be 500", func() {
				So(r.StatusCode, ShouldEqual, http.StatusInternalServerError)
			})
		})
	})
}

func TestWebsocketServerHelper_writeWebsocketResponse(t *testing.T) {

	Convey("Given I have response and a context", t, func() {

		resp := elemental.NewResponse()
		ctx := NewContext()
		ctx.Request = elemental.NewRequest()

		Convey("When I call writeWebsocketResponse using a context with no status code no data", func() {

			r := writeWebsocketResponse(resp, ctx)

			Convey("Then status should be 500", func() {
				So(r.StatusCode, ShouldEqual, http.StatusNoContent)
			})
		})

		Convey("When I call writeWebsocketResponse using a context with no status code and created data", func() {

			ctx.OutputData = &FakeIdentifiable{}
			ctx.Request.Operation = elemental.OperationCreate

			r := writeWebsocketResponse(resp, ctx)

			Convey("Then status should be 201", func() {
				So(r.StatusCode, ShouldEqual, http.StatusCreated)
			})
		})

		Convey("When I call writeWebsocketResponse using a context with no status code and no created data", func() {

			ctx.Request.Operation = elemental.OperationCreate

			r := writeWebsocketResponse(resp, ctx)

			Convey("Then status should be 201", func() {
				So(r.StatusCode, ShouldEqual, http.StatusNoContent)
			})
		})

		Convey("When I call writeWebsocketResponse for operation retrieve-many with a count of 5", func() {

			ctx.Request.Operation = elemental.OperationRetrieveMany
			ctx.CountTotal = 5

			r := writeWebsocketResponse(resp, ctx)

			Convey("Then status should be 201", func() {
				So(r.Total, ShouldEqual, 5)
			})
		})

		Convey("When I call writeWebsocketResponse for operation info with a count of 5", func() {

			ctx.Request.Operation = elemental.OperationInfo
			ctx.CountTotal = 5

			r := writeWebsocketResponse(resp, ctx)

			Convey("Then status should be 201", func() {
				So(r.Total, ShouldEqual, 5)
			})
		})

		Convey("When I call writeWebsocketResponse for operation create with output data", func() {

			ctx.OutputData = &List{ID: "a"}
			ctx.Request.Operation = elemental.OperationCreate

			r := writeWebsocketResponse(resp, ctx)

			Convey("Then status should be 201", func() {
				So(r.StatusCode, ShouldEqual, http.StatusCreated)
			})
			Convey("Then data should be correct", func() {
				So(string(r.Data), ShouldEqual, `{"ID":"a","creationOnly":"","description":"","name":"","parentID":"","parentType":"","readOnly":""}`)
			})
		})

		Convey("When I call writeWebsocketResponse for operation with messages", func() {

			ctx.AddMessage("a")
			ctx.AddMessage("b")

			r := writeWebsocketResponse(resp, ctx)

			Convey("Then messages should be correct", func() {
				So(r.Messages, ShouldResemble, []string{"a", "b"})
			})
		})

		Convey("When I call writeWebsocketResponse for operation with unmarshalable identity", func() {

			ctx.OutputData = &UnmarshalableList{}
			ctx.Request.Operation = elemental.OperationCreate

			Convey("Then it should panic", func() {
				So(func() { writeWebsocketResponse(resp, ctx) }, ShouldPanic) // nolint
			})
		})
	})
}

func TestWebsocketServerHelpers_runWSDispatcher(t *testing.T) {

	Convey("Given I have a fake dispatcher", t, func() {

		called := 0

		gctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
		defer cancel()

		ctx := NewContext()
		ctx.Request = elemental.NewRequestWithContext(gctx)

		resp := elemental.NewResponse()
		resp.Request = elemental.NewRequest()

		Convey("When I call runWSDispatcher", func() {

			d := func() error {
				called++
				return nil
			}

			r := runWSDispatcher(ctx, resp, d, true)

			Convey("Then the code should be 204", func() {
				So(r.StatusCode, ShouldEqual, 204)
			})

			Convey("Then the dispatcher should have been called once", func() {
				So(called, ShouldEqual, 1)
			})
		})

		Convey("When I call runWSDispatcher and it returns an error", func() {

			d := func() error {
				called++
				return elemental.NewError("nop", "nope", "test", 42)
			}

			r := runWSDispatcher(ctx, resp, d, true)

			Convey("Then the code should be 42", func() {
				So(r.StatusCode, ShouldEqual, 42)
			})

			Convey("Then the dispatcher should have been called once", func() {
				So(called, ShouldEqual, 1)
			})
		})

		Convey("When I call runWSDispatcher and cancel the context", func() {

			l := &sync.Mutex{}

			d := func() error {
				time.Sleep(2 * time.Second)
				l.Lock()
				called++
				l.Unlock()
				return nil
			}

			go func() { runWSDispatcher(ctx, resp, d, true) }()
			time.Sleep(300 * time.Millisecond)
			cancel()

			Convey("Then the dispatcher should have been called once", func() {
				l.Lock()
				So(called, ShouldEqual, 0)
				l.Unlock()
			})
		})
	})
}
