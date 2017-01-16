package bahamut

import (
	"net/http"

	"github.com/aporeto-inc/elemental"
)

func dispatchRetrieveManyOperation(
	request *elemental.Request,
	processorFinder processorFinder,
	factory elemental.IdentifiableFactory,
	authenticator Authenticator,
	authorizer Authorizer,
) (*Context, error) {

	ctx := NewContext()
	ctx.ReadElementalRequest(request)

	if err := CheckAuthentication(authenticator, ctx); err != nil {
		return nil, err
	}

	if err := CheckAuthorization(authorizer, ctx); err != nil {
		return nil, err
	}

	proc, _ := processorFinder(request.Identity)

	if _, ok := proc.(RetrieveManyProcessor); !ok {
		return nil, elemental.NewError("Not implemented", "No handler for retrieving many "+request.Identity.Name, "bahamut", http.StatusNotImplemented)
	}

	if err := proc.(RetrieveManyProcessor).ProcessRetrieveMany(ctx); err != nil {
		return nil, err
	}

	return ctx, nil
}

func dispatchRetrieveOperation(
	request *elemental.Request,
	processorFinder processorFinder,
	factory elemental.IdentifiableFactory,
	authenticator Authenticator,
	authorizer Authorizer,
) (*Context, error) {

	ctx := NewContext()
	ctx.ReadElementalRequest(request)

	if err := CheckAuthentication(authenticator, ctx); err != nil {
		return nil, err
	}

	if err := CheckAuthorization(authorizer, ctx); err != nil {
		return nil, err
	}

	proc, _ := processorFinder(request.Identity)

	if _, ok := proc.(RetrieveProcessor); !ok {
		return nil, elemental.NewError("Not implemented", "No handler for retrieving "+request.Identity.Name, "bahamut", http.StatusNotImplemented)
	}

	if err := proc.(RetrieveProcessor).ProcessRetrieve(ctx); err != nil {
		return nil, err
	}

	return ctx, nil
}

func dispatchCreateOperation(
	request *elemental.Request,
	processorFinder processorFinder,
	factory elemental.IdentifiableFactory,
	authenticator Authenticator,
	authorizer Authorizer,
	pusher eventPusher,
) (*Context, error) {

	ctx := NewContext()
	ctx.ReadElementalRequest(request)

	if err := CheckAuthentication(authenticator, ctx); err != nil {
		return nil, err
	}

	if err := CheckAuthorization(authorizer, ctx); err != nil {
		return nil, err
	}

	proc, _ := processorFinder(request.Identity)

	if _, ok := proc.(CreateProcessor); !ok {
		return nil, elemental.NewError("Not implemented", "No handler for creating "+request.Identity.Name, "bahamut", http.StatusNotImplemented)
	}

	obj := factory(request.Identity.Name)
	if err := request.Decode(&obj); err != nil {
		return nil, err
	}

	if v, ok := obj.(elemental.Validatable); ok {
		if err := v.Validate(); err != nil {
			return nil, err
		}
	}

	ctx.InputData = obj

	if err := proc.(CreateProcessor).ProcessCreate(ctx); err != nil {
		return nil, err
	}

	if ctx.HasEvents() {
		pusher(ctx.Events()...)
	}

	if ctx.OutputData != nil {
		pusher(elemental.NewEvent(elemental.EventCreate, ctx.OutputData.(elemental.Identifiable)))
	}

	return ctx, nil
}

func dispatchUpdateOperation(
	request *elemental.Request,
	processorFinder processorFinder,
	factory elemental.IdentifiableFactory,
	authenticator Authenticator,
	authorizer Authorizer,
	pusher eventPusher,
) (*Context, error) {

	ctx := NewContext()
	ctx.ReadElementalRequest(request)

	if err := CheckAuthentication(authenticator, ctx); err != nil {
		return nil, err
	}

	if err := CheckAuthorization(authorizer, ctx); err != nil {
		return nil, err
	}

	proc, _ := processorFinder(request.Identity)

	if _, ok := proc.(UpdateProcessor); !ok {
		return nil, elemental.NewError("Not implemented", "No handler for updating "+request.Identity.Name, "bahamut", http.StatusNotImplemented)
	}

	obj := factory(request.Identity.Name)
	if err := request.Decode(&obj); err != nil {
		return nil, err
	}

	if v, ok := obj.(elemental.Validatable); ok {
		if err := v.Validate(); err != nil {
			return nil, err
		}
	}

	ctx.InputData = obj

	if err := proc.(UpdateProcessor).ProcessUpdate(ctx); err != nil {
		return nil, err
	}

	if ctx.HasEvents() {
		pusher(ctx.Events()...)
	}

	if ctx.OutputData != nil {
		pusher(elemental.NewEvent(elemental.EventUpdate, ctx.OutputData.(elemental.Identifiable)))
	}

	return ctx, nil
}

func dispatchDeleteOperation(
	request *elemental.Request,
	processorFinder processorFinder,
	factory elemental.IdentifiableFactory,
	authenticator Authenticator,
	authorizer Authorizer,
	pusher eventPusher,
) (*Context, error) {

	ctx := NewContext()
	ctx.ReadElementalRequest(request)

	if err := CheckAuthentication(authenticator, ctx); err != nil {
		return nil, err
	}

	if err := CheckAuthorization(authorizer, ctx); err != nil {
		return nil, err
	}

	proc, _ := processorFinder(request.Identity)

	if _, ok := proc.(RetrieveProcessor); !ok {
		return nil, elemental.NewError("Not implemented", "No handler for deleting "+request.Identity.Name, "bahamut", http.StatusNotImplemented)
	}

	if err := proc.(DeleteProcessor).ProcessDelete(ctx); err != nil {
		return nil, err
	}

	if ctx.HasEvents() {
		pusher(ctx.Events()...)
	}

	if ctx.OutputData != nil {
		pusher(elemental.NewEvent(elemental.EventDelete, ctx.OutputData.(elemental.Identifiable)))
	}

	return ctx, nil
}

func dispatchPatchOperation(
	request *elemental.Request,
	processorFinder processorFinder,
	factory elemental.IdentifiableFactory,
	authenticator Authenticator,
	authorizer Authorizer,
	pusher eventPusher,
) (*Context, error) {

	ctx := NewContext()
	ctx.ReadElementalRequest(request)

	if err := CheckAuthentication(authenticator, ctx); err != nil {
		return nil, err
	}

	if err := CheckAuthorization(authorizer, ctx); err != nil {
		return nil, err
	}

	proc, _ := processorFinder(request.Identity)

	if _, ok := proc.(PatchProcessor); !ok {
		return nil, elemental.NewError("Not implemented", "No handler for patching "+request.Identity.Name, "bahamut", http.StatusNotImplemented)
	}

	var assignation *elemental.Assignation
	if err := request.Decode(&assignation); err != nil {
		return nil, elemental.NewError("Bad Request", "The request cannot be processed", "bahamut", http.StatusBadRequest)
	}

	ctx.InputData = assignation

	if err := proc.(PatchProcessor).ProcessPatch(ctx); err != nil {
		return nil, err
	}

	if ctx.HasEvents() {
		pusher(ctx.Events()...)
	}

	if ctx.OutputData != nil {
		pusher(elemental.NewEvent(elemental.EventCreate, ctx.OutputData.(*elemental.Assignation)))
	}

	return ctx, nil
}

func dispatchInfoOperation(
	request *elemental.Request,
	processorFinder processorFinder,
	factory elemental.IdentifiableFactory,
	authenticator Authenticator,
	authorizer Authorizer,
) (*Context, error) {

	ctx := NewContext()
	ctx.ReadElementalRequest(request)

	if err := CheckAuthentication(authenticator, ctx); err != nil {
		return nil, err
	}

	if err := CheckAuthorization(authorizer, ctx); err != nil {
		return nil, err
	}

	proc, _ := processorFinder(request.Identity)

	if _, ok := proc.(InfoProcessor); !ok {
		return nil, elemental.NewError("Not implemented", "No handler for info "+request.Identity.Name, "bahamut", http.StatusNotImplemented)
	}

	if err := proc.(InfoProcessor).ProcessInfo(ctx); err != nil {
		return nil, err
	}

	return ctx, nil
}
