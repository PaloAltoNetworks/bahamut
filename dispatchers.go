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

	// TODO: this is not very optimized as we do processError twice in the flow.
	// one here, and one after the dispatcher returns.
	// We need to refactor the code so we can do this only once.
	var e error
	if err != nil {
		e = processError(err, ctx.Request)
	}

	auditer.Audit(ctx, e)
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
	processorFinder processorFinderFunc,
	factory elemental.IdentifiableFactory,
	authenticators []RequestAuthenticator,
	authorizers []Authorizer,
	pusher eventPusherFunc,
	auditer Auditer,
) (ctx *Context, err error) {

	ctx = NewContextWithRequest(request)

	if err = CheckAuthentication(authenticators, ctx); err != nil {
		audit(auditer, ctx, err)
		return nil, err
	}

	if err = CheckAuthorization(authorizers, ctx); err != nil {
		audit(auditer, ctx, err)
		return nil, err
	}

	proc, _ := processorFinder(request.Identity)

	if _, ok := proc.(RetrieveManyProcessor); !ok {
		err = notImplementedErr(request)
		audit(auditer, ctx, err)
		return nil, err
	}

	if err = proc.(RetrieveManyProcessor).ProcessRetrieveMany(ctx); err != nil {
		audit(auditer, ctx, err)
		return nil, err
	}

	if ctx.HasEvents() {
		pusher(ctx.Events()...)
	}

	audit(auditer, ctx, nil)

	return ctx, nil
}

func dispatchRetrieveOperation(
	request *elemental.Request,
	processorFinder processorFinderFunc,
	factory elemental.IdentifiableFactory,
	authenticators []RequestAuthenticator,
	authorizers []Authorizer,
	pusher eventPusherFunc,
	auditer Auditer,
) (ctx *Context, err error) {

	ctx = NewContextWithRequest(request)

	if err = CheckAuthentication(authenticators, ctx); err != nil {
		audit(auditer, ctx, err)
		return nil, err
	}

	if err = CheckAuthorization(authorizers, ctx); err != nil {
		audit(auditer, ctx, err)
		return nil, err
	}

	proc, _ := processorFinder(request.Identity)

	if _, ok := proc.(RetrieveProcessor); !ok {
		err = notImplementedErr(request)
		audit(auditer, ctx, err)
		return nil, err
	}

	if err = proc.(RetrieveProcessor).ProcessRetrieve(ctx); err != nil {
		audit(auditer, ctx, err)
		return nil, err
	}

	if ctx.HasEvents() {
		pusher(ctx.Events()...)
	}

	audit(auditer, ctx, nil)

	return ctx, nil
}

func dispatchCreateOperation(
	request *elemental.Request,
	processorFinder processorFinderFunc,
	factory elemental.IdentifiableFactory,
	authenticators []RequestAuthenticator,
	authorizers []Authorizer,
	pusher eventPusherFunc,
	auditer Auditer,
	readOnlyMode bool,
	readOnlyExclusion []elemental.Identity,
) (ctx *Context, err error) {

	ctx = NewContextWithRequest(request)

	if err = CheckAuthentication(authenticators, ctx); err != nil {
		audit(auditer, ctx, err)
		return nil, err
	}

	if err = CheckAuthorization(authorizers, ctx); err != nil {
		audit(auditer, ctx, err)
		return nil, err
	}

	if readOnlyMode {
		if err = makeReadOnlyError(request.Identity, readOnlyExclusion); err != nil {
			return nil, err
		}
	}

	proc, _ := processorFinder(request.Identity)

	if _, ok := proc.(CreateProcessor); !ok {
		err = notImplementedErr(request)
		audit(auditer, ctx, err)
		return nil, err
	}

	obj := factory(request.Identity.Name, ctx.Request.Version)
	if err = request.Decode(&obj); err != nil {
		audit(auditer, ctx, err)
		return nil, err
	}

	if v, ok := obj.(elemental.Validatable); ok {
		if err = v.Validate(); err != nil {
			audit(auditer, ctx, err)
			return nil, err
		}
	}

	ctx.InputData = obj

	if err = proc.(CreateProcessor).ProcessCreate(ctx); err != nil {
		audit(auditer, ctx, err)
		return nil, err
	}

	if ctx.HasEvents() {
		pusher(ctx.Events()...)
	}

	if ctx.OutputData != nil {
		evt := elemental.NewEvent(elemental.EventCreate, ctx.OutputData.(elemental.Identifiable))
		evt.UserInfo = ctx.Metadata
		pusher(evt)
	}

	audit(auditer, ctx, nil)

	return ctx, nil
}

func dispatchUpdateOperation(
	request *elemental.Request,
	processorFinder processorFinderFunc,
	factory elemental.IdentifiableFactory,
	authenticators []RequestAuthenticator,
	authorizers []Authorizer,
	pusher eventPusherFunc,
	auditer Auditer,
	readOnlyMode bool,
	readOnlyExclusion []elemental.Identity,
) (ctx *Context, err error) {

	ctx = NewContextWithRequest(request)

	if err = CheckAuthentication(authenticators, ctx); err != nil {
		audit(auditer, ctx, err)
		return nil, err
	}

	if err = CheckAuthorization(authorizers, ctx); err != nil {
		audit(auditer, ctx, err)
		return nil, err
	}

	if readOnlyMode {
		if err = makeReadOnlyError(request.Identity, readOnlyExclusion); err != nil {
			return nil, err
		}
	}

	proc, _ := processorFinder(request.Identity)

	if _, ok := proc.(UpdateProcessor); !ok {
		err = notImplementedErr(request)
		audit(auditer, ctx, err)
		return nil, err
	}

	obj := factory(request.Identity.Name, ctx.Request.Version)
	if err = request.Decode(&obj); err != nil {
		audit(auditer, ctx, err)
		return nil, err
	}

	if v, ok := obj.(elemental.Validatable); ok {
		if err = v.Validate(); err != nil {
			audit(auditer, ctx, err)
			return nil, err
		}
	}

	ctx.InputData = obj

	if err = proc.(UpdateProcessor).ProcessUpdate(ctx); err != nil {
		audit(auditer, ctx, err)
		return nil, err
	}

	if ctx.HasEvents() {
		pusher(ctx.Events()...)
	}

	if ctx.OutputData != nil {
		evt := elemental.NewEvent(elemental.EventUpdate, ctx.OutputData.(elemental.Identifiable))
		evt.UserInfo = ctx.Metadata
		pusher(evt)
	}

	audit(auditer, ctx, nil)

	return ctx, nil
}

func dispatchDeleteOperation(
	request *elemental.Request,
	processorFinder processorFinderFunc,
	factory elemental.IdentifiableFactory,
	authenticators []RequestAuthenticator,
	authorizers []Authorizer,
	pusher eventPusherFunc,
	auditer Auditer,
	readOnlyMode bool,
	readOnlyExclusion []elemental.Identity,
) (ctx *Context, err error) {

	ctx = NewContextWithRequest(request)

	if err = CheckAuthentication(authenticators, ctx); err != nil {
		audit(auditer, ctx, err)
		return nil, err
	}

	if err = CheckAuthorization(authorizers, ctx); err != nil {
		audit(auditer, ctx, err)
		return nil, err
	}

	if readOnlyMode {
		if err = makeReadOnlyError(request.Identity, readOnlyExclusion); err != nil {
			return nil, err
		}
	}

	proc, _ := processorFinder(request.Identity)

	if _, ok := proc.(DeleteProcessor); !ok {
		err = notImplementedErr(request)
		audit(auditer, ctx, err)
		return nil, err
	}

	if err = proc.(DeleteProcessor).ProcessDelete(ctx); err != nil {
		audit(auditer, ctx, err)
		return nil, err
	}

	if ctx.HasEvents() {
		pusher(ctx.Events()...)
	}

	if ctx.OutputData != nil {
		evt := elemental.NewEvent(elemental.EventDelete, ctx.OutputData.(elemental.Identifiable))
		evt.UserInfo = ctx.Metadata
		pusher(evt)
	}

	audit(auditer, ctx, nil)

	return ctx, nil
}

func dispatchPatchOperation(
	request *elemental.Request,
	processorFinder processorFinderFunc,
	factory elemental.IdentifiableFactory,
	authenticators []RequestAuthenticator,
	authorizers []Authorizer,
	pusher eventPusherFunc,
	auditer Auditer,
	readOnlyMode bool,
	readOnlyExclusion []elemental.Identity,
) (ctx *Context, err error) {

	ctx = NewContextWithRequest(request)

	if err = CheckAuthentication(authenticators, ctx); err != nil {
		audit(auditer, ctx, err)
		return nil, err
	}

	if err = CheckAuthorization(authorizers, ctx); err != nil {
		audit(auditer, ctx, err)
		return nil, err
	}

	if readOnlyMode {
		if err = makeReadOnlyError(request.Identity, readOnlyExclusion); err != nil {
			return nil, err
		}
	}

	proc, _ := processorFinder(request.Identity)

	if _, ok := proc.(PatchProcessor); !ok {
		err = notImplementedErr(request)
		audit(auditer, ctx, err)
		return nil, err
	}

	var assignation *elemental.Assignation
	if err = request.Decode(&assignation); err != nil {
		audit(auditer, ctx, err)
		return nil, err
	}

	ctx.InputData = assignation

	if err = proc.(PatchProcessor).ProcessPatch(ctx); err != nil {
		audit(auditer, ctx, err)
		return nil, err
	}

	if ctx.HasEvents() {
		pusher(ctx.Events()...)
	}

	if ctx.OutputData != nil {
		evt := elemental.NewEvent(elemental.EventCreate, ctx.OutputData.(*elemental.Assignation))
		evt.UserInfo = ctx.Metadata
		pusher(evt)
	}

	audit(auditer, ctx, nil)

	return ctx, nil
}

func dispatchInfoOperation(
	request *elemental.Request,
	processorFinder processorFinderFunc,
	factory elemental.IdentifiableFactory,
	authenticators []RequestAuthenticator,
	authorizers []Authorizer,
	auditer Auditer,
) (ctx *Context, err error) {

	ctx = NewContextWithRequest(request)

	if err = CheckAuthentication(authenticators, ctx); err != nil {
		audit(auditer, ctx, err)
		return nil, err
	}

	if err = CheckAuthorization(authorizers, ctx); err != nil {
		audit(auditer, ctx, err)
		return nil, err
	}

	proc, _ := processorFinder(request.Identity)

	if _, ok := proc.(InfoProcessor); !ok {
		err = notImplementedErr(request)
		audit(auditer, ctx, err)
		return nil, err
	}

	if err = proc.(InfoProcessor).ProcessInfo(ctx); err != nil {
		audit(auditer, ctx, err)
		return nil, err
	}

	audit(auditer, ctx, nil)

	return ctx, nil
}

func makeReadOnlyError(identity elemental.Identity, readOnlyExclusion []elemental.Identity) error {

	for _, i := range readOnlyExclusion {
		if i.IsEqual(identity) {
			return nil
		}
	}

	return elemental.NewError("Locked", "This api is currently locked. Please try again later", "bahamut", http.StatusLocked)
}
