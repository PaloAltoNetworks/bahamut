// Author: Antoine Mercadal
// See LICENSE file for full LICENSE
// Copyright 2016 Aporeto.

package bahamut

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/aporeto-inc/elemental"

	opentracing "github.com/opentracing/opentracing-go"
)

type wsAPISession struct {
	processorFinder processorFinderFunc
	eventPusher     eventPusherFunc
	requests        chan *elemental.Request
	responses       chan *elemental.Response
	*wsSession
}

func newWSAPISession(request *http.Request, config Config, unregister unregisterFunc, processorFinder processorFinderFunc, eventPusher eventPusherFunc) *wsAPISession {

	return &wsAPISession{
		wsSession:       newWSSession(request, config, unregister, opentracing.StartSpan("bahamut.session.api")),
		processorFinder: processorFinder,
		eventPusher:     eventPusher,
		requests:        make(chan *elemental.Request),
		responses:       make(chan *elemental.Response),
	}
}

func (s *wsAPISession) String() string {

	return fmt.Sprintf("<apisession id:%s parameters: %v>",
		s.id,
		s.parameters,
	)
}

func (s *wsAPISession) read() {

	for {
		request := elemental.NewRequestWithContext(s.context)
		request.ClientIP = s.remoteAddr

		if err := s.conn.ReadJSON(request); err != nil {
			if _, ok := err.(*json.SyntaxError); !ok {
				s.stop()
				return
			}

			response := elemental.NewResponse()
			response.Request = request

			s.responses <- writeWebSocketError(response, elemental.NewError("Bad Request", "Invalid JSON", "bahamut", http.StatusBadRequest))
		}

		s.requests <- request
	}
}

func (s *wsAPISession) write() {

	for {
		select {
		case resp := <-s.responses:

			if err := s.conn.WriteJSON(resp); err != nil {
				s.stop()
				return
			}

		case <-s.closeCh:
			return
		}
	}
}

func (s *wsAPISession) listen() {

	go s.read()
	go s.write()

	// TODO: this is here for backward compat.
	// we should remvove this when all enforcers
	// are switched to at least manipulate 2.x
	s.responses <- &elemental.Response{
		StatusCode: http.StatusOK,
	}

	for {
		select {
		case request := <-s.requests:

			// We backport the token of the session into the request if we don't have an explicit one given in the request.
			if request.Password == "" {
				if t := s.GetToken(); t != "" {
					request.Username = "Bearer"
					request.Password = t
				}
			}

			// And we set the TLSConnectionState
			request.TLSConnectionState = s.TLSConnectionState()

			switch request.Operation {

			case elemental.OperationRetrieveMany:
				s.handleRetrieveMany(request)

			case elemental.OperationRetrieve:
				s.handleRetrieve(request)

			case elemental.OperationCreate:
				s.handleCreate(request)

			case elemental.OperationUpdate:
				s.handleUpdate(request)

			case elemental.OperationDelete:
				s.handleDelete(request)

			case elemental.OperationInfo:
				s.handleInfo(request)

			case elemental.OperationPatch:
				s.handlePatch(request)
			}

		case <-s.closeCh:
			return
		}
	}
}

func (s *wsAPISession) handleRetrieveMany(request *elemental.Request) {

	response := elemental.NewResponse()
	response.Request = request

	request.StartTracing()
	defer request.FinishTracing()

	parentIdentity := request.ParentIdentity
	if parentIdentity.IsEmpty() {
		parentIdentity = elemental.RootIdentity
	}

	if !elemental.IsRetrieveManyAllowed(s.config.Model.RelationshipsRegistry[request.Version], request.Identity, parentIdentity) {
		s.responses <- writeWebSocketError(response, elemental.NewError("Not allowed", "RetrieveMany operation not allowed on "+request.Identity.Category, "bahamut", http.StatusMethodNotAllowed))
		return
	}

	ctx := NewContextWithRequest(request)

	s.responses <- runWSDispatcher(
		ctx,
		response,
		func() error {
			return dispatchRetrieveManyOperation(
				ctx,
				s.processorFinder,
				s.config.Model.IdentifiablesFactory,
				s.config.Security.RequestAuthenticators,
				s.config.Security.Authorizers,
				s.eventPusher,
				s.config.Security.Auditer,
			)
		},
		s.config.WebSocketServer.PanicRecoveryDisabled,
	)
}

func (s *wsAPISession) handleRetrieve(request *elemental.Request) {

	response := elemental.NewResponse()
	response.Request = request

	request.StartTracing()
	defer request.FinishTracing()

	if !elemental.IsRetrieveAllowed(s.config.Model.RelationshipsRegistry[request.Version], request.Identity) || !request.ParentIdentity.IsEmpty() {
		s.responses <- writeWebSocketError(response, elemental.NewError("Not allowed", "Retrieve operation not allowed on "+request.Identity.Name, "bahamut", http.StatusMethodNotAllowed))
		return
	}

	ctx := NewContextWithRequest(request)

	s.responses <- runWSDispatcher(
		ctx,
		response,
		func() error {
			return dispatchRetrieveOperation(
				ctx,
				s.processorFinder,
				s.config.Model.IdentifiablesFactory,
				s.config.Security.RequestAuthenticators,
				s.config.Security.Authorizers,
				s.eventPusher,
				s.config.Security.Auditer,
			)
		},
		s.config.WebSocketServer.PanicRecoveryDisabled,
	)
}

func (s *wsAPISession) handleCreate(request *elemental.Request) {

	response := elemental.NewResponse()
	response.Request = request

	request.StartTracing()
	defer request.FinishTracing()

	parentIdentity := request.ParentIdentity
	if parentIdentity.IsEmpty() {
		parentIdentity = elemental.RootIdentity
	}

	if !elemental.IsCreateAllowed(s.config.Model.RelationshipsRegistry[request.Version], request.Identity, parentIdentity) {
		s.responses <- writeWebSocketError(response, elemental.NewError("Not allowed", "Create operation not allowed on "+request.Identity.Name, "bahamut", http.StatusMethodNotAllowed))
		return
	}

	ctx := NewContextWithRequest(request)

	s.responses <- runWSDispatcher(
		ctx,
		response,
		func() error {
			return dispatchCreateOperation(
				ctx,
				s.processorFinder,
				s.config.Model.IdentifiablesFactory,
				s.config.Security.RequestAuthenticators,
				s.config.Security.Authorizers,
				s.eventPusher,
				s.config.Security.Auditer,
				s.config.Model.ReadOnly,
				s.config.Model.ReadOnlyExcludedIdentities,
			)
		},
		s.config.WebSocketServer.PanicRecoveryDisabled,
	)
}

func (s *wsAPISession) handleUpdate(request *elemental.Request) {

	response := elemental.NewResponse()
	response.Request = request

	request.StartTracing()
	defer request.FinishTracing()

	if !elemental.IsUpdateAllowed(s.config.Model.RelationshipsRegistry[request.Version], request.Identity) || !request.ParentIdentity.IsEmpty() {
		s.responses <- writeWebSocketError(response, elemental.NewError("Not allowed", "Update operation not allowed on "+request.Identity.Name, "bahamut", http.StatusMethodNotAllowed))
		return
	}

	ctx := NewContextWithRequest(request)

	s.responses <- runWSDispatcher(
		ctx,
		response,
		func() error {
			return dispatchUpdateOperation(
				ctx,
				s.processorFinder,
				s.config.Model.IdentifiablesFactory,
				s.config.Security.RequestAuthenticators,
				s.config.Security.Authorizers,
				s.eventPusher,
				s.config.Security.Auditer,
				s.config.Model.ReadOnly,
				s.config.Model.ReadOnlyExcludedIdentities,
			)
		},
		s.config.WebSocketServer.PanicRecoveryDisabled,
	)
}

func (s *wsAPISession) handleDelete(request *elemental.Request) {

	response := elemental.NewResponse()
	response.Request = request

	request.StartTracing()
	defer request.FinishTracing()

	if !elemental.IsDeleteAllowed(s.config.Model.RelationshipsRegistry[request.Version], request.Identity) || !request.ParentIdentity.IsEmpty() {
		s.responses <- writeWebSocketError(response, elemental.NewError("Not allowed", "Delete operation not allowed on "+request.Identity.Name, "bahamut", http.StatusMethodNotAllowed))
		return
	}

	ctx := NewContextWithRequest(request)

	s.responses <- runWSDispatcher(
		ctx,
		response,
		func() error {
			return dispatchDeleteOperation(
				ctx,
				s.processorFinder,
				s.config.Model.IdentifiablesFactory,
				s.config.Security.RequestAuthenticators,
				s.config.Security.Authorizers,
				s.eventPusher,
				s.config.Security.Auditer,
				s.config.Model.ReadOnly,
				s.config.Model.ReadOnlyExcludedIdentities,
			)
		},
		s.config.WebSocketServer.PanicRecoveryDisabled,
	)
}

func (s *wsAPISession) handleInfo(request *elemental.Request) {

	response := elemental.NewResponse()
	response.Request = request

	request.StartTracing()
	defer request.FinishTracing()

	parentIdentity := request.ParentIdentity
	if parentIdentity.IsEmpty() {
		parentIdentity = elemental.RootIdentity
	}

	if !elemental.IsInfoAllowed(s.config.Model.RelationshipsRegistry[request.Version], request.Identity, parentIdentity) {
		s.responses <- writeWebSocketError(response, elemental.NewError("Not allowed", "Info operation not allowed on "+request.Identity.Category, "bahamut", http.StatusMethodNotAllowed))
		return
	}

	ctx := NewContextWithRequest(request)

	s.responses <- runWSDispatcher(
		ctx,
		response,
		func() error {
			return dispatchInfoOperation(
				ctx,
				s.processorFinder,
				s.config.Model.IdentifiablesFactory,
				s.config.Security.RequestAuthenticators,
				s.config.Security.Authorizers,
				s.config.Security.Auditer,
			)
		},
		s.config.WebSocketServer.PanicRecoveryDisabled,
	)
}

func (s *wsAPISession) handlePatch(request *elemental.Request) {

	response := elemental.NewResponse()
	response.Request = request

	request.StartTracing()
	defer request.FinishTracing()

	parentIdentity := request.ParentIdentity
	if parentIdentity.IsEmpty() {
		parentIdentity = elemental.RootIdentity
	}

	if !elemental.IsPatchAllowed(s.config.Model.RelationshipsRegistry[request.Version], request.Identity, parentIdentity) {
		s.responses <- writeWebSocketError(response, elemental.NewError("Not allowed", "Patch operation not allowed on "+request.Identity.Name, "bahamut", http.StatusMethodNotAllowed))
		return
	}

	ctx := NewContextWithRequest(request)

	s.responses <- runWSDispatcher(
		ctx,
		response,
		func() error {
			return dispatchPatchOperation(
				ctx,
				s.processorFinder,
				s.config.Model.IdentifiablesFactory,
				s.config.Security.RequestAuthenticators,
				s.config.Security.Authorizers,
				s.eventPusher,
				s.config.Security.Auditer,
				s.config.Model.ReadOnly,
				s.config.Model.ReadOnlyExcludedIdentities,
			)
		},
		s.config.WebSocketServer.PanicRecoveryDisabled,
	)
}
