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
	ctx *Context,
	processorFinder processorFinderFunc,
	factory elemental.IdentifiableFactory,
	authenticators []RequestAuthenticator,
	authorizers []Authorizer,
	pusher eventPusherFunc,
	auditer Auditer,
) (err error) {

	if err = CheckAuthentication(authenticators, ctx); err != nil {
		audit(auditer, ctx, err)
		return
	}

	if err = CheckAuthorization(authorizers, ctx); err != nil {
		audit(auditer, ctx, err)
		return
	}

	proc, _ := processorFinder(ctx.Request.Identity)

	if _, ok := proc.(RetrieveManyProcessor); !ok {
		err = notImplementedErr(ctx.Request)
		audit(auditer, ctx, err)
		return
	}

	if err = proc.(RetrieveManyProcessor).ProcessRetrieveMany(ctx); err != nil {
		audit(auditer, ctx, err)
		return
	}

	if ctx.HasEvents() {
		pusher(ctx.Events()...)
	}

	audit(auditer, ctx, nil)

	return
}

func dispatchRetrieveOperation(
	ctx *Context,
	processorFinder processorFinderFunc,
	factory elemental.IdentifiableFactory,
	authenticators []RequestAuthenticator,
	authorizers []Authorizer,
	pusher eventPusherFunc,
	auditer Auditer,
) (err error) {

	if err = CheckAuthentication(authenticators, ctx); err != nil {
		audit(auditer, ctx, err)
		return
	}

	if err = CheckAuthorization(authorizers, ctx); err != nil {
		audit(auditer, ctx, err)
		return
	}

	proc, _ := processorFinder(ctx.Request.Identity)

	if _, ok := proc.(RetrieveProcessor); !ok {
		err = notImplementedErr(ctx.Request)
		audit(auditer, ctx, err)
		return
	}

	if err = proc.(RetrieveProcessor).ProcessRetrieve(ctx); err != nil {
		audit(auditer, ctx, err)
		return
	}

	if ctx.HasEvents() {
		pusher(ctx.Events()...)
	}

	audit(auditer, ctx, nil)

	return
}

func dispatchCreateOperation(
	ctx *Context,
	processorFinder processorFinderFunc,
	factory elemental.IdentifiableFactory,
	authenticators []RequestAuthenticator,
	authorizers []Authorizer,
	pusher eventPusherFunc,
	auditer Auditer,
	readOnlyMode bool,
	readOnlyExclusion []elemental.Identity,
) (err error) {

	if err = CheckAuthentication(authenticators, ctx); err != nil {
		audit(auditer, ctx, err)
		return
	}

	if err = CheckAuthorization(authorizers, ctx); err != nil {
		audit(auditer, ctx, err)
		return
	}

	if readOnlyMode {
		if err = makeReadOnlyError(ctx.Request.Identity, readOnlyExclusion); err != nil {
			return
		}
	}

	proc, _ := processorFinder(ctx.Request.Identity)

	if _, ok := proc.(CreateProcessor); !ok {
		err = notImplementedErr(ctx.Request)
		audit(auditer, ctx, err)
		return
	}

	obj := factory(ctx.Request.Identity.Name, ctx.Request.Version)
	if err = ctx.Request.Decode(&obj); err != nil {
		audit(auditer, ctx, err)
		return
	}

	if v, ok := obj.(elemental.Validatable); ok {
		if err = v.Validate(); err != nil {
			audit(auditer, ctx, err)
			return
		}
	}

	ctx.InputData = obj

	if err = proc.(CreateProcessor).ProcessCreate(ctx); err != nil {
		audit(auditer, ctx, err)
		return
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

	return
}

func dispatchUpdateOperation(
	ctx *Context,
	processorFinder processorFinderFunc,
	factory elemental.IdentifiableFactory,
	authenticators []RequestAuthenticator,
	authorizers []Authorizer,
	pusher eventPusherFunc,
	auditer Auditer,
	readOnlyMode bool,
	readOnlyExclusion []elemental.Identity,
) (err error) {

	if err = CheckAuthentication(authenticators, ctx); err != nil {
		audit(auditer, ctx, err)
		return
	}

	if err = CheckAuthorization(authorizers, ctx); err != nil {
		audit(auditer, ctx, err)
		return
	}

	if readOnlyMode {
		if err = makeReadOnlyError(ctx.Request.Identity, readOnlyExclusion); err != nil {
			return
		}
	}

	proc, _ := processorFinder(ctx.Request.Identity)

	if _, ok := proc.(UpdateProcessor); !ok {
		err = notImplementedErr(ctx.Request)
		audit(auditer, ctx, err)
		return
	}

	obj := factory(ctx.Request.Identity.Name, ctx.Request.Version)
	if err = ctx.Request.Decode(&obj); err != nil {
		audit(auditer, ctx, err)
		return
	}

	if v, ok := obj.(elemental.Validatable); ok {
		if err = v.Validate(); err != nil {
			audit(auditer, ctx, err)
			return
		}
	}

	ctx.InputData = obj

	if err = proc.(UpdateProcessor).ProcessUpdate(ctx); err != nil {
		audit(auditer, ctx, err)
		return
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

	return
}

func dispatchDeleteOperation(
	ctx *Context,
	processorFinder processorFinderFunc,
	factory elemental.IdentifiableFactory,
	authenticators []RequestAuthenticator,
	authorizers []Authorizer,
	pusher eventPusherFunc,
	auditer Auditer,
	readOnlyMode bool,
	readOnlyExclusion []elemental.Identity,
) (err error) {

	if err = CheckAuthentication(authenticators, ctx); err != nil {
		audit(auditer, ctx, err)
		return
	}

	if err = CheckAuthorization(authorizers, ctx); err != nil {
		audit(auditer, ctx, err)
		return
	}

	if readOnlyMode {
		if err = makeReadOnlyError(ctx.Request.Identity, readOnlyExclusion); err != nil {
			return
		}
	}

	proc, _ := processorFinder(ctx.Request.Identity)

	if _, ok := proc.(DeleteProcessor); !ok {
		err = notImplementedErr(ctx.Request)
		audit(auditer, ctx, err)
		return
	}

	if err = proc.(DeleteProcessor).ProcessDelete(ctx); err != nil {
		audit(auditer, ctx, err)
		return
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

	return
}

func dispatchPatchOperation(
	ctx *Context,
	processorFinder processorFinderFunc,
	factory elemental.IdentifiableFactory,
	authenticators []RequestAuthenticator,
	authorizers []Authorizer,
	pusher eventPusherFunc,
	auditer Auditer,
	readOnlyMode bool,
	readOnlyExclusion []elemental.Identity,
) (err error) {

	if err = CheckAuthentication(authenticators, ctx); err != nil {
		audit(auditer, ctx, err)
		return
	}

	if err = CheckAuthorization(authorizers, ctx); err != nil {
		audit(auditer, ctx, err)
		return
	}

	if readOnlyMode {
		if err = makeReadOnlyError(ctx.Request.Identity, readOnlyExclusion); err != nil {
			return
		}
	}

	proc, _ := processorFinder(ctx.Request.Identity)

	if _, ok := proc.(PatchProcessor); !ok {
		err = notImplementedErr(ctx.Request)
		audit(auditer, ctx, err)
		return
	}

	var assignation *elemental.Assignation
	if err = ctx.Request.Decode(&assignation); err != nil {
		audit(auditer, ctx, err)
		return
	}

	ctx.InputData = assignation

	if err = proc.(PatchProcessor).ProcessPatch(ctx); err != nil {
		audit(auditer, ctx, err)
		return
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

	return
}

func dispatchInfoOperation(
	ctx *Context,
	processorFinder processorFinderFunc,
	factory elemental.IdentifiableFactory,
	authenticators []RequestAuthenticator,
	authorizers []Authorizer,
	auditer Auditer,
) (err error) {

	if err = CheckAuthentication(authenticators, ctx); err != nil {
		audit(auditer, ctx, err)
		return
	}

	if err = CheckAuthorization(authorizers, ctx); err != nil {
		audit(auditer, ctx, err)
		return
	}

	proc, _ := processorFinder(ctx.Request.Identity)

	if _, ok := proc.(InfoProcessor); !ok {
		err = notImplementedErr(ctx.Request)
		audit(auditer, ctx, err)
		return
	}

	if err = proc.(InfoProcessor).ProcessInfo(ctx); err != nil {
		audit(auditer, ctx, err)
		return
	}

	audit(auditer, ctx, nil)

	return
}

func makeReadOnlyError(identity elemental.Identity, readOnlyExclusion []elemental.Identity) error {

	for _, i := range readOnlyExclusion {
		if i.IsEqual(identity) {
			return nil
		}
	}

	return elemental.NewError("Locked", "This api is currently locked. Please try again later", "bahamut", http.StatusLocked)
}
