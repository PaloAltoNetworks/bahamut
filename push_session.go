// Author: Antoine Mercadal
// See LICENSE file for full LICENSE
// Copyright 2016 Aporeto.

package bahamut

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/Sirupsen/logrus"
	"github.com/aporeto-inc/elemental"
	"github.com/satori/go.uuid"
	"golang.org/x/net/websocket"
)

type pushSessionType int

const (
	pushSessionTypeEvent pushSessionType = iota + 1
	pushSessionTypeAPI
)

// PushSession represents a client session.
type PushSession struct {

	// UserInfo contains user opaque information.
	UserInfo interface{}

	Parameters url.Values
	Headers    http.Header

	id        string
	server    *pushServer
	socket    *websocket.Conn
	events    chan *elemental.Event
	requests  chan *elemental.Request
	stopAll   chan bool
	stopRead  chan bool
	stopWrite chan bool
	sType     pushSessionType
}

func newPushSession(ws *websocket.Conn, server *pushServer) *PushSession {

	return newSession(ws, server, pushSessionTypeEvent)
}

func newAPISession(ws *websocket.Conn, server *pushServer) *PushSession {

	return newSession(ws, server, pushSessionTypeAPI)
}

func newSession(ws *websocket.Conn, server *pushServer, sType pushSessionType) *PushSession {

	var parameters url.Values
	var headers http.Header

	if request := ws.Request(); request != nil {
		parameters = request.URL.Query()
	}

	if config := ws.Config(); config != nil {
		headers = config.Header
	}

	return &PushSession{
		id:         uuid.NewV4().String(),
		server:     server,
		socket:     ws,
		events:     make(chan *elemental.Event, 1024),
		requests:   make(chan *elemental.Request, 1024),
		stopRead:   make(chan bool, 2),
		stopWrite:  make(chan bool, 2),
		stopAll:    make(chan bool, 2),
		Parameters: parameters,
		Headers:    headers,
		sType:      sType,
	}
}

// Identifier returns the identifier of the push session.
func (s *PushSession) Identifier() string {

	return s.id
}

// continuously read data from the websocket
func (s *PushSession) read() {

	for {
		var request *elemental.Request

		if err := websocket.JSON.Receive(s.socket, &request); err != nil {
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

func (s *PushSession) write() {

	for {
		select {
		case data := <-s.events:
			if err := websocket.JSON.Send(s.socket, data); err != nil {
				s.stopAll <- true
				return
			}
		case <-s.stopWrite:
			return
		}
	}
}

// force close the current socket
func (s *PushSession) close() {

	s.stopAll <- true
}

// listens to events, either from kafka or from local events.
func (s *PushSession) listen() {

	switch s.sType {
	case pushSessionTypeAPI:
		s.listenToAPIRequest()
	case pushSessionTypeEvent:
		s.listenToPushEvents()
	default:
		panic("Unknown push session type")
	}
}

func (s *PushSession) listenToPushEvents() {

	publications := make(chan *Publication)
	errors := make(chan error)

	unsubscribe := s.server.config.WebSocketServer.Service.Subscribe(publications, errors, s.server.config.WebSocketServer.Topic)

	defer func() {
		s.server.unregisterSession(s)
		_ = s.socket.Close()
		unsubscribe()
	}()

	go s.read()
	go s.write()

	for {
		select {
		case message := <-publications:

			event := &elemental.Event{}
			if err := message.Decode(event); err != nil {
				log.WithFields(logrus.Fields{
					"session": s,
					"message": message,
				}).Error("Unable to decode event.")
				break
			}

			if s.server.config.WebSocketServer.SessionsHandler != nil {

				ok, err := s.server.config.WebSocketServer.SessionsHandler.ShouldPush(s, event)
				if err != nil {
					log.WithError(err).Error("Error during checking authorization.")
					break
				}

				if !ok {
					break
				}
			}

			select {
			case s.events <- event:
			default:
			}

		case err := <-errors:
			log.WithError(err).Error("Error during consuming pubsub topic.")

		case <-s.stopAll:
			s.stopRead <- true
			s.stopWrite <- true
			return
		}
	}
}

func (s *PushSession) listenToAPIRequest() {

	defer func() {
		s.server.unregisterSession(s)
		s.socket.Close()
	}()

	go s.write()
	go s.read()

	for {
		select {
		case request := <-s.requests:

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
			s.stopRead <- true
			s.stopWrite <- true
			return
		}
	}
}

func (s *PushSession) handleRetrieveMany(request *elemental.Request) {

	response := elemental.NewResponse()
	response.Request = request

	ctx, err := dispatchRetrieveManyOperation(
		request,
		s.server.processorFinder,
		s.server.config.Model.IdentifiablesFactory,
		s.server.config.Security.Authenticator,
		s.server.config.Security.Authorizer,
	)

	if err != nil {
		writeWebSocketError(s.socket, response, err)
		return
	}

	writeWebsocketResponse(s.socket, response, ctx)
}

func (s *PushSession) handleRetrieve(request *elemental.Request) {

	response := elemental.NewResponse()
	response.Request = request

	ctx, err := dispatchRetrieveOperation(
		response.Request,
		s.server.processorFinder,
		s.server.config.Model.IdentifiablesFactory,
		s.server.config.Security.Authenticator,
		s.server.config.Security.Authorizer,
	)

	if err != nil {
		writeWebSocketError(s.socket, response, err)
		return
	}

	writeWebsocketResponse(s.socket, response, ctx)
}

func (s *PushSession) handleCreate(request *elemental.Request) {

	response := elemental.NewResponse()
	response.Request = request

	ctx, err := dispatchCreateOperation(
		response.Request,
		s.server.processorFinder,
		s.server.config.Model.IdentifiablesFactory,
		s.server.config.Security.Authenticator,
		s.server.config.Security.Authorizer,
		s.server.pushEvents,
	)

	if err != nil {
		writeWebSocketError(s.socket, response, err)
		return
	}

	writeWebsocketResponse(s.socket, response, ctx)
}

func (s *PushSession) handleUpdate(request *elemental.Request) {

	response := elemental.NewResponse()
	response.Request = request

	ctx, err := dispatchUpdateOperation(
		response.Request,
		s.server.processorFinder,
		s.server.config.Model.IdentifiablesFactory,
		s.server.config.Security.Authenticator,
		s.server.config.Security.Authorizer,
		s.server.pushEvents,
	)

	if err != nil {
		writeWebSocketError(s.socket, response, err)
		return
	}

	writeWebsocketResponse(s.socket, response, ctx)
}

func (s *PushSession) handleDelete(request *elemental.Request) {

	response := elemental.NewResponse()
	response.Request = request

	ctx, err := dispatchDeleteOperation(
		response.Request,
		s.server.processorFinder,
		s.server.config.Model.IdentifiablesFactory,
		s.server.config.Security.Authenticator,
		s.server.config.Security.Authorizer,
		s.server.pushEvents,
	)

	if err != nil {
		writeWebSocketError(s.socket, response, err)
		return
	}

	writeWebsocketResponse(s.socket, response, ctx)
}

func (s *PushSession) handleInfo(request *elemental.Request) {

	response := elemental.NewResponse()
	response.Request = request

	ctx, err := dispatchInfoOperation(
		response.Request,
		s.server.processorFinder,
		s.server.config.Model.IdentifiablesFactory,
		s.server.config.Security.Authenticator,
		s.server.config.Security.Authorizer,
	)

	if err != nil {
		writeWebSocketError(s.socket, response, err)
		return
	}

	writeWebsocketResponse(s.socket, response, ctx)
}

func (s *PushSession) handlePatch(request *elemental.Request) {

	response := elemental.NewResponse()
	response.Request = request

	ctx, err := dispatchPatchOperation(
		response.Request,
		s.server.processorFinder,
		s.server.config.Model.IdentifiablesFactory,
		s.server.config.Security.Authenticator,
		s.server.config.Security.Authorizer,
		s.server.pushEvents,
	)

	if err != nil {
		writeWebSocketError(s.socket, response, err)
		return
	}

	writeWebsocketResponse(s.socket, response, ctx)
}

func (s *PushSession) String() string {

	return fmt.Sprintf("<session id:%s headers: %v parameters: %v>",
		s.id,
		s.Headers,
		s.Parameters,
	)
}
