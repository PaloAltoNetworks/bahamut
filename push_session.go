// Author: Antoine Mercadal
// See LICENSE file for full LICENSE
// Copyright 2016 Aporeto.

package bahamut

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"runtime/debug"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/aporeto-inc/elemental"
	"github.com/satori/go.uuid"
	"golang.org/x/net/websocket"
)

type sessionType int

const (
	sessionTypeEvent sessionType = iota + 1
	sessionTypeAPI
)

// Session represents a client session.
type Session struct {
	Parameters url.Values
	Headers    http.Header

	claims            []string
	config            Config
	events            chan *elemental.Event
	id                string
	processorFinder   processorFinder
	pushEventsFunc    func(...*elemental.Event)
	requests          chan *elemental.Request
	filters           chan *elemental.PushFilter
	socket            *websocket.Conn
	startTime         time.Time
	stopAll           chan bool
	stopRead          chan bool
	stopWrite         chan bool
	sType             sessionType
	unregisterFunc    func(*Session)
	filter            *elemental.PushFilter
	currentFilterLock *sync.Mutex
}

func newPushSession(ws *websocket.Conn, config Config, unregisterFunc func(*Session)) *Session {

	return newSession(ws, sessionTypeEvent, config, unregisterFunc, nil, nil)
}

func newAPISession(ws *websocket.Conn, config Config, unregisterFunc func(*Session), processorFinder processorFinder, pushEventsFunc func(...*elemental.Event)) *Session {

	return newSession(ws, sessionTypeAPI, config, unregisterFunc, processorFinder, pushEventsFunc)
}

func newSession(ws *websocket.Conn, sType sessionType, config Config, unregisterFunc func(*Session), processorFinder processorFinder, pushEventsFunc func(...*elemental.Event)) *Session {

	var parameters url.Values
	var headers http.Header

	if request := ws.Request(); request != nil {
		parameters = request.URL.Query()
	}

	if config := ws.Config(); config != nil {
		headers = config.Header
	}

	return &Session{
		config:            config,
		claims:            []string{},
		events:            make(chan *elemental.Event),
		Headers:           headers,
		id:                uuid.NewV4().String(),
		Parameters:        parameters,
		processorFinder:   processorFinder,
		pushEventsFunc:    pushEventsFunc,
		requests:          make(chan *elemental.Request, 8),
		filters:           make(chan *elemental.PushFilter, 8),
		currentFilterLock: &sync.Mutex{},
		socket:            ws,
		startTime:         time.Now(),
		stopAll:           make(chan bool, 2),
		stopRead:          make(chan bool, 2),
		stopWrite:         make(chan bool, 2),
		sType:             sType,
		unregisterFunc:    unregisterFunc,
	}
}

// Identifier returns the identifier of the push session.
func (s *Session) Identifier() string {

	return s.id
}

// SetClaims implements elemental.ClaimsHolder.
func (s *Session) SetClaims(claims []string) { s.claims = claims }

// GetClaims implements elemental.ClaimsHolder.
func (s *Session) GetClaims() []string { return s.claims }

// GetToken implements elemental.TokenHolder.
func (s *Session) GetToken() string { return s.Parameters.Get("token") }

// DirectPush will send given events to the session without any further control
// but ensuring the events did not happen before the session has been initialized.
// the ShouldPush method of the eventual bahamut.PushHandler will *not* be called.
//
// For performance reason, this method will *not* check that it is an session of type
// Event. If you direct push to an API session, you will fill up the internal channels until
// it blocks.
//
// This method should be used only if you know what you are doing, and you should not need it
// in the vast majority of all cases.
func (s *Session) DirectPush(events ...*elemental.Event) {

	for _, event := range events {

		if event.Timestamp.Before(s.startTime) {
			continue
		}

		f := s.currentFilter()
		if f != nil && f.IsFilteredOut(event.Identity, event.Type) {
			break
		}

		s.events <- event
	}

}

func (s *Session) readRequests() {

	for {
		var request *elemental.Request

		if err := websocket.JSON.Receive(s.socket, &request); err != nil {
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

func (s *Session) readFilters() {

	for {
		var filter *elemental.PushFilter

		if err := websocket.JSON.Receive(s.socket, &filter); err != nil {
			s.stopAll <- true
			return
		}

		select {
		case s.filters <- filter:
		case <-s.stopRead:
			return
		}
	}
}

func (s *Session) write() {

	for {
		select {
		case event := <-s.events:

			if err := websocket.JSON.Send(s.socket, event); err != nil {
				s.stopAll <- true
				return
			}

		case <-s.stopWrite:
			return
		}
	}
}

func (s *Session) close() {

	s.stopAll <- true
}

func (s *Session) listen() {

	switch s.sType {
	case sessionTypeAPI:
		s.listenToAPIRequest()
	case sessionTypeEvent:
		s.listenToPushEvents()
	default:
		panic("Unknown push session type")
	}
}

func (s *Session) currentFilter() *elemental.PushFilter {

	s.currentFilterLock.Lock()
	defer s.currentFilterLock.Unlock()

	if s.filter == nil {
		return nil
	}

	return s.filter.Duplicate()
}

func (s *Session) setCurrentFilter(f *elemental.PushFilter) {

	s.currentFilterLock.Lock()
	s.filter = f
	s.currentFilterLock.Unlock()
}

func (s *Session) listenToPushEvents() {

	go s.readFilters()
	go s.write()

	defer func() {
		s.stopRead <- true
		s.stopWrite <- true

		s.unregisterFunc(s)
		s.socket.Close() // nolint: errcheck
		s.processorFinder = nil
		s.pushEventsFunc = nil
		s.unregisterFunc = nil
	}()

	for {
		select {
		case filter := <-s.filters:
			s.setCurrentFilter(filter)

		case <-s.stopAll:
			return
		}
	}
}

func (s *Session) listenToAPIRequest() {

	go s.write()
	go s.readRequests()

	defer func() {
		s.stopRead <- true
		s.stopWrite <- true

		s.unregisterFunc(s)
		s.socket.Close() // nolint: errcheck
		s.processorFinder = nil
		s.pushEventsFunc = nil
		s.unregisterFunc = nil
	}()

	for {
		select {
		case request := <-s.requests:

			// We backport the token of the session into the request if we don't have an explicit one given in the request.
			if request.Password == "" {
				if t := s.Parameters.Get("token"); t != "" {
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

func (s *Session) handleEventualPanic(response *elemental.Response) {

	if r := recover(); r != nil {
		err := elemental.NewError(
			"Internal Server Error",
			fmt.Sprintf("%v", r),
			"bahamut",
			http.StatusInternalServerError,
		)

		st := string(debug.Stack())
		err.Data = st
		zap.L().Error("panic", zap.String("stacktrace", st), zap.Stringer("request", response.Request))

		writeWebSocketError(s.socket, response, err)
	}
}

func (s *Session) handleRetrieveMany(request *elemental.Request) {

	response := elemental.NewResponse()
	response.Request = request

	defer s.handleEventualPanic(response)

	parentIdentity := request.ParentIdentity
	if parentIdentity.IsEmpty() {
		parentIdentity = elemental.RootIdentity
	}

	if !elemental.IsRetrieveManyAllowed(s.config.Model.RelationshipsRegistry, request.Identity, parentIdentity) {
		writeWebSocketError(s.socket, response, elemental.NewError("Not allowed", "RetrieveMany operation not allowed on "+request.Identity.Category, "bahamut", http.StatusMethodNotAllowed))
		return
	}

	ctx, err := dispatchRetrieveManyOperation(
		request,
		s.processorFinder,
		s.config.Model.IdentifiablesFactory,
		s.config.Security.RequestAuthenticator,
		s.config.Security.Authorizer,
		s.pushEventsFunc,
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

func (s *Session) handleRetrieve(request *elemental.Request) {

	response := elemental.NewResponse()
	response.Request = request

	defer s.handleEventualPanic(response)

	if !elemental.IsRetrieveAllowed(s.config.Model.RelationshipsRegistry, request.Identity) || !request.ParentIdentity.IsEmpty() {
		writeWebSocketError(s.socket, response, elemental.NewError("Not allowed", "Retrieve operation not allowed on "+request.Identity.Name, "bahamut", http.StatusMethodNotAllowed))
		return
	}

	ctx, err := dispatchRetrieveOperation(
		response.Request,
		s.processorFinder,
		s.config.Model.IdentifiablesFactory,
		s.config.Security.RequestAuthenticator,
		s.config.Security.Authorizer,
		s.pushEventsFunc,
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

func (s *Session) handleCreate(request *elemental.Request) {

	response := elemental.NewResponse()
	response.Request = request

	defer s.handleEventualPanic(response)

	parentIdentity := request.ParentIdentity
	if parentIdentity.IsEmpty() {
		parentIdentity = elemental.RootIdentity
	}

	if !elemental.IsCreateAllowed(s.config.Model.RelationshipsRegistry, request.Identity, parentIdentity) {
		writeWebSocketError(s.socket, response, elemental.NewError("Not allowed", "Create operation not allowed on "+request.Identity.Name, "bahamut", http.StatusMethodNotAllowed))
		return
	}

	ctx, err := dispatchCreateOperation(
		response.Request,
		s.processorFinder,
		s.config.Model.IdentifiablesFactory,
		s.config.Security.RequestAuthenticator,
		s.config.Security.Authorizer,
		s.pushEventsFunc,
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

func (s *Session) handleUpdate(request *elemental.Request) {

	response := elemental.NewResponse()
	response.Request = request

	defer s.handleEventualPanic(response)

	if !elemental.IsUpdateAllowed(s.config.Model.RelationshipsRegistry, request.Identity) || !request.ParentIdentity.IsEmpty() {
		writeWebSocketError(s.socket, response, elemental.NewError("Not allowed", "Update operation not allowed on "+request.Identity.Name, "bahamut", http.StatusMethodNotAllowed))
		return
	}

	ctx, err := dispatchUpdateOperation(
		response.Request,
		s.processorFinder,
		s.config.Model.IdentifiablesFactory,
		s.config.Security.RequestAuthenticator,
		s.config.Security.Authorizer,
		s.pushEventsFunc,
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

func (s *Session) handleDelete(request *elemental.Request) {

	response := elemental.NewResponse()
	response.Request = request

	defer s.handleEventualPanic(response)

	if !elemental.IsDeleteAllowed(s.config.Model.RelationshipsRegistry, request.Identity) || !request.ParentIdentity.IsEmpty() {
		writeWebSocketError(s.socket, response, elemental.NewError("Not allowed", "Delete operation not allowed on "+request.Identity.Name, "bahamut", http.StatusMethodNotAllowed))
		return
	}

	ctx, err := dispatchDeleteOperation(
		response.Request,
		s.processorFinder,
		s.config.Model.IdentifiablesFactory,
		s.config.Security.RequestAuthenticator,
		s.config.Security.Authorizer,
		s.pushEventsFunc,
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

func (s *Session) handleInfo(request *elemental.Request) {

	response := elemental.NewResponse()
	response.Request = request

	defer s.handleEventualPanic(response)

	parentIdentity := request.ParentIdentity
	if parentIdentity.IsEmpty() {
		parentIdentity = elemental.RootIdentity
	}

	if !elemental.IsInfoAllowed(s.config.Model.RelationshipsRegistry, request.Identity, parentIdentity) {
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

func (s *Session) handlePatch(request *elemental.Request) {

	response := elemental.NewResponse()
	response.Request = request

	defer s.handleEventualPanic(response)

	parentIdentity := request.ParentIdentity
	if parentIdentity.IsEmpty() {
		parentIdentity = elemental.RootIdentity
	}

	if !elemental.IsPatchAllowed(s.config.Model.RelationshipsRegistry, request.Identity, parentIdentity) {
		writeWebSocketError(s.socket, response, elemental.NewError("Not allowed", "Patch operation not allowed on "+request.Identity.Name, "bahamut", http.StatusMethodNotAllowed))
		return
	}

	ctx, err := dispatchPatchOperation(
		response.Request,
		s.processorFinder,
		s.config.Model.IdentifiablesFactory,
		s.config.Security.RequestAuthenticator,
		s.config.Security.Authorizer,
		s.pushEventsFunc,
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

func (s *Session) String() string {

	return fmt.Sprintf("<session id:%s headers: %v parameters: %v>",
		s.id,
		s.Headers,
		s.Parameters,
	)
}
