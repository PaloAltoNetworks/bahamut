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
	"fmt"
	"net/http"

	"go.aporeto.io/elemental"
)

func audit(auditer Auditer, ctx *bcontext, err error) {

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
	ctx *bcontext,
	processorFinder processorFinderFunc,
	authenticators []RequestAuthenticator,
	authorizers []Authorizer,
	pusher eventPusherFunc,
	auditer Auditer,
) (err error) {

	if err = CheckAuthentication(authenticators, ctx); err != nil {
		audit(auditer, ctx, err)
		return err
	}

	if err = CheckAuthorization(authorizers, ctx); err != nil {
		audit(auditer, ctx, err)
		return err
	}

	proc, _ := processorFinder(ctx.request.Identity)

	if _, ok := proc.(RetrieveManyProcessor); !ok {
		err = notImplementedErr(ctx.request)
		audit(auditer, ctx, err)
		return err
	}

	if err = proc.(RetrieveManyProcessor).ProcessRetrieveMany(ctx); err != nil {
		audit(auditer, ctx, err)
		return err
	}

	if len(ctx.events) > 0 {
		pusher(ctx.events...)
	}

	audit(auditer, ctx, nil)

	return err
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
		return err
	}

	if err = CheckAuthorization(authorizers, ctx); err != nil {
		audit(auditer, ctx, err)
		return err
	}

	proc, _ := processorFinder(ctx.request.Identity)

	if _, ok := proc.(RetrieveProcessor); !ok {
		err = notImplementedErr(ctx.request)
		audit(auditer, ctx, err)
		return err
	}

	if err = proc.(RetrieveProcessor).ProcessRetrieve(ctx); err != nil {
		audit(auditer, ctx, err)
		return err
	}

	if len(ctx.events) > 0 {
		pusher(ctx.events...)
	}

	audit(auditer, ctx, nil)

	return err
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
		return err
	}

	if err = CheckAuthorization(authorizers, ctx); err != nil {
		audit(auditer, ctx, err)
		return err
	}

	if readOnlyMode {
		if err = makeReadOnlyError(ctx.request.Identity, readOnlyExclusion); err != nil {
			return err
		}
	}

	proc, _ := processorFinder(ctx.request.Identity)

	if _, ok := proc.(CreateProcessor); !ok {
		err = notImplementedErr(ctx.request)
		audit(auditer, ctx, err)
		return err
	}

	var obj elemental.Identifiable
	if unmarshaller != nil {
		if obj, err = unmarshaller(ctx.request); err != nil {
			audit(auditer, ctx, err)
			return elemental.NewError("Bad Request", err.Error(), "bahamut", http.StatusBadRequest)
		}
	} else {
		obj = modelManager.Identifiable(ctx.request.Identity)
		if len(ctx.Request().Data) > 0 {
			if err := ctx.Request().Decode(obj); err != nil {
				audit(auditer, ctx, err)
				return elemental.NewError("Bad Request", err.Error(), "bahamut", http.StatusBadRequest)
			}
		}
	}

	if v, ok := obj.(elemental.Validatable); ok {
		if err = v.Validate(); err != nil {
			audit(auditer, ctx, err)
			return err
		}
	}

	ctx.inputData = obj

	if err = proc.(CreateProcessor).ProcessCreate(ctx); err != nil {
		audit(auditer, ctx, err)
		return err
	}

	if len(ctx.events) > 0 {
		pusher(ctx.events...)
	}

	if o, ok := ctx.outputData.(elemental.Identifiable); ok && !ctx.disableOutputDataPush {
		pusher(elemental.NewEvent(elemental.EventCreate, o))
	}

	audit(auditer, ctx, nil)

	return err
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
		return err
	}

	if err = CheckAuthorization(authorizers, ctx); err != nil {
		audit(auditer, ctx, err)
		return err
	}

	if readOnlyMode {
		if err = makeReadOnlyError(ctx.request.Identity, readOnlyExclusion); err != nil {
			return err
		}
	}

	proc, _ := processorFinder(ctx.request.Identity)

	if _, ok := proc.(UpdateProcessor); !ok {
		err = notImplementedErr(ctx.request)
		audit(auditer, ctx, err)
		return err
	}
	var obj elemental.Identifiable

	if unmarshaller != nil {
		if obj, err = unmarshaller(ctx.request); err != nil {
			audit(auditer, ctx, err)
			return elemental.NewError("Bad Request", err.Error(), "bahamut", http.StatusBadRequest)
		}
	} else {
		obj = modelManager.Identifiable(ctx.request.Identity)
		if len(ctx.Request().Data) > 0 {
			if err := ctx.Request().Decode(obj); err != nil {
				audit(auditer, ctx, err)
				return elemental.NewError("Bad Request", err.Error(), "bahamut", http.StatusBadRequest)
			}
		}
	}

	if v, ok := obj.(elemental.Validatable); ok {
		if err = v.Validate(); err != nil {
			audit(auditer, ctx, err)
			return err
		}
	}

	ctx.inputData = obj

	if err = proc.(UpdateProcessor).ProcessUpdate(ctx); err != nil {
		audit(auditer, ctx, err)
		return err
	}

	if len(ctx.events) > 0 {
		pusher(ctx.events...)
	}

	if o, ok := ctx.outputData.(elemental.Identifiable); ok && !ctx.disableOutputDataPush {
		pusher(elemental.NewEvent(elemental.EventUpdate, o))
	}

	audit(auditer, ctx, nil)

	return err
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
		return err
	}

	if err = CheckAuthorization(authorizers, ctx); err != nil {
		audit(auditer, ctx, err)
		return err
	}

	if readOnlyMode {
		if err = makeReadOnlyError(ctx.request.Identity, readOnlyExclusion); err != nil {
			return err
		}
	}

	proc, _ := processorFinder(ctx.request.Identity)

	if _, ok := proc.(DeleteProcessor); !ok {
		err = notImplementedErr(ctx.request)
		audit(auditer, ctx, err)
		return err
	}

	if err = proc.(DeleteProcessor).ProcessDelete(ctx); err != nil {
		audit(auditer, ctx, err)
		return err
	}

	if len(ctx.events) > 0 {
		pusher(ctx.events...)
	}

	if o, ok := ctx.outputData.(elemental.Identifiable); ok && !ctx.disableOutputDataPush {
		pusher(elemental.NewEvent(elemental.EventDelete, o))
	}

	audit(auditer, ctx, nil)

	return err
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
	identifiableRetriever IdentifiableRetriever,
) (err error) {

	if err = CheckAuthentication(authenticators, ctx); err != nil {
		audit(auditer, ctx, err)
		return err
	}

	if err = CheckAuthorization(authorizers, ctx); err != nil {
		audit(auditer, ctx, err)
		return err
	}

	if readOnlyMode {
		if err = makeReadOnlyError(ctx.request.Identity, readOnlyExclusion); err != nil {
			return err
		}
	}

	proc, _ := processorFinder(ctx.request.Identity)

	if identifiableRetriever != nil {
		if _, ok := proc.(UpdateProcessor); !ok {
			err = notImplementedErr(ctx.request)
			audit(auditer, ctx, err)
			return err
		}
	} else {
		if _, ok := proc.(PatchProcessor); !ok {
			err = notImplementedErr(ctx.request)
			audit(auditer, ctx, err)
			return err
		}
	}
	var sparse elemental.Identifiable

	if unmarshaller != nil {
		if sparse, err = unmarshaller(ctx.request); err != nil {
			audit(auditer, ctx, err)
			return elemental.NewError("Bad Request", err.Error(), "bahamut", http.StatusBadRequest)
		}
	} else {
		sparse = modelManager.SparseIdentifiable(ctx.request.Identity)
		if err := ctx.Request().Decode(sparse); err != nil {
			audit(auditer, ctx, err)
			return elemental.NewError("Bad Request", err.Error(), "bahamut", http.StatusBadRequest)
		}
	}

	if identifiableRetriever != nil {
		identifiable, err := identifiableRetriever(ctx.Request())
		if err != nil {
			audit(auditer, ctx, err)
			return err
		}

		patchable, ok := identifiable.(elemental.Patchable)
		if !ok {
			audit(auditer, ctx, err)
			return elemental.NewError("Bad Request", "Identifiable is not patchable", "bahamut", http.StatusBadRequest)
		}

		patchable.Patch(sparse.(elemental.SparseIdentifiable))

		if v, ok := patchable.(elemental.Validatable); ok {
			if err = v.Validate(); err != nil {
				audit(auditer, ctx, err)
				return err
			}
		}

		ctx.inputData = patchable

		if err = proc.(UpdateProcessor).ProcessUpdate(ctx); err != nil {
			audit(auditer, ctx, err)
			return err
		}
	} else {
		ctx.inputData = sparse
		if err = proc.(PatchProcessor).ProcessPatch(ctx); err != nil {
			audit(auditer, ctx, err)
			return err
		}
	}

	if len(ctx.events) > 0 {
		pusher(ctx.events...)
	}

	if o, ok := ctx.outputData.(elemental.Identifiable); ok && !ctx.disableOutputDataPush {
		pusher(elemental.NewEvent(elemental.EventUpdate, o))
	}

	audit(auditer, ctx, nil)

	return err
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
		return err
	}

	if err = CheckAuthorization(authorizers, ctx); err != nil {
		audit(auditer, ctx, err)
		return err
	}

	proc, _ := processorFinder(ctx.request.Identity)

	if _, ok := proc.(InfoProcessor); !ok {
		err = notImplementedErr(ctx.request)
		audit(auditer, ctx, err)
		return err
	}

	if err = proc.(InfoProcessor).ProcessInfo(ctx); err != nil {
		audit(auditer, ctx, err)
		return err
	}

	if len(ctx.events) > 0 {
		pusher(ctx.events...)
	}

	audit(auditer, ctx, nil)

	return err
}

func makeReadOnlyError(identity elemental.Identity, readOnlyExclusion []elemental.Identity) error {

	for _, i := range readOnlyExclusion {
		if i.IsEqual(identity) {
			return nil
		}
	}

	return elemental.NewError("Locked", "This api is currently locked. Please try again later", "bahamut", http.StatusLocked)
}
