package bahamut

import (
	"fmt"
	"net/http"

	"github.com/aporeto-inc/elemental"
)

func audit(auditer Auditer, ctx *Context, err error) {
	if auditer == nil {
		return
	}
	auditer.Audit(ctx, err)
}

func notImplementedErr(request *elemental.Request) error {
	return elemental.NewError(
		"Not implemented",
		fmt.Sprintf("No handler for operation %s on %s", request.Operation, request.Identity.Name),
		"bahamut",
		http.StatusNotImplemented,
	)
}

func dispatchRetrieveManyOperation(
	request *elemental.Request,
	processorFinder processorFinder,
	factory elemental.IdentifiableFactory,
	authenticator Authenticator,
	authorizer Authorizer,
	auditer Auditer,
) (*Context, error) {

	ctx := NewContext()
	ctx.ReadElementalRequest(request)

	if err := CheckAuthentication(authenticator, ctx); err != nil {
		audit(auditer, ctx, err)
		return nil, err
	}

	if err := CheckAuthorization(authorizer, ctx); err != nil {
		audit(auditer, ctx, err)
		return nil, err
	}

	proc, _ := processorFinder(request.Identity)

	if _, ok := proc.(RetrieveManyProcessor); !ok {
		err := notImplementedErr(request)
		audit(auditer, ctx, err)
		return nil, err
	}

	if err := proc.(RetrieveManyProcessor).ProcessRetrieveMany(ctx); err != nil {
		audit(auditer, ctx, err)
		return nil, err
	}

	audit(auditer, ctx, nil)

	return ctx, nil
}

func dispatchRetrieveOperation(
	request *elemental.Request,
	processorFinder processorFinder,
	factory elemental.IdentifiableFactory,
	authenticator Authenticator,
	authorizer Authorizer,
	auditer Auditer,
) (*Context, error) {

	ctx := NewContext()
	ctx.ReadElementalRequest(request)

	if err := CheckAuthentication(authenticator, ctx); err != nil {
		audit(auditer, ctx, err)
		return nil, err
	}

	if err := CheckAuthorization(authorizer, ctx); err != nil {
		audit(auditer, ctx, err)
		return nil, err
	}

	proc, _ := processorFinder(request.Identity)

	if _, ok := proc.(RetrieveProcessor); !ok {
		err := notImplementedErr(request)
		audit(auditer, ctx, err)
		return nil, err
	}

	if err := proc.(RetrieveProcessor).ProcessRetrieve(ctx); err != nil {
		audit(auditer, ctx, err)
		return nil, err
	}

	audit(auditer, ctx, nil)

	return ctx, nil
}

func dispatchCreateOperation(
	request *elemental.Request,
	processorFinder processorFinder,
	factory elemental.IdentifiableFactory,
	authenticator Authenticator,
	authorizer Authorizer,
	pusher eventPusher,
	auditer Auditer,
) (*Context, error) {

	ctx := NewContext()
	ctx.ReadElementalRequest(request)

	if err := CheckAuthentication(authenticator, ctx); err != nil {
		audit(auditer, ctx, err)
		return nil, err
	}

	if err := CheckAuthorization(authorizer, ctx); err != nil {
		audit(auditer, ctx, err)
		return nil, err
	}

	proc, _ := processorFinder(request.Identity)

	if _, ok := proc.(CreateProcessor); !ok {
		err := notImplementedErr(request)
		audit(auditer, ctx, err)
		return nil, err
	}

	obj := factory(request.Identity.Name)
	if err := request.Decode(&obj); err != nil {
		audit(auditer, ctx, err)
		return nil, err
	}

	if v, ok := obj.(elemental.Validatable); ok {
		if err := v.Validate(); err != nil {
			audit(auditer, ctx, err)
			return nil, err
		}
	}

	ctx.InputData = obj

	if err := proc.(CreateProcessor).ProcessCreate(ctx); err != nil {
		audit(auditer, ctx, err)
		return nil, err
	}

	if ctx.HasEvents() {
		pusher(ctx.Events()...)
	}

	if ctx.OutputData != nil {
		pusher(elemental.NewEvent(elemental.EventCreate, ctx.OutputData.(elemental.Identifiable)))
	}

	audit(auditer, ctx, nil)

	return ctx, nil
}

func dispatchUpdateOperation(
	request *elemental.Request,
	processorFinder processorFinder,
	factory elemental.IdentifiableFactory,
	authenticator Authenticator,
	authorizer Authorizer,
	pusher eventPusher,
	auditer Auditer,
) (*Context, error) {

	ctx := NewContext()
	ctx.ReadElementalRequest(request)

	if err := CheckAuthentication(authenticator, ctx); err != nil {
		audit(auditer, ctx, err)
		return nil, err
	}

	if err := CheckAuthorization(authorizer, ctx); err != nil {
		audit(auditer, ctx, err)
		return nil, err
	}

	proc, _ := processorFinder(request.Identity)

	if _, ok := proc.(UpdateProcessor); !ok {
		err := notImplementedErr(request)
		audit(auditer, ctx, err)
		return nil, err
	}

	obj := factory(request.Identity.Name)
	if err := request.Decode(&obj); err != nil {
		audit(auditer, ctx, err)
		return nil, err
	}

	if v, ok := obj.(elemental.Validatable); ok {
		if err := v.Validate(); err != nil {
			audit(auditer, ctx, err)
			return nil, err
		}
	}

	ctx.InputData = obj

	if err := proc.(UpdateProcessor).ProcessUpdate(ctx); err != nil {
		audit(auditer, ctx, err)
		return nil, err
	}

	if ctx.HasEvents() {
		pusher(ctx.Events()...)
	}

	if ctx.OutputData != nil {
		pusher(elemental.NewEvent(elemental.EventUpdate, ctx.OutputData.(elemental.Identifiable)))
	}

	audit(auditer, ctx, nil)

	return ctx, nil
}

func dispatchDeleteOperation(
	request *elemental.Request,
	processorFinder processorFinder,
	factory elemental.IdentifiableFactory,
	authenticator Authenticator,
	authorizer Authorizer,
	pusher eventPusher,
	auditer Auditer,
) (*Context, error) {

	ctx := NewContext()
	ctx.ReadElementalRequest(request)

	if err := CheckAuthentication(authenticator, ctx); err != nil {
		audit(auditer, ctx, err)
		return nil, err
	}

	if err := CheckAuthorization(authorizer, ctx); err != nil {
		audit(auditer, ctx, err)
		return nil, err
	}

	proc, _ := processorFinder(request.Identity)

	if _, ok := proc.(DeleteProcessor); !ok {
		err := notImplementedErr(request)
		audit(auditer, ctx, err)
		return nil, err
	}

	if err := proc.(DeleteProcessor).ProcessDelete(ctx); err != nil {
		audit(auditer, ctx, err)
		return nil, err
	}

	if ctx.HasEvents() {
		pusher(ctx.Events()...)
	}

	if ctx.OutputData != nil {
		pusher(elemental.NewEvent(elemental.EventDelete, ctx.OutputData.(elemental.Identifiable)))
	}

	audit(auditer, ctx, nil)

	return ctx, nil
}

func dispatchPatchOperation(
	request *elemental.Request,
	processorFinder processorFinder,
	factory elemental.IdentifiableFactory,
	authenticator Authenticator,
	authorizer Authorizer,
	pusher eventPusher,
	auditer Auditer,
) (*Context, error) {

	ctx := NewContext()
	ctx.ReadElementalRequest(request)

	if err := CheckAuthentication(authenticator, ctx); err != nil {
		audit(auditer, ctx, err)
		return nil, err
	}

	if err := CheckAuthorization(authorizer, ctx); err != nil {
		audit(auditer, ctx, err)
		return nil, err
	}

	proc, _ := processorFinder(request.Identity)

	if _, ok := proc.(PatchProcessor); !ok {
		err := notImplementedErr(request)
		audit(auditer, ctx, err)
		return nil, err
	}

	var assignation *elemental.Assignation
	if err := request.Decode(&assignation); err != nil {
		audit(auditer, ctx, err)
		return nil, err
	}

	ctx.InputData = assignation

	if err := proc.(PatchProcessor).ProcessPatch(ctx); err != nil {
		audit(auditer, ctx, err)
		return nil, err
	}

	if ctx.HasEvents() {
		pusher(ctx.Events()...)
	}

	if ctx.OutputData != nil {
		pusher(elemental.NewEvent(elemental.EventCreate, ctx.OutputData.(*elemental.Assignation)))
	}

	audit(auditer, ctx, nil)

	return ctx, nil
}

func dispatchInfoOperation(
	request *elemental.Request,
	processorFinder processorFinder,
	factory elemental.IdentifiableFactory,
	authenticator Authenticator,
	authorizer Authorizer,
	auditer Auditer,
) (*Context, error) {

	ctx := NewContext()
	ctx.ReadElementalRequest(request)

	if err := CheckAuthentication(authenticator, ctx); err != nil {
		audit(auditer, ctx, err)
		return nil, err
	}

	if err := CheckAuthorization(authorizer, ctx); err != nil {
		audit(auditer, ctx, err)
		return nil, err
	}

	proc, _ := processorFinder(request.Identity)

	if _, ok := proc.(InfoProcessor); !ok {
		err := notImplementedErr(request)
		audit(auditer, ctx, err)
		return nil, err
	}

	if err := proc.(InfoProcessor).ProcessInfo(ctx); err != nil {
		audit(auditer, ctx, err)
		return nil, err
	}

	audit(auditer, ctx, nil)

	return ctx, nil
}
