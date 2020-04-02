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
	"fmt"
	"net/http"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
	"go.aporeto.io/elemental"
	testmodel "go.aporeto.io/elemental/test/model"
)

func TestHandlers_makeResponse(t *testing.T) {

	Convey("Given I have context with a redirect and a response", t, func() {

		ctx := newContext(context.Background(), elemental.NewRequest())
		response := elemental.NewResponse(elemental.NewRequest())

		ctx.redirect = "http://ici"

		Convey("When I call makeResponse", func() {

			makeResponse(ctx, response, nil)

			Convey("Then response.Redirect should be set", func() {
				So(response.Redirect, ShouldEqual, "http://ici")
			})
		})
	})

	Convey("Given I have context with a a count and a response", t, func() {

		ctx := newContext(context.Background(), elemental.NewRequest())
		response := elemental.NewResponse(elemental.NewRequest())

		ctx.count = 42

		Convey("When I call makeResponse on a retrieveMany operation", func() {

			ctx.request.Operation = elemental.OperationRetrieveMany

			makeResponse(ctx, response, nil)

			Convey("Then response.Total should be set", func() {
				So(response.Total, ShouldEqual, 42)
			})
		})

		Convey("When I call makeResponse on a info operation", func() {

			ctx.request.Operation = elemental.OperationInfo

			makeResponse(ctx, response, nil)

			Convey("Then response.Total should be set", func() {
				So(response.Total, ShouldEqual, 42)
			})
		})

		Convey("When I call makeResponse on with a cookie", func() {

			ctx.request.Operation = elemental.OperationInfo
			ctx.outputCookies = []*http.Cookie{
				{
					Name: "a",
				},
				{
					Name: "b",
				},
			}

			makeResponse(ctx, response, nil)

			Convey("Then response.Cookie should be set", func() {
				So(len(response.Cookies), ShouldEqual, 2)
			})
		})

		Convey("When I call makeResponse on a create operation", func() {

			ctx.request.Operation = elemental.OperationCreate

			makeResponse(ctx, response, nil)

			Convey("Then response.Total should not be set", func() {
				So(response.Total, ShouldEqual, 0)
			})
		})

		Convey("When I call makeResponse on a update operation", func() {

			ctx.request.Operation = elemental.OperationUpdate

			makeResponse(ctx, response, nil)

			Convey("Then response.Total should not be set", func() {
				So(response.Total, ShouldEqual, 0)
			})
		})

		Convey("When I call makeResponse on a delete operation", func() {

			ctx.request.Operation = elemental.OperationDelete

			makeResponse(ctx, response, nil)

			Convey("Then response.Total should not be set", func() {
				So(response.Total, ShouldEqual, 0)
			})
		})

		Convey("When I call makeResponse on a patch operation", func() {

			ctx.request.Operation = elemental.OperationPatch

			makeResponse(ctx, response, nil)

			Convey("Then response.Total should not be set", func() {
				So(response.Total, ShouldEqual, 0)
			})
		})
	})

	Convey("Given I have context with a status code set to 0 and a response", t, func() {

		ctx := newContext(context.Background(), elemental.NewRequest())
		response := elemental.NewResponse(elemental.NewRequest())
		ctx.outputData = []string{}
		ctx.statusCode = 0

		Convey("When I set the operation to Create and I call makeResponse", func() {

			ctx.request.Operation = elemental.OperationCreate

			makeResponse(ctx, response, nil)

			Convey("Then response.StatusCode should be http.StatusOK", func() {
				So(response.StatusCode, ShouldEqual, http.StatusOK)
			})
		})

		Convey("When I set the operation to Info and I call makeResponse", func() {

			ctx.request.Operation = elemental.OperationInfo

			makeResponse(ctx, response, nil)

			Convey("Then response.StatusCode should be http.StatusNoContent", func() {
				So(response.StatusCode, ShouldEqual, http.StatusNoContent)
			})
		})

		Convey("When I set the operation to Retrieve and I call makeResponse", func() {

			ctx.request.Operation = elemental.OperationRetrieve

			makeResponse(ctx, response, nil)

			Convey("Then response.StatusCode should be http.StatusOK", func() {
				So(response.StatusCode, ShouldEqual, http.StatusOK)
			})
		})

		Convey("When I set the operation to Create, status code OK, but no data, and I call makeResponse", func() {

			ctx.request.Operation = elemental.OperationCreate
			ctx.statusCode = http.StatusOK
			ctx.outputData = nil

			makeResponse(ctx, response, nil)

			Convey("Then response.StatusCode should be http.StatusNoContent", func() {
				So(response.StatusCode, ShouldEqual, http.StatusNoContent)
			})
		})
	})

	Convey("Given I have context with messages and a response", t, func() {

		ctx := newContext(context.Background(), elemental.NewRequest())
		response := elemental.NewResponse(elemental.NewRequest())
		ctx.AddMessage("hello world")

		Convey("When I call makeResponse", func() {

			makeResponse(ctx, response, nil)

			Convey("Then response.Message should be set", func() {
				So(response.Messages, ShouldResemble, []string{"hello world"})
			})
		})
	})

	Convey("Given I have context with next and a response", t, func() {

		ctx := newContext(context.Background(), elemental.NewRequest())
		response := elemental.NewResponse(elemental.NewRequest())
		ctx.SetNext("a")

		Convey("When I call makeResponse", func() {

			makeResponse(ctx, response, nil)

			Convey("Then response.Message should be set", func() {
				So(response.Next, ShouldEqual, "a")
			})
		})
	})

	Convey("Given I have context with unmarshalable data and a response", t, func() {

		ctx := newContext(context.Background(), elemental.NewRequest())
		response := elemental.NewResponse(elemental.NewRequest())
		ctx.outputData = testmodel.NewUnmarshalableList()

		Convey("When I call makeResponse", func() {

			Convey("Then it should panic", func() {
				So(func() { makeResponse(ctx, response, nil) }, ShouldPanic)
			})
		})
	})

	Convey("Given I have context with X-Fields header and indentifiable output data", t, func() {

		req := elemental.NewRequest()
		req.Headers.Add("X-Fields", "name")
		req.Headers.Add("X-Fields", "ID")

		ctx := newContext(context.Background(), req)
		response := elemental.NewResponse(req)
		ctx.outputData = &testmodel.List{
			Name:        "the name",
			ID:          "xxx",
			Description: " the description",
		}

		Convey("When I call makeResponse", func() {

			resp := makeResponse(ctx, response, nil)

			Convey("Then output data should be correct", func() {
				So(string(resp.Data), ShouldEqual, `{"ID":"xxx","name":"the name"}`)
			})
		})
	})

	Convey("Given I have context with X-Fields header and indentifiable output data", t, func() {

		req := elemental.NewRequest()
		req.Headers.Add("X-Fields", "name")
		req.Headers.Add("X-Fields", "ID")

		ctx := newContext(context.Background(), req)
		response := elemental.NewResponse(req)
		ctx.outputData = testmodel.ListsList{
			&testmodel.List{
				Name:        "the name",
				ID:          "xxx",
				Description: " the description",
			},
			&testmodel.List{
				Name:        "the name2",
				ID:          "xxx2",
				Description: " the description2",
			},
		}

		Convey("When I call makeResponse", func() {

			resp := makeResponse(ctx, response, nil)

			Convey("Then output data should be correct", func() {
				So(string(resp.Data), ShouldEqual, `[{"ID":"xxx","name":"the name"},{"ID":"xxx2","name":"the name2"}]`)
			})
		})
	})

	Convey("Given I have context indentifiable output data and custom working marshaller", t, func() {

		req := elemental.NewRequest()
		req.Identity = testmodel.ListIdentity
		req.Headers.Add("X-Fields", "name")
		req.Headers.Add("X-Fields", "ID")

		ctx := newContext(context.Background(), req)

		response := elemental.NewResponse(req)
		ctx.outputData = &testmodel.List{
			Name:        "the name",
			ID:          "xxx",
			Description: " the description",
		}

		Convey("When I call makeResponse", func() {

			resp := makeResponse(ctx, response, map[elemental.Identity]CustomMarshaller{
				testmodel.ListIdentity: func(*elemental.Response, interface{}, error) ([]byte, error) { return []byte("coucou"), nil },
			})

			Convey("Then output data should be correct", func() {
				So(string(resp.Data), ShouldEqual, `coucou`)
			})
		})
	})

	Convey("Given I have context indentifiable output data and custom non working marshaller", t, func() {

		req := elemental.NewRequest()
		req.Identity = testmodel.ListIdentity
		req.Headers.Add("X-Fields", "name")
		req.Headers.Add("X-Fields", "ID")

		ctx := newContext(context.Background(), req)

		response := elemental.NewResponse(req)
		ctx.outputData = &testmodel.List{
			Name:        "the name",
			ID:          "xxx",
			Description: " the description",
		}

		Convey("When I call makeResponse should panic", func() {

			So(func() {
				makeResponse(ctx, response, map[elemental.Identity]CustomMarshaller{
					testmodel.ListIdentity: func(*elemental.Response, interface{}, error) ([]byte, error) { return nil, fmt.Errorf("boom") },
				})
			}, ShouldPanicWith, "unable to encode output data using custom marshaller: boom")

		})
	})
}

func TestHandlers_makeErrorResponse(t *testing.T) {

	Convey("Given I a response and an error", t, func() {

		resp := elemental.NewResponse(elemental.NewRequest())
		err := fmt.Errorf("paf")

		Convey("When I call makeErrorResponse", func() {

			r := makeErrorResponse(context.Background(), resp, err, nil)

			Convey("Then the returned response should be the same", func() {
				So(resp, ShouldEqual, r)
			})

			Convey("Then response should be correctly encoded", func() {
				So(string(resp.Data), ShouldEqual, `[{"code":500,"description":"paf","subject":"elemental","title":"Internal Server Error","trace":"unknown"}]`)
			})
		})
	})

	Convey("Given I a response and an context canceled error", t, func() {

		err := context.Canceled

		Convey("When I call makeErrorResponse", func() {

			r := makeErrorResponse(context.Background(), nil, err, nil)

			Convey("Then the returned response should be the same", func() {
				So(r, ShouldEqual, nil)
			})
		})
	})

	Convey("Given I have context indentifiable output data and custom working marshaller", t, func() {

		req := elemental.NewRequest()
		req.Identity = testmodel.ListIdentity
		req.Headers.Add("X-Fields", "name")
		req.Headers.Add("X-Fields", "ID")

		resp := elemental.NewResponse(req)

		err := fmt.Errorf("paf")

		Convey("When I call makeErrorResponse", func() {

			r := makeErrorResponse(context.Background(), resp, err, map[elemental.Identity]CustomMarshaller{
				testmodel.ListIdentity: func(*elemental.Response, interface{}, error) ([]byte, error) { return []byte("coucou"), nil },
			})

			Convey("Then the returned response should be the same", func() {
				So(r.Data, ShouldResemble, []byte("coucou"))
			})
		})
	})

	Convey("Given I have context indentifiable output data and custom working marshaller", t, func() {

		req := elemental.NewRequest()
		req.Identity = testmodel.ListIdentity
		req.Headers.Add("X-Fields", "name")
		req.Headers.Add("X-Fields", "ID")

		resp := elemental.NewResponse(req)

		err := fmt.Errorf("paf")

		Convey("When I call makeErrorResponse should panic", func() {

			So(func() {
				makeErrorResponse(context.Background(), resp, err, map[elemental.Identity]CustomMarshaller{
					testmodel.ListIdentity: func(*elemental.Response, interface{}, error) ([]byte, error) { return nil, fmt.Errorf("boom") },
				})
			}, ShouldPanicWith, "unable to encode error using custom marshaller: boom")
		})
	})
}

func TestHandlers_handleEventualPanic(t *testing.T) {

	Convey("Given I have a response and a channel", t, func() {

		c := make(chan error)

		Convey("When I call my function that panics with handleEventualPanic installed with recover", func() {

			f := func() {
				defer handleEventualPanic(context.Background(), c, false)
				panic("Noooooooooooooooooo")
			}

			go f()

			err := <-c

			Convey("Then err should be correct", func() {
				So(err.Error(), ShouldEqual, "error 500 (bahamut): Internal Server Error: panic: Noooooooooooooooooo")
			})
		})

		Convey("When I call my function that panics with handleEventualPanic installed with no recover", func() {

			f := func() {
				defer handleEventualPanic(context.Background(), c, true)
				panic("Noooooooooooooooooo")
			}

			Convey("Then it should panic", func() {
				So(f, ShouldPanic)
			})
		})
	})
}

func TestHandlers_runDispatcher(t *testing.T) {

	Convey("When I call runDispatcher", t, func() {

		calledCounter := &counter{}

		hreq, err := http.NewRequest(http.MethodGet, "https://127.0.0.1/list", bytes.NewBuffer([]byte("hello")))
		if err != nil {
			panic(err)
		}
		gctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
		defer cancel()

		ctx := newContext(context.Background(), elemental.NewRequest())
		ctx.ctx = gctx

		ctx.request, err = elemental.NewRequestFromHTTPRequest(hreq, testmodel.Manager())
		if err != nil {
			panic(err)
		}

		response := elemental.NewResponse(elemental.NewRequest())

		d := func() error {
			calledCounter.Add(1)
			return nil
		}

		r := runDispatcher(ctx, response, d, true, nil)

		Convey("Then the code should be 204", func() {
			So(r.StatusCode, ShouldEqual, 204)
		})

		Convey("Then the dispatcher should have been called once", func() {
			So(calledCounter.Value(), ShouldEqual, 1)
		})
	})

	Convey("When I call runDispatcher and it returns an error", t, func() {

		calledCounter := &counter{}

		gctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
		defer cancel()

		ctx := newContext(context.Background(), elemental.NewRequest())
		ctx.request = elemental.NewRequest()
		ctx.ctx = gctx

		response := elemental.NewResponse(elemental.NewRequest())

		d := func() error {
			calledCounter.Add(1)
			return elemental.NewError("nop", "nope", "test", 42)
		}

		r := runDispatcher(ctx, response, d, true, nil)

		Convey("Then the code should be 42", func() {
			So(r.StatusCode, ShouldEqual, 42)
		})

		Convey("Then the dispatcher should have been called once", func() {
			So(calledCounter.Value(), ShouldEqual, 1)
		})
	})

	Convey("When I call runDispatcher and cancel the context", t, func() {

		calledCounter := &counter{}

		gctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
		defer cancel()

		ctx := newContext(context.Background(), elemental.NewRequest())
		ctx.request = elemental.NewRequest()
		ctx.ctx = gctx

		d := func() error {
			time.Sleep(300 * time.Millisecond)
			calledCounter.Add(1)
			return nil
		}

		r := elemental.NewResponse(elemental.NewRequest())
		go func() { runDispatcher(ctx, r, d, true, nil) }()
		time.Sleep(30 * time.Millisecond)
		cancel()

		Convey("Then the dispatcher should have been called once", func() {
			So(calledCounter.Value(), ShouldEqual, 0)
		})
	})
}

func TestHandlers_handleRetrieveMany(t *testing.T) {

	Convey("Given I have a config", t, func() {

		cfg := config{}
		cfg.model.modelManagers = map[int]elemental.ModelManager{
			0: testmodel.Manager(),
			1: testmodel.Manager(),
		}

		Convey("Given I a have fake processor finder that return no error", func() {

			calledCounter := &counter{}
			pf := func(identity elemental.Identity) (Processor, error) {
				calledCounter.Add(1)
				return struct{}{}, nil
			}

			Convey("When I call handleRetrieveMany on valid api call", func() {

				ctx := newContext(context.Background(), elemental.NewRequest())
				ctx.request = elemental.NewRequest()
				ctx.request.ParentIdentity = elemental.RootIdentity
				ctx.request.Identity = testmodel.UserIdentity
				ctx.request.Operation = elemental.OperationRetrieveMany
				ctx.statusCode = http.StatusAccepted

				resp := handleRetrieveMany(ctx, cfg, pf, nil)

				Convey("Then resp should be correct", func() {
					So(resp, ShouldNotBeNil)
					So(string(resp.Data), ShouldEqual, `[{"code":501,"description":"No handler for operation retrieve-many on user","subject":"bahamut","title":"Not implemented","trace":"unknown"}]`)
				})

				Convey("Then the dispactcher should have been called once", func() {
					So(calledCounter.Value(), ShouldEqual, 1)
				})
			})

			Convey("When I call handleRetrieveMany on invalid api call", func() {

				ctx := newContext(context.Background(), elemental.NewRequest())
				ctx.request = elemental.NewRequest()
				ctx.request.Identity = testmodel.UserIdentity
				ctx.request.Operation = elemental.OperationRetrieveMany
				ctx.request.ParentIdentity = testmodel.UserIdentity

				resp := handleRetrieveMany(ctx, cfg, pf, nil)

				Convey("Then resp should be correct", func() {
					So(resp, ShouldNotBeNil)
					So(string(resp.Data), ShouldEqual, `[{"code":405,"description":"RetrieveMany operation not allowed on users","subject":"bahamut","title":"Not allowed","trace":"unknown"}]`)
				})

				Convey("Then the dispactcher should have been called once", func() {
					So(calledCounter.Value(), ShouldEqual, 0)
				})
			})
		})
	})
}

func TestHandlers_handleRetrieve(t *testing.T) {

	Convey("Given I have a config", t, func() {

		cfg := config{}
		cfg.model.modelManagers = map[int]elemental.ModelManager{
			0: testmodel.Manager(),
			1: testmodel.Manager(),
		}

		Convey("Given I a have fake processor finder that return no error", func() {

			calledCounter := &counter{}
			pf := func(identity elemental.Identity) (Processor, error) {
				calledCounter.Add(1)
				return struct{}{}, nil
			}

			Convey("When I call handleRetrieve on valid api call", func() {

				ctx := newContext(context.Background(), elemental.NewRequest())
				ctx.request = elemental.NewRequest()
				ctx.request.Identity = testmodel.UserIdentity
				ctx.request.Operation = elemental.OperationRetrieve
				ctx.statusCode = http.StatusAccepted
				ctx.request.ParentIdentity = elemental.RootIdentity

				resp := handleRetrieve(ctx, cfg, pf, nil)

				Convey("Then resp should be correct", func() {
					So(resp, ShouldNotBeNil)
					So(string(resp.Data), ShouldEqual, `[{"code":501,"description":"No handler for operation retrieve on user","subject":"bahamut","title":"Not implemented","trace":"unknown"}]`)
				})

				Convey("Then the dispactcher should have been called once", func() {
					So(calledCounter.Value(), ShouldEqual, 1)
				})
			})

			Convey("When I call handleRetrieve on invalid api call", func() {

				ctx := newContext(context.Background(), elemental.NewRequest())
				ctx.request = elemental.NewRequest()
				ctx.request.Identity = testmodel.UserIdentity
				ctx.request.Operation = elemental.OperationRetrieve
				ctx.request.ParentIdentity = testmodel.UserIdentity

				resp := handleRetrieve(ctx, cfg, pf, nil)

				Convey("Then resp should be correct", func() {
					So(resp, ShouldNotBeNil)
					So(string(resp.Data), ShouldEqual, `[{"code":405,"description":"Retrieve operation not allowed on user","subject":"bahamut","title":"Not allowed","trace":"unknown"}]`)
				})

				Convey("Then the dispactcher should have been called once", func() {
					So(calledCounter.Value(), ShouldEqual, 0)
				})
			})
		})
	})
}

func TestHandlers_handleCreate(t *testing.T) {

	Convey("Given I have a config", t, func() {

		cfg := config{}
		cfg.model.modelManagers = map[int]elemental.ModelManager{
			0: testmodel.Manager(),
			1: testmodel.Manager(),
		}

		Convey("Given I a have fake processor finder that return no error", func() {

			calledCounter := &counter{}
			pf := func(identity elemental.Identity) (Processor, error) {
				calledCounter.Add(1)
				return struct{}{}, nil
			}

			Convey("When I call handleCreate on valid api call", func() {

				ctx := newContext(context.Background(), elemental.NewRequest())
				ctx.request = elemental.NewRequest()
				ctx.request.Identity = testmodel.UserIdentity
				ctx.request.Operation = elemental.OperationCreate
				ctx.statusCode = http.StatusAccepted
				ctx.request.ParentIdentity = elemental.RootIdentity

				resp := handleCreate(ctx, cfg, pf, nil)

				Convey("Then resp should be correct", func() {
					So(resp, ShouldNotBeNil)
					So(string(resp.Data), ShouldEqual, `[{"code":501,"description":"No handler for operation create on user","subject":"bahamut","title":"Not implemented","trace":"unknown"}]`)
				})

				Convey("Then the dispactcher should have been called once", func() {
					So(calledCounter.Value(), ShouldEqual, 1)
				})
			})

			Convey("When I call handleCreate on invalid api call", func() {

				ctx := newContext(context.Background(), elemental.NewRequest())
				ctx.request = elemental.NewRequest()
				ctx.request.Identity = testmodel.UserIdentity
				ctx.request.Operation = elemental.OperationCreate
				ctx.request.ParentIdentity = testmodel.UserIdentity

				resp := handleCreate(ctx, cfg, pf, nil)

				Convey("Then resp should be correct", func() {
					So(resp, ShouldNotBeNil)
					So(string(resp.Data), ShouldEqual, `[{"code":405,"description":"Create operation not allowed on user","subject":"bahamut","title":"Not allowed","trace":"unknown"}]`)
				})

				Convey("Then the dispactcher should have been called once", func() {
					So(calledCounter.Value(), ShouldEqual, 0)
				})
			})
		})
	})
}

func TestHandlers_handleUpdate(t *testing.T) {

	Convey("Given I have a config", t, func() {

		cfg := config{}
		cfg.model.modelManagers = map[int]elemental.ModelManager{
			0: testmodel.Manager(),
			1: testmodel.Manager(),
		}

		Convey("Given I a have fake processor finder that return no error", func() {

			calledCounter := &counter{}
			pf := func(identity elemental.Identity) (Processor, error) {
				calledCounter.Add(1)
				return struct{}{}, nil
			}

			Convey("When I call handleUpdate on valid api call", func() {

				ctx := newContext(context.Background(), elemental.NewRequest())
				ctx.request = elemental.NewRequest()
				ctx.request.Identity = testmodel.UserIdentity
				ctx.request.Operation = elemental.OperationUpdate
				ctx.statusCode = http.StatusAccepted
				ctx.request.ParentIdentity = elemental.RootIdentity

				resp := handleUpdate(ctx, cfg, pf, nil)

				Convey("Then resp should be correct", func() {
					So(resp, ShouldNotBeNil)
					So(string(resp.Data), ShouldEqual, `[{"code":501,"description":"No handler for operation update on user","subject":"bahamut","title":"Not implemented","trace":"unknown"}]`)
				})

				Convey("Then the dispactcher should have been called once", func() {
					So(calledCounter.Value(), ShouldEqual, 1)
				})
			})

			Convey("When I call handleUpdate on invalid api call", func() {

				ctx := newContext(context.Background(), elemental.NewRequest())
				ctx.request = elemental.NewRequest()
				ctx.request.Identity = testmodel.UserIdentity
				ctx.request.Operation = elemental.OperationUpdate
				ctx.request.ParentIdentity = testmodel.UserIdentity

				resp := handleUpdate(ctx, cfg, pf, nil)

				Convey("Then resp should be correct", func() {
					So(resp, ShouldNotBeNil)
					So(string(resp.Data), ShouldEqual, `[{"code":405,"description":"Update operation not allowed on user","subject":"bahamut","title":"Not allowed","trace":"unknown"}]`)
				})

				Convey("Then the dispactcher should have been called once", func() {
					So(calledCounter.Value(), ShouldEqual, 0)
				})
			})
		})
	})
}

func TestHandlers_handleDelete(t *testing.T) {

	Convey("Given I have a config", t, func() {

		cfg := config{}
		cfg.model.modelManagers = map[int]elemental.ModelManager{
			0: testmodel.Manager(),
			1: testmodel.Manager(),
		}

		Convey("Given I a have fake processor finder that return no error", func() {

			calledCounter := &counter{}
			pf := func(identity elemental.Identity) (Processor, error) {
				calledCounter.Add(1)
				return struct{}{}, nil
			}

			Convey("When I call handleDelete on valid api call", func() {

				ctx := newContext(context.Background(), elemental.NewRequest())
				ctx.request = elemental.NewRequest()
				ctx.request.Identity = testmodel.UserIdentity
				ctx.request.Operation = elemental.OperationDelete
				ctx.statusCode = http.StatusAccepted
				ctx.request.ParentIdentity = elemental.RootIdentity

				resp := handleDelete(ctx, cfg, pf, nil)

				Convey("Then resp should be correct", func() {
					So(resp, ShouldNotBeNil)
					So(string(resp.Data), ShouldEqual, `[{"code":501,"description":"No handler for operation delete on user","subject":"bahamut","title":"Not implemented","trace":"unknown"}]`)
				})

				Convey("Then the dispactcher should have been called once", func() {
					So(calledCounter.Value(), ShouldEqual, 1)
				})
			})

			Convey("When I call handleDelete on invalid api call", func() {

				ctx := newContext(context.Background(), elemental.NewRequest())
				ctx.request = elemental.NewRequest()
				ctx.request.Identity = testmodel.UserIdentity
				ctx.request.Operation = elemental.OperationDelete
				ctx.request.ParentIdentity = testmodel.UserIdentity

				resp := handleDelete(ctx, cfg, pf, nil)

				Convey("Then resp should be correct", func() {
					So(resp, ShouldNotBeNil)
					So(string(resp.Data), ShouldEqual, `[{"code":405,"description":"Delete operation not allowed on user","subject":"bahamut","title":"Not allowed","trace":"unknown"}]`)
				})

				Convey("Then the dispactcher should have been called once", func() {
					So(calledCounter.Value(), ShouldEqual, 0)
				})
			})
		})
	})
}

func TestHandlers_handleInfo(t *testing.T) {

	Convey("Given I have a config", t, func() {

		cfg := config{}
		cfg.model.modelManagers = map[int]elemental.ModelManager{
			0: testmodel.Manager(),
			1: testmodel.Manager(),
		}

		Convey("Given I a have fake processor finder that return no error", func() {

			calledCounter := &counter{}
			pf := func(identity elemental.Identity) (Processor, error) {
				calledCounter.Add(1)
				return struct{}{}, nil
			}

			Convey("When I call handleInfo on valid api call", func() {

				ctx := newContext(context.Background(), elemental.NewRequest())
				ctx.request = elemental.NewRequest()
				ctx.request.Identity = testmodel.UserIdentity
				ctx.request.Operation = elemental.OperationInfo
				ctx.statusCode = http.StatusAccepted
				ctx.request.ParentIdentity = elemental.RootIdentity

				resp := handleInfo(ctx, cfg, pf, nil)

				Convey("Then resp should be correct", func() {
					So(resp, ShouldNotBeNil)
					So(string(resp.Data), ShouldEqual, `[{"code":501,"description":"No handler for operation info on user","subject":"bahamut","title":"Not implemented","trace":"unknown"}]`)
				})

				Convey("Then the dispactcher should have been called once", func() {
					So(calledCounter.Value(), ShouldEqual, 1)
				})
			})

			Convey("When I call handleInfo on invalid api call", func() {

				ctx := newContext(context.Background(), elemental.NewRequest())
				ctx.request = elemental.NewRequest()
				ctx.request.Identity = testmodel.UserIdentity
				ctx.request.Operation = elemental.OperationInfo
				ctx.request.ParentIdentity = testmodel.UserIdentity

				resp := handleInfo(ctx, cfg, pf, nil)

				Convey("Then resp should be correct", func() {
					So(resp, ShouldNotBeNil)
					So(string(resp.Data), ShouldEqual, `[{"code":405,"description":"Info operation not allowed on users","subject":"bahamut","title":"Not allowed","trace":"unknown"}]`)
				})

				Convey("Then the dispactcher should have been called once", func() {
					So(calledCounter.Value(), ShouldEqual, 0)
				})
			})
		})
	})
}

func TestHandlers_handlePatch(t *testing.T) {

	Convey("Given I have a config", t, func() {

		cfg := config{}
		cfg.model.modelManagers = map[int]elemental.ModelManager{
			0: testmodel.Manager(),
			1: testmodel.Manager(),
		}

		Convey("Given I a have fake processor finder that return no error", func() {

			calledCounter := &counter{}
			pf := func(identity elemental.Identity) (Processor, error) {
				calledCounter.Add(1)
				return struct{}{}, nil
			}

			Convey("When I call handlePatch on valid api call", func() {

				ctx := newContext(context.Background(), elemental.NewRequest())
				ctx.request = elemental.NewRequest()
				ctx.request.Identity = testmodel.UserIdentity
				ctx.request.Operation = elemental.OperationPatch
				ctx.statusCode = http.StatusAccepted
				ctx.request.ParentIdentity = elemental.RootIdentity

				resp := handlePatch(ctx, cfg, pf, nil)

				Convey("Then resp should be correct", func() {
					So(resp, ShouldNotBeNil)
					So(string(resp.Data), ShouldEqual, `[{"code":501,"description":"No handler for operation patch on user","subject":"bahamut","title":"Not implemented","trace":"unknown"}]`)
				})

				Convey("Then the dispactcher should have been called once", func() {
					So(calledCounter.Value(), ShouldEqual, 1)
				})
			})

			Convey("When I call handlePatch on invalid api call", func() {

				ctx := newContext(context.Background(), elemental.NewRequest())
				ctx.request = elemental.NewRequest()
				ctx.request.Identity = testmodel.UserIdentity
				ctx.request.Operation = elemental.OperationPatch
				ctx.request.ParentIdentity = testmodel.UserIdentity

				resp := handlePatch(ctx, cfg, pf, nil)

				Convey("Then resp should be correct", func() {
					So(resp, ShouldNotBeNil)
					So(string(resp.Data), ShouldEqual, `[{"code":405,"description":"Patch operation not allowed on users","subject":"bahamut","title":"Not allowed","trace":"unknown"}]`)
				})

				Convey("Then the dispactcher should have been called once", func() {
					So(calledCounter.Value(), ShouldEqual, 0)
				})
			})
		})
	})
}
