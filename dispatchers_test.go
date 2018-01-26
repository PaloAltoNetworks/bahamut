package bahamut

import (
	"net/http"
	"sync"
	"testing"

	"github.com/aporeto-inc/elemental"
	"github.com/aporeto-inc/elemental/test/model"
	. "github.com/smartystreets/goconvey/convey"
)

// FakeAuditer
type FakeAuditer struct {
	nbCalls int
}

func (p *FakeAuditer) Audit(*Context, error) {
	p.nbCalls++
}

// FakeIdentifiable
type FakeIdentifiable struct {
	identity        string
	identifier      string
	validationError error
}

func (p *FakeIdentifiable) Identity() elemental.Identity {
	return elemental.MakeIdentity(p.identity, "FakeCategory")
}
func (p *FakeIdentifiable) Identifier() string {
	return p.identifier
}
func (p *FakeIdentifiable) SetIdentifier(identifier string) {
	p.identifier = identifier
}
func (p *FakeIdentifiable) Version() int {
	return 1
}
func (p *FakeIdentifiable) Validate() error {
	return p.validationError
}

// FakeCompleteProcessor
type FakeCompleteProcessor struct {
	err    error
	output interface{}
	events []*elemental.Event
}

func (p *FakeCompleteProcessor) ProcessRetrieveMany(ctx *Context) error {
	ctx.OutputData = p.output
	ctx.EnqueueEvents(p.events...)
	return p.err
}
func (p *FakeCompleteProcessor) ProcessRetrieve(ctx *Context) error {
	ctx.OutputData = p.output
	ctx.EnqueueEvents(p.events...)
	return p.err
}
func (p *FakeCompleteProcessor) ProcessCreate(ctx *Context) error {
	ctx.OutputData = p.output
	ctx.EnqueueEvents(p.events...)
	return p.err
}
func (p *FakeCompleteProcessor) ProcessUpdate(ctx *Context) error {
	ctx.OutputData = p.output
	ctx.EnqueueEvents(p.events...)
	return p.err
}
func (p *FakeCompleteProcessor) ProcessDelete(ctx *Context) error {
	ctx.OutputData = p.output
	ctx.EnqueueEvents(p.events...)
	return p.err
}
func (p *FakeCompleteProcessor) ProcessPatch(ctx *Context) error {
	ctx.OutputData = p.output
	ctx.EnqueueEvents(p.events...)
	return p.err
}
func (p *FakeCompleteProcessor) ProcessInfo(ctx *Context) error {
	ctx.OutputData = p.output
	ctx.EnqueueEvents(p.events...)
	return p.err
}

type FakePusher struct {
	events []*elemental.Event
	sync.Mutex
}

func (f *FakePusher) Push(evt ...*elemental.Event) {
	f.Lock()
	defer f.Unlock()
	f.events = append(f.events, evt...)
}

// TestDispatchers_dispatchRetrieveManyOperation tests dispatchRetrieveManyOperation method
func TestDispatchers_dispatchRetrieveManyOperation(t *testing.T) {

	Convey("Given I have a processor that handle ProcessRetrieveMany function", t, func() {
		request := elemental.NewRequest()

		processorFinder := func(identity elemental.Identity) (Processor, error) {
			return &FakeCompleteProcessor{
				output: "hello",
				events: []*elemental.Event{elemental.NewEvent(elemental.EventUpdate, &FakeIdentifiable{})},
			}, nil
		}

		factory := func(identity string, version int) elemental.Identifiable {
			return &FakeIdentifiable{}
		}

		auditer := &FakeAuditer{}
		pusher := &FakePusher{}

		ctx := NewContextWithRequest(request)
		err := dispatchRetrieveManyOperation(ctx, processorFinder, factory, nil, nil, pusher.Push, auditer)

		expectedNbCalls := 1

		Convey("Then I should have no error and context should be initiated", func() {
			So(err, ShouldBeNil)
			So(auditer.nbCalls, ShouldEqual, expectedNbCalls)
			So(ctx.OutputData, ShouldResemble, "hello")
			So(len(pusher.events), ShouldEqual, 1)
			So(pusher.events[0].Type, ShouldEqual, elemental.EventUpdate)
		})
	})

	Convey("Given I have a processor that handle ProcessRetrieveMany function with error", t, func() {
		request := elemental.NewRequest()

		processorFinder := func(identity elemental.Identity) (Processor, error) {
			return &FakeCompleteProcessor{
				err: elemental.NewError("Error", "Bad request.", "bahamut-test", http.StatusBadRequest),
			}, nil
		}

		factory := func(identity string, version int) elemental.Identifiable {
			return &FakeIdentifiable{}
		}

		auditer := &FakeAuditer{}

		ctx := NewContextWithRequest(request)
		err := dispatchRetrieveManyOperation(ctx, processorFinder, factory, nil, nil, nil, auditer)

		expectedError := "error 400 (bahamut-test): Error: Bad request."
		expectedNbCalls := 1

		Convey("Then I should get a bahamut error and no context", func() {
			So(err.Error(), ShouldEqual, expectedError)
			So(auditer.nbCalls, ShouldEqual, expectedNbCalls)
		})
	})

	Convey("Given I have a processor that does not handle ProcessRetrieveMany function", t, func() {
		request := elemental.NewRequest()
		request.Operation = elemental.OperationRetrieveMany
		request.Identity = elemental.MakeIdentity("Fake", "Test")

		processorFinder := func(identity elemental.Identity) (Processor, error) {
			return &FakeProcessor{}, nil
		}

		factory := func(identity string, version int) elemental.Identifiable {
			return &FakeIdentifiable{}
		}

		auditer := &FakeAuditer{}

		expectedError := "error 501 (bahamut): Not implemented: No handler for operation retrieve-many on Fake"
		expectedNbCalls := 1

		ctx := NewContextWithRequest(request)
		err := dispatchRetrieveManyOperation(ctx, processorFinder, factory, nil, nil, nil, auditer)

		Convey("Then I should get a bahamut error and no context", func() {
			So(err.Error(), ShouldEqual, expectedError)
			So(auditer.nbCalls, ShouldEqual, expectedNbCalls)
		})
	})

	Convey("Given I have a processor that handle ProcessRetrieveMany function and an authenticator that is not authenticated", t, func() {
		request := elemental.NewRequest()

		processorFinder := func(identity elemental.Identity) (Processor, error) {
			return &FakeProcessor{}, nil
		}

		factory := func(identity string, version int) elemental.Identifiable {
			return &FakeIdentifiable{}
		}

		authenticators := []RequestAuthenticator{
			&Auth{
				authenticated: true,
			},
			&Auth{
				errored: true,
				err:     elemental.NewError("Error", "Authenticator does not authenticate.", "bahamut-test", http.StatusInternalServerError),
			},
		}

		auditer := &FakeAuditer{}

		expectedError := "error 500 (bahamut-test): Error: Authenticator does not authenticate."
		expectedNbCalls := 1

		ctx := NewContextWithRequest(request)
		err := dispatchRetrieveManyOperation(ctx, processorFinder, factory, authenticators, nil, nil, auditer)

		Convey("Then I should get a bahamut error and no context", func() {
			So(err.Error(), ShouldEqual, expectedError)
			So(auditer.nbCalls, ShouldEqual, expectedNbCalls)
		})
	})

	Convey("Given I have a processor that handle ProcessRetrieveMany function and an authorizer that is not authorize", t, func() {
		request := elemental.NewRequest()

		processorFinder := func(identity elemental.Identity) (Processor, error) {
			return &FakeProcessor{}, nil
		}

		factory := func(identity string, version int) elemental.Identifiable {
			return &FakeIdentifiable{}
		}

		authenticators := []RequestAuthenticator{
			&Auth{
				authenticated: true,
			},
			&Auth{
				authenticated: true,
			},
		}

		authorizers := []Authorizer{
			&Auth{
				authorized: true,
			},
			&Auth{
				errored: true,
				err:     elemental.NewError("Error", "Authorizer does not authorize.", "bahamut-test", http.StatusInternalServerError),
			},
		}

		auditer := &FakeAuditer{}

		expectedError := "error 500 (bahamut-test): Error: Authorizer does not authorize."
		expectedNbCalls := 1

		ctx := NewContextWithRequest(request)
		err := dispatchRetrieveManyOperation(ctx, processorFinder, factory, authenticators, authorizers, nil, auditer)

		Convey("Then I should get a bahamut error and no context", func() {
			So(err.Error(), ShouldEqual, expectedError)
			So(auditer.nbCalls, ShouldEqual, expectedNbCalls)
		})
	})
}

// TestDispatchers_dispatchRetrieveOperation tests dispatchRetrieveOperation method
func TestDispatchers_dispatchRetrieveOperation(t *testing.T) {

	Convey("Given I have a processor that handle ProcessRetrieve function", t, func() {
		request := elemental.NewRequest()

		processorFinder := func(identity elemental.Identity) (Processor, error) {
			return &FakeCompleteProcessor{
				output: "hello",
				events: []*elemental.Event{elemental.NewEvent(elemental.EventUpdate, &FakeIdentifiable{})},
			}, nil
		}

		factory := func(identity string, version int) elemental.Identifiable {
			return &FakeIdentifiable{}
		}

		auditer := &FakeAuditer{}
		pusher := &FakePusher{}

		ctx := NewContextWithRequest(request)
		err := dispatchRetrieveOperation(ctx, processorFinder, factory, nil, nil, pusher.Push, auditer)

		expectedNbCalls := 1

		Convey("Then I should have no error and context should be initiated", func() {
			So(err, ShouldBeNil)
			So(auditer.nbCalls, ShouldEqual, expectedNbCalls)
			So(ctx.OutputData, ShouldResemble, "hello")
			So(len(pusher.events), ShouldEqual, 1)
			So(pusher.events[0].Type, ShouldEqual, elemental.EventUpdate)
		})
	})

	Convey("Given I have a processor that handle ProcessRetrieve function with error", t, func() {
		request := elemental.NewRequest()

		processorFinder := func(identity elemental.Identity) (Processor, error) {
			return &FakeCompleteProcessor{
				err: elemental.NewError("Error", "Bad request.", "bahamut-test", http.StatusBadRequest),
			}, nil
		}

		factory := func(identity string, version int) elemental.Identifiable {
			return &FakeIdentifiable{}
		}

		auditer := &FakeAuditer{}

		ctx := NewContextWithRequest(request)
		err := dispatchRetrieveOperation(ctx, processorFinder, factory, nil, nil, nil, auditer)

		expectedError := "error 400 (bahamut-test): Error: Bad request."
		expectedNbCalls := 1

		Convey("Then I should get a bahamut error and no context", func() {
			So(err.Error(), ShouldEqual, expectedError)
			So(auditer.nbCalls, ShouldEqual, expectedNbCalls)
		})
	})

	Convey("Given I have a processor that does not handle ProcessRetrieve function", t, func() {
		request := elemental.NewRequest()
		request.Operation = elemental.OperationRetrieveMany
		request.Identity = elemental.MakeIdentity("Fake", "Test")

		processorFinder := func(identity elemental.Identity) (Processor, error) {
			return &FakeProcessor{}, nil
		}

		factory := func(identity string, version int) elemental.Identifiable {
			return &FakeIdentifiable{}
		}

		auditer := &FakeAuditer{}

		expectedError := "error 501 (bahamut): Not implemented: No handler for operation retrieve-many on Fake"
		expectedNbCalls := 1

		ctx := NewContextWithRequest(request)
		err := dispatchRetrieveOperation(ctx, processorFinder, factory, nil, nil, nil, auditer)

		Convey("Then I should get a bahamut error and no context", func() {
			So(err.Error(), ShouldEqual, expectedError)
			So(auditer.nbCalls, ShouldEqual, expectedNbCalls)
		})
	})

	Convey("Given I have a processor that does not handle ProcessRetrieve function and an authenticator that is not authenticated", t, func() {
		request := elemental.NewRequest()

		processorFinder := func(identity elemental.Identity) (Processor, error) {
			return &FakeProcessor{}, nil
		}

		factory := func(identity string, version int) elemental.Identifiable {
			return &FakeIdentifiable{}
		}

		authenticators := []RequestAuthenticator{
			&Auth{
				authenticated: true,
			},
			&Auth{
				errored: true,
				err:     elemental.NewError("Error", "Authenticator does not authenticate.", "bahamut-test", http.StatusInternalServerError),
			},
		}

		auditer := &FakeAuditer{}

		expectedError := "error 500 (bahamut-test): Error: Authenticator does not authenticate."
		expectedNbCalls := 1

		ctx := NewContextWithRequest(request)
		err := dispatchRetrieveOperation(ctx, processorFinder, factory, authenticators, nil, nil, auditer)

		Convey("Then I should get a bahamut error and no context", func() {
			So(err.Error(), ShouldEqual, expectedError)
			So(auditer.nbCalls, ShouldEqual, expectedNbCalls)
		})
	})

	Convey("Given I have a processor that does not handle ProcessRetrieve function and an authorizer that is not authorize", t, func() {
		request := elemental.NewRequest()

		processorFinder := func(identity elemental.Identity) (Processor, error) {
			return &FakeProcessor{}, nil
		}

		factory := func(identity string, version int) elemental.Identifiable {
			return &FakeIdentifiable{}
		}

		authenticators := []RequestAuthenticator{
			&Auth{
				authenticated: true,
			},
			&Auth{
				authenticated: true,
			},
		}

		authorizers := []Authorizer{
			&Auth{
				authorized: true,
			},
			&Auth{
				errored: true,
				err:     elemental.NewError("Error", "Authorizer does not authorize.", "bahamut-test", http.StatusInternalServerError),
			},
		}

		auditer := &FakeAuditer{}

		expectedError := "error 500 (bahamut-test): Error: Authorizer does not authorize."
		expectedNbCalls := 1

		ctx := NewContextWithRequest(request)
		err := dispatchRetrieveOperation(ctx, processorFinder, factory, authenticators, authorizers, nil, auditer)

		Convey("Then I should get a bahamut error and no context", func() {
			So(err.Error(), ShouldEqual, expectedError)
			So(auditer.nbCalls, ShouldEqual, expectedNbCalls)
		})
	})
}

// TestDispatchers_dispatchCreateOperation tests dispatchCreateOperation method
func TestDispatchers_dispatchCreateOperation(t *testing.T) {

	Convey("Given I have a processor that handle ProcessCreate function", t, func() {
		request := elemental.NewRequest()
		request.Data = []byte(`{"ID": "1234", "Name": "Fake"}`)

		processorFinder := func(identity elemental.Identity) (Processor, error) {
			return &FakeCompleteProcessor{
				output: &FakeIdentifiable{identifier: "a"},
				events: []*elemental.Event{elemental.NewEvent(elemental.EventUpdate, &FakeIdentifiable{})},
			}, nil
		}

		factory := func(identity string, version int) elemental.Identifiable {
			return &FakeIdentifiable{}
		}

		auditer := &FakeAuditer{}
		pusher := &FakePusher{}

		ctx := NewContextWithRequest(request)
		err := dispatchCreateOperation(ctx, processorFinder, factory, nil, nil, pusher.Push, auditer, false, nil)

		expectedNbCalls := 1

		Convey("Then I should have no error and context should be initiated", func() {
			So(err, ShouldBeNil)
			So(ctx.InputData, ShouldNotBeNil)
			So(auditer.nbCalls, ShouldEqual, expectedNbCalls)
			So(ctx.OutputData, ShouldResemble, &FakeIdentifiable{identifier: "a"})
			So(len(pusher.events), ShouldEqual, 2)
			So(pusher.events[0].Type, ShouldEqual, elemental.EventUpdate)
			So(pusher.events[1].Type, ShouldEqual, elemental.EventCreate)
		})
	})

	Convey("Given I have a processor that handle ProcessCreate function with read only mode", t, func() {
		request := elemental.NewRequest()
		request.Data = []byte(`{"ID": "1234", "Name": "Fake"}`)

		processorFinder := func(identity elemental.Identity) (Processor, error) {
			return &FakeCompleteProcessor{}, nil
		}

		factory := func(identity string, version int) elemental.Identifiable {
			return &FakeIdentifiable{}
		}

		auditer := &FakeAuditer{}

		ctx := NewContextWithRequest(request)
		err := dispatchCreateOperation(ctx, processorFinder, factory, nil, nil, nil, auditer, true, nil)

		Convey("Then I should have a 423 error and context should be nil", func() {
			So(err, ShouldNotBeNil)
			So(err.(elemental.Error).Code, ShouldEqual, http.StatusLocked)
		})
	})

	Convey("Given I have a processor that handle ProcessCreate function with error", t, func() {
		request := elemental.NewRequest()
		request.Data = []byte(`{"ID": "1234", "Name": "Fake"}`)

		processorFinder := func(identity elemental.Identity) (Processor, error) {
			return &FakeCompleteProcessor{
				err: elemental.NewError("Error", "Bad request.", "bahamut-test", http.StatusBadRequest),
			}, nil
		}

		factory := func(identity string, version int) elemental.Identifiable {
			return &FakeIdentifiable{}
		}

		auditer := &FakeAuditer{}

		ctx := NewContextWithRequest(request)
		err := dispatchCreateOperation(ctx, processorFinder, factory, nil, nil, nil, auditer, false, nil)

		expectedError := "error 400 (bahamut-test): Error: Bad request."
		expectedNbCalls := 1

		Convey("Then I should get a bahamut error and no context", func() {
			So(err.Error(), ShouldEqual, expectedError)
			So(auditer.nbCalls, ShouldEqual, expectedNbCalls)
		})
	})

	Convey("Given I have a processor that handle ProcessCreate function with an empty JSON", t, func() {
		request := elemental.NewRequest()

		processorFinder := func(identity elemental.Identity) (Processor, error) {
			return &FakeCompleteProcessor{}, nil
		}

		factory := func(identity string, version int) elemental.Identifiable {
			return &FakeIdentifiable{}
		}

		auditer := &FakeAuditer{}

		ctx := NewContextWithRequest(request)
		err := dispatchCreateOperation(ctx, processorFinder, factory, nil, nil, nil, auditer, false, nil)

		expectedError := "error 400 (elemental): Bad Request: Something went wrong in the server when reading the body of the request"
		expectedNbCalls := 1

		Convey("Then I should get a bahamut error and no context", func() {
			So(err.Error(), ShouldEqual, expectedError)
			So(auditer.nbCalls, ShouldEqual, expectedNbCalls)
		})
	})

	Convey("Given I have a processor that handle ProcessCreate function with an invalid JSON", t, func() {
		request := elemental.NewRequest()
		request.Data = []byte(`An invalid JSON`)

		processorFinder := func(identity elemental.Identity) (Processor, error) {
			return &FakeCompleteProcessor{}, nil
		}

		factory := func(identity string, version int) elemental.Identifiable {
			return &FakeIdentifiable{}
		}

		auditer := &FakeAuditer{}

		ctx := NewContextWithRequest(request)
		err := dispatchCreateOperation(ctx, processorFinder, factory, nil, nil, nil, auditer, false, nil)

		expectedError := "error 400 (elemental): Bad Request: Invalid JSON"
		expectedNbCalls := 1

		Convey("Then I should get a bahamut error and no context", func() {
			So(err.Error(), ShouldEqual, expectedError)
			So(auditer.nbCalls, ShouldEqual, expectedNbCalls)
		})
	})

	Convey("Given I have a processor that handle ProcessCreate function with an invalid object", t, func() {
		request := elemental.NewRequest()
		request.Data = []byte(`{"ID": "1234", "Name": "Fake"}`)

		processorFinder := func(identity elemental.Identity) (Processor, error) {
			return &FakeCompleteProcessor{}, nil
		}

		factory := func(identity string, version int) elemental.Identifiable {
			return &FakeIdentifiable{
				validationError: elemental.NewError("Error", "Object validation has failed.", "bahamut-test", http.StatusBadRequest),
			}
		}

		auditer := &FakeAuditer{}

		ctx := NewContextWithRequest(request)
		err := dispatchCreateOperation(ctx, processorFinder, factory, nil, nil, nil, auditer, false, nil)

		expectedError := "error 400 (bahamut-test): Error: Object validation has failed."
		expectedNbCalls := 1

		Convey("Then I should get a bahamut error and no context", func() {
			So(err.Error(), ShouldEqual, expectedError)
			So(auditer.nbCalls, ShouldEqual, expectedNbCalls)
		})
	})

	Convey("Given I have a processor that does not handle ProcessCreate function", t, func() {
		request := elemental.NewRequest()
		request.Operation = elemental.OperationRetrieveMany
		request.Identity = elemental.MakeIdentity("Fake", "Test")

		processorFinder := func(identity elemental.Identity) (Processor, error) {
			return &FakeProcessor{}, nil
		}

		factory := func(identity string, version int) elemental.Identifiable {
			return &FakeIdentifiable{}
		}

		auditer := &FakeAuditer{}

		expectedError := "error 501 (bahamut): Not implemented: No handler for operation retrieve-many on Fake"
		expectedNbCalls := 1

		ctx := NewContextWithRequest(request)
		err := dispatchCreateOperation(ctx, processorFinder, factory, nil, nil, nil, auditer, false, nil)

		Convey("Then I should get a bahamut error and no context", func() {
			So(err.Error(), ShouldEqual, expectedError)
			So(auditer.nbCalls, ShouldEqual, expectedNbCalls)
		})
	})

	Convey("Given I have a processor that does not handle ProcessCreate function and an authenticator that is not authenticated", t, func() {
		request := elemental.NewRequest()

		processorFinder := func(identity elemental.Identity) (Processor, error) {
			return &FakeProcessor{}, nil
		}

		factory := func(identity string, version int) elemental.Identifiable {
			return &FakeIdentifiable{}
		}

		authenticators := []RequestAuthenticator{
			&Auth{
				authenticated: true,
			},
			&Auth{
				errored: true,
				err:     elemental.NewError("Error", "Authenticator does not authenticate.", "bahamut-test", http.StatusInternalServerError),
			},
		}

		auditer := &FakeAuditer{}

		expectedError := "error 500 (bahamut-test): Error: Authenticator does not authenticate."
		expectedNbCalls := 1

		ctx := NewContextWithRequest(request)
		err := dispatchCreateOperation(ctx, processorFinder, factory, authenticators, nil, nil, auditer, false, nil)

		Convey("Then I should get a bahamut error and no context", func() {
			So(err.Error(), ShouldEqual, expectedError)
			So(auditer.nbCalls, ShouldEqual, expectedNbCalls)
		})
	})

	Convey("Given I have a processor that does not handle ProcessCreate function and an authorizer that is not authorize", t, func() {
		request := elemental.NewRequest()

		processorFinder := func(identity elemental.Identity) (Processor, error) {
			return &FakeProcessor{}, nil
		}

		factory := func(identity string, version int) elemental.Identifiable {
			return &FakeIdentifiable{}
		}

		authenticators := []RequestAuthenticator{
			&Auth{
				authenticated: true,
			},
			&Auth{
				authenticated: true,
			},
		}

		authorizers := []Authorizer{
			&Auth{
				authorized: true,
			},
			&Auth{
				errored: true,
				err:     elemental.NewError("Error", "Authorizer does not authorize.", "bahamut-test", http.StatusInternalServerError),
			},
		}

		auditer := &FakeAuditer{}

		expectedError := "error 500 (bahamut-test): Error: Authorizer does not authorize."
		expectedNbCalls := 1

		ctx := NewContextWithRequest(request)
		err := dispatchCreateOperation(ctx, processorFinder, factory, authenticators, authorizers, nil, auditer, false, nil)

		Convey("Then I should get a bahamut error and no context", func() {
			So(err.Error(), ShouldEqual, expectedError)
			So(auditer.nbCalls, ShouldEqual, expectedNbCalls)
		})
	})
}

// TestDispatchers_dispatchUpdateOperation tests dispatchUpdateOperation method
func TestDispatchers_dispatchUpdateOperation(t *testing.T) {

	Convey("Given I have a processor that handle ProcessUpdate function", t, func() {
		request := elemental.NewRequest()
		request.Data = []byte(`{"ID": "1234", "Name": "Fake"}`)

		processorFinder := func(identity elemental.Identity) (Processor, error) {
			return &FakeCompleteProcessor{
				output: &FakeIdentifiable{identifier: "a"},
				events: []*elemental.Event{elemental.NewEvent(elemental.EventDelete, &FakeIdentifiable{})},
			}, nil
		}

		factory := func(identity string, version int) elemental.Identifiable {
			return &FakeIdentifiable{}
		}

		auditer := &FakeAuditer{}
		pusher := &FakePusher{}

		ctx := NewContextWithRequest(request)
		err := dispatchUpdateOperation(ctx, processorFinder, factory, nil, nil, pusher.Push, auditer, false, nil)

		expectedNbCalls := 1

		Convey("Then I should have no error and context should be initiated", func() {
			So(err, ShouldBeNil)
			So(ctx.InputData, ShouldNotBeNil)
			So(auditer.nbCalls, ShouldEqual, expectedNbCalls)
			So(ctx.OutputData, ShouldResemble, &FakeIdentifiable{identifier: "a"})
			So(len(pusher.events), ShouldEqual, 2)
			So(pusher.events[0].Type, ShouldEqual, elemental.EventDelete)
			So(pusher.events[1].Type, ShouldEqual, elemental.EventUpdate)
		})
	})

	Convey("Given I have a processor that handle ProcessUpdate function with read only mode", t, func() {
		request := elemental.NewRequest()
		request.Data = []byte(`{"ID": "1234", "Name": "Fake"}`)

		processorFinder := func(identity elemental.Identity) (Processor, error) {
			return &FakeCompleteProcessor{}, nil
		}

		factory := func(identity string, version int) elemental.Identifiable {
			return &FakeIdentifiable{}
		}

		auditer := &FakeAuditer{}

		ctx := NewContextWithRequest(request)
		err := dispatchUpdateOperation(ctx, processorFinder, factory, nil, nil, nil, auditer, true, nil)

		Convey("Then I should have a 423 error and context should be nil", func() {
			So(err, ShouldNotBeNil)
			So(err.(elemental.Error).Code, ShouldEqual, http.StatusLocked)
		})
	})

	Convey("Given I have a processor that handle ProcessUpdate function with error", t, func() {
		request := elemental.NewRequest()
		request.Data = []byte(`{"ID": "1234", "Name": "Fake"}`)

		processorFinder := func(identity elemental.Identity) (Processor, error) {
			return &FakeCompleteProcessor{
				err: elemental.NewError("Error", "Bad request.", "bahamut-test", http.StatusBadRequest),
			}, nil
		}

		factory := func(identity string, version int) elemental.Identifiable {
			return &FakeIdentifiable{}
		}

		auditer := &FakeAuditer{}

		ctx := NewContextWithRequest(request)
		err := dispatchUpdateOperation(ctx, processorFinder, factory, nil, nil, nil, auditer, false, nil)

		expectedError := "error 400 (bahamut-test): Error: Bad request."
		expectedNbCalls := 1

		Convey("Then I should get a bahamut error and no context", func() {
			So(err.Error(), ShouldEqual, expectedError)
			So(auditer.nbCalls, ShouldEqual, expectedNbCalls)
		})
	})

	Convey("Given I have a processor that handle ProcessUpdate function with an empty JSON", t, func() {
		request := elemental.NewRequest()

		processorFinder := func(identity elemental.Identity) (Processor, error) {
			return &FakeCompleteProcessor{}, nil
		}

		factory := func(identity string, version int) elemental.Identifiable {
			return &FakeIdentifiable{}
		}

		auditer := &FakeAuditer{}

		ctx := NewContextWithRequest(request)
		err := dispatchUpdateOperation(ctx, processorFinder, factory, nil, nil, nil, auditer, false, nil)

		expectedError := "error 400 (elemental): Bad Request: Something went wrong in the server when reading the body of the request"
		expectedNbCalls := 1

		Convey("Then I should get a bahamut error and no context", func() {
			So(err.Error(), ShouldEqual, expectedError)
			So(auditer.nbCalls, ShouldEqual, expectedNbCalls)
		})
	})

	Convey("Given I have a processor that handle ProcessUpdate function with an invalid JSON", t, func() {
		request := elemental.NewRequest()
		request.Data = []byte(`An invalid JSON`)

		processorFinder := func(identity elemental.Identity) (Processor, error) {
			return &FakeCompleteProcessor{}, nil
		}

		factory := func(identity string, version int) elemental.Identifiable {
			return &FakeIdentifiable{}
		}

		auditer := &FakeAuditer{}

		ctx := NewContextWithRequest(request)
		err := dispatchUpdateOperation(ctx, processorFinder, factory, nil, nil, nil, auditer, false, nil)

		expectedError := "error 400 (elemental): Bad Request: Invalid JSON"
		expectedNbCalls := 1

		Convey("Then I should get a bahamut error and no context", func() {
			So(err.Error(), ShouldEqual, expectedError)
			So(auditer.nbCalls, ShouldEqual, expectedNbCalls)
		})
	})

	Convey("Given I have a processor that handle ProcessUpdate function with an invalid object", t, func() {
		request := elemental.NewRequest()
		request.Data = []byte(`{"ID": "1234", "Name": "Fake"}`)

		processorFinder := func(identity elemental.Identity) (Processor, error) {
			return &FakeCompleteProcessor{}, nil
		}

		factory := func(identity string, version int) elemental.Identifiable {
			return &FakeIdentifiable{
				validationError: elemental.NewError("Error", "Object validation has failed.", "bahamut-test", http.StatusBadRequest),
			}
		}

		auditer := &FakeAuditer{}

		ctx := NewContextWithRequest(request)
		err := dispatchUpdateOperation(ctx, processorFinder, factory, nil, nil, nil, auditer, false, nil)

		expectedError := "error 400 (bahamut-test): Error: Object validation has failed."
		expectedNbCalls := 1

		Convey("Then I should get a bahamut error and no context", func() {
			So(err.Error(), ShouldEqual, expectedError)
			So(auditer.nbCalls, ShouldEqual, expectedNbCalls)
		})
	})

	Convey("Given I have a processor that does not handle ProcessUpdate function", t, func() {
		request := elemental.NewRequest()
		request.Operation = elemental.OperationRetrieveMany
		request.Identity = elemental.MakeIdentity("Fake", "Test")

		processorFinder := func(identity elemental.Identity) (Processor, error) {
			return &FakeProcessor{}, nil
		}

		factory := func(identity string, version int) elemental.Identifiable {
			return &FakeIdentifiable{}
		}

		auditer := &FakeAuditer{}

		expectedError := "error 501 (bahamut): Not implemented: No handler for operation retrieve-many on Fake"
		expectedNbCalls := 1

		ctx := NewContextWithRequest(request)
		err := dispatchUpdateOperation(ctx, processorFinder, factory, nil, nil, nil, auditer, false, nil)

		Convey("Then I should get a bahamut error and no context", func() {
			So(err.Error(), ShouldEqual, expectedError)
			So(auditer.nbCalls, ShouldEqual, expectedNbCalls)
		})
	})

	Convey("Given I have a processor that does not handle ProcessUpdate function and an authenticator that is not authenticated", t, func() {
		request := elemental.NewRequest()

		processorFinder := func(identity elemental.Identity) (Processor, error) {
			return &FakeProcessor{}, nil
		}

		factory := func(identity string, version int) elemental.Identifiable {
			return &FakeIdentifiable{}
		}

		authenticators := []RequestAuthenticator{
			&Auth{
				authenticated: true,
			},
			&Auth{
				errored: true,
				err:     elemental.NewError("Error", "Authenticator does not authenticate.", "bahamut-test", http.StatusInternalServerError),
			},
		}

		auditer := &FakeAuditer{}

		expectedError := "error 500 (bahamut-test): Error: Authenticator does not authenticate."
		expectedNbCalls := 1

		ctx := NewContextWithRequest(request)
		err := dispatchUpdateOperation(ctx, processorFinder, factory, authenticators, nil, nil, auditer, false, nil)

		Convey("Then I should get a bahamut error and no context", func() {
			So(err.Error(), ShouldEqual, expectedError)
			So(auditer.nbCalls, ShouldEqual, expectedNbCalls)
		})
	})

	Convey("Given I have a processor that does not handle ProcessUpdate function and an authorizer that is not authorize", t, func() {
		request := elemental.NewRequest()

		processorFinder := func(identity elemental.Identity) (Processor, error) {
			return &FakeProcessor{}, nil
		}

		factory := func(identity string, version int) elemental.Identifiable {
			return &FakeIdentifiable{}
		}

		authenticators := []RequestAuthenticator{
			&Auth{
				authenticated: true,
			},
			&Auth{
				authenticated: true,
			},
		}

		authorizers := []Authorizer{
			&Auth{
				authorized: true,
			},
			&Auth{
				errored: true,
				err:     elemental.NewError("Error", "Authorizer does not authorize.", "bahamut-test", http.StatusInternalServerError),
			},
		}

		auditer := &FakeAuditer{}

		expectedError := "error 500 (bahamut-test): Error: Authorizer does not authorize."
		expectedNbCalls := 1

		ctx := NewContextWithRequest(request)
		err := dispatchUpdateOperation(ctx, processorFinder, factory, authenticators, authorizers, nil, auditer, false, nil)

		Convey("Then I should get a bahamut error and no context", func() {
			So(err.Error(), ShouldEqual, expectedError)
			So(auditer.nbCalls, ShouldEqual, expectedNbCalls)
		})
	})
}

// TestDispatchers_dispatchDeleteOperation tests dispatchDeleteOperation method
func TestDispatchers_dispatchDeleteOperation(t *testing.T) {

	Convey("Given I have a processor that handle ProcessDelete function", t, func() {
		request := elemental.NewRequest()
		request.Data = []byte(`{"ID": "1234", "Name": "Fake"}`)

		processorFinder := func(identity elemental.Identity) (Processor, error) {
			return &FakeCompleteProcessor{
				output: &FakeIdentifiable{identifier: "a"},
				events: []*elemental.Event{elemental.NewEvent(elemental.EventCreate, &FakeIdentifiable{})},
			}, nil
		}

		factory := func(identity string, version int) elemental.Identifiable {
			return &FakeIdentifiable{}
		}

		auditer := &FakeAuditer{}
		pusher := &FakePusher{}

		ctx := NewContextWithRequest(request)
		err := dispatchDeleteOperation(ctx, processorFinder, factory, nil, nil, pusher.Push, auditer, false, nil)

		expectedNbCalls := 1

		Convey("Then I should have no error and context should be initiated", func() {
			So(err, ShouldBeNil)
			So(auditer.nbCalls, ShouldEqual, expectedNbCalls)
			So(ctx.OutputData, ShouldResemble, &FakeIdentifiable{identifier: "a"})
			So(len(pusher.events), ShouldEqual, 2)
			So(pusher.events[0].Type, ShouldEqual, elemental.EventCreate)
			So(pusher.events[1].Type, ShouldEqual, elemental.EventDelete)
		})
	})

	Convey("Given I have a processor that handle ProcessDelete function with read only mode", t, func() {
		request := elemental.NewRequest()
		request.Data = []byte(`{"ID": "1234", "Name": "Fake"}`)

		processorFinder := func(identity elemental.Identity) (Processor, error) {
			return &FakeCompleteProcessor{}, nil
		}

		factory := func(identity string, version int) elemental.Identifiable {
			return &FakeIdentifiable{}
		}

		auditer := &FakeAuditer{}

		ctx := NewContextWithRequest(request)
		err := dispatchDeleteOperation(ctx, processorFinder, factory, nil, nil, nil, auditer, true, nil)

		Convey("Then I should have a 423 error and context should be nil", func() {
			So(err, ShouldNotBeNil)
			So(err.(elemental.Error).Code, ShouldEqual, http.StatusLocked)
		})
	})

	Convey("Given I have a processor that handle ProcessDelete function with error", t, func() {
		request := elemental.NewRequest()
		request.Data = []byte(`{"ID": "1234", "Name": "Fake"}`)

		processorFinder := func(identity elemental.Identity) (Processor, error) {
			return &FakeCompleteProcessor{
				err: elemental.NewError("Error", "Bad request.", "bahamut-test", http.StatusBadRequest),
			}, nil
		}

		factory := func(identity string, version int) elemental.Identifiable {
			return &FakeIdentifiable{}
		}

		auditer := &FakeAuditer{}

		ctx := NewContextWithRequest(request)
		err := dispatchDeleteOperation(ctx, processorFinder, factory, nil, nil, nil, auditer, false, nil)

		expectedError := "error 400 (bahamut-test): Error: Bad request."
		expectedNbCalls := 1

		Convey("Then I should get a bahamut error and no context", func() {
			So(err.Error(), ShouldEqual, expectedError)
			So(auditer.nbCalls, ShouldEqual, expectedNbCalls)
		})
	})

	Convey("Given I have a processor that does not handle ProcessDelete function", t, func() {
		request := elemental.NewRequest()
		request.Operation = elemental.OperationRetrieveMany
		request.Identity = elemental.MakeIdentity("Fake", "Test")

		processorFinder := func(identity elemental.Identity) (Processor, error) {
			return &FakeProcessor{}, nil
		}

		factory := func(identity string, version int) elemental.Identifiable {
			return &FakeIdentifiable{}
		}

		auditer := &FakeAuditer{}

		expectedError := "error 501 (bahamut): Not implemented: No handler for operation retrieve-many on Fake"
		expectedNbCalls := 1

		ctx := NewContextWithRequest(request)
		err := dispatchDeleteOperation(ctx, processorFinder, factory, nil, nil, nil, auditer, false, nil)

		Convey("Then I should get a bahamut error and no context", func() {
			So(err.Error(), ShouldEqual, expectedError)
			So(auditer.nbCalls, ShouldEqual, expectedNbCalls)
		})
	})

	Convey("Given I have a processor that does not handle ProcessDelete function and an authenticator that is not authenticated", t, func() {
		request := elemental.NewRequest()

		processorFinder := func(identity elemental.Identity) (Processor, error) {
			return &FakeProcessor{}, nil
		}

		factory := func(identity string, version int) elemental.Identifiable {
			return &FakeIdentifiable{}
		}

		authenticators := []RequestAuthenticator{
			&Auth{
				authenticated: true,
			},
			&Auth{
				errored: true,
				err:     elemental.NewError("Error", "Authenticator does not authenticate.", "bahamut-test", http.StatusInternalServerError),
			},
		}

		auditer := &FakeAuditer{}

		expectedError := "error 500 (bahamut-test): Error: Authenticator does not authenticate."
		expectedNbCalls := 1

		ctx := NewContextWithRequest(request)
		err := dispatchDeleteOperation(ctx, processorFinder, factory, authenticators, nil, nil, auditer, false, nil)

		Convey("Then I should get a bahamut error and no context", func() {
			So(err.Error(), ShouldEqual, expectedError)
			So(auditer.nbCalls, ShouldEqual, expectedNbCalls)
		})
	})

	Convey("Given I have a processor that does not handle ProcessDelete function and an authorizer that is not authorize", t, func() {
		request := elemental.NewRequest()

		processorFinder := func(identity elemental.Identity) (Processor, error) {
			return &FakeProcessor{}, nil
		}

		factory := func(identity string, version int) elemental.Identifiable {
			return &FakeIdentifiable{}
		}

		authenticators := []RequestAuthenticator{
			&Auth{
				authenticated: true,
			},
			&Auth{
				authenticated: true,
			},
		}

		authorizers := []Authorizer{
			&Auth{
				authorized: true,
			},
			&Auth{
				errored: true,
				err:     elemental.NewError("Error", "Authorizer does not authorize.", "bahamut-test", http.StatusInternalServerError),
			},
		}

		auditer := &FakeAuditer{}

		expectedError := "error 500 (bahamut-test): Error: Authorizer does not authorize."
		expectedNbCalls := 1

		ctx := NewContextWithRequest(request)
		err := dispatchDeleteOperation(ctx, processorFinder, factory, authenticators, authorizers, nil, auditer, false, nil)

		Convey("Then I should get a bahamut error and no context", func() {
			So(err.Error(), ShouldEqual, expectedError)
			So(auditer.nbCalls, ShouldEqual, expectedNbCalls)
		})
	})
}

// TestDispatchers_dispatchPatchOperation tests dispatchPatchOperation method
func TestDispatchers_dispatchPatchOperation(t *testing.T) {

	Convey("Given I have a processor that handle ProcessPatch function", t, func() {
		request := elemental.NewRequest()
		request.Data = []byte(`{"ID": "1234", "Name": "Fake"}`)

		processorFinder := func(identity elemental.Identity) (Processor, error) {
			return &FakeCompleteProcessor{
				output: &elemental.Patch{Type: elemental.PatchTypeSetIfZero},
				events: []*elemental.Event{elemental.NewEvent(elemental.EventDelete, &FakeIdentifiable{})},
			}, nil
		}

		factory := func(identity string, version int) elemental.Identifiable {
			return &FakeIdentifiable{}
		}

		auditer := &FakeAuditer{}
		pusher := &FakePusher{}

		ctx := NewContextWithRequest(request)
		err := dispatchPatchOperation(ctx, processorFinder, factory, nil, nil, pusher.Push, auditer, false, nil)

		expectedNbCalls := 1

		Convey("Then I should have no error and context should be initiated", func() {
			So(err, ShouldBeNil)
			So(auditer.nbCalls, ShouldEqual, expectedNbCalls)
			So(ctx.OutputData, ShouldResemble, &elemental.Patch{Type: elemental.PatchTypeSetIfZero})
			So(len(pusher.events), ShouldEqual, 2)
			So(pusher.events[0].Type, ShouldEqual, elemental.EventDelete)
			So(pusher.events[1].Type, ShouldEqual, elemental.EventUpdate)
		})
	})

	Convey("Given I have a processor that handle ProcessPatch function with read only mode", t, func() {
		request := elemental.NewRequest()
		request.Data = []byte(`{"ID": "1234", "Name": "Fake"}`)

		processorFinder := func(identity elemental.Identity) (Processor, error) {
			return &FakeCompleteProcessor{}, nil
		}

		factory := func(identity string, version int) elemental.Identifiable {
			return &FakeIdentifiable{}
		}

		auditer := &FakeAuditer{}

		ctx := NewContextWithRequest(request)
		err := dispatchPatchOperation(ctx, processorFinder, factory, nil, nil, nil, auditer, true, nil)

		Convey("Then I should have a 423 error and context should be nil", func() {
			So(err, ShouldNotBeNil)
			So(err.(elemental.Error).Code, ShouldEqual, http.StatusLocked)
		})
	})

	Convey("Given I have a processor that handle ProcessPatch function with an invalid JSON", t, func() {
		request := elemental.NewRequest()
		request.Data = []byte(`Invalid JSON`)

		processorFinder := func(identity elemental.Identity) (Processor, error) {
			return &FakeCompleteProcessor{}, nil
		}

		factory := func(identity string, version int) elemental.Identifiable {
			return &FakeIdentifiable{}
		}

		auditer := &FakeAuditer{}

		ctx := NewContextWithRequest(request)
		err := dispatchPatchOperation(ctx, processorFinder, factory, nil, nil, nil, auditer, false, nil)

		expectedError := "error 400 (elemental): Bad Request: Invalid JSON"
		expectedNbCalls := 1

		Convey("Then I should have no error and context should be initiated", func() {
			So(err.Error(), ShouldEqual, expectedError)
			So(auditer.nbCalls, ShouldEqual, expectedNbCalls)
		})
	})

	Convey("Given I have a processor that handle ProcessPatch function with error", t, func() {
		request := elemental.NewRequest()
		request.Data = []byte(`{"ID": "1234", "Name": "Fake"}`)

		processorFinder := func(identity elemental.Identity) (Processor, error) {
			return &FakeCompleteProcessor{
				err: elemental.NewError("Error", "Bad request.", "bahamut-test", http.StatusBadRequest),
			}, nil
		}

		factory := func(identity string, version int) elemental.Identifiable {
			return &FakeIdentifiable{}
		}

		auditer := &FakeAuditer{}

		ctx := NewContextWithRequest(request)
		err := dispatchPatchOperation(ctx, processorFinder, factory, nil, nil, nil, auditer, false, nil)

		expectedError := "error 400 (bahamut-test): Error: Bad request."
		expectedNbCalls := 1

		Convey("Then I should get a bahamut error and no context", func() {
			So(err.Error(), ShouldEqual, expectedError)
			So(auditer.nbCalls, ShouldEqual, expectedNbCalls)
		})
	})

	Convey("Given I have a processor that does not handle ProcessPatch function", t, func() {
		request := elemental.NewRequest()
		request.Operation = elemental.OperationRetrieveMany
		request.Identity = elemental.MakeIdentity("Fake", "Test")

		processorFinder := func(identity elemental.Identity) (Processor, error) {
			return &FakeProcessor{}, nil
		}

		factory := func(identity string, version int) elemental.Identifiable {
			return &FakeIdentifiable{}
		}

		auditer := &FakeAuditer{}

		expectedError := "error 501 (bahamut): Not implemented: No handler for operation retrieve-many on Fake"
		expectedNbCalls := 1

		ctx := NewContextWithRequest(request)
		err := dispatchPatchOperation(ctx, processorFinder, factory, nil, nil, nil, auditer, false, nil)

		Convey("Then I should get a bahamut error and no context", func() {
			So(err.Error(), ShouldEqual, expectedError)
			So(auditer.nbCalls, ShouldEqual, expectedNbCalls)
		})
	})

	Convey("Given I have a processor that does not handle ProcessPatch function and an authenticator that is not authenticated", t, func() {
		request := elemental.NewRequest()

		processorFinder := func(identity elemental.Identity) (Processor, error) {
			return &FakeProcessor{}, nil
		}

		factory := func(identity string, version int) elemental.Identifiable {
			return &FakeIdentifiable{}
		}

		authenticators := []RequestAuthenticator{
			&Auth{
				authenticated: true,
			},
			&Auth{
				errored: true,
				err:     elemental.NewError("Error", "Authenticator does not authenticate.", "bahamut-test", http.StatusInternalServerError),
			},
		}

		auditer := &FakeAuditer{}

		expectedError := "error 500 (bahamut-test): Error: Authenticator does not authenticate."
		expectedNbCalls := 1

		ctx := NewContextWithRequest(request)
		err := dispatchPatchOperation(ctx, processorFinder, factory, authenticators, nil, nil, auditer, false, nil)

		Convey("Then I should get a bahamut error and no context", func() {
			So(err.Error(), ShouldEqual, expectedError)
			So(auditer.nbCalls, ShouldEqual, expectedNbCalls)
		})
	})

	Convey("Given I have a processor that does not handle ProcessPatch function and an authorizer that is not authorize", t, func() {
		request := elemental.NewRequest()

		processorFinder := func(identity elemental.Identity) (Processor, error) {
			return &FakeProcessor{}, nil
		}

		factory := func(identity string, version int) elemental.Identifiable {
			return &FakeIdentifiable{}
		}

		authenticators := []RequestAuthenticator{
			&Auth{
				authenticated: true,
			},
			&Auth{
				authenticated: true,
			},
		}

		authorizers := []Authorizer{
			&Auth{
				authorized: true,
			},
			&Auth{
				errored: true,
				err:     elemental.NewError("Error", "Authorizer does not authorize.", "bahamut-test", http.StatusInternalServerError),
			},
		}

		auditer := &FakeAuditer{}

		expectedError := "error 500 (bahamut-test): Error: Authorizer does not authorize."
		expectedNbCalls := 1

		ctx := NewContextWithRequest(request)
		err := dispatchPatchOperation(ctx, processorFinder, factory, authenticators, authorizers, nil, auditer, false, nil)

		Convey("Then I should get a bahamut error and no context", func() {
			So(err.Error(), ShouldEqual, expectedError)
			So(auditer.nbCalls, ShouldEqual, expectedNbCalls)
		})
	})
}

// TestDispatchers_dispatchInfoOperation tests dispatchInfoOperation method
func TestDispatchers_dispatchInfoOperation(t *testing.T) {

	Convey("Given I have a processor that handle ProcessInfo function", t, func() {
		request := elemental.NewRequest()
		request.Data = []byte(`{"ID": "1234", "Name": "Fake"}`)

		processorFinder := func(identity elemental.Identity) (Processor, error) {
			return &FakeCompleteProcessor{}, nil
		}

		factory := func(identity string, version int) elemental.Identifiable {
			return &FakeIdentifiable{}
		}

		auditer := &FakeAuditer{}
		pusher := &FakePusher{}

		ctx := NewContextWithRequest(request)
		err := dispatchInfoOperation(ctx, processorFinder, factory, nil, nil, pusher.Push, auditer)

		expectedNbCalls := 1

		Convey("Then I should have no error and context should be initiated", func() {
			So(err, ShouldBeNil)
			So(auditer.nbCalls, ShouldEqual, expectedNbCalls)
		})
	})

	Convey("Given I have a processor that handle ProcessInfo function with error", t, func() {
		request := elemental.NewRequest()
		request.Data = []byte(`{"ID": "1234", "Name": "Fake"}`)

		processorFinder := func(identity elemental.Identity) (Processor, error) {
			return &FakeCompleteProcessor{
				err: elemental.NewError("Error", "Bad request.", "bahamut-test", http.StatusBadRequest),
			}, nil
		}

		factory := func(identity string, version int) elemental.Identifiable {
			return &FakeIdentifiable{}
		}

		auditer := &FakeAuditer{}
		pusher := &FakePusher{}

		ctx := NewContextWithRequest(request)
		err := dispatchInfoOperation(ctx, processorFinder, factory, nil, nil, pusher.Push, auditer)

		expectedError := "error 400 (bahamut-test): Error: Bad request."
		expectedNbCalls := 1

		Convey("Then I should get a bahamut error and no context", func() {
			So(err.Error(), ShouldEqual, expectedError)
			So(auditer.nbCalls, ShouldEqual, expectedNbCalls)
		})
	})

	Convey("Given I have a processor that does not handle ProcessInfo function", t, func() {
		request := elemental.NewRequest()
		request.Operation = elemental.OperationRetrieveMany
		request.Identity = elemental.MakeIdentity("Fake", "Test")

		processorFinder := func(identity elemental.Identity) (Processor, error) {
			return &FakeProcessor{}, nil
		}

		factory := func(identity string, version int) elemental.Identifiable {
			return &FakeIdentifiable{}
		}

		auditer := &FakeAuditer{}
		pusher := &FakePusher{}

		expectedError := "error 501 (bahamut): Not implemented: No handler for operation retrieve-many on Fake"
		expectedNbCalls := 1

		ctx := NewContextWithRequest(request)
		err := dispatchInfoOperation(ctx, processorFinder, factory, nil, nil, pusher.Push, auditer)

		Convey("Then I should get a bahamut error and no context", func() {
			So(err.Error(), ShouldEqual, expectedError)
			So(auditer.nbCalls, ShouldEqual, expectedNbCalls)
		})
	})

	Convey("Given I have a processor that does not handle ProcessInfo function and an authenticator that is not authenticated", t, func() {
		request := elemental.NewRequest()

		processorFinder := func(identity elemental.Identity) (Processor, error) {
			return &FakeProcessor{}, nil
		}

		factory := func(identity string, version int) elemental.Identifiable {
			return &FakeIdentifiable{}
		}

		authenticators := []RequestAuthenticator{
			&Auth{
				authenticated: true,
			},
			&Auth{
				errored: true,
				err:     elemental.NewError("Error", "Authenticator does not authenticate.", "bahamut-test", http.StatusInternalServerError),
			},
		}

		auditer := &FakeAuditer{}
		pusher := &FakePusher{}

		expectedError := "error 500 (bahamut-test): Error: Authenticator does not authenticate."
		expectedNbCalls := 1

		ctx := NewContextWithRequest(request)
		err := dispatchInfoOperation(ctx, processorFinder, factory, authenticators, nil, pusher.Push, auditer)

		Convey("Then I should get a bahamut error and no context", func() {
			So(err.Error(), ShouldEqual, expectedError)
			So(auditer.nbCalls, ShouldEqual, expectedNbCalls)
		})
	})

	Convey("Given I have a processor that does not handle ProcessInfo function and an authorizer that is not authorize", t, func() {
		request := elemental.NewRequest()

		processorFinder := func(identity elemental.Identity) (Processor, error) {
			return &FakeProcessor{}, nil
		}

		factory := func(identity string, version int) elemental.Identifiable {
			return &FakeIdentifiable{}
		}

		authenticators := []RequestAuthenticator{
			&Auth{
				authenticated: true,
			},
			&Auth{
				authenticated: true,
			},
		}

		authorizers := []Authorizer{
			&Auth{
				authorized: true,
			},
			&Auth{
				errored: true,
				err:     elemental.NewError("Error", "Authorizer does not authorize.", "bahamut-test", http.StatusInternalServerError),
			},
		}

		auditer := &FakeAuditer{}
		pusher := &FakePusher{}

		expectedError := "error 500 (bahamut-test): Error: Authorizer does not authorize."
		expectedNbCalls := 1

		ctx := NewContextWithRequest(request)
		err := dispatchInfoOperation(ctx, processorFinder, factory, authenticators, authorizers, pusher.Push, auditer)

		Convey("Then I should get a bahamut error and no context", func() {
			So(err.Error(), ShouldEqual, expectedError)
			So(auditer.nbCalls, ShouldEqual, expectedNbCalls)
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
