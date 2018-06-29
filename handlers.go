package bahamut

import (
	"context"
	"net/http"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
	"go.aporeto.io/elemental"
	"go.uber.org/zap"
)

type handlerFunc func(*bcontext, config, processorFinderFunc, eventPusherFunc) *elemental.Response

func makeResponse(ctx *bcontext, response *elemental.Response) *elemental.Response {

	if ctx.redirect != "" {
		response.Redirect = ctx.redirect
		return response
	}

	var fields []log.Field
	defer func() {
		span := opentracing.SpanFromContext(ctx.ctx)
		if span != nil {
			span.LogFields(fields...)
		}
	}()

	response.StatusCode = ctx.statusCode
	if response.StatusCode == 0 {
		switch ctx.request.Operation {
		case elemental.OperationCreate:
			response.StatusCode = http.StatusCreated
		case elemental.OperationInfo:
			response.StatusCode = http.StatusNoContent
		default:
			response.StatusCode = http.StatusOK
		}
	}

	if ctx.request.Operation == elemental.OperationRetrieveMany || ctx.request.Operation == elemental.OperationInfo {
		response.Total = ctx.countTotal
		fields = append(fields, (log.Int("count-total", ctx.countTotal)))
	}

	if msgs := ctx.messages; len(msgs) > 0 {
		response.Messages = msgs
		fields = append(fields, (log.Object("messages", msgs)))
	}

	if ctx.outputData != nil {
		if err := response.Encode(ctx.outputData); err != nil {
			zap.L().Panic("Unable to encode output data", zap.Error(err))
		}
		fields = append(fields, (log.Object("response", string(response.Data))))
	} else {
		response.StatusCode = http.StatusNoContent
	}

	fields = append(fields, (log.Int("status.code", response.StatusCode)))

	return response
}

func makeErrorResponse(ctx context.Context, response *elemental.Response, err error) *elemental.Response {

	outError := processError(ctx, err, response)

	response.StatusCode = outError.Code()
	if e := response.Encode(outError); e != nil {
		zap.L().Panic("Unable to encode error", zap.Error(err))
	}

	return response
}

func handleEventualPanic(ctx context.Context, response *elemental.Response, c chan error, reco bool) {

	if err := handleRecoveredPanic(ctx, recover(), response, reco); err != nil {
		c <- err
	}
}

func runDispatcher(ctx *bcontext, r *elemental.Response, d func() error, recover bool) *elemental.Response {

	e := make(chan error)

	go func() {
		defer handleEventualPanic(ctx.ctx, r, e, recover)
		e <- d()
	}()

	select {

	case <-ctx.ctx.Done():
		return makeErrorResponse(ctx.ctx, r, ctx.ctx.Err())

	case err := <-e:
		if err != nil {
			return makeErrorResponse(ctx.ctx, r, err)
		}

		return makeResponse(ctx, r)
	}
}

func handleRetrieveMany(ctx *bcontext, cfg config, processorFinder processorFinderFunc, pusherFunc eventPusherFunc) (response *elemental.Response) {

	response = elemental.NewResponse(ctx.request)

	parentIdentity := ctx.request.ParentIdentity
	if parentIdentity.IsEmpty() {
		parentIdentity = elemental.RootIdentity
	}

	if !elemental.IsRetrieveManyAllowed(cfg.model.modelManagers[ctx.request.Version].Relationships(), ctx.request.Identity, parentIdentity) {
		return makeErrorResponse(
			ctx.ctx,
			response,
			elemental.NewError(
				"Not allowed",
				"RetrieveMany operation not allowed on "+ctx.request.Identity.Category,
				"bahamut",
				http.StatusMethodNotAllowed,
			),
		)
	}

	return runDispatcher(
		ctx,
		response,
		func() error {
			return dispatchRetrieveManyOperation(
				ctx,
				processorFinder,
				cfg.security.requestAuthenticators,
				cfg.security.authorizers,
				pusherFunc,
				cfg.security.auditer,
			)
		},
		cfg.general.panicRecoveryDisabled,
	)
}

func handleRetrieve(ctx *bcontext, cfg config, processorFinder processorFinderFunc, pusherFunc eventPusherFunc) (response *elemental.Response) {

	response = elemental.NewResponse(ctx.request)

	if !elemental.IsRetrieveAllowed(cfg.model.modelManagers[ctx.request.Version].Relationships(), ctx.request.Identity) || !ctx.request.ParentIdentity.IsEmpty() {
		return makeErrorResponse(
			ctx.ctx,
			response,
			elemental.NewError(
				"Not allowed",
				"Retrieve operation not allowed on "+ctx.request.Identity.Name, "bahamut",
				http.StatusMethodNotAllowed,
			),
		)
	}

	return runDispatcher(
		ctx,
		response,
		func() error {
			return dispatchRetrieveOperation(
				ctx,
				processorFinder,
				cfg.security.requestAuthenticators,
				cfg.security.authorizers,
				pusherFunc,
				cfg.security.auditer,
			)
		},
		cfg.general.panicRecoveryDisabled,
	)
}

func handleCreate(ctx *bcontext, cfg config, processorFinder processorFinderFunc, pusherFunc eventPusherFunc) (response *elemental.Response) {

	response = elemental.NewResponse(ctx.request)

	parentIdentity := ctx.request.ParentIdentity
	if parentIdentity.IsEmpty() {
		parentIdentity = elemental.RootIdentity
	}

	if !elemental.IsCreateAllowed(cfg.model.modelManagers[ctx.request.Version].Relationships(), ctx.request.Identity, parentIdentity) {
		return makeErrorResponse(
			ctx.ctx,
			response,
			elemental.NewError(
				"Not allowed",
				"Create operation not allowed on "+ctx.request.Identity.Name, "bahamut",
				http.StatusMethodNotAllowed,
			),
		)
	}

	return runDispatcher(
		ctx,
		response,
		func() error {
			return dispatchCreateOperation(
				ctx,
				processorFinder,
				cfg.model.modelManagers[ctx.request.Version],
				cfg.model.unmarshallers[ctx.request.Identity],
				cfg.security.requestAuthenticators,
				cfg.security.authorizers,
				pusherFunc,
				cfg.security.auditer,
				cfg.model.readOnly,
				cfg.model.readOnlyExcludedIdentities,
			)
		},
		cfg.general.panicRecoveryDisabled,
	)
}

func handleUpdate(ctx *bcontext, cfg config, processorFinder processorFinderFunc, pusherFunc eventPusherFunc) (response *elemental.Response) {

	response = elemental.NewResponse(ctx.request)

	if !elemental.IsUpdateAllowed(cfg.model.modelManagers[ctx.request.Version].Relationships(), ctx.request.Identity) || !ctx.request.ParentIdentity.IsEmpty() {
		return makeErrorResponse(
			ctx.ctx,
			response,
			elemental.NewError(
				"Not allowed",
				"Update operation not allowed on "+ctx.request.Identity.Name, "bahamut",
				http.StatusMethodNotAllowed,
			),
		)
	}

	return runDispatcher(
		ctx,
		response,
		func() error {
			return dispatchUpdateOperation(
				ctx,
				processorFinder,
				cfg.model.modelManagers[ctx.request.Version],
				cfg.model.unmarshallers[ctx.request.Identity],
				cfg.security.requestAuthenticators,
				cfg.security.authorizers,
				pusherFunc,
				cfg.security.auditer,
				cfg.model.readOnly,
				cfg.model.readOnlyExcludedIdentities,
			)
		},
		cfg.general.panicRecoveryDisabled,
	)
}

func handleDelete(ctx *bcontext, cfg config, processorFinder processorFinderFunc, pusherFunc eventPusherFunc) (response *elemental.Response) {

	response = elemental.NewResponse(ctx.request)

	if !elemental.IsDeleteAllowed(cfg.model.modelManagers[ctx.request.Version].Relationships(), ctx.request.Identity) || !ctx.request.ParentIdentity.IsEmpty() {
		return makeErrorResponse(
			ctx.ctx,
			response,
			elemental.NewError(
				"Not allowed",
				"Delete operation not allowed on "+ctx.request.Identity.Name, "bahamut",
				http.StatusMethodNotAllowed,
			),
		)
	}

	return runDispatcher(
		ctx,
		response,
		func() error {
			return dispatchDeleteOperation(
				ctx,
				processorFinder,
				cfg.security.requestAuthenticators,
				cfg.security.authorizers,
				pusherFunc,
				cfg.security.auditer,
				cfg.model.readOnly,
				cfg.model.readOnlyExcludedIdentities,
			)
		},
		cfg.general.panicRecoveryDisabled,
	)
}

func handleInfo(ctx *bcontext, cfg config, processorFinder processorFinderFunc, pusherFunc eventPusherFunc) (response *elemental.Response) {

	response = elemental.NewResponse(ctx.request)

	parentIdentity := ctx.request.ParentIdentity
	if parentIdentity.IsEmpty() {
		parentIdentity = elemental.RootIdentity
	}

	if !elemental.IsInfoAllowed(cfg.model.modelManagers[ctx.request.Version].Relationships(), ctx.request.Identity, parentIdentity) {
		return makeErrorResponse(
			ctx.ctx,
			response,
			elemental.NewError(
				"Not allowed",
				"Info operation not allowed on "+ctx.request.Identity.Category, "bahamut",
				http.StatusMethodNotAllowed,
			),
		)
	}

	return runDispatcher(
		ctx,
		response,
		func() error {
			return dispatchInfoOperation(
				ctx,
				processorFinder,
				cfg.security.requestAuthenticators,
				cfg.security.authorizers,
				pusherFunc,
				cfg.security.auditer,
			)
		},
		cfg.general.panicRecoveryDisabled,
	)
}

func handlePatch(ctx *bcontext, cfg config, processorFinder processorFinderFunc, pusherFunc eventPusherFunc) (response *elemental.Response) {

	response = elemental.NewResponse(ctx.request)

	parentIdentity := ctx.request.ParentIdentity
	if parentIdentity.IsEmpty() {
		parentIdentity = elemental.RootIdentity
	}

	if !elemental.IsPatchAllowed(cfg.model.modelManagers[ctx.request.Version].Relationships(), ctx.request.Identity, parentIdentity) {
		return makeErrorResponse(
			ctx.ctx,
			response,
			elemental.NewError(
				"Not allowed",
				"Patch operation not allowed on "+ctx.request.Identity.Category, "bahamut",
				http.StatusMethodNotAllowed,
			),
		)
	}

	return runDispatcher(
		ctx,
		response,
		func() error {
			return dispatchPatchOperation(
				ctx,
				processorFinder,
				cfg.security.requestAuthenticators,
				cfg.security.authorizers,
				pusherFunc,
				cfg.security.auditer,
				cfg.model.readOnly,
				cfg.model.readOnlyExcludedIdentities,
			)
		},
		cfg.general.panicRecoveryDisabled,
	)
}
