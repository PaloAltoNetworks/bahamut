package bahamut

import (
	"fmt"
	"net/http"

	"go.aporeto.io/elemental"
)

func audit(auditer Auditer, ctx *bcontext, err error) {

	if auditer == nil {
		return
	}

	go auditer.Audit(ctx, err)
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
	ctx *bcontext,
	processorFinder processorFinderFunc,
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

	proc, _ := processorFinder(ctx.request.Identity)

	if _, ok := proc.(RetrieveManyProcessor); !ok {
		err = notImplementedErr(ctx.request)
		audit(auditer, ctx, err)
		return
	}

	if err = proc.(RetrieveManyProcessor).ProcessRetrieveMany(ctx); err != nil {
		audit(auditer, ctx, err)
		return
	}

	if len(ctx.events) > 0 {
		pusher(ctx.events...)
	}

	audit(auditer, ctx, nil)

	return
}

func dispatchRetrieveOperation(
	ctx *bcontext,
	processorFinder processorFinderFunc,
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

	proc, _ := processorFinder(ctx.request.Identity)

	if _, ok := proc.(RetrieveProcessor); !ok {
		err = notImplementedErr(ctx.request)
		audit(auditer, ctx, err)
		return
	}

	if err = proc.(RetrieveProcessor).ProcessRetrieve(ctx); err != nil {
		audit(auditer, ctx, err)
		return
	}

	if len(ctx.events) > 0 {
		pusher(ctx.events...)
	}

	audit(auditer, ctx, nil)

	return
}

func dispatchCreateOperation(
	ctx *bcontext,
	processorFinder processorFinderFunc,
	modelManager elemental.ModelManager,
	unmarshaller CustomUmarshaller,
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
		if err = makeReadOnlyError(ctx.request.Identity, readOnlyExclusion); err != nil {
			return
		}
	}

	proc, _ := processorFinder(ctx.request.Identity)

	if _, ok := proc.(CreateProcessor); !ok {
		err = notImplementedErr(ctx.request)
		audit(auditer, ctx, err)
		return
	}

	var obj elemental.Identifiable
	if unmarshaller != nil {
		if obj, err = unmarshaller(ctx.request); err != nil {
			audit(auditer, ctx, err)
			return
		}
	} else {
		obj = modelManager.Identifiable(ctx.request.Identity)
		if err = ctx.request.Decode(&obj); err != nil {
			audit(auditer, ctx, err)
			return
		}
	}

	if v, ok := obj.(elemental.Validatable); ok {
		if err = v.Validate(); err != nil {
			audit(auditer, ctx, err)
			return
		}
	}

	ctx.inputData = obj

	if err = proc.(CreateProcessor).ProcessCreate(ctx); err != nil {
		audit(auditer, ctx, err)
		return
	}

	if len(ctx.events) > 0 {
		pusher(ctx.events...)
	}

	if ctx.outputData != nil {
		evt := elemental.NewEvent(elemental.EventCreate, ctx.outputData.(elemental.Identifiable))
		pusher(evt)
	}

	audit(auditer, ctx, nil)

	return
}

func dispatchUpdateOperation(
	ctx *bcontext,
	processorFinder processorFinderFunc,
	modelManager elemental.ModelManager,
	unmarshaller CustomUmarshaller,
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
		if err = makeReadOnlyError(ctx.request.Identity, readOnlyExclusion); err != nil {
			return
		}
	}

	proc, _ := processorFinder(ctx.request.Identity)

	if _, ok := proc.(UpdateProcessor); !ok {
		err = notImplementedErr(ctx.request)
		audit(auditer, ctx, err)
		return
	}
	var obj elemental.Identifiable

	if unmarshaller != nil {
		if obj, err = unmarshaller(ctx.request); err != nil {
			audit(auditer, ctx, err)
			return
		}
	} else {
		obj = modelManager.Identifiable(ctx.request.Identity)
		if err = ctx.request.Decode(&obj); err != nil {
			audit(auditer, ctx, err)
			return
		}
	}

	if v, ok := obj.(elemental.Validatable); ok {
		if err = v.Validate(); err != nil {
			audit(auditer, ctx, err)
			return
		}
	}

	ctx.inputData = obj

	if err = proc.(UpdateProcessor).ProcessUpdate(ctx); err != nil {
		audit(auditer, ctx, err)
		return
	}

	if len(ctx.events) > 0 {
		pusher(ctx.events...)
	}

	if ctx.outputData != nil {
		evt := elemental.NewEvent(elemental.EventUpdate, ctx.outputData.(elemental.Identifiable))
		pusher(evt)
	}

	audit(auditer, ctx, nil)

	return
}

func dispatchDeleteOperation(
	ctx *bcontext,
	processorFinder processorFinderFunc,
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
		if err = makeReadOnlyError(ctx.request.Identity, readOnlyExclusion); err != nil {
			return
		}
	}

	proc, _ := processorFinder(ctx.request.Identity)

	if _, ok := proc.(DeleteProcessor); !ok {
		err = notImplementedErr(ctx.request)
		audit(auditer, ctx, err)
		return
	}

	if err = proc.(DeleteProcessor).ProcessDelete(ctx); err != nil {
		audit(auditer, ctx, err)
		return
	}

	if len(ctx.events) > 0 {
		pusher(ctx.events...)
	}

	if ctx.outputData != nil {
		evt := elemental.NewEvent(elemental.EventDelete, ctx.outputData.(elemental.Identifiable))
		pusher(evt)
	}

	audit(auditer, ctx, nil)

	return
}

func dispatchPatchOperation(
	ctx *bcontext,
	processorFinder processorFinderFunc,
	modelManager elemental.ModelManager,
	unmarshaller CustomUmarshaller,
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
		if err = makeReadOnlyError(ctx.request.Identity, readOnlyExclusion); err != nil {
			return
		}
	}

	proc, _ := processorFinder(ctx.request.Identity)

	if _, ok := proc.(PatchProcessor); !ok {
		err = notImplementedErr(ctx.request)
		audit(auditer, ctx, err)
		return
	}
	var sparse elemental.Identifiable

	if unmarshaller != nil {
		if sparse, err = unmarshaller(ctx.request); err != nil {
			audit(auditer, ctx, err)
			return
		}
	} else {
		sparse = modelManager.SparseIdentifiable(ctx.request.Identity)
		if err = ctx.request.Decode(&sparse); err != nil {
			audit(auditer, ctx, err)
			return
		}
	}

	ctx.inputData = sparse

	if err = proc.(PatchProcessor).ProcessPatch(ctx); err != nil {
		audit(auditer, ctx, err)
		return
	}

	if len(ctx.events) > 0 {
		pusher(ctx.events...)
	}

	if ctx.outputData != nil {
		evt := elemental.NewEvent(elemental.EventUpdate, ctx.outputData.(elemental.Identifiable))
		pusher(evt)
	}

	audit(auditer, ctx, nil)

	return
}

func dispatchInfoOperation(
	ctx *bcontext,
	processorFinder processorFinderFunc,
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

	proc, _ := processorFinder(ctx.request.Identity)

	if _, ok := proc.(InfoProcessor); !ok {
		err = notImplementedErr(ctx.request)
		audit(auditer, ctx, err)
		return
	}

	if err = proc.(InfoProcessor).ProcessInfo(ctx); err != nil {
		audit(auditer, ctx, err)
		return
	}

	if len(ctx.events) > 0 {
		pusher(ctx.events...)
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
