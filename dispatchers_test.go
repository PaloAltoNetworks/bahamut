package bahamut

import (
	"net/http"
	"testing"

	"github.com/aporeto-inc/elemental"
	. "github.com/smartystreets/goconvey/convey"
)

// TestDispatchers_dispatchRetrieveManyOperation tests dispatchRetrieveManyOperation method
func TestDispatchers_dispatchRetrieveManyOperation(t *testing.T) {

	Convey("Given I have a processor that handle ProcessRetrieveMany function", t, func() {
		request := elemental.NewRequest()

		processorFinder := func(identity elemental.Identity) (Processor, error) {
			return &MockCompleteProcessor{}, nil
		}

		factory := func(identity string, version int) elemental.Identifiable {
			return &elemental.MockIdentifiable{}
		}

		auditer := &MockAuditer{}

		ctx, err := dispatchRetrieveManyOperation(request, processorFinder, factory, nil, nil, nil, auditer)

		expectedHits := 1

		Convey("Then I should have no error and context should be initiated", func() {
			So(err, ShouldBeNil)
			So(ctx, ShouldNotBeNil)
			So(auditer.Hits, ShouldEqual, expectedHits)
		})
	})

	Convey("Given I have a processor that handle ProcessRetrieveMany function with error", t, func() {
		request := elemental.NewRequest()

		processorFinder := func(identity elemental.Identity) (Processor, error) {
			return &MockCompleteProcessor{
				ExpectedError: elemental.NewError("Error", "Bad request.", "bahamut-test", http.StatusBadRequest),
			}, nil
		}

		factory := func(identity string, version int) elemental.Identifiable {
			return &elemental.MockIdentifiable{}
		}

		auditer := &MockAuditer{}

		ctx, err := dispatchRetrieveManyOperation(request, processorFinder, factory, nil, nil, nil, auditer)

		expectedError := "error 400 (bahamut-test): Error: Bad request."
		expectedHits := 1

		Convey("Then I should get a bahamut error and no context", func() {
			So(err.Error(), ShouldEqual, expectedError)
			So(ctx, ShouldBeNil)
			So(auditer.Hits, ShouldEqual, expectedHits)
		})
	})

	Convey("Given I have a processor that does not handle ProcessRetrieveMany function", t, func() {
		request := elemental.NewRequest()
		request.Operation = elemental.OperationRetrieveMany
		request.Identity = elemental.MakeIdentity("Fake", "Test")

		processorFinder := func(identity elemental.Identity) (Processor, error) {
			return &MockProcessor{}, nil
		}

		factory := func(identity string, version int) elemental.Identifiable {
			return &elemental.MockIdentifiable{}
		}

		auditer := &MockAuditer{}

		expectedError := "error 501 (bahamut): Not implemented: No handler for operation retrieve-many on Fake"
		expectedHits := 1

		ctx, err := dispatchRetrieveManyOperation(request, processorFinder, factory, nil, nil, nil, auditer)

		Convey("Then I should get a bahamut error and no context", func() {
			So(err.Error(), ShouldEqual, expectedError)
			So(ctx, ShouldBeNil)
			So(auditer.Hits, ShouldEqual, expectedHits)
		})
	})

	Convey("Given I have a processor that handle ProcessRetrieveMany function and an authenticator that is not ExpectedAuthenticatedResult", t, func() {
		request := elemental.NewRequest()

		processorFinder := func(identity elemental.Identity) (Processor, error) {
			return &MockProcessor{}, nil
		}

		factory := func(identity string, version int) elemental.Identifiable {
			return &elemental.MockIdentifiable{}
		}

		authenticator := &MockAuth{
			DefinedHasError: true,
			ExpectedError:   elemental.NewError("Error", "Authenticator does not authenticate.", "bahamut-test", http.StatusInternalServerError),
		}

		auditer := &MockAuditer{}

		expectedError := "error 500 (bahamut-test): Error: Authenticator does not authenticate."
		expectedHits := 1

		ctx, err := dispatchRetrieveManyOperation(request, processorFinder, factory, authenticator, nil, nil, auditer)

		Convey("Then I should get a bahamut error and no context", func() {
			So(err.Error(), ShouldEqual, expectedError)
			So(ctx, ShouldBeNil)
			So(auditer.Hits, ShouldEqual, expectedHits)
		})
	})

	Convey("Given I have a processor that handle ProcessRetrieveMany function and an authorizer that is not authorize", t, func() {
		request := elemental.NewRequest()

		processorFinder := func(identity elemental.Identity) (Processor, error) {
			return &MockProcessor{}, nil
		}

		factory := func(identity string, version int) elemental.Identifiable {
			return &elemental.MockIdentifiable{}
		}

		authenticator := &MockAuth{
			ExpectedAuthenticatedResult: true,
		}

		authorizer := &MockAuth{
			DefinedHasError: true,
			ExpectedError:   elemental.NewError("Error", "Authorizer does not authorize.", "bahamut-test", http.StatusInternalServerError),
		}

		auditer := &MockAuditer{}

		expectedError := "error 500 (bahamut-test): Error: Authorizer does not authorize."
		expectedHits := 1

		ctx, err := dispatchRetrieveManyOperation(request, processorFinder, factory, authenticator, authorizer, nil, auditer)

		Convey("Then I should get a bahamut error and no context", func() {
			So(err.Error(), ShouldEqual, expectedError)
			So(ctx, ShouldBeNil)
			So(auditer.Hits, ShouldEqual, expectedHits)
		})
	})
}

// TestDispatchers_dispatchRetrieveOperation tests dispatchRetrieveOperation method
func TestDispatchers_dispatchRetrieveOperation(t *testing.T) {

	Convey("Given I have a processor that handle ProcessRetrieve function", t, func() {
		request := elemental.NewRequest()

		processorFinder := func(identity elemental.Identity) (Processor, error) {
			return &MockCompleteProcessor{}, nil
		}

		factory := func(identity string, version int) elemental.Identifiable {
			return &elemental.MockIdentifiable{}
		}

		auditer := &MockAuditer{}

		ctx, err := dispatchRetrieveOperation(request, processorFinder, factory, nil, nil, nil, auditer)

		expectedHits := 1

		Convey("Then I should have no error and context should be initiated", func() {
			So(err, ShouldBeNil)
			So(ctx, ShouldNotBeNil)
			So(auditer.Hits, ShouldEqual, expectedHits)
		})
	})

	Convey("Given I have a processor that handle ProcessRetrieve function with error", t, func() {
		request := elemental.NewRequest()

		processorFinder := func(identity elemental.Identity) (Processor, error) {
			return &MockCompleteProcessor{
				ExpectedError: elemental.NewError("Error", "Bad request.", "bahamut-test", http.StatusBadRequest),
			}, nil
		}

		factory := func(identity string, version int) elemental.Identifiable {
			return &elemental.MockIdentifiable{}
		}

		auditer := &MockAuditer{}

		ctx, err := dispatchRetrieveOperation(request, processorFinder, factory, nil, nil, nil, auditer)

		expectedError := "error 400 (bahamut-test): Error: Bad request."
		expectedHits := 1

		Convey("Then I should get a bahamut error and no context", func() {
			So(err.Error(), ShouldEqual, expectedError)
			So(ctx, ShouldBeNil)
			So(auditer.Hits, ShouldEqual, expectedHits)
		})
	})

	Convey("Given I have a processor that does not handle ProcessRetrieve function", t, func() {
		request := elemental.NewRequest()
		request.Operation = elemental.OperationRetrieveMany
		request.Identity = elemental.MakeIdentity("Fake", "Test")

		processorFinder := func(identity elemental.Identity) (Processor, error) {
			return &MockProcessor{}, nil
		}

		factory := func(identity string, version int) elemental.Identifiable {
			return &elemental.MockIdentifiable{}
		}

		auditer := &MockAuditer{}

		expectedError := "error 501 (bahamut): Not implemented: No handler for operation retrieve-many on Fake"
		expectedHits := 1

		ctx, err := dispatchRetrieveOperation(request, processorFinder, factory, nil, nil, nil, auditer)

		Convey("Then I should get a bahamut error and no context", func() {
			So(err.Error(), ShouldEqual, expectedError)
			So(ctx, ShouldBeNil)
			So(auditer.Hits, ShouldEqual, expectedHits)
		})
	})

	Convey("Given I have a processor that does not handle ProcessRetrieve function and an authenticator that is not ExpectedAuthenticatedResult", t, func() {
		request := elemental.NewRequest()

		processorFinder := func(identity elemental.Identity) (Processor, error) {
			return &MockProcessor{}, nil
		}

		factory := func(identity string, version int) elemental.Identifiable {
			return &elemental.MockIdentifiable{}
		}

		authenticator := &MockAuth{
			DefinedHasError: true,
			ExpectedError:   elemental.NewError("Error", "Authenticator does not authenticate.", "bahamut-test", http.StatusInternalServerError),
		}

		auditer := &MockAuditer{}

		expectedError := "error 500 (bahamut-test): Error: Authenticator does not authenticate."
		expectedHits := 1

		ctx, err := dispatchRetrieveOperation(request, processorFinder, factory, authenticator, nil, nil, auditer)

		Convey("Then I should get a bahamut error and no context", func() {
			So(err.Error(), ShouldEqual, expectedError)
			So(ctx, ShouldBeNil)
			So(auditer.Hits, ShouldEqual, expectedHits)
		})
	})

	Convey("Given I have a processor that does not handle ProcessRetrieve function and an authorizer that is not authorize", t, func() {
		request := elemental.NewRequest()

		processorFinder := func(identity elemental.Identity) (Processor, error) {
			return &MockProcessor{}, nil
		}

		factory := func(identity string, version int) elemental.Identifiable {
			return &elemental.MockIdentifiable{}
		}

		authenticator := &MockAuth{
			ExpectedAuthenticatedResult: true,
		}

		authorizer := &MockAuth{
			DefinedHasError: true,
			ExpectedError:   elemental.NewError("Error", "Authorizer does not authorize.", "bahamut-test", http.StatusInternalServerError),
		}

		auditer := &MockAuditer{}

		expectedError := "error 500 (bahamut-test): Error: Authorizer does not authorize."
		expectedHits := 1

		ctx, err := dispatchRetrieveOperation(request, processorFinder, factory, authenticator, authorizer, nil, auditer)

		Convey("Then I should get a bahamut error and no context", func() {
			So(err.Error(), ShouldEqual, expectedError)
			So(ctx, ShouldBeNil)
			So(auditer.Hits, ShouldEqual, expectedHits)
		})
	})
}

// TestDispatchers_dispatchCreateOperation tests dispatchCreateOperation method
func TestDispatchers_dispatchCreateOperation(t *testing.T) {

	Convey("Given I have a processor that handle ProcessCreate function", t, func() {
		request := elemental.NewRequest()
		request.Data = []byte(`{"ID": "1234", "Name": "Fake"}`)

		processorFinder := func(identity elemental.Identity) (Processor, error) {
			return &MockCompleteProcessor{}, nil
		}

		factory := func(identity string, version int) elemental.Identifiable {
			return &elemental.MockIdentifiable{}
		}

		auditer := &MockAuditer{}

		ctx, err := dispatchCreateOperation(request, processorFinder, factory, nil, nil, nil, auditer, false, nil)

		expectedHits := 1

		Convey("Then I should have no error and context should be initiated", func() {
			So(err, ShouldBeNil)
			So(ctx, ShouldNotBeNil)
			So(ctx.InputData, ShouldNotBeNil)
			So(auditer.Hits, ShouldEqual, expectedHits)
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

		ctx, err := dispatchCreateOperation(request, processorFinder, factory, nil, nil, nil, auditer, true, nil)

		Convey("Then I should have a 423 error and context should be nil", func() {
			So(err, ShouldNotBeNil)
			So(err.(elemental.Error).Code, ShouldEqual, http.StatusLocked)
			So(ctx, ShouldBeNil)
		})
	})

	Convey("Given I have a processor that handle ProcessCreate function with error", t, func() {
		request := elemental.NewRequest()
		request.Data = []byte(`{"ID": "1234", "Name": "Fake"}`)

		processorFinder := func(identity elemental.Identity) (Processor, error) {
			return &MockCompleteProcessor{
				ExpectedError: elemental.NewError("Error", "Bad request.", "bahamut-test", http.StatusBadRequest),
			}, nil
		}

		factory := func(identity string, version int) elemental.Identifiable {
			return &elemental.MockIdentifiable{}
		}

		auditer := &MockAuditer{}

		ctx, err := dispatchCreateOperation(request, processorFinder, factory, nil, nil, nil, auditer, false, nil)

		expectedError := "error 400 (bahamut-test): Error: Bad request."
		expectedHits := 1

		Convey("Then I should get a bahamut error and no context", func() {
			So(err.Error(), ShouldEqual, expectedError)
			So(ctx, ShouldBeNil)
			So(auditer.Hits, ShouldEqual, expectedHits)
		})
	})

	Convey("Given I have a processor that handle ProcessCreate function with an empty JSON", t, func() {
		request := elemental.NewRequest()

		processorFinder := func(identity elemental.Identity) (Processor, error) {
			return &MockCompleteProcessor{}, nil
		}

		factory := func(identity string, version int) elemental.Identifiable {
			return &elemental.MockIdentifiable{}
		}

		auditer := &MockAuditer{}

		ctx, err := dispatchCreateOperation(request, processorFinder, factory, nil, nil, nil, auditer, false, nil)

		expectedError := "error 400 (elemental): Bad Request: Something went wrong in the server when reading the body of the request"
		expectedHits := 1

		Convey("Then I should get a bahamut error and no context", func() {
			So(err.Error(), ShouldEqual, expectedError)
			So(ctx, ShouldBeNil)
			So(auditer.Hits, ShouldEqual, expectedHits)
		})
	})

	Convey("Given I have a processor that handle ProcessCreate function with an invalid JSON", t, func() {
		request := elemental.NewRequest()
		request.Data = []byte(`An invalid JSON`)

		processorFinder := func(identity elemental.Identity) (Processor, error) {
			return &MockCompleteProcessor{}, nil
		}

		factory := func(identity string, version int) elemental.Identifiable {
			return &elemental.MockIdentifiable{}
		}

		auditer := &MockAuditer{}

		ctx, err := dispatchCreateOperation(request, processorFinder, factory, nil, nil, nil, auditer, false, nil)

		expectedError := "error 400 (elemental): Bad Request: Invalid JSON"
		expectedHits := 1

		Convey("Then I should get a bahamut error and no context", func() {
			So(err.Error(), ShouldEqual, expectedError)
			So(ctx, ShouldBeNil)
			So(auditer.Hits, ShouldEqual, expectedHits)
		})
	})

	Convey("Given I have a processor that handle ProcessCreate function with an invalid object", t, func() {
		request := elemental.NewRequest()
		request.Data = []byte(`{"ID": "1234", "Name": "Fake"}`)

		processorFinder := func(identity elemental.Identity) (Processor, error) {
			return &MockCompleteProcessor{}, nil
		}

		factory := func(identity string, version int) elemental.Identifiable {
			return &elemental.MockIdentifiable{
				ExpectedValidationError: elemental.NewError("Error", "Object validation has failed.", "bahamut-test", http.StatusBadRequest),
			}
		}

		auditer := &MockAuditer{}

		ctx, err := dispatchCreateOperation(request, processorFinder, factory, nil, nil, nil, auditer, false, nil)

		expectedError := "error 400 (bahamut-test): Error: Object validation has failed."
		expectedHits := 1

		Convey("Then I should get a bahamut error and no context", func() {
			So(err.Error(), ShouldEqual, expectedError)
			So(ctx, ShouldBeNil)
			So(auditer.Hits, ShouldEqual, expectedHits)
		})
	})

	Convey("Given I have a processor that does not handle ProcessCreate function", t, func() {
		request := elemental.NewRequest()
		request.Operation = elemental.OperationRetrieveMany
		request.Identity = elemental.MakeIdentity("Fake", "Test")

		processorFinder := func(identity elemental.Identity) (Processor, error) {
			return &MockProcessor{}, nil
		}

		factory := func(identity string, version int) elemental.Identifiable {
			return &elemental.MockIdentifiable{}
		}

		auditer := &MockAuditer{}

		expectedError := "error 501 (bahamut): Not implemented: No handler for operation retrieve-many on Fake"
		expectedHits := 1

		ctx, err := dispatchCreateOperation(request, processorFinder, factory, nil, nil, nil, auditer, false, nil)

		Convey("Then I should get a bahamut error and no context", func() {
			So(err.Error(), ShouldEqual, expectedError)
			So(ctx, ShouldBeNil)
			So(auditer.Hits, ShouldEqual, expectedHits)
		})
	})

	Convey("Given I have a processor that does not handle ProcessCreate function and an authenticator that is not ExpectedAuthenticatedResult", t, func() {
		request := elemental.NewRequest()

		processorFinder := func(identity elemental.Identity) (Processor, error) {
			return &MockProcessor{}, nil
		}

		factory := func(identity string, version int) elemental.Identifiable {
			return &elemental.MockIdentifiable{}
		}

		authenticator := &MockAuth{
			DefinedHasError: true,
			ExpectedError:   elemental.NewError("Error", "Authenticator does not authenticate.", "bahamut-test", http.StatusInternalServerError),
		}

		auditer := &MockAuditer{}

		expectedError := "error 500 (bahamut-test): Error: Authenticator does not authenticate."
		expectedHits := 1

		ctx, err := dispatchCreateOperation(request, processorFinder, factory, authenticator, nil, nil, auditer, false, nil)

		Convey("Then I should get a bahamut error and no context", func() {
			So(err.Error(), ShouldEqual, expectedError)
			So(ctx, ShouldBeNil)
			So(auditer.Hits, ShouldEqual, expectedHits)
		})
	})

	Convey("Given I have a processor that does not handle ProcessCreate function and an authorizer that is not authorize", t, func() {
		request := elemental.NewRequest()

		processorFinder := func(identity elemental.Identity) (Processor, error) {
			return &MockProcessor{}, nil
		}

		factory := func(identity string, version int) elemental.Identifiable {
			return &elemental.MockIdentifiable{}
		}

		authenticator := &MockAuth{
			ExpectedAuthenticatedResult: true,
		}

		authorizer := &MockAuth{
			DefinedHasError: true,
			ExpectedError:   elemental.NewError("Error", "Authorizer does not authorize.", "bahamut-test", http.StatusInternalServerError),
		}

		auditer := &MockAuditer{}

		expectedError := "error 500 (bahamut-test): Error: Authorizer does not authorize."
		expectedHits := 1

		ctx, err := dispatchCreateOperation(request, processorFinder, factory, authenticator, authorizer, nil, auditer, false, nil)

		Convey("Then I should get a bahamut error and no context", func() {
			So(err.Error(), ShouldEqual, expectedError)
			So(ctx, ShouldBeNil)
			So(auditer.Hits, ShouldEqual, expectedHits)
		})
	})
}

// TestDispatchers_dispatchUpdateOperation tests dispatchUpdateOperation method
func TestDispatchers_dispatchUpdateOperation(t *testing.T) {

	Convey("Given I have a processor that handle ProcessUpdate function", t, func() {
		request := elemental.NewRequest()
		request.Data = []byte(`{"ID": "1234", "Name": "Fake"}`)

		processorFinder := func(identity elemental.Identity) (Processor, error) {
			return &MockCompleteProcessor{}, nil
		}

		factory := func(identity string, version int) elemental.Identifiable {
			return &elemental.MockIdentifiable{}
		}

		auditer := &MockAuditer{}

		ctx, err := dispatchUpdateOperation(request, processorFinder, factory, nil, nil, nil, auditer, false, nil)

		expectedHits := 1

		Convey("Then I should have no error and context should be initiated", func() {
			So(err, ShouldBeNil)
			So(ctx, ShouldNotBeNil)
			So(ctx.InputData, ShouldNotBeNil)
			So(auditer.Hits, ShouldEqual, expectedHits)
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

		ctx, err := dispatchUpdateOperation(request, processorFinder, factory, nil, nil, nil, auditer, true, nil)

		Convey("Then I should have a 423 error and context should be nil", func() {
			So(err, ShouldNotBeNil)
			So(err.(elemental.Error).Code, ShouldEqual, http.StatusLocked)
			So(ctx, ShouldBeNil)
		})
	})

	Convey("Given I have a processor that handle ProcessUpdate function with error", t, func() {
		request := elemental.NewRequest()
		request.Data = []byte(`{"ID": "1234", "Name": "Fake"}`)

		processorFinder := func(identity elemental.Identity) (Processor, error) {
			return &MockCompleteProcessor{
				ExpectedError: elemental.NewError("Error", "Bad request.", "bahamut-test", http.StatusBadRequest),
			}, nil
		}

		factory := func(identity string, version int) elemental.Identifiable {
			return &elemental.MockIdentifiable{}
		}

		auditer := &MockAuditer{}

		ctx, err := dispatchUpdateOperation(request, processorFinder, factory, nil, nil, nil, auditer, false, nil)

		expectedError := "error 400 (bahamut-test): Error: Bad request."
		expectedHits := 1

		Convey("Then I should get a bahamut error and no context", func() {
			So(err.Error(), ShouldEqual, expectedError)
			So(ctx, ShouldBeNil)
			So(auditer.Hits, ShouldEqual, expectedHits)
		})
	})

	Convey("Given I have a processor that handle ProcessUpdate function with an empty JSON", t, func() {
		request := elemental.NewRequest()

		processorFinder := func(identity elemental.Identity) (Processor, error) {
			return &MockCompleteProcessor{}, nil
		}

		factory := func(identity string, version int) elemental.Identifiable {
			return &elemental.MockIdentifiable{}
		}

		auditer := &MockAuditer{}

		ctx, err := dispatchUpdateOperation(request, processorFinder, factory, nil, nil, nil, auditer, false, nil)

		expectedError := "error 400 (elemental): Bad Request: Something went wrong in the server when reading the body of the request"
		expectedHits := 1

		Convey("Then I should get a bahamut error and no context", func() {
			So(err.Error(), ShouldEqual, expectedError)
			So(ctx, ShouldBeNil)
			So(auditer.Hits, ShouldEqual, expectedHits)
		})
	})

	Convey("Given I have a processor that handle ProcessUpdate function with an invalid JSON", t, func() {
		request := elemental.NewRequest()
		request.Data = []byte(`An invalid JSON`)

		processorFinder := func(identity elemental.Identity) (Processor, error) {
			return &MockCompleteProcessor{}, nil
		}

		factory := func(identity string, version int) elemental.Identifiable {
			return &elemental.MockIdentifiable{}
		}

		auditer := &MockAuditer{}

		ctx, err := dispatchUpdateOperation(request, processorFinder, factory, nil, nil, nil, auditer, false, nil)

		expectedError := "error 400 (elemental): Bad Request: Invalid JSON"
		expectedHits := 1

		Convey("Then I should get a bahamut error and no context", func() {
			So(err.Error(), ShouldEqual, expectedError)
			So(ctx, ShouldBeNil)
			So(auditer.Hits, ShouldEqual, expectedHits)
		})
	})

	Convey("Given I have a processor that handle ProcessUpdate function with an invalid object", t, func() {
		request := elemental.NewRequest()
		request.Data = []byte(`{"ID": "1234", "Name": "Fake"}`)

		processorFinder := func(identity elemental.Identity) (Processor, error) {
			return &MockCompleteProcessor{}, nil
		}

		factory := func(identity string, version int) elemental.Identifiable {
			return &elemental.MockIdentifiable{
				ExpectedValidationError: elemental.NewError("Error", "Object validation has failed.", "bahamut-test", http.StatusBadRequest),
			}
		}

		auditer := &MockAuditer{}

		ctx, err := dispatchUpdateOperation(request, processorFinder, factory, nil, nil, nil, auditer, false, nil)

		expectedError := "error 400 (bahamut-test): Error: Object validation has failed."
		expectedHits := 1

		Convey("Then I should get a bahamut error and no context", func() {
			So(err.Error(), ShouldEqual, expectedError)
			So(ctx, ShouldBeNil)
			So(auditer.Hits, ShouldEqual, expectedHits)
		})
	})

	Convey("Given I have a processor that does not handle ProcessUpdate function", t, func() {
		request := elemental.NewRequest()
		request.Operation = elemental.OperationRetrieveMany
		request.Identity = elemental.MakeIdentity("Fake", "Test")

		processorFinder := func(identity elemental.Identity) (Processor, error) {
			return &MockProcessor{}, nil
		}

		factory := func(identity string, version int) elemental.Identifiable {
			return &elemental.MockIdentifiable{}
		}

		auditer := &MockAuditer{}

		expectedError := "error 501 (bahamut): Not implemented: No handler for operation retrieve-many on Fake"
		expectedHits := 1

		ctx, err := dispatchUpdateOperation(request, processorFinder, factory, nil, nil, nil, auditer, false, nil)

		Convey("Then I should get a bahamut error and no context", func() {
			So(err.Error(), ShouldEqual, expectedError)
			So(ctx, ShouldBeNil)
			So(auditer.Hits, ShouldEqual, expectedHits)
		})
	})

	Convey("Given I have a processor that does not handle ProcessUpdate function and an authenticator that is not ExpectedAuthenticatedResult", t, func() {
		request := elemental.NewRequest()

		processorFinder := func(identity elemental.Identity) (Processor, error) {
			return &MockProcessor{}, nil
		}

		factory := func(identity string, version int) elemental.Identifiable {
			return &elemental.MockIdentifiable{}
		}

		authenticator := &MockAuth{
			DefinedHasError: true,
			ExpectedError:   elemental.NewError("Error", "Authenticator does not authenticate.", "bahamut-test", http.StatusInternalServerError),
		}

		auditer := &MockAuditer{}

		expectedError := "error 500 (bahamut-test): Error: Authenticator does not authenticate."
		expectedHits := 1

		ctx, err := dispatchUpdateOperation(request, processorFinder, factory, authenticator, nil, nil, auditer, false, nil)

		Convey("Then I should get a bahamut error and no context", func() {
			So(err.Error(), ShouldEqual, expectedError)
			So(ctx, ShouldBeNil)
			So(auditer.Hits, ShouldEqual, expectedHits)
		})
	})

	Convey("Given I have a processor that does not handle ProcessUpdate function and an authorizer that is not authorize", t, func() {
		request := elemental.NewRequest()

		processorFinder := func(identity elemental.Identity) (Processor, error) {
			return &MockProcessor{}, nil
		}

		factory := func(identity string, version int) elemental.Identifiable {
			return &elemental.MockIdentifiable{}
		}

		authenticator := &MockAuth{
			ExpectedAuthenticatedResult: true,
		}

		authorizer := &MockAuth{
			DefinedHasError: true,
			ExpectedError:   elemental.NewError("Error", "Authorizer does not authorize.", "bahamut-test", http.StatusInternalServerError),
		}

		auditer := &MockAuditer{}

		expectedError := "error 500 (bahamut-test): Error: Authorizer does not authorize."
		expectedHits := 1

		ctx, err := dispatchUpdateOperation(request, processorFinder, factory, authenticator, authorizer, nil, auditer, false, nil)

		Convey("Then I should get a bahamut error and no context", func() {
			So(err.Error(), ShouldEqual, expectedError)
			So(ctx, ShouldBeNil)
			So(auditer.Hits, ShouldEqual, expectedHits)
		})
	})
}

// TestDispatchers_dispatchDeleteOperation tests dispatchDeleteOperation method
func TestDispatchers_dispatchDeleteOperation(t *testing.T) {

	Convey("Given I have a processor that handle ProcessDelete function", t, func() {
		request := elemental.NewRequest()
		request.Data = []byte(`{"ID": "1234", "Name": "Fake"}`)

		processorFinder := func(identity elemental.Identity) (Processor, error) {
			return &MockCompleteProcessor{}, nil
		}

		factory := func(identity string, version int) elemental.Identifiable {
			return &elemental.MockIdentifiable{}
		}

		auditer := &MockAuditer{}

		ctx, err := dispatchDeleteOperation(request, processorFinder, factory, nil, nil, nil, auditer, false, nil)

		expectedHits := 1

		Convey("Then I should have no error and context should be initiated", func() {
			So(err, ShouldBeNil)
			So(ctx, ShouldNotBeNil)
			So(auditer.Hits, ShouldEqual, expectedHits)
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

		ctx, err := dispatchDeleteOperation(request, processorFinder, factory, nil, nil, nil, auditer, true, nil)

		Convey("Then I should have a 423 error and context should be nil", func() {
			So(err, ShouldNotBeNil)
			So(err.(elemental.Error).Code, ShouldEqual, http.StatusLocked)
			So(ctx, ShouldBeNil)
		})
	})

	Convey("Given I have a processor that handle ProcessDelete function with error", t, func() {
		request := elemental.NewRequest()
		request.Data = []byte(`{"ID": "1234", "Name": "Fake"}`)

		processorFinder := func(identity elemental.Identity) (Processor, error) {
			return &MockCompleteProcessor{
				ExpectedError: elemental.NewError("Error", "Bad request.", "bahamut-test", http.StatusBadRequest),
			}, nil
		}

		factory := func(identity string, version int) elemental.Identifiable {
			return &elemental.MockIdentifiable{}
		}

		auditer := &MockAuditer{}

		ctx, err := dispatchDeleteOperation(request, processorFinder, factory, nil, nil, nil, auditer, false, nil)

		expectedError := "error 400 (bahamut-test): Error: Bad request."
		expectedHits := 1

		Convey("Then I should get a bahamut error and no context", func() {
			So(err.Error(), ShouldEqual, expectedError)
			So(ctx, ShouldBeNil)
			So(auditer.Hits, ShouldEqual, expectedHits)
		})
	})

	Convey("Given I have a processor that does not handle ProcessDelete function", t, func() {
		request := elemental.NewRequest()
		request.Operation = elemental.OperationRetrieveMany
		request.Identity = elemental.MakeIdentity("Fake", "Test")

		processorFinder := func(identity elemental.Identity) (Processor, error) {
			return &MockProcessor{}, nil
		}

		factory := func(identity string, version int) elemental.Identifiable {
			return &elemental.MockIdentifiable{}
		}

		auditer := &MockAuditer{}

		expectedError := "error 501 (bahamut): Not implemented: No handler for operation retrieve-many on Fake"
		expectedHits := 1

		ctx, err := dispatchDeleteOperation(request, processorFinder, factory, nil, nil, nil, auditer, false, nil)

		Convey("Then I should get a bahamut error and no context", func() {
			So(err.Error(), ShouldEqual, expectedError)
			So(ctx, ShouldBeNil)
			So(auditer.Hits, ShouldEqual, expectedHits)
		})
	})

	Convey("Given I have a processor that does not handle ProcessDelete function and an authenticator that is not ExpectedAuthenticatedResult", t, func() {
		request := elemental.NewRequest()

		processorFinder := func(identity elemental.Identity) (Processor, error) {
			return &MockProcessor{}, nil
		}

		factory := func(identity string, version int) elemental.Identifiable {
			return &elemental.MockIdentifiable{}
		}

		authenticator := &MockAuth{
			DefinedHasError: true,
			ExpectedError:   elemental.NewError("Error", "Authenticator does not authenticate.", "bahamut-test", http.StatusInternalServerError),
		}

		auditer := &MockAuditer{}

		expectedError := "error 500 (bahamut-test): Error: Authenticator does not authenticate."
		expectedHits := 1

		ctx, err := dispatchDeleteOperation(request, processorFinder, factory, authenticator, nil, nil, auditer, false, nil)

		Convey("Then I should get a bahamut error and no context", func() {
			So(err.Error(), ShouldEqual, expectedError)
			So(ctx, ShouldBeNil)
			So(auditer.Hits, ShouldEqual, expectedHits)
		})
	})

	Convey("Given I have a processor that does not handle ProcessDelete function and an authorizer that is not authorize", t, func() {
		request := elemental.NewRequest()

		processorFinder := func(identity elemental.Identity) (Processor, error) {
			return &MockProcessor{}, nil
		}

		factory := func(identity string, version int) elemental.Identifiable {
			return &elemental.MockIdentifiable{}
		}

		authenticator := &MockAuth{
			ExpectedAuthenticatedResult: true,
		}

		authorizer := &MockAuth{
			DefinedHasError: true,
			ExpectedError:   elemental.NewError("Error", "Authorizer does not authorize.", "bahamut-test", http.StatusInternalServerError),
		}

		auditer := &MockAuditer{}

		expectedError := "error 500 (bahamut-test): Error: Authorizer does not authorize."
		expectedHits := 1

		ctx, err := dispatchDeleteOperation(request, processorFinder, factory, authenticator, authorizer, nil, auditer, false, nil)

		Convey("Then I should get a bahamut error and no context", func() {
			So(err.Error(), ShouldEqual, expectedError)
			So(ctx, ShouldBeNil)
			So(auditer.Hits, ShouldEqual, expectedHits)
		})
	})
}

// TestDispatchers_dispatchPatchOperation tests dispatchPatchOperation method
func TestDispatchers_dispatchPatchOperation(t *testing.T) {

	Convey("Given I have a processor that handle ProcessPatch function", t, func() {
		request := elemental.NewRequest()
		request.Data = []byte(`{"ID": "1234", "Name": "Fake"}`)

		processorFinder := func(identity elemental.Identity) (Processor, error) {
			return &MockCompleteProcessor{}, nil
		}

		factory := func(identity string, version int) elemental.Identifiable {
			return &elemental.MockIdentifiable{}
		}

		auditer := &MockAuditer{}

		ctx, err := dispatchPatchOperation(request, processorFinder, factory, nil, nil, nil, auditer, false, nil)

		expectedHits := 1

		Convey("Then I should have no error and context should be initiated", func() {
			So(err, ShouldBeNil)
			So(ctx, ShouldNotBeNil)
			So(auditer.Hits, ShouldEqual, expectedHits)
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

		ctx, err := dispatchPatchOperation(request, processorFinder, factory, nil, nil, nil, auditer, true, nil)

		Convey("Then I should have a 423 error and context should be nil", func() {
			So(err, ShouldNotBeNil)
			So(err.(elemental.Error).Code, ShouldEqual, http.StatusLocked)
			So(ctx, ShouldBeNil)
		})
	})

	Convey("Given I have a processor that handle ProcessPatch function with an invalid JSON", t, func() {
		request := elemental.NewRequest()
		request.Data = []byte(`Invalid JSON`)

		processorFinder := func(identity elemental.Identity) (Processor, error) {
			return &MockCompleteProcessor{}, nil
		}

		factory := func(identity string, version int) elemental.Identifiable {
			return &elemental.MockIdentifiable{}
		}

		auditer := &MockAuditer{}

		ctx, err := dispatchPatchOperation(request, processorFinder, factory, nil, nil, nil, auditer, false, nil)

		expectedError := "error 400 (elemental): Bad Request: Invalid JSON"
		expectedHits := 1

		Convey("Then I should have no error and context should be initiated", func() {
			So(err.Error(), ShouldEqual, expectedError)
			So(ctx, ShouldBeNil)
			So(auditer.Hits, ShouldEqual, expectedHits)
		})
	})

	Convey("Given I have a processor that handle ProcessPatch function with error", t, func() {
		request := elemental.NewRequest()
		request.Data = []byte(`{"ID": "1234", "Name": "Fake"}`)

		processorFinder := func(identity elemental.Identity) (Processor, error) {
			return &MockCompleteProcessor{
				ExpectedError: elemental.NewError("Error", "Bad request.", "bahamut-test", http.StatusBadRequest),
			}, nil
		}

		factory := func(identity string, version int) elemental.Identifiable {
			return &elemental.MockIdentifiable{}
		}

		auditer := &MockAuditer{}

		ctx, err := dispatchPatchOperation(request, processorFinder, factory, nil, nil, nil, auditer, false, nil)

		expectedError := "error 400 (bahamut-test): Error: Bad request."
		expectedHits := 1

		Convey("Then I should get a bahamut error and no context", func() {
			So(err.Error(), ShouldEqual, expectedError)
			So(ctx, ShouldBeNil)
			So(auditer.Hits, ShouldEqual, expectedHits)
		})
	})

	Convey("Given I have a processor that does not handle ProcessPatch function", t, func() {
		request := elemental.NewRequest()
		request.Operation = elemental.OperationRetrieveMany
		request.Identity = elemental.MakeIdentity("Fake", "Test")

		processorFinder := func(identity elemental.Identity) (Processor, error) {
			return &MockProcessor{}, nil
		}

		factory := func(identity string, version int) elemental.Identifiable {
			return &elemental.MockIdentifiable{}
		}

		auditer := &MockAuditer{}

		expectedError := "error 501 (bahamut): Not implemented: No handler for operation retrieve-many on Fake"
		expectedHits := 1

		ctx, err := dispatchPatchOperation(request, processorFinder, factory, nil, nil, nil, auditer, false, nil)

		Convey("Then I should get a bahamut error and no context", func() {
			So(err.Error(), ShouldEqual, expectedError)
			So(ctx, ShouldBeNil)
			So(auditer.Hits, ShouldEqual, expectedHits)
		})
	})

	Convey("Given I have a processor that does not handle ProcessPatch function and an authenticator that is not ExpectedAuthenticatedResult", t, func() {
		request := elemental.NewRequest()

		processorFinder := func(identity elemental.Identity) (Processor, error) {
			return &MockProcessor{}, nil
		}

		factory := func(identity string, version int) elemental.Identifiable {
			return &elemental.MockIdentifiable{}
		}

		authenticator := &MockAuth{
			DefinedHasError: true,
			ExpectedError:   elemental.NewError("Error", "Authenticator does not authenticate.", "bahamut-test", http.StatusInternalServerError),
		}

		auditer := &MockAuditer{}

		expectedError := "error 500 (bahamut-test): Error: Authenticator does not authenticate."
		expectedHits := 1

		ctx, err := dispatchPatchOperation(request, processorFinder, factory, authenticator, nil, nil, auditer, false, nil)

		Convey("Then I should get a bahamut error and no context", func() {
			So(err.Error(), ShouldEqual, expectedError)
			So(ctx, ShouldBeNil)
			So(auditer.Hits, ShouldEqual, expectedHits)
		})
	})

	Convey("Given I have a processor that does not handle ProcessPatch function and an authorizer that is not authorize", t, func() {
		request := elemental.NewRequest()

		processorFinder := func(identity elemental.Identity) (Processor, error) {
			return &MockProcessor{}, nil
		}

		factory := func(identity string, version int) elemental.Identifiable {
			return &elemental.MockIdentifiable{}
		}

		authenticator := &MockAuth{
			ExpectedAuthenticatedResult: true,
		}

		authorizer := &MockAuth{
			DefinedHasError: true,
			ExpectedError:   elemental.NewError("Error", "Authorizer does not authorize.", "bahamut-test", http.StatusInternalServerError),
		}

		auditer := &MockAuditer{}

		expectedError := "error 500 (bahamut-test): Error: Authorizer does not authorize."
		expectedHits := 1

		ctx, err := dispatchPatchOperation(request, processorFinder, factory, authenticator, authorizer, nil, auditer, false, nil)

		Convey("Then I should get a bahamut error and no context", func() {
			So(err.Error(), ShouldEqual, expectedError)
			So(ctx, ShouldBeNil)
			So(auditer.Hits, ShouldEqual, expectedHits)
		})
	})
}

// TestDispatchers_dispatchInfoOperation tests dispatchInfoOperation method
func TestDispatchers_dispatchInfoOperation(t *testing.T) {

	Convey("Given I have a processor that handle ProcessInfo function", t, func() {
		request := elemental.NewRequest()
		request.Data = []byte(`{"ID": "1234", "Name": "Fake"}`)

		processorFinder := func(identity elemental.Identity) (Processor, error) {
			return &MockCompleteProcessor{}, nil
		}

		factory := func(identity string, version int) elemental.Identifiable {
			return &elemental.MockIdentifiable{}
		}

		auditer := &MockAuditer{}

		ctx, err := dispatchInfoOperation(request, processorFinder, factory, nil, nil, auditer)

		expectedHits := 1

		Convey("Then I should have no error and context should be initiated", func() {
			So(err, ShouldBeNil)
			So(ctx, ShouldNotBeNil)
			So(auditer.Hits, ShouldEqual, expectedHits)
		})
	})

	Convey("Given I have a processor that handle ProcessInfo function with error", t, func() {
		request := elemental.NewRequest()
		request.Data = []byte(`{"ID": "1234", "Name": "Fake"}`)

		processorFinder := func(identity elemental.Identity) (Processor, error) {
			return &MockCompleteProcessor{
				ExpectedError: elemental.NewError("Error", "Bad request.", "bahamut-test", http.StatusBadRequest),
			}, nil
		}

		factory := func(identity string, version int) elemental.Identifiable {
			return &elemental.MockIdentifiable{}
		}

		auditer := &MockAuditer{}

		ctx, err := dispatchInfoOperation(request, processorFinder, factory, nil, nil, auditer)

		expectedError := "error 400 (bahamut-test): Error: Bad request."
		expectedHits := 1

		Convey("Then I should get a bahamut error and no context", func() {
			So(err.Error(), ShouldEqual, expectedError)
			So(ctx, ShouldBeNil)
			So(auditer.Hits, ShouldEqual, expectedHits)
		})
	})

	Convey("Given I have a processor that does not handle ProcessInfo function", t, func() {
		request := elemental.NewRequest()
		request.Operation = elemental.OperationRetrieveMany
		request.Identity = elemental.MakeIdentity("Fake", "Test")

		processorFinder := func(identity elemental.Identity) (Processor, error) {
			return &MockProcessor{}, nil
		}

		factory := func(identity string, version int) elemental.Identifiable {
			return &elemental.MockIdentifiable{}
		}

		auditer := &MockAuditer{}

		expectedError := "error 501 (bahamut): Not implemented: No handler for operation retrieve-many on Fake"
		expectedHits := 1

		ctx, err := dispatchInfoOperation(request, processorFinder, factory, nil, nil, auditer)

		Convey("Then I should get a bahamut error and no context", func() {
			So(err.Error(), ShouldEqual, expectedError)
			So(ctx, ShouldBeNil)
			So(auditer.Hits, ShouldEqual, expectedHits)
		})
	})

	Convey("Given I have a processor that does not handle ProcessInfo function and an authenticator that is not ExpectedAuthenticatedResult", t, func() {
		request := elemental.NewRequest()

		processorFinder := func(identity elemental.Identity) (Processor, error) {
			return &MockProcessor{}, nil
		}

		factory := func(identity string, version int) elemental.Identifiable {
			return &elemental.MockIdentifiable{}
		}

		authenticator := &MockAuth{
			DefinedHasError: true,
			ExpectedError:   elemental.NewError("Error", "Authenticator does not authenticate.", "bahamut-test", http.StatusInternalServerError),
		}

		auditer := &MockAuditer{}

		expectedError := "error 500 (bahamut-test): Error: Authenticator does not authenticate."
		expectedHits := 1

		ctx, err := dispatchInfoOperation(request, processorFinder, factory, authenticator, nil, auditer)

		Convey("Then I should get a bahamut error and no context", func() {
			So(err.Error(), ShouldEqual, expectedError)
			So(ctx, ShouldBeNil)
			So(auditer.Hits, ShouldEqual, expectedHits)
		})
	})

	Convey("Given I have a processor that does not handle ProcessInfo function and an authorizer that is not authorize", t, func() {
		request := elemental.NewRequest()

		processorFinder := func(identity elemental.Identity) (Processor, error) {
			return &MockProcessor{}, nil
		}

		factory := func(identity string, version int) elemental.Identifiable {
			return &elemental.MockIdentifiable{}
		}

		authenticator := &MockAuth{
			ExpectedAuthenticatedResult: true,
		}

		authorizer := &MockAuth{
			DefinedHasError: true,
			ExpectedError:   elemental.NewError("Error", "Authorizer does not authorize.", "bahamut-test", http.StatusInternalServerError),
		}

		auditer := &MockAuditer{}

		expectedError := "error 500 (bahamut-test): Error: Authorizer does not authorize."
		expectedHits := 1

		ctx, err := dispatchInfoOperation(request, processorFinder, factory, authenticator, authorizer, auditer)

		Convey("Then I should get a bahamut error and no context", func() {
			So(err.Error(), ShouldEqual, expectedError)
			So(ctx, ShouldBeNil)
			So(auditer.Hits, ShouldEqual, expectedHits)
		})
	})
}
