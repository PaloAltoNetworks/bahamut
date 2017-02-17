// Author: Antoine Mercadal
// See LICENSE file for full LICENSE
// Copyright 2016 Aporeto.

package bahamut

import (
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"

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

	config          Config
	events          chan *elemental.Event
	id              string
	processorFinder processorFinder
	pushEventsFunc  func(...*elemental.Event)
	ready           bool
	readyLock       *sync.Mutex
	requests        chan *elemental.Request
	socket          *websocket.Conn
	startTime       time.Time
	stopAll         chan bool
	stopRead        chan bool
	stopWrite       chan bool
	sType           pushSessionType
	unregisterFunc  func(*PushSession)
}

func newPushSession(ws *websocket.Conn, config Config, unregisterFunc func(*PushSession)) *PushSession {

	return newSession(ws, pushSessionTypeEvent, config, unregisterFunc, nil, nil)
}

func newAPISession(ws *websocket.Conn, config Config, unregisterFunc func(*PushSession), processorFinder processorFinder, pushEventsFunc func(...*elemental.Event)) *PushSession {

	return newSession(ws, pushSessionTypeAPI, config, unregisterFunc, processorFinder, pushEventsFunc)
}

func newSession(ws *websocket.Conn, sType pushSessionType, config Config, unregisterFunc func(*PushSession), processorFinder processorFinder, pushEventsFunc func(...*elemental.Event)) *PushSession {

	var parameters url.Values
	var headers http.Header

	if request := ws.Request(); request != nil {
		parameters = request.URL.Query()
	}

	if config := ws.Config(); config != nil {
		headers = config.Header
	}

	return &PushSession{
		config:          config,
		events:          make(chan *elemental.Event),
		Headers:         headers,
		id:              uuid.NewV4().String(),
		Parameters:      parameters,
		processorFinder: processorFinder,
		pushEventsFunc:  pushEventsFunc,
		readyLock:       &sync.Mutex{},
		requests:        make(chan *elemental.Request),
		socket:          ws,
		startTime:       time.Now(),
		stopAll:         make(chan bool, 2),
		stopRead:        make(chan bool, 2),
		stopWrite:       make(chan bool, 2),
		sType:           sType,
		unregisterFunc:  unregisterFunc,
	}
}

// Identifier returns the identifier of the push session.
func (s *PushSession) Identifier() string {

	return s.id
}

func (s *PushSession) isReady() bool {

	s.readyLock.Lock()
	defer s.readyLock.Unlock()

	return s.ready
}

func (s *PushSession) setReady(ok bool) {
	s.readyLock.Lock()
	s.ready = ok
	s.readyLock.Unlock()
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
		case event := <-s.events:

			// is the event happened before the initial push session start time, we ignore.
			if event.Timestamp.Before(s.startTime) {
				break
			}

			if s.config.WebSocketServer.SessionsHandler != nil {

				ok, err := s.config.WebSocketServer.SessionsHandler.ShouldPush(s, event)
				if err != nil {
					log.WithError(err).Error("Error while checking authorization.")
					break
				}

				if !ok {
					break
				}
			}

			if err := websocket.JSON.Send(s.socket, event); err != nil {
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

	s.setReady(true)

	go s.read()
	go s.write()

	<-s.stopAll

	s.setReady(false)

	s.stopRead <- true
	s.stopWrite <- true

	s.unregisterFunc(s)
	s.socket.Close()
	s.processorFinder = nil
	s.pushEventsFunc = nil
	s.unregisterFunc = nil

}

func (s *PushSession) listenToAPIRequest() {

	s.setReady(true)

	go s.write()
	go s.read()

L:
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
			break L
		}
	}

	s.setReady(false)

	s.stopRead <- true
	s.stopWrite <- true

	s.unregisterFunc(s)
	s.socket.Close()
	s.processorFinder = nil
	s.pushEventsFunc = nil
	s.unregisterFunc = nil
}

func (s *PushSession) handleEventualPanic(response *elemental.Response) {

	if r := recover(); r != nil {
		writeWebSocketError(
			s.socket,
			response,
			elemental.NewError(
				"Internal Server Error",
				fmt.Sprintf("%v", r),
				"bahamut",
				http.StatusInternalServerError,
			),
		)
	}
}

func (s *PushSession) handleRetrieveMany(request *elemental.Request) {

	response := elemental.NewResponse()
	response.Request = request

	defer s.handleEventualPanic(response)

	ctx, err := dispatchRetrieveManyOperation(
		request,
		s.processorFinder,
		s.config.Model.IdentifiablesFactory,
		s.config.Security.Authenticator,
		s.config.Security.Authorizer,
		s.config.Security.Auditer,
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

	defer s.handleEventualPanic(response)

	ctx, err := dispatchRetrieveOperation(
		response.Request,
		s.processorFinder,
		s.config.Model.IdentifiablesFactory,
		s.config.Security.Authenticator,
		s.config.Security.Authorizer,
		s.config.Security.Auditer,
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

	defer s.handleEventualPanic(response)

	ctx, err := dispatchCreateOperation(
		response.Request,
		s.processorFinder,
		s.config.Model.IdentifiablesFactory,
		s.config.Security.Authenticator,
		s.config.Security.Authorizer,
		s.pushEventsFunc,
		s.config.Security.Auditer,
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

	defer s.handleEventualPanic(response)

	ctx, err := dispatchUpdateOperation(
		response.Request,
		s.processorFinder,
		s.config.Model.IdentifiablesFactory,
		s.config.Security.Authenticator,
		s.config.Security.Authorizer,
		s.pushEventsFunc,
		s.config.Security.Auditer,
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

	defer s.handleEventualPanic(response)

	ctx, err := dispatchDeleteOperation(
		response.Request,
		s.processorFinder,
		s.config.Model.IdentifiablesFactory,
		s.config.Security.Authenticator,
		s.config.Security.Authorizer,
		s.pushEventsFunc,
		s.config.Security.Auditer,
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

	defer s.handleEventualPanic(response)

	ctx, err := dispatchInfoOperation(
		response.Request,
		s.processorFinder,
		s.config.Model.IdentifiablesFactory,
		s.config.Security.Authenticator,
		s.config.Security.Authorizer,
		s.config.Security.Auditer,
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

	defer s.handleEventualPanic(response)

	ctx, err := dispatchPatchOperation(
		response.Request,
		s.processorFinder,
		s.config.Model.IdentifiablesFactory,
		s.config.Security.Authenticator,
		s.config.Security.Authorizer,
		s.pushEventsFunc,
		s.config.Security.Auditer,
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
