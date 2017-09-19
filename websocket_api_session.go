// Author: Antoine Mercadal
// See LICENSE file for full LICENSE
// Copyright 2016 Aporeto.

package bahamut

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/aporeto-inc/elemental"
	"golang.org/x/net/websocket"
)

type wsAPISession struct {
	processorFinder processorFinderFunc
	eventPusher     eventPusherFunc
	requests        chan *elemental.Request

	*wsSession
}

func newWSAPISession(ws *websocket.Conn, config Config, unregister unregisterFunc, processorFinder processorFinderFunc, eventPusher eventPusherFunc) internalWSSession {

	return &wsAPISession{
		wsSession:       newWSSession(ws, config, unregister),
		processorFinder: processorFinder,
		eventPusher:     eventPusher,
		requests:        make(chan *elemental.Request, 8),
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
		request := elemental.NewRequest()
		request.ClientIP = s.remoteAddr

		if err := websocket.JSON.Receive(s.socket, request); err != nil {
			if _, ok := err.(*json.SyntaxError); ok {
				response := elemental.NewResponse()
				response.Request = request
				writeWebSocketError(s.socket, response, elemental.NewError("Bad Request", "Invalid JSON", "bahamut", http.StatusBadRequest))
				continue
			}

			s.stopAll <- true
			return
		}

		select {
		case s.requests <- request:
		case <-s.stopRead:
			return
		}
	}
}

func (s *wsAPISession) listen() {

	go s.read()
	defer s.stop()

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

			switch request.Operation {

			case elemental.OperationRetrieveMany:
				go s.handleRetrieveMany(request)

			case elemental.OperationRetrieve:
				go s.handleRetrieve(request)

			case elemental.OperationCreate:
				go s.handleCreate(request)

			case elemental.OperationUpdate:
				go s.handleUpdate(request)

			case elemental.OperationDelete:
				go s.handleDelete(request)

			case elemental.OperationInfo:
				go s.handleInfo(request)

			case elemental.OperationPatch:
				go s.handlePatch(request)
			}

		case <-s.stopAll:
			return
		}
	}
}

func (s *wsAPISession) handleEventualPanic(response *elemental.Response) {

	err := handleRecoveredPanic(recover(), response.Request)
	if err == nil {
		return
	}

	writeWebSocketError(s.socket, response, err)
}

func (s *wsAPISession) handleRetrieveMany(request *elemental.Request) {

	response := elemental.NewResponse()
	response.Request = request

	request.StartTracing()
	defer request.FinishTracing()
	defer s.handleEventualPanic(response)

	parentIdentity := request.ParentIdentity
	if parentIdentity.IsEmpty() {
		parentIdentity = elemental.RootIdentity
	}

	if !elemental.IsRetrieveManyAllowed(s.config.Model.RelationshipsRegistry[request.Version], request.Identity, parentIdentity) {
		writeWebSocketError(s.socket, response, elemental.NewError("Not allowed", "RetrieveMany operation not allowed on "+request.Identity.Category, "bahamut", http.StatusMethodNotAllowed))
		return
	}

	ctx, err := dispatchRetrieveManyOperation(
		request,
		s.processorFinder,
		s.config.Model.IdentifiablesFactory,
		s.config.Security.RequestAuthenticator,
		s.config.Security.Authorizer,
		s.eventPusher,
		s.config.Security.Auditer,
	)

	if err != nil {
		writeWebSocketError(s.socket, response, err)
		return
	}

	if err := writeWebsocketResponse(s.socket, response, ctx); err != nil {
		writeWebSocketError(s.socket, response, err)
	}
}

func (s *wsAPISession) handleRetrieve(request *elemental.Request) {

	response := elemental.NewResponse()
	response.Request = request

	request.StartTracing()
	defer request.FinishTracing()
	defer s.handleEventualPanic(response)

	if !elemental.IsRetrieveAllowed(s.config.Model.RelationshipsRegistry[request.Version], request.Identity) || !request.ParentIdentity.IsEmpty() {
		writeWebSocketError(s.socket, response, elemental.NewError("Not allowed", "Retrieve operation not allowed on "+request.Identity.Name, "bahamut", http.StatusMethodNotAllowed))
		return
	}

	ctx, err := dispatchRetrieveOperation(
		response.Request,
		s.processorFinder,
		s.config.Model.IdentifiablesFactory,
		s.config.Security.RequestAuthenticator,
		s.config.Security.Authorizer,
		s.eventPusher,
		s.config.Security.Auditer,
	)

	if err != nil {
		writeWebSocketError(s.socket, response, err)
		return
	}

	if err := writeWebsocketResponse(s.socket, response, ctx); err != nil {
		writeWebSocketError(s.socket, response, err)
	}
}

func (s *wsAPISession) handleCreate(request *elemental.Request) {

	response := elemental.NewResponse()
	response.Request = request

	request.StartTracing()
	defer request.FinishTracing()
	defer s.handleEventualPanic(response)

	parentIdentity := request.ParentIdentity
	if parentIdentity.IsEmpty() {
		parentIdentity = elemental.RootIdentity
	}

	if !elemental.IsCreateAllowed(s.config.Model.RelationshipsRegistry[request.Version], request.Identity, parentIdentity) {
		writeWebSocketError(s.socket, response, elemental.NewError("Not allowed", "Create operation not allowed on "+request.Identity.Name, "bahamut", http.StatusMethodNotAllowed))
		return
	}

	ctx, err := dispatchCreateOperation(
		response.Request,
		s.processorFinder,
		s.config.Model.IdentifiablesFactory,
		s.config.Security.RequestAuthenticator,
		s.config.Security.Authorizer,
		s.eventPusher,
		s.config.Security.Auditer,
		s.config.Model.ReadOnly,
	)

	if err != nil {
		writeWebSocketError(s.socket, response, err)
		return
	}

	if err := writeWebsocketResponse(s.socket, response, ctx); err != nil {
		writeWebSocketError(s.socket, response, err)
	}
}

func (s *wsAPISession) handleUpdate(request *elemental.Request) {

	response := elemental.NewResponse()
	response.Request = request

	request.StartTracing()
	defer request.FinishTracing()
	defer s.handleEventualPanic(response)

	if !elemental.IsUpdateAllowed(s.config.Model.RelationshipsRegistry[request.Version], request.Identity) || !request.ParentIdentity.IsEmpty() {
		writeWebSocketError(s.socket, response, elemental.NewError("Not allowed", "Update operation not allowed on "+request.Identity.Name, "bahamut", http.StatusMethodNotAllowed))
		return
	}

	ctx, err := dispatchUpdateOperation(
		response.Request,
		s.processorFinder,
		s.config.Model.IdentifiablesFactory,
		s.config.Security.RequestAuthenticator,
		s.config.Security.Authorizer,
		s.eventPusher,
		s.config.Security.Auditer,
		s.config.Model.ReadOnly,
	)

	if err != nil {
		writeWebSocketError(s.socket, response, err)
		return
	}

	if err := writeWebsocketResponse(s.socket, response, ctx); err != nil {
		writeWebSocketError(s.socket, response, err)
	}
}

func (s *wsAPISession) handleDelete(request *elemental.Request) {

	response := elemental.NewResponse()
	response.Request = request

	request.StartTracing()
	defer request.FinishTracing()
	defer s.handleEventualPanic(response)

	if !elemental.IsDeleteAllowed(s.config.Model.RelationshipsRegistry[request.Version], request.Identity) || !request.ParentIdentity.IsEmpty() {
		writeWebSocketError(s.socket, response, elemental.NewError("Not allowed", "Delete operation not allowed on "+request.Identity.Name, "bahamut", http.StatusMethodNotAllowed))
		return
	}

	ctx, err := dispatchDeleteOperation(
		response.Request,
		s.processorFinder,
		s.config.Model.IdentifiablesFactory,
		s.config.Security.RequestAuthenticator,
		s.config.Security.Authorizer,
		s.eventPusher,
		s.config.Security.Auditer,
		s.config.Model.ReadOnly,
	)

	if err != nil {
		writeWebSocketError(s.socket, response, err)
		return
	}

	if err := writeWebsocketResponse(s.socket, response, ctx); err != nil {
		writeWebSocketError(s.socket, response, err)
	}
}

func (s *wsAPISession) handleInfo(request *elemental.Request) {

	response := elemental.NewResponse()
	response.Request = request

	request.StartTracing()
	defer request.FinishTracing()
	defer s.handleEventualPanic(response)

	parentIdentity := request.ParentIdentity
	if parentIdentity.IsEmpty() {
		parentIdentity = elemental.RootIdentity
	}

	if !elemental.IsInfoAllowed(s.config.Model.RelationshipsRegistry[request.Version], request.Identity, parentIdentity) {
		writeWebSocketError(s.socket, response, elemental.NewError("Not allowed", "Info operation not allowed on "+request.Identity.Category, "bahamut", http.StatusMethodNotAllowed))
		return
	}

	ctx, err := dispatchInfoOperation(
		response.Request,
		s.processorFinder,
		s.config.Model.IdentifiablesFactory,
		s.config.Security.RequestAuthenticator,
		s.config.Security.Authorizer,
		s.config.Security.Auditer,
	)

	if err != nil {
		writeWebSocketError(s.socket, response, err)
		return
	}

	if err := writeWebsocketResponse(s.socket, response, ctx); err != nil {
		writeWebSocketError(s.socket, response, err)
	}
}

func (s *wsAPISession) handlePatch(request *elemental.Request) {

	response := elemental.NewResponse()
	response.Request = request

	request.StartTracing()
	defer request.FinishTracing()
	defer s.handleEventualPanic(response)

	parentIdentity := request.ParentIdentity
	if parentIdentity.IsEmpty() {
		parentIdentity = elemental.RootIdentity
	}

	if !elemental.IsPatchAllowed(s.config.Model.RelationshipsRegistry[request.Version], request.Identity, parentIdentity) {
		writeWebSocketError(s.socket, response, elemental.NewError("Not allowed", "Patch operation not allowed on "+request.Identity.Name, "bahamut", http.StatusMethodNotAllowed))
		return
	}

	ctx, err := dispatchPatchOperation(
		response.Request,
		s.processorFinder,
		s.config.Model.IdentifiablesFactory,
		s.config.Security.RequestAuthenticator,
		s.config.Security.Authorizer,
		s.eventPusher,
		s.config.Security.Auditer,
		s.config.Model.ReadOnly,
	)

	if err != nil {
		writeWebSocketError(s.socket, response, err)
		return
	}

	if err := writeWebsocketResponse(s.socket, response, ctx); err != nil {
		writeWebSocketError(s.socket, response, err)
	}
}
