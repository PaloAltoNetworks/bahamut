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
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"go.aporeto.io/elemental"
	testmodel "go.aporeto.io/elemental/test/model"
)

// TestDispatchers_dispatchRetrieveManyOperation tests dispatchRetrieveManyOperation method
func TestDispatchers_dispatchRetrieveManyOperation(t *testing.T) {

	Convey("Given I have a processor that handle ProcessRetrieveMany function", t, func() {
		request := elemental.NewRequest()

		processorFinder := func(identity elemental.Identity) (Processor, error) {
			return &mockProcessor{
				output: "hello",
				events: []*elemental.Event{elemental.NewEvent(elemental.EventUpdate, &testmodel.List{})},
			}, nil
		}

		auditer := &mockAuditer{}
		pusher := &mockPusher{}

		ctx := newContext(context.TODO(), request)
		err := dispatchRetrieveManyOperation(ctx, processorFinder, nil, nil, pusher.Push, auditer)

		expectedNbCalls := 1

		Convey("Then I should have no error and context should be initiated", func() {
			So(err, ShouldBeNil)
			So(auditer.GetCallCount(), ShouldEqual, expectedNbCalls)
			So(ctx.outputData, ShouldResemble, "hello")
			So(len(pusher.events), ShouldEqual, 1)
			So(pusher.events[0].Type, ShouldEqual, elemental.EventUpdate)
		})
	})

	Convey("Given I have a processor that handle ProcessRetrieveMany function with error", t, func() {
		request := elemental.NewRequest()

		processorFinder := func(identity elemental.Identity) (Processor, error) {
			return &mockProcessor{
				err: elemental.NewError("Error", "Bad request.", "bahamut-test", http.StatusBadRequest),
			}, nil
		}

		auditer := &mockAuditer{}

		ctx := newContext(context.TODO(), request)
		err := dispatchRetrieveManyOperation(ctx, processorFinder, nil, nil, nil, auditer)

		expectedError := "error 400 (bahamut-test): Error: Bad request."
		expectedNbCalls := 1

		Convey("Then I should get a bahamut error and no context", func() {
			So(err.Error(), ShouldEqual, expectedError)
			So(auditer.GetCallCount(), ShouldEqual, expectedNbCalls)
		})
	})

	Convey("Given I have a processor that does not handle ProcessRetrieveMany function", t, func() {
		request := elemental.NewRequest()
		request.Operation = elemental.OperationRetrieveMany
		request.Identity = elemental.MakeIdentity("Fake", "Test")

		processorFinder := func(identity elemental.Identity) (Processor, error) {
			return &mockEmptyProcessor{}, nil
		}

		auditer := &mockAuditer{}

		expectedError := "error 501 (bahamut): Not implemented: No handler for operation retrieve-many on Fake"
		expectedNbCalls := 1

		ctx := newContext(context.TODO(), request)
		err := dispatchRetrieveManyOperation(ctx, processorFinder, nil, nil, nil, auditer)

		Convey("Then I should get a bahamut error and no context", func() {
			So(err.Error(), ShouldEqual, expectedError)
			So(auditer.GetCallCount(), ShouldEqual, expectedNbCalls)
		})
	})

	Convey("Given I have a processor that handle ProcessRetrieveMany function and an authenticator that is not authenticated", t, func() {
		request := elemental.NewRequest()

		processorFinder := func(identity elemental.Identity) (Processor, error) {
			return &mockEmptyProcessor{}, nil
		}

		authenticators := []RequestAuthenticator{
			&mockAuth{
				action: AuthActionContinue,
			},
			&mockAuth{
				errored: true,
				err:     elemental.NewError("Error", "Authenticator does not authenticate.", "bahamut-test", http.StatusInternalServerError),
			},
		}

		auditer := &mockAuditer{}

		expectedError := "error 500 (bahamut-test): Error: Authenticator does not authenticate."
		expectedNbCalls := 1

		ctx := newContext(context.TODO(), request)
		err := dispatchRetrieveManyOperation(ctx, processorFinder, authenticators, nil, nil, auditer)

		Convey("Then I should get a bahamut error and no context", func() {
			So(err.Error(), ShouldEqual, expectedError)
			So(auditer.GetCallCount(), ShouldEqual, expectedNbCalls)
		})
	})

	Convey("Given I have a processor that handle ProcessRetrieveMany function and an authorizer that is not authorize", t, func() {
		request := elemental.NewRequest()

		processorFinder := func(identity elemental.Identity) (Processor, error) {
			return &mockEmptyProcessor{}, nil
		}

		authenticators := []RequestAuthenticator{
			&mockAuth{
				action: AuthActionContinue,
			},
			&mockAuth{
				action: AuthActionContinue,
			},
		}

		authorizers := []Authorizer{
			&mockAuth{
				action: AuthActionContinue,
			},
			&mockAuth{
				errored: true,
				err:     elemental.NewError("Error", "Authorizer does not authorize.", "bahamut-test", http.StatusInternalServerError),
			},
		}

		auditer := &mockAuditer{}

		expectedError := "error 500 (bahamut-test): Error: Authorizer does not authorize."
		expectedNbCalls := 1

		ctx := newContext(context.TODO(), request)
		err := dispatchRetrieveManyOperation(ctx, processorFinder, authenticators, authorizers, nil, auditer)

		Convey("Then I should get a bahamut error and no context", func() {
			So(err.Error(), ShouldEqual, expectedError)
			So(auditer.GetCallCount(), ShouldEqual, expectedNbCalls)
		})
	})
}

// TestDispatchers_dispatchRetrieveOperation tests dispatchRetrieveOperation method
func TestDispatchers_dispatchRetrieveOperation(t *testing.T) {

	Convey("Given I have a processor that handle ProcessRetrieve function", t, func() {
		request := elemental.NewRequest()

		processorFinder := func(identity elemental.Identity) (Processor, error) {
			return &mockProcessor{
				output: "hello",
				events: []*elemental.Event{elemental.NewEvent(elemental.EventUpdate, &testmodel.List{})},
			}, nil
		}

		auditer := &mockAuditer{}
		pusher := &mockPusher{}

		ctx := newContext(context.TODO(), request)
		err := dispatchRetrieveOperation(ctx, processorFinder, nil, nil, pusher.Push, auditer)

		expectedNbCalls := 1

		Convey("Then I should have no error and context should be initiated", func() {
			So(err, ShouldBeNil)
			So(auditer.GetCallCount(), ShouldEqual, expectedNbCalls)
			So(ctx.outputData, ShouldResemble, "hello")
			So(len(pusher.events), ShouldEqual, 1)
			So(pusher.events[0].Type, ShouldEqual, elemental.EventUpdate)
		})
	})

	Convey("Given I have a processor that handle ProcessRetrieve function with error", t, func() {
		request := elemental.NewRequest()

		processorFinder := func(identity elemental.Identity) (Processor, error) {
			return &mockProcessor{
				err: elemental.NewError("Error", "Bad request.", "bahamut-test", http.StatusBadRequest),
			}, nil
		}

		auditer := &mockAuditer{}

		ctx := newContext(context.TODO(), request)
		err := dispatchRetrieveOperation(ctx, processorFinder, nil, nil, nil, auditer)

		expectedError := "error 400 (bahamut-test): Error: Bad request."
		expectedNbCalls := 1

		Convey("Then I should get a bahamut error and no context", func() {
			So(err.Error(), ShouldEqual, expectedError)
			So(auditer.GetCallCount(), ShouldEqual, expectedNbCalls)
		})
	})

	Convey("Given I have a processor that does not handle ProcessRetrieve function", t, func() {
		request := elemental.NewRequest()
		request.Operation = elemental.OperationRetrieveMany
		request.Identity = elemental.MakeIdentity("Fake", "Test")

		processorFinder := func(identity elemental.Identity) (Processor, error) {
			return &mockEmptyProcessor{}, nil
		}

		auditer := &mockAuditer{}

		expectedError := "error 501 (bahamut): Not implemented: No handler for operation retrieve-many on Fake"
		expectedNbCalls := 1

		ctx := newContext(context.TODO(), request)
		err := dispatchRetrieveOperation(ctx, processorFinder, nil, nil, nil, auditer)

		Convey("Then I should get a bahamut error and no context", func() {
			So(err.Error(), ShouldEqual, expectedError)
			So(auditer.GetCallCount(), ShouldEqual, expectedNbCalls)
		})
	})

	Convey("Given I have a processor that does not handle ProcessRetrieve function and an authenticator that is not authenticated", t, func() {
		request := elemental.NewRequest()

		processorFinder := func(identity elemental.Identity) (Processor, error) {
			return &mockEmptyProcessor{}, nil
		}

		authenticators := []RequestAuthenticator{
			&mockAuth{
				action: AuthActionContinue,
			},
			&mockAuth{
				errored: true,
				err:     elemental.NewError("Error", "Authenticator does not authenticate.", "bahamut-test", http.StatusInternalServerError),
			},
		}

		auditer := &mockAuditer{}

		expectedError := "error 500 (bahamut-test): Error: Authenticator does not authenticate."
		expectedNbCalls := 1

		ctx := newContext(context.TODO(), request)
		err := dispatchRetrieveOperation(ctx, processorFinder, authenticators, nil, nil, auditer)

		Convey("Then I should get a bahamut error and no context", func() {
			So(err.Error(), ShouldEqual, expectedError)
			So(auditer.GetCallCount(), ShouldEqual, expectedNbCalls)
		})
	})

	Convey("Given I have a processor that does not handle ProcessRetrieve function and an authorizer that is not authorize", t, func() {
		request := elemental.NewRequest()

		processorFinder := func(identity elemental.Identity) (Processor, error) {
			return &mockEmptyProcessor{}, nil
		}

		authenticators := []RequestAuthenticator{
			&mockAuth{
				action: AuthActionContinue,
			},
			&mockAuth{
				action: AuthActionContinue,
			},
		}

		authorizers := []Authorizer{
			&mockAuth{
				action: AuthActionContinue,
			},
			&mockAuth{
				errored: true,
				err:     elemental.NewError("Error", "Authorizer does not authorize.", "bahamut-test", http.StatusInternalServerError),
			},
		}

		auditer := &mockAuditer{}

		expectedError := "error 500 (bahamut-test): Error: Authorizer does not authorize."
		expectedNbCalls := 1

		ctx := newContext(context.TODO(), request)
		err := dispatchRetrieveOperation(ctx, processorFinder, authenticators, authorizers, nil, auditer)

		Convey("Then I should get a bahamut error and no context", func() {
			So(err.Error(), ShouldEqual, expectedError)
			So(auditer.GetCallCount(), ShouldEqual, expectedNbCalls)
		})
	})
}

// TestDispatchers_dispatchCreateOperation tests dispatchCreateOperation method
func TestDispatchers_dispatchCreateOperation(t *testing.T) {

	Convey("Given I have a processor that handle ProcessCreate function", t, func() {
		request := elemental.NewRequest()
		request.Identity = testmodel.ListIdentity
		request.Data = []byte(`{"ID": "1234", "name": "Fake"}`)

		processorFinder := func(identity elemental.Identity) (Processor, error) {
			return &mockProcessor{
				output: &testmodel.List{ID: "a"},
				events: []*elemental.Event{elemental.NewEvent(elemental.EventUpdate, &testmodel.List{})},
			}, nil
		}

		auditer := &mockAuditer{}
		pusher := &mockPusher{}

		ctx := newContext(context.TODO(), request)
		err := dispatchCreateOperation(ctx, processorFinder, testmodel.Manager(), nil, nil, nil, pusher.Push, auditer, false, nil)

		expectedNbCalls := 1

		Convey("Then I should have no error and context should be initiated", func() {
			So(err, ShouldBeNil)
			So(ctx.InputData, ShouldNotBeNil)
			So(auditer.GetCallCount(), ShouldEqual, expectedNbCalls)
			So(ctx.outputData, ShouldResemble, &testmodel.List{ID: "a"})
			So(len(pusher.events), ShouldEqual, 2)
			So(pusher.events[0].Type, ShouldEqual, elemental.EventUpdate)
			So(pusher.events[1].Type, ShouldEqual, elemental.EventCreate)
		})
	})

	// bug fix: https://github.com/aporeto-inc/bahamut/issues/64
	Convey("Setup request and fresh context", t, func() {

		request := elemental.NewRequest()
		request.Identity = testmodel.ListIdentity
		request.Data = []byte(`{"ID": "1234", "name": "Fake"}`)
		ctx := newContext(context.TODO(), request)

		auditer := &mockAuditer{}
		pusher := &mockPusher{}

		Convey("Given a processor that can handle ProcessCreate function with a context output that does not satisfy the elemental.Identifiable interface", func() {

			var err error
			processorFinder := func(identity elemental.Identity) (Processor, error) {
				return &mockProcessor{
					// notice how this output will NOT satisfy the elemental.Identifiable interface
					output: json.RawMessage("some random bytes!"),
				}, nil
			}

			Convey("Then I should not panic no events should be pushed", func() {
				So(func() {
					err = dispatchCreateOperation(ctx, processorFinder, testmodel.Manager(), nil, nil, nil, pusher.Push, auditer, false, nil)
				}, ShouldNotPanic)
				So(err, ShouldBeNil)
				So(ctx.InputData, ShouldNotBeNil)
				So(auditer.GetCallCount(), ShouldEqual, 1)
				So(ctx.outputData, ShouldResemble, json.RawMessage("some random bytes!"))
				So(len(pusher.events), ShouldEqual, 0)
			})

		})

		Convey("Given a processor that can handle ProcessCreate function with a context output that contains a nil elemental.Identifiable", func() {

			// notice how this is a type that satisfies the elemental.Identifiable interface, but it is not a nil interface!
			var testIdentifiable *testmodel.List
			var _ elemental.Identifiable = testIdentifiable
			ctx.SetOutputData(testIdentifiable)

			var err error
			processorFinder := func(identity elemental.Identity) (Processor, error) {
				return &mockProcessor{
					// notice how this output satisfies the elemental.Identifiable interface
					output: testIdentifiable,
				}, nil
			}

			Convey("Then I should not panic and an event should be pushed", func() {
				So(func() {
					err = dispatchCreateOperation(ctx, processorFinder, testmodel.Manager(), nil, nil, nil, pusher.Push, auditer, false, nil)
				}, ShouldNotPanic)
				So(err, ShouldBeNil)
				So(ctx.InputData, ShouldNotBeNil)
				So(auditer.GetCallCount(), ShouldEqual, 1)
				So(ctx.outputData, ShouldResemble, (*testmodel.List)(nil))
				So(len(pusher.events), ShouldEqual, 1)
			})
		})
	})

	Convey("Given I have a processor that handle ProcessCreate function with read only mode", t, func() {
		request := elemental.NewRequest()
		request.Identity = testmodel.ListIdentity
		request.Data = []byte(`{"ID": "1234", "name": "Fake"}`)

		processorFinder := func(identity elemental.Identity) (Processor, error) {
			return &mockProcessor{}, nil
		}

		auditer := &mockAuditer{}

		ctx := newContext(context.TODO(), request)
		err := dispatchCreateOperation(ctx, processorFinder, testmodel.Manager(), nil, nil, nil, nil, auditer, true, nil)

		Convey("Then I should have a 423 error and context should be nil", func() {
			So(err, ShouldNotBeNil)
			So(err.(elemental.Error).Code, ShouldEqual, http.StatusLocked)
		})
	})

	Convey("Given I have a processor that handle ProcessCreate function with error", t, func() {
		request := elemental.NewRequest()
		request.Identity = testmodel.ListIdentity
		request.Data = []byte(`{"ID": "1234", "name": "Fake"}`)

		processorFinder := func(identity elemental.Identity) (Processor, error) {
			return &mockProcessor{
				err: elemental.NewError("Error", "Bad request.", "bahamut-test", http.StatusBadRequest),
			}, nil
		}

		auditer := &mockAuditer{}

		ctx := newContext(context.TODO(), request)
		err := dispatchCreateOperation(ctx, processorFinder, testmodel.Manager(), nil, nil, nil, nil, auditer, false, nil)

		expectedError := "error 400 (bahamut-test): Error: Bad request."
		expectedNbCalls := 1

		Convey("Then I should get a bahamut error and no context", func() {
			So(err.Error(), ShouldEqual, expectedError)
			So(auditer.GetCallCount(), ShouldEqual, expectedNbCalls)
		})
	})

	Convey("Given I have a processor that handle ProcessCreate function with an empty JSON", t, func() {
		request := elemental.NewRequest()
		request.Identity = testmodel.ListIdentity

		processorFinder := func(identity elemental.Identity) (Processor, error) {
			return &mockProcessor{}, nil
		}

		auditer := &mockAuditer{}

		ctx := newContext(context.TODO(), request)
		err := dispatchCreateOperation(ctx, processorFinder, testmodel.Manager(), nil, nil, nil, nil, auditer, false, nil)

		expectedError := "error 400 (bahamut): Bad Request: unable to decode application/json: json decode error [pos 0]: only encoded map or array can be decoded into a struct"
		expectedNbCalls := 1

		Convey("Then I should get a bahamut error and no context", func() {
			So(err.Error(), ShouldEqual, expectedError)
			So(auditer.GetCallCount(), ShouldEqual, expectedNbCalls)
		})
	})

	Convey("Given I have a processor that handle ProcessCreate function with an invalid JSON", t, func() {
		request := elemental.NewRequest()
		request.Data = []byte(`An invalid JSON`)

		processorFinder := func(identity elemental.Identity) (Processor, error) {
			return &mockProcessor{}, nil
		}

		auditer := &mockAuditer{}

		ctx := newContext(context.TODO(), request)
		err := dispatchCreateOperation(ctx, processorFinder, testmodel.Manager(), nil, nil, nil, nil, auditer, false, nil)

		expectedError := "error 400 (bahamut): Bad Request: unable to decode application/json: json decode error [pos 1]: cannot decode into nil"
		expectedNbCalls := 1

		Convey("Then I should get a bahamut error and no context", func() {
			So(err.Error(), ShouldEqual, expectedError)
			So(auditer.GetCallCount(), ShouldEqual, expectedNbCalls)
		})
	})

	Convey("Given I have a processor that handle ProcessCreate function with an invalid object", t, func() {
		request := elemental.NewRequest()
		request.Identity = testmodel.TaskIdentity
		request.Data = []byte(`{"ID": "1234", "name": "Fake", "status": "not-good"}`)

		processorFinder := func(identity elemental.Identity) (Processor, error) {
			return &mockProcessor{}, nil
		}

		auditer := &mockAuditer{}

		ctx := newContext(context.TODO(), request)
		err := dispatchCreateOperation(ctx, processorFinder, testmodel.Manager(), nil, nil, nil, nil, auditer, false, nil)

		expectedError := "error 422 (elemental): Validation Error: Data 'not-good' of attribute 'status' is not in list '[DONE PROGRESS TODO]'"
		expectedNbCalls := 1

		Convey("Then I should get a bahamut error and no context", func() {
			So(err.Error(), ShouldEqual, expectedError)
			So(auditer.GetCallCount(), ShouldEqual, expectedNbCalls)
		})
	})

	Convey("Given I have a processor that does not handle ProcessCreate function", t, func() {
		request := elemental.NewRequest()
		request.Operation = elemental.OperationRetrieveMany
		request.Identity = elemental.MakeIdentity("Fake", "Test")

		processorFinder := func(identity elemental.Identity) (Processor, error) {
			return &mockEmptyProcessor{}, nil
		}

		auditer := &mockAuditer{}

		expectedError := "error 501 (bahamut): Not implemented: No handler for operation retrieve-many on Fake"
		expectedNbCalls := 1

		ctx := newContext(context.TODO(), request)
		err := dispatchCreateOperation(ctx, processorFinder, testmodel.Manager(), nil, nil, nil, nil, auditer, false, nil)

		Convey("Then I should get a bahamut error and no context", func() {
			So(err.Error(), ShouldEqual, expectedError)
			So(auditer.GetCallCount(), ShouldEqual, expectedNbCalls)
		})
	})

	Convey("Given I have a processor with custom unmarshal function that fails", t, func() {
		request := elemental.NewRequest()
		request.Operation = elemental.OperationCreate
		request.Identity = elemental.MakeIdentity("Fake", "Test")

		processorFinder := func(identity elemental.Identity) (Processor, error) {
			return &mockProcessor{}, nil
		}

		expectedError := "boom"

		ctx := newContext(context.TODO(), request)
		err := dispatchCreateOperation(
			ctx,
			processorFinder,
			testmodel.Manager(),
			func(*elemental.Request) (elemental.Identifiable, error) {
				return nil, fmt.Errorf("boom")
			},
			nil,
			nil,
			nil,
			nil,
			false,
			nil,
		)

		Convey("Then I should get a bahamut error and no context", func() {
			So(err.Error(), ShouldEqual, expectedError)
		})
	})

	Convey("Given I have a processor that does not handle ProcessCreate function and an authenticator that is not authenticated", t, func() {
		request := elemental.NewRequest()

		processorFinder := func(identity elemental.Identity) (Processor, error) {
			return &mockEmptyProcessor{}, nil
		}

		authenticators := []RequestAuthenticator{
			&mockAuth{
				action: AuthActionContinue,
			},
			&mockAuth{
				errored: true,
				err:     elemental.NewError("Error", "Authenticator does not authenticate.", "bahamut-test", http.StatusInternalServerError),
			},
		}

		auditer := &mockAuditer{}

		expectedError := "error 500 (bahamut-test): Error: Authenticator does not authenticate."
		expectedNbCalls := 1

		ctx := newContext(context.TODO(), request)
		err := dispatchCreateOperation(ctx, processorFinder, testmodel.Manager(), nil, authenticators, nil, nil, auditer, false, nil)

		Convey("Then I should get a bahamut error and no context", func() {
			So(err.Error(), ShouldEqual, expectedError)
			So(auditer.GetCallCount(), ShouldEqual, expectedNbCalls)
		})
	})

	Convey("Given I have a processor that does not handle ProcessCreate function and an authorizer that is not authorize", t, func() {
		request := elemental.NewRequest()

		processorFinder := func(identity elemental.Identity) (Processor, error) {
			return &mockEmptyProcessor{}, nil
		}

		authenticators := []RequestAuthenticator{
			&mockAuth{
				action: AuthActionContinue,
			},
			&mockAuth{
				action: AuthActionContinue,
			},
		}

		authorizers := []Authorizer{
			&mockAuth{
				action: AuthActionContinue,
			},
			&mockAuth{
				errored: true,
				err:     elemental.NewError("Error", "Authorizer does not authorize.", "bahamut-test", http.StatusInternalServerError),
			},
		}

		auditer := &mockAuditer{}

		expectedError := "error 500 (bahamut-test): Error: Authorizer does not authorize."
		expectedNbCalls := 1

		ctx := newContext(context.TODO(), request)
		err := dispatchCreateOperation(ctx, processorFinder, testmodel.Manager(), nil, authenticators, authorizers, nil, auditer, false, nil)

		Convey("Then I should get a bahamut error and no context", func() {
			So(err.Error(), ShouldEqual, expectedError)
			So(auditer.GetCallCount(), ShouldEqual, expectedNbCalls)
		})
	})
}

// TestDispatchers_dispatchUpdateOperation tests dispatchUpdateOperation method
func TestDispatchers_dispatchUpdateOperation(t *testing.T) {

	Convey("Given I have a processor that handle ProcessUpdate function", t, func() {
		request := elemental.NewRequest()
		request.Identity = testmodel.ListIdentity
		request.Data = []byte(`{"ID": "1234", "name": "Fake"}`)

		processorFinder := func(identity elemental.Identity) (Processor, error) {
			return &mockProcessor{
				output: &testmodel.List{ID: "a"},
				events: []*elemental.Event{elemental.NewEvent(elemental.EventDelete, &testmodel.List{})},
			}, nil
		}

		auditer := &mockAuditer{}
		pusher := &mockPusher{}

		ctx := newContext(context.TODO(), request)
		err := dispatchUpdateOperation(ctx, processorFinder, testmodel.Manager(), nil, nil, nil, pusher.Push, auditer, false, nil)

		expectedNbCalls := 1

		Convey("Then I should have no error and context should be initiated", func() {
			So(err, ShouldBeNil)
			So(ctx.InputData, ShouldNotBeNil)
			So(auditer.GetCallCount(), ShouldEqual, expectedNbCalls)
			So(ctx.outputData, ShouldResemble, &testmodel.List{ID: "a"})
			So(len(pusher.events), ShouldEqual, 2)
			So(pusher.events[0].Type, ShouldEqual, elemental.EventDelete)
			So(pusher.events[1].Type, ShouldEqual, elemental.EventUpdate)
		})
	})

	Convey("Setup request and fresh context", t, func() {

		request := elemental.NewRequest()
		request.Identity = testmodel.ListIdentity
		request.Data = []byte(`{"ID": "1234", "name": "Fake"}`)
		ctx := newContext(context.TODO(), request)

		auditer := &mockAuditer{}
		pusher := &mockPusher{}

		// bug fix: https://github.com/aporeto-inc/bahamut/issues/64
		Convey("Given I have a processor that handle ProcessUpdate function with a context output that does not satisfy elemental.Identifiable", func() {

			processorFinder := func(identity elemental.Identity) (Processor, error) {
				return &mockProcessor{
					// notice how this output will NOT satisfy the elemental.Identifiable interface
					output: json.RawMessage("some random bytes!"),
				}, nil
			}

			Convey("Then I should not panic no events should be pushed", func() {
				var err error
				So(func() {
					err = dispatchUpdateOperation(ctx, processorFinder, testmodel.Manager(), nil, nil, nil, pusher.Push, auditer, false, nil)
				}, ShouldNotPanic)
				So(err, ShouldBeNil)
				So(ctx.InputData, ShouldNotBeNil)
				So(auditer.GetCallCount(), ShouldEqual, 1)
				So(ctx.outputData, ShouldResemble, json.RawMessage("some random bytes!"))
				So(len(pusher.events), ShouldEqual, 0)
			})
		})

		Convey("Given I have a processor that handle ProcessUpdate function with a context output that contains a nil elemental.Identifiable", func() {

			// notice how this is a type that satisfies the elemental.Identifiable interface, but it is not a nil interface!
			var testIdentifiable *testmodel.List
			var _ elemental.Identifiable = testIdentifiable
			ctx.SetOutputData(testIdentifiable)

			processorFinder := func(identity elemental.Identity) (Processor, error) {
				return &mockProcessor{
					// notice how this output satisfies the elemental.Identifiable interface
					output: testIdentifiable,
				}, nil
			}

			Convey("Then I should not panic and an event should be pushed", func() {
				var err error
				So(func() {
					err = dispatchUpdateOperation(ctx, processorFinder, testmodel.Manager(), nil, nil, nil, pusher.Push, auditer, false, nil)
				}, ShouldNotPanic)
				So(err, ShouldBeNil)
				So(ctx.InputData, ShouldNotBeNil)
				So(auditer.GetCallCount(), ShouldEqual, 1)
				So(ctx.outputData, ShouldResemble, (*testmodel.List)(nil))
				So(len(pusher.events), ShouldEqual, 1)
			})
		})
	})

	Convey("Given I have a processor that handle ProcessUpdate function with read only mode", t, func() {
		request := elemental.NewRequest()
		request.Identity = testmodel.ListIdentity
		request.Data = []byte(`{"ID": "1234", "name": "Fake"}`)

		processorFinder := func(identity elemental.Identity) (Processor, error) {
			return &mockProcessor{}, nil
		}

		auditer := &mockAuditer{}

		ctx := newContext(context.TODO(), request)
		err := dispatchUpdateOperation(ctx, processorFinder, testmodel.Manager(), nil, nil, nil, nil, auditer, true, nil)

		Convey("Then I should have a 423 error and context should be nil", func() {
			So(err, ShouldNotBeNil)
			So(err.(elemental.Error).Code, ShouldEqual, http.StatusLocked)
		})
	})

	Convey("Given I have a processor that handle ProcessUpdate function with error", t, func() {
		request := elemental.NewRequest()
		request.Identity = testmodel.ListIdentity
		request.Data = []byte(`{"ID": "1234", "name": "Fake"}`)

		processorFinder := func(identity elemental.Identity) (Processor, error) {
			return &mockProcessor{
				err: elemental.NewError("Error", "Bad request.", "bahamut-test", http.StatusBadRequest),
			}, nil
		}

		auditer := &mockAuditer{}

		ctx := newContext(context.TODO(), request)
		err := dispatchUpdateOperation(ctx, processorFinder, testmodel.Manager(), nil, nil, nil, nil, auditer, false, nil)

		expectedError := "error 400 (bahamut-test): Error: Bad request."
		expectedNbCalls := 1

		Convey("Then I should get a bahamut error and no context", func() {
			So(err.Error(), ShouldEqual, expectedError)
			So(auditer.GetCallCount(), ShouldEqual, expectedNbCalls)
		})
	})

	Convey("Given I have a processor that handle ProcessUpdate function with an empty JSON", t, func() {
		request := elemental.NewRequest()
		request.Identity = testmodel.ListIdentity

		processorFinder := func(identity elemental.Identity) (Processor, error) {
			return &mockProcessor{}, nil
		}

		auditer := &mockAuditer{}

		ctx := newContext(context.TODO(), request)
		err := dispatchUpdateOperation(ctx, processorFinder, testmodel.Manager(), nil, nil, nil, nil, auditer, false, nil)

		expectedError := "error 400 (bahamut): Bad Request: unable to decode application/json: json decode error [pos 0]: only encoded map or array can be decoded into a struct"
		expectedNbCalls := 1

		Convey("Then I should get a bahamut error and no context", func() {
			So(err.Error(), ShouldEqual, expectedError)
			So(auditer.GetCallCount(), ShouldEqual, expectedNbCalls)
		})
	})

	Convey("Given I have a processor that handle ProcessUpdate function with an invalid JSON", t, func() {
		request := elemental.NewRequest()
		request.Data = []byte(`An invalid JSON`)
		request.Identity = testmodel.ListIdentity

		processorFinder := func(identity elemental.Identity) (Processor, error) {
			return &mockProcessor{}, nil
		}

		auditer := &mockAuditer{}

		ctx := newContext(context.TODO(), request)
		err := dispatchUpdateOperation(ctx, processorFinder, testmodel.Manager(), nil, nil, nil, nil, auditer, false, nil)

		expectedError := "error 400 (bahamut): Bad Request: unable to decode application/json: json decode error [pos 1]: only encoded map or array can be decoded into a struct"
		expectedNbCalls := 1

		Convey("Then I should get a bahamut error and no context", func() {
			So(err.Error(), ShouldEqual, expectedError)
			So(auditer.GetCallCount(), ShouldEqual, expectedNbCalls)
		})
	})

	Convey("Given I have a processor that handle ProcessUpdate function with an invalid object", t, func() {
		request := elemental.NewRequest()
		request.Identity = testmodel.TaskIdentity
		request.Data = []byte(`{"ID": "1234", "name": "Fake", "status": "not-good"}`)

		processorFinder := func(identity elemental.Identity) (Processor, error) {
			return &mockProcessor{}, nil
		}

		auditer := &mockAuditer{}

		ctx := newContext(context.TODO(), request)
		err := dispatchUpdateOperation(ctx, processorFinder, testmodel.Manager(), nil, nil, nil, nil, auditer, false, nil)

		expectedError := "error 422 (elemental): Validation Error: Data 'not-good' of attribute 'status' is not in list '[DONE PROGRESS TODO]'"
		expectedNbCalls := 1

		Convey("Then I should get a bahamut error and no context", func() {
			So(err.Error(), ShouldEqual, expectedError)
			So(auditer.GetCallCount(), ShouldEqual, expectedNbCalls)
		})
	})

	Convey("Given I have a processor that does not handle ProcessUpdate function", t, func() {
		request := elemental.NewRequest()
		request.Operation = elemental.OperationRetrieveMany
		request.Identity = elemental.MakeIdentity("Fake", "Test")

		processorFinder := func(identity elemental.Identity) (Processor, error) {
			return &mockEmptyProcessor{}, nil
		}

		auditer := &mockAuditer{}

		expectedError := "error 501 (bahamut): Not implemented: No handler for operation retrieve-many on Fake"
		expectedNbCalls := 1

		ctx := newContext(context.TODO(), request)
		err := dispatchUpdateOperation(ctx, processorFinder, testmodel.Manager(), nil, nil, nil, nil, auditer, false, nil)

		Convey("Then I should get a bahamut error and no context", func() {
			So(err.Error(), ShouldEqual, expectedError)
			So(auditer.GetCallCount(), ShouldEqual, expectedNbCalls)
		})
	})

	Convey("Given I have a processor that does not handle ProcessUpdate function and an authenticator that is not authenticated", t, func() {
		request := elemental.NewRequest()

		processorFinder := func(identity elemental.Identity) (Processor, error) {
			return &mockEmptyProcessor{}, nil
		}

		authenticators := []RequestAuthenticator{
			&mockAuth{
				action: AuthActionContinue,
			},
			&mockAuth{
				errored: true,
				err:     elemental.NewError("Error", "Authenticator does not authenticate.", "bahamut-test", http.StatusInternalServerError),
			},
		}

		auditer := &mockAuditer{}

		expectedError := "error 500 (bahamut-test): Error: Authenticator does not authenticate."
		expectedNbCalls := 1

		ctx := newContext(context.TODO(), request)
		err := dispatchUpdateOperation(ctx, processorFinder, testmodel.Manager(), nil, authenticators, nil, nil, auditer, false, nil)

		Convey("Then I should get a bahamut error and no context", func() {
			So(err.Error(), ShouldEqual, expectedError)
			So(auditer.GetCallCount(), ShouldEqual, expectedNbCalls)
		})
	})

	Convey("Given I have a processor that does not handle ProcessUpdate function and an authorizer that is not authorize", t, func() {
		request := elemental.NewRequest()

		processorFinder := func(identity elemental.Identity) (Processor, error) {
			return &mockEmptyProcessor{}, nil
		}

		authenticators := []RequestAuthenticator{
			&mockAuth{
				action: AuthActionContinue,
			},
			&mockAuth{
				action: AuthActionContinue,
			},
		}

		authorizers := []Authorizer{
			&mockAuth{
				action: AuthActionContinue,
			},
			&mockAuth{
				errored: true,
				err:     elemental.NewError("Error", "Authorizer does not authorize.", "bahamut-test", http.StatusInternalServerError),
			},
		}

		auditer := &mockAuditer{}

		expectedError := "error 500 (bahamut-test): Error: Authorizer does not authorize."
		expectedNbCalls := 1

		ctx := newContext(context.TODO(), request)
		err := dispatchUpdateOperation(ctx, processorFinder, testmodel.Manager(), nil, authenticators, authorizers, nil, auditer, false, nil)

		Convey("Then I should get a bahamut error and no context", func() {
			So(err.Error(), ShouldEqual, expectedError)
			So(auditer.GetCallCount(), ShouldEqual, expectedNbCalls)
		})
	})

	Convey("Given I have a processor with custom unmarshal function that fails", t, func() {
		request := elemental.NewRequest()
		request.Operation = elemental.OperationUpdate
		request.Identity = elemental.MakeIdentity("Fake", "Test")

		processorFinder := func(identity elemental.Identity) (Processor, error) {
			return &mockProcessor{}, nil
		}

		expectedError := "boom"

		ctx := newContext(context.TODO(), request)
		err := dispatchUpdateOperation(
			ctx,
			processorFinder,
			testmodel.Manager(),
			func(*elemental.Request) (elemental.Identifiable, error) {
				return nil, fmt.Errorf("boom")
			},
			nil,
			nil,
			nil,
			nil,
			false,
			nil,
		)

		Convey("Then I should get a bahamut error and no context", func() {
			So(err.Error(), ShouldEqual, expectedError)
		})
	})
}

// TestDispatchers_dispatchDeleteOperation tests dispatchDeleteOperation method
func TestDispatchers_dispatchDeleteOperation(t *testing.T) {

	Convey("Given I have a processor that handle ProcessDelete function", t, func() {
		request := elemental.NewRequest()
		request.Data = []byte(`{"ID": "1234", "name": "Fake"}`)

		processorFinder := func(identity elemental.Identity) (Processor, error) {
			return &mockProcessor{
				output: &testmodel.List{ID: "a"},
				events: []*elemental.Event{elemental.NewEvent(elemental.EventCreate, &testmodel.List{})},
			}, nil
		}

		auditer := &mockAuditer{}
		pusher := &mockPusher{}

		ctx := newContext(context.TODO(), request)
		err := dispatchDeleteOperation(ctx, processorFinder, nil, nil, pusher.Push, auditer, false, nil)

		expectedNbCalls := 1

		Convey("Then I should have no error and context should be initiated", func() {
			So(err, ShouldBeNil)
			So(auditer.GetCallCount(), ShouldEqual, expectedNbCalls)
			So(ctx.outputData, ShouldResemble, &testmodel.List{ID: "a"})
			So(len(pusher.events), ShouldEqual, 2)
			So(pusher.events[0].Type, ShouldEqual, elemental.EventCreate)
			So(pusher.events[1].Type, ShouldEqual, elemental.EventDelete)
		})
	})

	Convey("Setup request and fresh context", t, func() {

		request := elemental.NewRequest()
		request.Data = []byte(`{"ID": "1234", "name": "Fake"}`)
		ctx := newContext(context.TODO(), request)

		auditer := &mockAuditer{}
		pusher := &mockPusher{}

		// bug fix: https://github.com/aporeto-inc/bahamut/issues/64
		Convey("Given I have a processor that handle ProcessDelete function with a context output that does not satisfy elemental.Identifiable", func() {

			processorFinder := func(identity elemental.Identity) (Processor, error) {
				return &mockProcessor{
					// notice how this output will NOT satisfy the elemental.Identifiable interface
					output: json.RawMessage("some random bytes!"),
				}, nil
			}

			Convey("Then I should not panic no events should be pushed", func() {
				var err error
				So(func() {
					err = dispatchDeleteOperation(ctx, processorFinder, nil, nil, pusher.Push, auditer, false, nil)
				}, ShouldNotPanic)
				So(err, ShouldBeNil)
				So(auditer.GetCallCount(), ShouldEqual, 1)
				So(ctx.outputData, ShouldResemble, json.RawMessage("some random bytes!"))
				So(len(pusher.events), ShouldEqual, 0)
			})
		})

		Convey("Given I have a processor that handle ProcessDelete function with a context output that contains a nil elemental.Identifiable", func() {

			// notice how this is a type that satisfies the elemental.Identifiable interface, but it is not a nil interface!
			var testIdentifiable *testmodel.List
			var _ elemental.Identifiable = testIdentifiable
			ctx.SetOutputData(testIdentifiable)

			processorFinder := func(identity elemental.Identity) (Processor, error) {
				return &mockProcessor{
					// notice how this output satisfies the elemental.Identifiable interface
					output: testIdentifiable,
				}, nil
			}

			Convey("Then I should not panic and an event should be pushed", func() {
				var err error
				So(func() {
					err = dispatchDeleteOperation(ctx, processorFinder, nil, nil, pusher.Push, auditer, false, nil)
				}, ShouldNotPanic)
				So(err, ShouldBeNil)
				So(auditer.GetCallCount(), ShouldEqual, 1)
				So(ctx.outputData, ShouldResemble, (*testmodel.List)(nil))
				So(len(pusher.events), ShouldEqual, 1)
			})
		})
	})

	Convey("Given I have a processor that handle ProcessDelete function with read only mode", t, func() {
		request := elemental.NewRequest()
		request.Data = []byte(`{"ID": "1234", "name": "Fake"}`)

		processorFinder := func(identity elemental.Identity) (Processor, error) {
			return &mockProcessor{}, nil
		}

		auditer := &mockAuditer{}

		ctx := newContext(context.TODO(), request)
		err := dispatchDeleteOperation(ctx, processorFinder, nil, nil, nil, auditer, true, nil)

		Convey("Then I should have a 423 error and context should be nil", func() {
			So(err, ShouldNotBeNil)
			So(err.(elemental.Error).Code, ShouldEqual, http.StatusLocked)
		})
	})

	Convey("Given I have a processor that handle ProcessDelete function with error", t, func() {
		request := elemental.NewRequest()
		request.Data = []byte(`{"ID": "1234", "name": "Fake"}`)

		processorFinder := func(identity elemental.Identity) (Processor, error) {
			return &mockProcessor{
				err: elemental.NewError("Error", "Bad request.", "bahamut-test", http.StatusBadRequest),
			}, nil
		}

		auditer := &mockAuditer{}

		ctx := newContext(context.TODO(), request)
		err := dispatchDeleteOperation(ctx, processorFinder, nil, nil, nil, auditer, false, nil)

		expectedError := "error 400 (bahamut-test): Error: Bad request."
		expectedNbCalls := 1

		Convey("Then I should get a bahamut error and no context", func() {
			So(err.Error(), ShouldEqual, expectedError)
			So(auditer.GetCallCount(), ShouldEqual, expectedNbCalls)
		})
	})

	Convey("Given I have a processor that does not handle ProcessDelete function", t, func() {
		request := elemental.NewRequest()
		request.Operation = elemental.OperationRetrieveMany
		request.Identity = elemental.MakeIdentity("Fake", "Test")

		processorFinder := func(identity elemental.Identity) (Processor, error) {
			return &mockEmptyProcessor{}, nil
		}

		auditer := &mockAuditer{}

		expectedError := "error 501 (bahamut): Not implemented: No handler for operation retrieve-many on Fake"
		expectedNbCalls := 1

		ctx := newContext(context.TODO(), request)
		err := dispatchDeleteOperation(ctx, processorFinder, nil, nil, nil, auditer, false, nil)

		Convey("Then I should get a bahamut error and no context", func() {
			So(err.Error(), ShouldEqual, expectedError)
			So(auditer.GetCallCount(), ShouldEqual, expectedNbCalls)
		})
	})

	Convey("Given I have a processor that does not handle ProcessDelete function and an authenticator that is not authenticated", t, func() {
		request := elemental.NewRequest()

		processorFinder := func(identity elemental.Identity) (Processor, error) {
			return &mockEmptyProcessor{}, nil
		}

		authenticators := []RequestAuthenticator{
			&mockAuth{
				action: AuthActionContinue,
			},
			&mockAuth{
				errored: true,
				err:     elemental.NewError("Error", "Authenticator does not authenticate.", "bahamut-test", http.StatusInternalServerError),
			},
		}

		auditer := &mockAuditer{}

		expectedError := "error 500 (bahamut-test): Error: Authenticator does not authenticate."
		expectedNbCalls := 1

		ctx := newContext(context.TODO(), request)
		err := dispatchDeleteOperation(ctx, processorFinder, authenticators, nil, nil, auditer, false, nil)

		Convey("Then I should get a bahamut error and no context", func() {
			So(err.Error(), ShouldEqual, expectedError)
			So(auditer.GetCallCount(), ShouldEqual, expectedNbCalls)
		})
	})

	Convey("Given I have a processor that does not handle ProcessDelete function and an authorizer that is not authorize", t, func() {
		request := elemental.NewRequest()

		processorFinder := func(identity elemental.Identity) (Processor, error) {
			return &mockEmptyProcessor{}, nil
		}

		authenticators := []RequestAuthenticator{
			&mockAuth{
				action: AuthActionContinue,
			},
			&mockAuth{
				action: AuthActionContinue,
			},
		}

		authorizers := []Authorizer{
			&mockAuth{
				action: AuthActionContinue,
			},
			&mockAuth{
				errored: true,
				err:     elemental.NewError("Error", "Authorizer does not authorize.", "bahamut-test", http.StatusInternalServerError),
			},
		}

		auditer := &mockAuditer{}

		expectedError := "error 500 (bahamut-test): Error: Authorizer does not authorize."
		expectedNbCalls := 1

		ctx := newContext(context.TODO(), request)
		err := dispatchDeleteOperation(ctx, processorFinder, authenticators, authorizers, nil, auditer, false, nil)

		Convey("Then I should get a bahamut error and no context", func() {
			So(err.Error(), ShouldEqual, expectedError)
			So(auditer.GetCallCount(), ShouldEqual, expectedNbCalls)
		})
	})
}

// TestDispatchers_dispatchPatchOperation tests dispatchPatchOperation method
func TestDispatchers_dispatchPatchOperation(t *testing.T) {

	Convey("Given I have a processor that handle ProcessPatch function", t, func() {
		request := elemental.NewRequest()
		request.Identity = testmodel.ListIdentity
		request.Data = []byte(`{"ID": "1234", "name": "Fake"}`)

		expectedID := "a"
		expectedName := "Fake"

		processorFinder := func(identity elemental.Identity) (Processor, error) {
			return &mockProcessor{
				output: &testmodel.SparseList{ID: &expectedID, Name: &expectedName},
				events: []*elemental.Event{elemental.NewEvent(elemental.EventDelete, &testmodel.List{})},
			}, nil
		}

		auditer := &mockAuditer{}
		pusher := &mockPusher{}

		ctx := newContext(context.TODO(), request)
		err := dispatchPatchOperation(ctx, processorFinder, testmodel.Manager(), nil, nil, nil, pusher.Push, auditer, false, nil)

		expectedNbCalls := 1

		Convey("Then I should have no error and context should be initiated", func() {
			So(err, ShouldBeNil)
			So(auditer.GetCallCount(), ShouldEqual, expectedNbCalls)
			So(ctx.outputData, ShouldResemble, &testmodel.SparseList{ID: &expectedID, Name: &expectedName})
			So(len(pusher.events), ShouldEqual, 2)
			So(pusher.events[0].Type, ShouldEqual, elemental.EventDelete)
			So(pusher.events[1].Type, ShouldEqual, elemental.EventUpdate)
		})
	})

	Convey("Setup request and fresh context", t, func() {

		request := elemental.NewRequest()
		request.Identity = testmodel.ListIdentity
		request.Data = []byte(`{"ID": "1234", "name": "Fake"}`)
		ctx := newContext(context.TODO(), request)

		auditer := &mockAuditer{}
		pusher := &mockPusher{}

		// bug fix: https://github.com/aporeto-inc/bahamut/issues/64
		Convey("Given I have a processor that handle ProcessPatch function with a context output that does not satisfy elemental.Identifiable", func() {

			processorFinder := func(identity elemental.Identity) (Processor, error) {
				return &mockProcessor{
					// notice how this output will NOT satisfy the elemental.Identifiable interface
					output: json.RawMessage("some random bytes!"),
				}, nil
			}

			Convey("Then I should not panic no events should be pushed", func() {
				var err error
				So(func() {
					err = dispatchPatchOperation(ctx, processorFinder, testmodel.Manager(), nil, nil, nil, pusher.Push, auditer, false, nil)
				}, ShouldNotPanic)
				So(err, ShouldBeNil)
				So(auditer.GetCallCount(), ShouldEqual, 1)
				So(ctx.outputData, ShouldResemble, json.RawMessage("some random bytes!"))
				So(len(pusher.events), ShouldEqual, 0)
			})
		})

		Convey("Given I have a processor that handle ProcessPatch function with a context output that contains a nil elemental.Identifiable", func() {

			// notice how this is a type that satisfies the elemental.Identifiable interface, but it is not a nil interface!
			var testIdentifiable *testmodel.List
			var _ elemental.Identifiable = testIdentifiable
			ctx.SetOutputData(testIdentifiable)

			processorFinder := func(identity elemental.Identity) (Processor, error) {
				return &mockProcessor{
					// notice how this output satisfies the elemental.Identifiable interface
					output: testIdentifiable,
				}, nil
			}

			Convey("Then I should not panic and an event should be pushed", func() {
				var err error
				So(func() {
					err = dispatchPatchOperation(ctx, processorFinder, testmodel.Manager(), nil, nil, nil, pusher.Push, auditer, false, nil)
				}, ShouldNotPanic)
				So(err, ShouldBeNil)
				So(auditer.GetCallCount(), ShouldEqual, 1)
				So(ctx.outputData, ShouldResemble, (*testmodel.List)(nil))
				So(len(pusher.events), ShouldEqual, 1)
			})
		})
	})

	Convey("Given I have a processor that handle ProcessPatch function with an invalid JSON", t, func() {
		request := elemental.NewRequest()
		request.Identity = testmodel.ListIdentity
		request.Data = []byte(`Invalid JSON`)

		processorFinder := func(identity elemental.Identity) (Processor, error) {
			return &mockProcessor{}, nil
		}

		auditer := &mockAuditer{}

		ctx := newContext(context.TODO(), request)
		err := dispatchPatchOperation(ctx, processorFinder, testmodel.Manager(), nil, nil, nil, nil, auditer, false, nil)

		expectedError := "error 400 (bahamut): Bad Request: unable to decode application/json: json decode error [pos 1]: only encoded map or array can be decoded into a struct"
		expectedNbCalls := 1

		Convey("Then I should have no error and context should be initiated", func() {
			So(err.Error(), ShouldEqual, expectedError)
			So(auditer.GetCallCount(), ShouldEqual, expectedNbCalls)
		})
	})

	Convey("Given I have a processor that handle ProcessPatch function with error", t, func() {
		request := elemental.NewRequest()
		request.Identity = testmodel.ListIdentity
		request.Data = []byte(`{"ID": "1234", "name": "Fake"}`)

		processorFinder := func(identity elemental.Identity) (Processor, error) {
			return &mockProcessor{
				err: elemental.NewError("Error", "Bad request.", "bahamut-test", http.StatusBadRequest),
			}, nil
		}

		auditer := &mockAuditer{}

		ctx := newContext(context.TODO(), request)
		err := dispatchPatchOperation(ctx, processorFinder, testmodel.Manager(), nil, nil, nil, nil, auditer, false, nil)

		expectedError := "error 400 (bahamut-test): Error: Bad request."
		expectedNbCalls := 1

		Convey("Then I should get a bahamut error and no context", func() {
			So(err.Error(), ShouldEqual, expectedError)
			So(auditer.GetCallCount(), ShouldEqual, expectedNbCalls)
		})
	})

	Convey("Given I have a processor that does not handle ProcessPatch function", t, func() {
		request := elemental.NewRequest()
		request.Identity = testmodel.ListIdentity
		request.Operation = elemental.OperationRetrieveMany
		request.Identity = elemental.MakeIdentity("Fake", "Test")

		processorFinder := func(identity elemental.Identity) (Processor, error) {
			return &mockEmptyProcessor{}, nil
		}

		auditer := &mockAuditer{}

		expectedError := "error 501 (bahamut): Not implemented: No handler for operation retrieve-many on Fake"
		expectedNbCalls := 1

		ctx := newContext(context.TODO(), request)
		err := dispatchPatchOperation(ctx, processorFinder, testmodel.Manager(), nil, nil, nil, nil, auditer, false, nil)

		Convey("Then I should get a bahamut error and no context", func() {
			So(err.Error(), ShouldEqual, expectedError)
			So(auditer.GetCallCount(), ShouldEqual, expectedNbCalls)
		})
	})

	Convey("Given I have a processor that does not handle ProcessPatch function and an authenticator that is not authenticated", t, func() {
		request := elemental.NewRequest()
		request.Identity = testmodel.ListIdentity

		processorFinder := func(identity elemental.Identity) (Processor, error) {
			return &mockEmptyProcessor{}, nil
		}

		authenticators := []RequestAuthenticator{
			&mockAuth{
				action: AuthActionContinue,
			},
			&mockAuth{
				errored: true,
				err:     elemental.NewError("Error", "Authenticator does not authenticate.", "bahamut-test", http.StatusInternalServerError),
			},
		}

		auditer := &mockAuditer{}

		expectedError := "error 500 (bahamut-test): Error: Authenticator does not authenticate."
		expectedNbCalls := 1

		ctx := newContext(context.TODO(), request)
		err := dispatchPatchOperation(ctx, processorFinder, testmodel.Manager(), nil, authenticators, nil, nil, auditer, false, nil)

		Convey("Then I should get a bahamut error and no context", func() {
			So(err.Error(), ShouldEqual, expectedError)
			So(auditer.GetCallCount(), ShouldEqual, expectedNbCalls)
		})
	})

	Convey("Given I have a processor that does not handle ProcessPatch function and an authorizer that is not authorize", t, func() {
		request := elemental.NewRequest()
		request.Identity = testmodel.ListIdentity

		processorFinder := func(identity elemental.Identity) (Processor, error) {
			return &mockEmptyProcessor{}, nil
		}

		authenticators := []RequestAuthenticator{
			&mockAuth{
				action: AuthActionContinue,
			},
			&mockAuth{
				action: AuthActionContinue,
			},
		}

		authorizers := []Authorizer{
			&mockAuth{
				action: AuthActionContinue,
			},
			&mockAuth{
				errored: true,
				err:     elemental.NewError("Error", "Authorizer does not authorize.", "bahamut-test", http.StatusInternalServerError),
			},
		}

		auditer := &mockAuditer{}

		expectedError := "error 500 (bahamut-test): Error: Authorizer does not authorize."
		expectedNbCalls := 1

		ctx := newContext(context.TODO(), request)
		err := dispatchPatchOperation(ctx, processorFinder, testmodel.Manager(), nil, authenticators, authorizers, nil, auditer, false, nil)

		Convey("Then I should get a bahamut error and no context", func() {
			So(err.Error(), ShouldEqual, expectedError)
			So(auditer.GetCallCount(), ShouldEqual, expectedNbCalls)
		})
	})

	Convey("Given I have a processor that handle ProcessPatch function with read only mode", t, func() {
		request := elemental.NewRequest()
		request.Identity = testmodel.ListIdentity
		request.Data = []byte(`{"ID": "1234", "name": "Fake"}`)

		processorFinder := func(identity elemental.Identity) (Processor, error) {
			return &mockProcessor{}, nil
		}

		auditer := &mockAuditer{}

		ctx := newContext(context.TODO(), request)
		err := dispatchPatchOperation(ctx, processorFinder, testmodel.Manager(), nil, nil, nil, nil, auditer, true, nil)

		Convey("Then I should have a 423 error and context should be nil", func() {
			So(err, ShouldNotBeNil)
			So(err.(elemental.Error).Code, ShouldEqual, http.StatusLocked)
		})
	})

	Convey("Given I have a processor with custom unmarshal function that fails", t, func() {
		request := elemental.NewRequest()
		request.Operation = elemental.OperationUpdate
		request.Identity = elemental.MakeIdentity("Fake", "Test")

		processorFinder := func(identity elemental.Identity) (Processor, error) {
			return &mockProcessor{}, nil
		}

		expectedError := "boom"

		ctx := newContext(context.TODO(), request)
		err := dispatchPatchOperation(
			ctx,
			processorFinder,
			testmodel.Manager(),
			func(*elemental.Request) (elemental.Identifiable, error) {
				return nil, fmt.Errorf("boom")
			},
			nil,
			nil,
			nil,
			nil,
			false,
			nil,
		)

		Convey("Then I should get a bahamut error and no context", func() {
			So(err.Error(), ShouldEqual, expectedError)
		})
	})
}

// TestDispatchers_dispatchInfoOperation tests dispatchInfoOperation method
func TestDispatchers_dispatchInfoOperation(t *testing.T) {

	Convey("Given I have a processor that handle ProcessInfo function", t, func() {
		request := elemental.NewRequest()
		request.Data = []byte(`{"ID": "1234", "name": "Fake"}`)

		processorFinder := func(identity elemental.Identity) (Processor, error) {
			return &mockProcessor{
				events: []*elemental.Event{elemental.NewEvent(elemental.EventUpdate, &testmodel.List{})},
			}, nil
		}

		auditer := &mockAuditer{}
		pusher := &mockPusher{}

		ctx := newContext(context.TODO(), request)
		err := dispatchInfoOperation(ctx, processorFinder, nil, nil, pusher.Push, auditer)

		expectedNbCalls := 1

		Convey("Then I should have no error and context should be initiated", func() {
			So(err, ShouldBeNil)
			So(auditer.GetCallCount(), ShouldEqual, expectedNbCalls)
			So(pusher.events[0].Type, ShouldEqual, elemental.EventUpdate)
		})
	})

	Convey("Given I have a processor that handle ProcessInfo function with error", t, func() {
		request := elemental.NewRequest()
		request.Data = []byte(`{"ID": "1234", "name": "Fake"}`)

		processorFinder := func(identity elemental.Identity) (Processor, error) {
			return &mockProcessor{
				err: elemental.NewError("Error", "Bad request.", "bahamut-test", http.StatusBadRequest),
			}, nil
		}

		auditer := &mockAuditer{}
		pusher := &mockPusher{}

		ctx := newContext(context.TODO(), request)
		err := dispatchInfoOperation(ctx, processorFinder, nil, nil, pusher.Push, auditer)

		expectedError := "error 400 (bahamut-test): Error: Bad request."
		expectedNbCalls := 1

		Convey("Then I should get a bahamut error and no context", func() {
			So(err.Error(), ShouldEqual, expectedError)
			So(auditer.GetCallCount(), ShouldEqual, expectedNbCalls)
		})
	})

	Convey("Given I have a processor that does not handle ProcessInfo function", t, func() {
		request := elemental.NewRequest()
		request.Operation = elemental.OperationRetrieveMany
		request.Identity = elemental.MakeIdentity("Fake", "Test")

		processorFinder := func(identity elemental.Identity) (Processor, error) {
			return &mockEmptyProcessor{}, nil
		}

		auditer := &mockAuditer{}
		pusher := &mockPusher{}

		expectedError := "error 501 (bahamut): Not implemented: No handler for operation retrieve-many on Fake"
		expectedNbCalls := 1

		ctx := newContext(context.TODO(), request)
		err := dispatchInfoOperation(ctx, processorFinder, nil, nil, pusher.Push, auditer)

		Convey("Then I should get a bahamut error and no context", func() {
			So(err.Error(), ShouldEqual, expectedError)
			So(auditer.GetCallCount(), ShouldEqual, expectedNbCalls)
		})
	})

	Convey("Given I have a processor that does not handle ProcessInfo function and an authenticator that is not authenticated", t, func() {
		request := elemental.NewRequest()

		processorFinder := func(identity elemental.Identity) (Processor, error) {
			return &mockEmptyProcessor{}, nil
		}

		authenticators := []RequestAuthenticator{
			&mockAuth{
				action: AuthActionContinue,
			},
			&mockAuth{
				errored: true,
				err:     elemental.NewError("Error", "Authenticator does not authenticate.", "bahamut-test", http.StatusInternalServerError),
			},
		}

		auditer := &mockAuditer{}
		pusher := &mockPusher{}

		expectedError := "error 500 (bahamut-test): Error: Authenticator does not authenticate."
		expectedNbCalls := 1

		ctx := newContext(context.TODO(), request)
		err := dispatchInfoOperation(ctx, processorFinder, authenticators, nil, pusher.Push, auditer)

		Convey("Then I should get a bahamut error and no context", func() {
			So(err.Error(), ShouldEqual, expectedError)
			So(auditer.GetCallCount(), ShouldEqual, expectedNbCalls)
		})
	})

	Convey("Given I have a processor that does not handle ProcessInfo function and an authorizer that is not authorize", t, func() {
		request := elemental.NewRequest()

		processorFinder := func(identity elemental.Identity) (Processor, error) {
			return &mockEmptyProcessor{}, nil
		}

		authenticators := []RequestAuthenticator{
			&mockAuth{
				action: AuthActionContinue,
			},
			&mockAuth{
				action: AuthActionContinue,
			},
		}

		authorizers := []Authorizer{
			&mockAuth{
				action: AuthActionContinue,
			},
			&mockAuth{
				errored: true,
				err:     elemental.NewError("Error", "Authorizer does not authorize.", "bahamut-test", http.StatusInternalServerError),
			},
		}

		auditer := &mockAuditer{}
		pusher := &mockPusher{}

		expectedError := "error 500 (bahamut-test): Error: Authorizer does not authorize."
		expectedNbCalls := 1

		ctx := newContext(context.TODO(), request)
		err := dispatchInfoOperation(ctx, processorFinder, authenticators, authorizers, pusher.Push, auditer)

		Convey("Then I should get a bahamut error and no context", func() {
			So(err.Error(), ShouldEqual, expectedError)
			So(auditer.GetCallCount(), ShouldEqual, expectedNbCalls)
		})
	})
}

func TestDispatchers_makeReadOnlyError(t *testing.T) {

	Convey("Given I have an exclustion list", t, func() {

		ex := []elemental.Identity{testmodel.ListIdentity}

		Convey("When I call makeReadOnlyError on an identity that is not excluded", func() {

			err := makeReadOnlyError(testmodel.UserIdentity, ex)

			Convey("Then err should not be nil", func() {
				So(err, ShouldNotBeNil)
			})
		})

		Convey("When I call makeReadOnlyError on an identity that is excluded", func() {

			err := makeReadOnlyError(testmodel.ListIdentity, ex)

			Convey("Then err should be nil", func() {
				So(err, ShouldBeNil)
			})
		})
	})
}
