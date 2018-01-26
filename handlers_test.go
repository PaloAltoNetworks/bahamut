package bahamut

import (
	"context"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/aporeto-inc/elemental/test/model"

	"github.com/aporeto-inc/elemental"
	. "github.com/smartystreets/goconvey/convey"
)

func TestHandlers_runDispatcher(t *testing.T) {

	Convey("Given I have a fake dispatcher", t, func() {

		called := 0

		gctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
		defer cancel()

		ctx := NewContext()
		ctx.Request = elemental.NewRequestWithContext(gctx)

		response := elemental.NewResponse(gctx)

		Convey("When I call runDispatcher", func() {

			d := func() error {
				called++
				return nil
			}

			r := runDispatcher(ctx, response, d, true)

			Convey("Then the code should be 204", func() {
				So(r.StatusCode, ShouldEqual, 204)
			})

			Convey("Then the dispatcher should have been called once", func() {
				So(called, ShouldEqual, 1)
			})
		})

		Convey("When I call runDispatcher and it returns an error", func() {

			d := func() error {
				called++
				return elemental.NewError("nop", "nope", "test", 42)
			}

			r := runDispatcher(ctx, response, d, true)

			Convey("Then the code should be 42", func() {
				So(r.StatusCode, ShouldEqual, 42)
			})

			Convey("Then the dispatcher should have been called once", func() {
				So(called, ShouldEqual, 1)
			})
		})

		Convey("When I call runDispatcher and cancel the context", func() {

			l := &sync.Mutex{}

			d := func() error {
				time.Sleep(2 * time.Second)
				l.Lock()
				called++
				l.Unlock()
				return nil
			}

			go func() { runDispatcher(ctx, nil, d, true) }()
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

func TestHandlers_makeResponse(t *testing.T) {

	Convey("Given I have context with a redirect and a response", t, func() {

		ctx := NewContext()
		response := elemental.NewResponse(context.TODO())

		ctx.Redirect = "http://ici"

		Convey("When I call makeResponse", func() {

			makeResponse(ctx, response)

			Convey("Then response.Redirect should be set", func() {
				So(response.Redirect, ShouldEqual, "http://ici")
			})
		})
	})

	Convey("Given I have context with a a count and a response", t, func() {

		ctx := NewContext()
		response := elemental.NewResponse(context.TODO())

		ctx.CountTotal = 42

		Convey("When I call makeResponse on a retrieveMany operation", func() {

			ctx.Request.Operation = elemental.OperationRetrieveMany

			makeResponse(ctx, response)

			Convey("Then response.Count should be set", func() {
				So(response.Count, ShouldEqual, 42)
			})
		})

		Convey("When I call makeResponse on a info operation", func() {

			ctx.Request.Operation = elemental.OperationInfo

			makeResponse(ctx, response)

			Convey("Then response.Count should be set", func() {
				So(response.Count, ShouldEqual, 42)
			})
		})

		Convey("When I call makeResponse on a create operation", func() {

			ctx.Request.Operation = elemental.OperationCreate

			makeResponse(ctx, response)

			Convey("Then response.Count should not be set", func() {
				So(response.Count, ShouldEqual, 0)
			})
		})

		Convey("When I call makeResponse on a create update", func() {

			ctx.Request.Operation = elemental.OperationUpdate

			makeResponse(ctx, response)

			Convey("Then response.Count should not be set", func() {
				So(response.Count, ShouldEqual, 0)
			})
		})

		Convey("When I call makeResponse on a create delete", func() {

			ctx.Request.Operation = elemental.OperationDelete

			makeResponse(ctx, response)

			Convey("Then response.Count should not be set", func() {
				So(response.Count, ShouldEqual, 0)
			})
		})

		Convey("When I call makeResponse on a create patch", func() {

			ctx.Request.Operation = elemental.OperationPatch

			makeResponse(ctx, response)

			Convey("Then response.Count should not be set", func() {
				So(response.Count, ShouldEqual, 0)
			})
		})
	})

	Convey("Given I have context with a status code set to 0 and a response", t, func() {

		ctx := NewContext()
		response := elemental.NewResponse(context.TODO())
		ctx.OutputData = []string{}
		ctx.StatusCode = 0

		Convey("When I set the operation to Create and I call makeResponse", func() {

			ctx.Request.Operation = elemental.OperationCreate

			makeResponse(ctx, response)

			Convey("Then response.StatusCode should equal", func() {
				So(response.StatusCode, ShouldEqual, http.StatusCreated)
			})
		})

		Convey("When I set the operation to Info and I call makeResponse", func() {

			ctx.Request.Operation = elemental.OperationInfo

			makeResponse(ctx, response)

			Convey("Then response.StatusCode should equal", func() {
				So(response.StatusCode, ShouldEqual, http.StatusNoContent)
			})
		})

		Convey("When I set the operation to Retrieve and I call makeResponse", func() {

			ctx.Request.Operation = elemental.OperationRetrieve

			makeResponse(ctx, response)

			Convey("Then response.StatusCode should equal", func() {
				So(response.StatusCode, ShouldEqual, http.StatusOK)
			})
		})

		Convey("When I set the operation to Create, status code created, but no data, and I call makeResponse", func() {

			ctx.Request.Operation = elemental.OperationCreate
			ctx.StatusCode = http.StatusCreated
			ctx.OutputData = nil

			makeResponse(ctx, response)

			Convey("Then response.StatusCode should equal", func() {
				So(response.StatusCode, ShouldEqual, http.StatusNoContent)
			})
		})
	})

	Convey("Given I have context with messages and a response", t, func() {

		ctx := NewContext()
		response := elemental.NewResponse(context.TODO())
		ctx.AddMessage("hello world")

		Convey("When I call makeResponse", func() {

			makeResponse(ctx, response)

			Convey("Then response.Message should be set", func() {
				So(response.Messages, ShouldResemble, []string{"hello world"})
			})
		})
	})

	Convey("Given I have context with unmarshalable data and a response", t, func() {

		ctx := NewContext()
		response := elemental.NewResponse(context.TODO())
		ctx.OutputData = testmodel.NewUnmarshalableList()

		Convey("When I call makeResponse", func() {

			Convey("Then it should panic", func() {
				So(func() { makeResponse(ctx, response) }, ShouldPanic)
			})
		})
	})
}
