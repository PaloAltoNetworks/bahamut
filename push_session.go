// Author: Antoine Mercadal
// See LICENSE file for full LICENSE
// Copyright 2016 Aporeto.

package bahamut

import (
	"fmt"
	"net/http"

	"github.com/aporeto-inc/elemental"
	"github.com/satori/go.uuid"
	"golang.org/x/net/websocket"

	log "github.com/Sirupsen/logrus"
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

	// Info contains various request related information.
	Info *Info

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

	info := &Info{}

	if request := ws.Request(); request != nil {
		info.Parameters = request.URL.Query()
	}

	if config := ws.Config(); config != nil {
		info.Headers = config.Header
	}

	return &PushSession{
		id:        uuid.NewV4().String(),
		server:    server,
		socket:    ws,
		events:    make(chan *elemental.Event, 1024),
		requests:  make(chan *elemental.Request, 1024),
		stopRead:  make(chan bool, 2),
		stopWrite: make(chan bool, 2),
		stopAll:   make(chan bool, 2),
		Info:      info,
		sType:     sType,
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

	unsubscribe := s.server.config.Service.Subscribe(publications, errors, s.server.config.Topic)

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
				log.WithFields(log.Fields{"session": s, "message": message, "package": "bahamut"}).Error("Unable to decode event.")
				break
			}

			if s.server.config.SessionsHandler != nil {

				ok, err := s.server.config.SessionsHandler.ShouldPush(s, event)
				if err != nil {
					log.WithFields(log.Fields{"error": err, "package": "bahamut"}).Error("Error during checking authorization.")
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
			log.WithFields(log.Fields{
				"package": "bahamut",
				"error":   err.Error(),
			}).Error("Error during consuming pubsub topic.")

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
				go s.handleRetrieveManyOperation(request)

			case elemental.OperationRetrieve:
				go s.handleRetrieveOperation(request)

			case elemental.OperationCreate:
				go s.handleCreateOperation(request)

			case elemental.OperationUpdate:
				go s.handleUpdateOperation(request)

			case elemental.OperationDelete:
				go s.handleDeleteOperation(request)
			}

		case <-s.stopAll:
			s.stopRead <- true
			s.stopWrite <- true
			return
		}
	}
}

func (s *PushSession) String() string {

	return fmt.Sprintf("<session id:%s info: %s>",
		s.id,
		s.Info,
	)
}

func (s *PushSession) handleRetrieveManyOperation(request *elemental.Request) {

	proc, _ := s.server.processorFinder(request.Identity)

	response := elemental.NewResponse()
	response.Request = request

	ctx := NewContext(elemental.OperationRetrieveMany)
	ctx.ReadElementalRequest(request)

	if !CheckWebSocketAuthentication(s.server.config.Authenticator, ctx, response, s.socket) {
		return
	}

	if !CheckWebSocketAuthorization(s.server.config.Authorizer, ctx, response, s.socket) {
		return
	}

	if _, ok := proc.(RetrieveManyProcessor); !ok {
		writeWebSocketError(s.socket, response, elemental.NewError("Not implemented", "No handler for retrieving many "+request.Identity.Name, "bahamut", http.StatusNotImplemented))
		return
	}

	if err := proc.(RetrieveManyProcessor).ProcessRetrieveMany(ctx); err != nil {
		writeWebSocketError(s.socket, response, err)
		return
	}

	if err := ctx.WriteWebsocketResponse(s.socket, response); err != nil {
		writeWebSocketError(s.socket, response, elemental.NewError("Cannot Write Response", err.Error(), "bahamut", http.StatusInternalServerError))
		return
	}
}

func (s *PushSession) handleRetrieveOperation(request *elemental.Request) {

	proc, _ := s.server.processorFinder(request.Identity)

	response := elemental.NewResponse()
	response.Request = request

	ctx := NewContext(elemental.OperationRetrieve)
	ctx.ReadElementalRequest(request)

	if !CheckWebSocketAuthentication(s.server.config.Authenticator, ctx, response, s.socket) {
		return
	}

	if !CheckWebSocketAuthorization(s.server.config.Authorizer, ctx, response, s.socket) {
		return
	}

	if _, ok := proc.(RetrieveProcessor); !ok {
		writeWebSocketError(s.socket, response, elemental.NewError("Not implemented", "No handler for retrieving many "+request.Identity.Name, "bahamut", http.StatusNotImplemented))
		return
	}

	if err := proc.(RetrieveProcessor).ProcessRetrieve(ctx); err != nil {
		writeWebSocketError(s.socket, response, err)
		return
	}

	if err := ctx.WriteWebsocketResponse(s.socket, response); err != nil {
		writeWebSocketError(s.socket, response, elemental.NewError("Cannot Write Response", err.Error(), "bahamut", http.StatusInternalServerError))
		return
	}
}

func (s *PushSession) handleCreateOperation(request *elemental.Request) {

	proc, _ := s.server.processorFinder(request.Identity)

	response := elemental.NewResponse()
	response.Request = request

	ctx := NewContext(elemental.OperationCreate)
	ctx.ReadElementalRequest(request)

	if !CheckWebSocketAuthentication(s.server.config.Authenticator, ctx, response, s.socket) {
		return
	}

	if !CheckWebSocketAuthorization(s.server.config.Authorizer, ctx, response, s.socket) {
		return
	}

	if _, ok := proc.(CreateProcessor); !ok {
		writeWebSocketError(s.socket, response, elemental.NewError("Not implemented", "No handler for retrieving many "+request.Identity.Name, "bahamut", http.StatusNotImplemented))
		return
	}

	obj := s.server.config.IdentifiablesFactory(request.Identity.Name)

	if err := request.Decode(&obj); err != nil {
		writeWebSocketError(s.socket, response, err)
		return
	}

	if v, ok := obj.(elemental.Validatable); ok {
		if err := v.Validate(); err != nil {
			writeWebSocketError(s.socket, response, err)
			return
		}
	}

	ctx.InputData = obj

	if err := proc.(CreateProcessor).ProcessCreate(ctx); err != nil {
		writeWebSocketError(s.socket, response, err)
		return
	}

	if ctx.HasEvents() {
		s.server.pushEvents(ctx.Events()...)
	}

	if ctx.OutputData != nil {
		s.server.pushEvents(elemental.NewEvent(elemental.EventCreate, ctx.OutputData.(elemental.Identifiable)))
	}

	if err := ctx.WriteWebsocketResponse(s.socket, response); err != nil {
		writeWebSocketError(s.socket, response, elemental.NewError("Cannot Write Response", err.Error(), "bahamut", http.StatusInternalServerError))
		return
	}
}

func (s *PushSession) handleUpdateOperation(request *elemental.Request) {

	proc, _ := s.server.processorFinder(request.Identity)

	response := elemental.NewResponse()
	response.Request = request

	ctx := NewContext(elemental.OperationUpdate)
	ctx.ReadElementalRequest(request)

	if !CheckWebSocketAuthentication(s.server.config.Authenticator, ctx, response, s.socket) {
		return
	}

	if !CheckWebSocketAuthorization(s.server.config.Authorizer, ctx, response, s.socket) {
		return
	}

	if _, ok := proc.(UpdateProcessor); !ok {
		writeWebSocketError(s.socket, response, elemental.NewError("Not implemented", "No handler for retrieving many "+request.Identity.Name, "bahamut", http.StatusNotImplemented))
		return
	}

	obj := s.server.config.IdentifiablesFactory(request.Identity.Name).(elemental.Validatable)

	if err := request.Decode(&obj); err != nil {
		writeWebSocketError(s.socket, response, err)
		return
	}

	if v, ok := obj.(elemental.Validatable); ok {
		if err := v.Validate(); err != nil {
			writeWebSocketError(s.socket, response, err)
			return
		}
	}

	ctx.InputData = obj

	if err := proc.(UpdateProcessor).ProcessUpdate(ctx); err != nil {
		writeWebSocketError(s.socket, response, err)
		return
	}

	if ctx.HasEvents() {
		s.server.pushEvents(ctx.Events()...)
	}

	if ctx.OutputData != nil {
		s.server.pushEvents(elemental.NewEvent(elemental.EventUpdate, ctx.OutputData.(elemental.Identifiable)))
	}

	if err := ctx.WriteWebsocketResponse(s.socket, response); err != nil {
		writeWebSocketError(s.socket, response, elemental.NewError("Cannot Write Response", err.Error(), "bahamut", http.StatusInternalServerError))
		return
	}
}

func (s *PushSession) handleDeleteOperation(request *elemental.Request) {

	proc, _ := s.server.processorFinder(request.Identity)

	response := elemental.NewResponse()
	response.Request = request

	ctx := NewContext(elemental.OperationUpdate)
	ctx.ReadElementalRequest(request)

	if !CheckWebSocketAuthentication(s.server.config.Authenticator, ctx, response, s.socket) {
		return
	}

	if !CheckWebSocketAuthorization(s.server.config.Authorizer, ctx, response, s.socket) {
		return
	}

	if _, ok := proc.(DeleteProcessor); !ok {
		writeWebSocketError(s.socket, response, elemental.NewError("Not implemented", "No handler for retrieving many "+request.Identity.Name, "bahamut", http.StatusNotImplemented))
		return
	}

	if err := proc.(DeleteProcessor).ProcessDelete(ctx); err != nil {
		writeWebSocketError(s.socket, response, err)
		return
	}

	if ctx.HasEvents() {
		s.server.pushEvents(ctx.Events()...)
	}

	if ctx.OutputData != nil {
		s.server.pushEvents(elemental.NewEvent(elemental.EventDelete, ctx.OutputData.(elemental.Identifiable)))
	}

	if err := ctx.WriteWebsocketResponse(s.socket, response); err != nil {
		writeWebSocketError(s.socket, response, elemental.NewError("Cannot Write Response", err.Error(), "bahamut", http.StatusInternalServerError))
		return
	}
}
