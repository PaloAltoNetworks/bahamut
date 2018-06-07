package bahamut

import (
	"context"
	"net/http"

	"github.com/aporeto-inc/elemental"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
	"go.uber.org/zap"
)

type handlerFunc func(*Context, config, processorFinderFunc, eventPusherFunc) *elemental.Response

func makeResponse(ctx *Context, response *elemental.Response) *elemental.Response {

	if ctx.Redirect != "" {
		response.Redirect = ctx.Redirect
		return response
	}

	var fields []log.Field
	defer func() {
		span := opentracing.SpanFromContext(ctx.Context())
		if span != nil {
			span.LogFields(fields...)
		}
	}()

	response.StatusCode = ctx.StatusCode
	if response.StatusCode == 0 {
		switch ctx.Request.Operation {
		case elemental.OperationCreate:
			response.StatusCode = http.StatusCreated
		case elemental.OperationInfo:
			response.StatusCode = http.StatusNoContent
		default:
			response.StatusCode = http.StatusOK
		}
	}

	if ctx.Request.Operation == elemental.OperationRetrieveMany || ctx.Request.Operation == elemental.OperationInfo {
		response.Total = ctx.CountTotal
		fields = append(fields, (log.Int("count-total", ctx.CountTotal)))
	}

	if msgs := ctx.messages(); len(msgs) > 0 {
		response.Messages = msgs
		fields = append(fields, (log.Object("messages", msgs)))
	}

	if ctx.OutputData != nil {
		if err := response.Encode(ctx.OutputData); err != nil {
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

func runDispatcher(ctx *Context, r *elemental.Response, d func() error, recover bool) *elemental.Response {

	e := make(chan error)

	go func() {
		defer handleEventualPanic(ctx.Context(), r, e, recover)
		e <- d()
	}()

	select {

	case <-ctx.Context().Done():
		return makeErrorResponse(ctx.Context(), r, ctx.Context().Err())

	case err := <-e:
		if err != nil {
			if _, ok := err.(errMockPanicRequested); ok {
				panic(err.Error())
			}
			return makeErrorResponse(ctx.Context(), r, err)
		}

		return makeResponse(ctx, r)
	}
}

func handleRetrieveMany(ctx *Context, cfg config, processorFinder processorFinderFunc, pusherFunc eventPusherFunc) (response *elemental.Response) {

	response = elemental.NewResponse(ctx.Request)

	parentIdentity := ctx.Request.ParentIdentity
	if parentIdentity.IsEmpty() {
		parentIdentity = elemental.RootIdentity
	}

	if !elemental.IsRetrieveManyAllowed(cfg.model.relationshipsRegistry[ctx.Request.Version], ctx.Request.Identity, parentIdentity) {
		return makeErrorResponse(
			ctx.Context(),
			response,
			elemental.NewError(
				"Not allowed",
				"RetrieveMany operation not allowed on "+ctx.Request.Identity.Category,
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

func handleRetrieve(ctx *Context, cfg config, processorFinder processorFinderFunc, pusherFunc eventPusherFunc) (response *elemental.Response) {

	response = elemental.NewResponse(ctx.Request)

	if !elemental.IsRetrieveAllowed(cfg.model.relationshipsRegistry[ctx.Request.Version], ctx.Request.Identity) || !ctx.Request.ParentIdentity.IsEmpty() {
		return makeErrorResponse(
			ctx.Context(),
			response,
			elemental.NewError(
				"Not allowed",
				"Retrieve operation not allowed on "+ctx.Request.Identity.Name, "bahamut",
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

func handleCreate(ctx *Context, cfg config, processorFinder processorFinderFunc, pusherFunc eventPusherFunc) (response *elemental.Response) {

	response = elemental.NewResponse(ctx.Request)

	parentIdentity := ctx.Request.ParentIdentity
	if parentIdentity.IsEmpty() {
		parentIdentity = elemental.RootIdentity
	}

	if !elemental.IsCreateAllowed(cfg.model.relationshipsRegistry[ctx.Request.Version], ctx.Request.Identity, parentIdentity) {
		return makeErrorResponse(
			ctx.Context(),
			response,
			elemental.NewError(
				"Not allowed",
				"Create operation not allowed on "+ctx.Request.Identity.Name, "bahamut",
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
				cfg.model.identifiableFactories[ctx.Request.Version],
				cfg.model.unmarshallers[ctx.Request.Identity],
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

func handleUpdate(ctx *Context, cfg config, processorFinder processorFinderFunc, pusherFunc eventPusherFunc) (response *elemental.Response) {

	response = elemental.NewResponse(ctx.Request)

	if !elemental.IsUpdateAllowed(cfg.model.relationshipsRegistry[ctx.Request.Version], ctx.Request.Identity) || !ctx.Request.ParentIdentity.IsEmpty() {
		return makeErrorResponse(
			ctx.Context(),
			response,
			elemental.NewError(
				"Not allowed",
				"Update operation not allowed on "+ctx.Request.Identity.Name, "bahamut",
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
				cfg.model.identifiableFactories[ctx.Request.Version],
				cfg.model.unmarshallers[ctx.Request.Identity],
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

func handleDelete(ctx *Context, cfg config, processorFinder processorFinderFunc, pusherFunc eventPusherFunc) (response *elemental.Response) {

	response = elemental.NewResponse(ctx.Request)

	if !elemental.IsDeleteAllowed(cfg.model.relationshipsRegistry[ctx.Request.Version], ctx.Request.Identity) || !ctx.Request.ParentIdentity.IsEmpty() {
		return makeErrorResponse(
			ctx.Context(),
			response,
			elemental.NewError(
				"Not allowed",
				"Delete operation not allowed on "+ctx.Request.Identity.Name, "bahamut",
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

func handleInfo(ctx *Context, cfg config, processorFinder processorFinderFunc, pusherFunc eventPusherFunc) (response *elemental.Response) {

	response = elemental.NewResponse(ctx.Request)

	parentIdentity := ctx.Request.ParentIdentity
	if parentIdentity.IsEmpty() {
		parentIdentity = elemental.RootIdentity
	}

	if !elemental.IsInfoAllowed(cfg.model.relationshipsRegistry[ctx.Request.Version], ctx.Request.Identity, parentIdentity) {
		return makeErrorResponse(
			ctx.Context(),
			response,
			elemental.NewError(
				"Not allowed",
				"Info operation not allowed on "+ctx.Request.Identity.Category, "bahamut",
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

func handlePatch(ctx *Context, cfg config, processorFinder processorFinderFunc, pusherFunc eventPusherFunc) (response *elemental.Response) {

	response = elemental.NewResponse(ctx.Request)

	parentIdentity := ctx.Request.ParentIdentity
	if parentIdentity.IsEmpty() {
		parentIdentity = elemental.RootIdentity
	}

	if !elemental.IsPatchAllowed(cfg.model.relationshipsRegistry[ctx.Request.Version], ctx.Request.Identity, parentIdentity) {
		return makeErrorResponse(
			ctx.Context(),
			response,
			elemental.NewError(
				"Not allowed",
				"Patch operation not allowed on "+ctx.Request.Identity.Category, "bahamut",
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
