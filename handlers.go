package bahamut

import (
	"context"
	"net/http"

	"github.com/opentracing/opentracing-go"

	"github.com/aporeto-inc/elemental"
	"github.com/opentracing/opentracing-go/log"
	"go.uber.org/zap"
)

type handlerFunc func(*Context, Config, *elemental.Request, processorFinderFunc, eventPusherFunc) *elemental.Response

func makeResponse(c *Context, response *elemental.Response) *elemental.Response {

	if c.Redirect != "" {
		response.Redirect = c.Redirect
		return response
	}

	var fields []log.Field
	defer func() {
		span := opentracing.SpanFromContext(c)
		if span != nil {
			span.LogFields(fields...)
		}
	}()

	response.StatusCode = c.StatusCode
	if response.StatusCode == 0 {
		switch c.Request.Operation {
		case elemental.OperationCreate:
			response.StatusCode = http.StatusCreated
		case elemental.OperationInfo:
			response.StatusCode = http.StatusNoContent
		default:
			response.StatusCode = http.StatusOK
		}
	}

	if c.Request.Operation == elemental.OperationRetrieveMany || c.Request.Operation == elemental.OperationInfo {
		response.Count = c.CountTotal
		fields = append(fields, (log.Int("count-total", c.CountTotal)))
	}

	if msgs := c.messages(); len(msgs) > 0 {
		response.Messages = msgs
		fields = append(fields, (log.Object("messages", msgs)))
	}

	if c.OutputData != nil {
		if err := response.Encode(c.OutputData); err != nil {
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
		defer handleEventualPanic(ctx, r, e, recover)
		e <- d()
	}()

	select {

	case <-ctx.Done():
		return nil

	case err := <-e:
		if err != nil {
			if _, ok := err.(errMockPanicRequested); ok {
				panic(err.Error())
			}
			return makeErrorResponse(ctx, r, err)
		}

		return makeResponse(ctx, r)
	}
}

func handleRetrieveMany(ctx *Context, config Config, request *elemental.Request, processorFinder processorFinderFunc, pusherFunc eventPusherFunc) (response *elemental.Response) {

	response = elemental.NewResponse()

	parentIdentity := request.ParentIdentity
	if parentIdentity.IsEmpty() {
		parentIdentity = elemental.RootIdentity
	}

	if !elemental.IsRetrieveManyAllowed(config.Model.RelationshipsRegistry[request.Version], request.Identity, parentIdentity) {
		return makeErrorResponse(
			ctx,
			response,
			elemental.NewError(
				"Not allowed",
				"RetrieveMany operation not allowed on "+request.Identity.Category,
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
				config.Model.IdentifiablesFactory,
				config.Security.RequestAuthenticators,
				config.Security.Authorizers,
				pusherFunc,
				config.Security.Auditer,
			)
		},
		config.WebSocketServer.PanicRecoveryDisabled,
	)
}

func handleRetrieve(ctx *Context, config Config, request *elemental.Request, processorFinder processorFinderFunc, pusherFunc eventPusherFunc) (response *elemental.Response) {

	response = elemental.NewResponse()

	if !elemental.IsRetrieveAllowed(config.Model.RelationshipsRegistry[request.Version], request.Identity) || !request.ParentIdentity.IsEmpty() {
		return makeErrorResponse(
			ctx,
			response,
			elemental.NewError(
				"Not allowed",
				"Retrieve operation not allowed on "+request.Identity.Name, "bahamut",
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
				config.Model.IdentifiablesFactory,
				config.Security.RequestAuthenticators,
				config.Security.Authorizers,
				pusherFunc,
				config.Security.Auditer,
			)
		},
		config.WebSocketServer.PanicRecoveryDisabled,
	)
}

func handleCreate(ctx *Context, config Config, request *elemental.Request, processorFinder processorFinderFunc, pusherFunc eventPusherFunc) (response *elemental.Response) {

	response = elemental.NewResponse()

	parentIdentity := request.ParentIdentity
	if parentIdentity.IsEmpty() {
		parentIdentity = elemental.RootIdentity
	}

	if !elemental.IsCreateAllowed(config.Model.RelationshipsRegistry[request.Version], request.Identity, parentIdentity) {
		return makeErrorResponse(
			ctx,
			response,
			elemental.NewError(
				"Not allowed",
				"Create operation not allowed on "+request.Identity.Name, "bahamut",
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
				config.Model.IdentifiablesFactory,
				config.Security.RequestAuthenticators,
				config.Security.Authorizers,
				pusherFunc,
				config.Security.Auditer,
				config.Model.ReadOnly,
				config.Model.ReadOnlyExcludedIdentities,
			)
		},
		config.WebSocketServer.PanicRecoveryDisabled,
	)
}

func handleUpdate(ctx *Context, config Config, request *elemental.Request, processorFinder processorFinderFunc, pusherFunc eventPusherFunc) (response *elemental.Response) {

	response = elemental.NewResponse()

	if !elemental.IsUpdateAllowed(config.Model.RelationshipsRegistry[request.Version], request.Identity) || !request.ParentIdentity.IsEmpty() {
		return makeErrorResponse(
			ctx,
			response,
			elemental.NewError(
				"Not allowed",
				"Update operation not allowed on "+request.Identity.Name, "bahamut",
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
				config.Model.IdentifiablesFactory,
				config.Security.RequestAuthenticators,
				config.Security.Authorizers,
				pusherFunc,
				config.Security.Auditer,
				config.Model.ReadOnly,
				config.Model.ReadOnlyExcludedIdentities,
			)
		},
		config.WebSocketServer.PanicRecoveryDisabled,
	)
}

func handleDelete(ctx *Context, config Config, request *elemental.Request, processorFinder processorFinderFunc, pusherFunc eventPusherFunc) (response *elemental.Response) {

	response = elemental.NewResponse()

	if !elemental.IsDeleteAllowed(config.Model.RelationshipsRegistry[request.Version], request.Identity) || !request.ParentIdentity.IsEmpty() {
		return makeErrorResponse(
			ctx,
			response,
			elemental.NewError(
				"Not allowed",
				"Delete operation not allowed on "+request.Identity.Name, "bahamut",
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
				config.Model.IdentifiablesFactory,
				config.Security.RequestAuthenticators,
				config.Security.Authorizers,
				pusherFunc,
				config.Security.Auditer,
				config.Model.ReadOnly,
				config.Model.ReadOnlyExcludedIdentities,
			)
		},
		config.WebSocketServer.PanicRecoveryDisabled,
	)
}

func handleInfo(ctx *Context, config Config, request *elemental.Request, processorFinder processorFinderFunc, pusherFunc eventPusherFunc) (response *elemental.Response) {

	response = elemental.NewResponse()

	parentIdentity := request.ParentIdentity
	if parentIdentity.IsEmpty() {
		parentIdentity = elemental.RootIdentity
	}

	if !elemental.IsInfoAllowed(config.Model.RelationshipsRegistry[request.Version], request.Identity, parentIdentity) {
		return makeErrorResponse(
			ctx,
			response,
			elemental.NewError(
				"Not allowed",
				"Info operation not allowed on "+request.Identity.Category, "bahamut",
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
				config.Model.IdentifiablesFactory,
				config.Security.RequestAuthenticators,
				config.Security.Authorizers,
				pusherFunc,
				config.Security.Auditer,
			)
		},
		config.WebSocketServer.PanicRecoveryDisabled,
	)
}

func handlePatch(ctx *Context, config Config, request *elemental.Request, processorFinder processorFinderFunc, pusherFunc eventPusherFunc) (response *elemental.Response) {

	response = elemental.NewResponse()

	parentIdentity := request.ParentIdentity
	if parentIdentity.IsEmpty() {
		parentIdentity = elemental.RootIdentity
	}

	if !elemental.IsPatchAllowed(config.Model.RelationshipsRegistry[request.Version], request.Identity, parentIdentity) {
		return makeErrorResponse(
			ctx,
			response,
			elemental.NewError(
				"Not allowed",
				"Patch operation not allowed on "+request.Identity.Name, "bahamut",
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
				config.Model.IdentifiablesFactory,
				config.Security.RequestAuthenticators,
				config.Security.Authorizers,
				pusherFunc,
				config.Security.Auditer,
				config.Model.ReadOnly,
				config.Model.ReadOnlyExcludedIdentities,
			)
		},
		config.WebSocketServer.PanicRecoveryDisabled,
	)
}
