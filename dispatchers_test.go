package bahamut

import (
	"net/http"
	"testing"

	"github.com/aporeto-inc/elemental"
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
	identity   string
	identifier string
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

// FakeCompleteProcessor
type FakeCompleteProcessor struct {
	err error
}

func (p *FakeCompleteProcessor) ProcessRetrieveMany(*Context) error {
	return p.err
}

func (p *FakeCompleteProcessor) ProcessRetrieve(*Context) error {
	return p.err
}

// TestDispatchers_dispatchRetrieveManyOperation tests dispatchRetrieveManyOperation method
func TestDispatchers_dispatchRetrieveManyOperation(t *testing.T) {

	Convey("Given I have a processor that handle ProcessRetrieveMany function", t, func() {
		request := elemental.NewRequest()

		processorFinder := func(identity elemental.Identity) (Processor, error) {
			return &FakeCompleteProcessor{}, nil
		}

		factory := func(identity string) elemental.Identifiable {
			return &FakeIdentifiable{}
		}

		auditer := &FakeAuditer{}

		ctx, err := dispatchRetrieveManyOperation(request, processorFinder, factory, nil, nil, nil, auditer)

		expectedNbCalls := 1

		Convey("Then I should have no error and context should be initiated", func() {
			So(err, ShouldBeNil)
			So(ctx, ShouldNotBeNil)
			So(auditer.nbCalls, ShouldEqual, expectedNbCalls)
		})
	})

	Convey("Given I have a processor that handle ProcessRetrieveMany function with error", t, func() {
		request := elemental.NewRequest()

		processorFinder := func(identity elemental.Identity) (Processor, error) {
			return &FakeCompleteProcessor{
				err: elemental.NewError("Error", "Bad request.", "bahamut-test", http.StatusBadRequest),
			}, nil
		}

		factory := func(identity string) elemental.Identifiable {
			return &FakeIdentifiable{}
		}

		auditer := &FakeAuditer{}

		ctx, err := dispatchRetrieveManyOperation(request, processorFinder, factory, nil, nil, nil, auditer)

		expectedError := "error 400 (bahamut-test): Error: Bad request."
		expectedNbCalls := 1

		Convey("Then I should get a bahamut error and no context", func() {
			So(err.Error(), ShouldEqual, expectedError)
			So(ctx, ShouldBeNil)
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

		factory := func(identity string) elemental.Identifiable {
			return &FakeIdentifiable{}
		}

		auditer := &FakeAuditer{}

		expectedError := "error 501 (bahamut): Not implemented: No handler for operation retrieve-many on Fake"
		expectedNbCalls := 1

		ctx, err := dispatchRetrieveManyOperation(request, processorFinder, factory, nil, nil, nil, auditer)

		Convey("Then I should get a bahamut error and no context", func() {
			So(err.Error(), ShouldEqual, expectedError)
			So(ctx, ShouldBeNil)
			So(auditer.nbCalls, ShouldEqual, expectedNbCalls)
		})
	})

	Convey("Given I have a processor that handle ProcessRetrieveMany function and an authenticator that is not authenticated", t, func() {
		request := elemental.NewRequest()

		processorFinder := func(identity elemental.Identity) (Processor, error) {
			return &FakeProcessor{}, nil
		}

		factory := func(identity string) elemental.Identifiable {
			return &FakeIdentifiable{}
		}

		authenticator := &Auth{
			errored: true,
			err:     elemental.NewError("Error", "Authenticator does not authenticate.", "bahamut-test", http.StatusInternalServerError),
		}

		auditer := &FakeAuditer{}

		expectedError := "error 500 (bahamut-test): Error: Authenticator does not authenticate."
		expectedNbCalls := 1

		ctx, err := dispatchRetrieveManyOperation(request, processorFinder, factory, authenticator, nil, nil, auditer)

		Convey("Then I should get a bahamut error and no context", func() {
			So(err.Error(), ShouldEqual, expectedError)
			So(ctx, ShouldBeNil)
			So(auditer.nbCalls, ShouldEqual, expectedNbCalls)
		})
	})

	Convey("Given I have a processor that handle ProcessRetrieveMany function and an authorizer that is not authorize", t, func() {
		request := elemental.NewRequest()

		processorFinder := func(identity elemental.Identity) (Processor, error) {
			return &FakeProcessor{}, nil
		}

		factory := func(identity string) elemental.Identifiable {
			return &FakeIdentifiable{}
		}

		authenticator := &Auth{
			authenticated: true,
		}

		authorizer := &Auth{
			errored: true,
			err:     elemental.NewError("Error", "Authorizer does not authorize.", "bahamut-test", http.StatusInternalServerError),
		}

		auditer := &FakeAuditer{}

		expectedError := "error 500 (bahamut-test): Error: Authorizer does not authorize."
		expectedNbCalls := 1

		ctx, err := dispatchRetrieveManyOperation(request, processorFinder, factory, authenticator, authorizer, nil, auditer)

		Convey("Then I should get a bahamut error and no context", func() {
			So(err.Error(), ShouldEqual, expectedError)
			So(ctx, ShouldBeNil)
			So(auditer.nbCalls, ShouldEqual, expectedNbCalls)
		})
	})
}

// TestDispatchers_dispatchRetrieveManyOperation tests dispatchRetrieveManyOperation method
func TestDispatchers_dispatchRetrieveOperation(t *testing.T) {

	Convey("Given I have a processor that handle ProcessRetrieve function", t, func() {
		request := elemental.NewRequest()

		processorFinder := func(identity elemental.Identity) (Processor, error) {
			return &FakeCompleteProcessor{}, nil
		}

		factory := func(identity string) elemental.Identifiable {
			return &FakeIdentifiable{}
		}

		auditer := &FakeAuditer{}

		ctx, err := dispatchRetrieveOperation(request, processorFinder, factory, nil, nil, nil, auditer)

		expectedNbCalls := 1

		Convey("Then I should have no error and context should be initiated", func() {
			So(err, ShouldBeNil)
			So(ctx, ShouldNotBeNil)
			So(auditer.nbCalls, ShouldEqual, expectedNbCalls)
		})
	})

	Convey("Given I have a processor that handle ProcessRetrieve function with error", t, func() {
		request := elemental.NewRequest()

		processorFinder := func(identity elemental.Identity) (Processor, error) {
			return &FakeCompleteProcessor{
				err: elemental.NewError("Error", "Bad request.", "bahamut-test", http.StatusBadRequest),
			}, nil
		}

		factory := func(identity string) elemental.Identifiable {
			return &FakeIdentifiable{}
		}

		auditer := &FakeAuditer{}

		ctx, err := dispatchRetrieveOperation(request, processorFinder, factory, nil, nil, nil, auditer)

		expectedError := "error 400 (bahamut-test): Error: Bad request."
		expectedNbCalls := 1

		Convey("Then I should get a bahamut error and no context", func() {
			So(err.Error(), ShouldEqual, expectedError)
			So(ctx, ShouldBeNil)
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

		factory := func(identity string) elemental.Identifiable {
			return &FakeIdentifiable{}
		}

		auditer := &FakeAuditer{}

		expectedError := "error 501 (bahamut): Not implemented: No handler for operation retrieve-many on Fake"
		expectedNbCalls := 1

		ctx, err := dispatchRetrieveOperation(request, processorFinder, factory, nil, nil, nil, auditer)

		Convey("Then I should get a bahamut error and no context", func() {
			So(err.Error(), ShouldEqual, expectedError)
			So(ctx, ShouldBeNil)
			So(auditer.nbCalls, ShouldEqual, expectedNbCalls)
		})
	})

	Convey("Given I have a processor that does not handle ProcessRetrieve function and an authenticator that is not authenticated", t, func() {
		request := elemental.NewRequest()

		processorFinder := func(identity elemental.Identity) (Processor, error) {
			return &FakeProcessor{}, nil
		}

		factory := func(identity string) elemental.Identifiable {
			return &FakeIdentifiable{}
		}

		authenticator := &Auth{
			errored: true,
			err:     elemental.NewError("Error", "Authenticator does not authenticate.", "bahamut-test", http.StatusInternalServerError),
		}

		auditer := &FakeAuditer{}

		expectedError := "error 500 (bahamut-test): Error: Authenticator does not authenticate."
		expectedNbCalls := 1

		ctx, err := dispatchRetrieveOperation(request, processorFinder, factory, authenticator, nil, nil, auditer)

		Convey("Then I should get a bahamut error and no context", func() {
			So(err.Error(), ShouldEqual, expectedError)
			So(ctx, ShouldBeNil)
			So(auditer.nbCalls, ShouldEqual, expectedNbCalls)
		})
	})

	Convey("Given I have a processor that does not handle ProcessRetrieve function and an authorizer that is not authorize", t, func() {
		request := elemental.NewRequest()

		processorFinder := func(identity elemental.Identity) (Processor, error) {
			return &FakeProcessor{}, nil
		}

		factory := func(identity string) elemental.Identifiable {
			return &FakeIdentifiable{}
		}

		authenticator := &Auth{
			authenticated: true,
		}

		authorizer := &Auth{
			errored: true,
			err:     elemental.NewError("Error", "Authorizer does not authorize.", "bahamut-test", http.StatusInternalServerError),
		}

		auditer := &FakeAuditer{}

		expectedError := "error 500 (bahamut-test): Error: Authorizer does not authorize."
		expectedNbCalls := 1

		ctx, err := dispatchRetrieveOperation(request, processorFinder, factory, authenticator, authorizer, nil, auditer)

		Convey("Then I should get a bahamut error and no context", func() {
			So(err.Error(), ShouldEqual, expectedError)
			So(ctx, ShouldBeNil)
			So(auditer.nbCalls, ShouldEqual, expectedNbCalls)
		})
	})
}

// func dispatchRetrieveOperation(
// 	request *elemental.Request,
// 	processorFinder processorFinder,
// 	factory elemental.IdentifiableFactory,
// 	authenticator RequestAuthenticator,
// 	authorizer Authorizer,
// 	pusher eventPusher,
// 	auditer Auditer,
// ) (*Context, error) {
//
// 	ctx := NewContext()
// 	if err := ctx.ReadElementalRequest(request); err != nil {
// 		return nil, err
// 	}
//
// 	ctx.Request.StartTracing()
// 	defer ctx.Request.FinishTracing()
//
// 	if err := CheckAuthentication(authenticator, ctx); err != nil {
// 		audit(auditer, ctx, err)
// 		return nil, err
// 	}
//
// 	if err := CheckAuthorization(authorizer, ctx); err != nil {
// 		audit(auditer, ctx, err)
// 		return nil, err
// 	}
//
// 	proc, _ := processorFinder(request.Identity)
//
// 	if _, ok := proc.(RetrieveProcessor); !ok {
// 		err := notImplementedErr(request)
// 		audit(auditer, ctx, err)
// 		return nil, err
// 	}
//
// 	if err := proc.(RetrieveProcessor).ProcessRetrieve(ctx); err != nil {
// 		audit(auditer, ctx, err)
// 		return nil, err
// 	}
//
// 	if ctx.HasEvents() {
// 		pusher(ctx.Events()...)
// 	}
//
// 	audit(auditer, ctx, nil)
//
// 	return ctx, nil
// }
