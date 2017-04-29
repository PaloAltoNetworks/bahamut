// Author: Antoine Mercadal
// See LICENSE file for full LICENSE
// Copyright 2016 Aporeto.

package bahamut

import (
	"fmt"
	"net/http"
	"sync"

	"go.uber.org/zap"

	"github.com/aporeto-inc/elemental"
	"github.com/go-zoo/bone"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/opentracing/opentracing-go/log"
	"golang.org/x/net/websocket"
)

type sessionTracer struct {
	span opentracing.Span
}

func newSessionTracer(session *Session) *sessionTracer {

	sp := opentracing.StartSpan(fmt.Sprintf("bahamut.session.authentication"))
	sp.SetTag("bahamut.session.id", session.Identifier())

	return &sessionTracer{
		span: sp,
	}
}

func (t *sessionTracer) Span() opentracing.Span {
	return t.span
}

func (t *sessionTracer) NewChildSpan(name string) opentracing.Span {

	return opentracing.StartSpan(name, opentracing.ChildOf(t.span.Context()))
}

type pushServer struct {
	address         string
	sessions        map[string]*Session
	close           chan bool
	multiplexer     *bone.Mux
	config          Config
	processorFinder processorFinder
	sessionsLock    *sync.Mutex
}

func newPushServer(config Config, multiplexer *bone.Mux) *pushServer {

	srv := &pushServer{
		sessions:     map[string]*Session{},
		close:        make(chan bool, 2),
		multiplexer:  multiplexer,
		config:       config,
		sessionsLock: &sync.Mutex{},
	}

	srv.multiplexer.Handle("/events", websocket.Handler(srv.handlePushConnection))
	srv.multiplexer.Handle("/wsapi", websocket.Handler(srv.handleAPIConnection))

	return srv
}

// adds a new push session to register in the push server
func (n *pushServer) registerSession(session *Session) {

	n.sessionsLock.Lock()
	defer n.sessionsLock.Unlock()

	n.sessions[session.id] = session

	if handler := n.config.WebSocketServer.SessionsHandler; handler != nil {
		handler.OnPushSessionStart(session)
	}
}

// adds a new push session to unregister from the push server
func (n *pushServer) unregisterSession(session *Session) {

	n.sessionsLock.Lock()
	defer n.sessionsLock.Unlock()

	delete(n.sessions, session.id)

	if handler := n.config.WebSocketServer.SessionsHandler; handler != nil {
		handler.OnPushSessionStop(session)
	}
}

// handlePushConnection handle connection for push events
func (n *pushServer) handlePushConnection(ws *websocket.Conn) {

	n.runSession(ws, newPushSession(ws, n.config, n.unregisterSession))
}

// handleAPIConnection handle connection for push events
func (n *pushServer) handleAPIConnection(ws *websocket.Conn) {

	n.runSession(ws, newAPISession(ws, n.config, n.unregisterSession, n.processorFinder, n.pushEvents))
}

func (n *pushServer) runSession(ws *websocket.Conn, session *Session) {

	if n.config.Security.SessionAuthenticator != nil {

		spanHolder := newSessionTracer(session)

		ok, err := n.config.Security.SessionAuthenticator.AuthenticateSession(session, spanHolder)
		if err != nil {
			ext.Error.Set(spanHolder.Span(), true)
			spanHolder.Span().LogFields(log.Error(err))
			spanHolder.Span().Finish()
		}

		if !ok {
			if session.sType == sessionTypeAPI {
				response := elemental.NewResponse()
				writeWebSocketError(ws, response, elemental.NewError("Unauthorized", "You are not authorized to access this api", "bahamut", http.StatusUnauthorized))
			}
			ws.Close() // nolint: errcheck
			spanHolder.Span().Finish()
			return
		}

		spanHolder.Span().Finish()
	}

	// Send the first hello message.
	if session.sType == sessionTypeAPI {
		response := elemental.NewResponse()
		response.StatusCode = http.StatusOK
		if err := websocket.JSON.Send(ws, response); err != nil {
			zap.L().Error("Error while sending hello message", zap.Error(err))
			return
		}
	}

	n.registerSession(session)
	session.listen()
}

// push a new event. If the global push system is available, it will be used.
// otherwise, only local sessions will receive the push
func (n *pushServer) pushEvents(events ...*elemental.Event) {

	if n.config.WebSocketServer.Service == nil {
		return
	}

	for _, event := range events {

		publication := NewPublication(n.config.WebSocketServer.Topic)
		if err := publication.Encode(event); err != nil {
			zap.L().Error("Unable to encode event. Message dropped", zap.Error(err))
			break
		}

		var err error
		for i := 0; i < 3; i++ {
			err = n.config.WebSocketServer.Service.Publish(publication)
			if err == nil {
				break
			}
		}

		if err != nil {
			zap.L().Warn("Unable to publish. Message dropped",
				zap.String("topic", publication.Topic),
				zap.Stringer("event", event),
				zap.Error(err),
			)
		}
	}
}

func (n *pushServer) closeAllSessions() {

	n.sessionsLock.Lock()
	for _, session := range n.sessions {
		session.close()
	}
	n.sessionsLock.Unlock()
}

// starts the push server
func (n *pushServer) start() {

	publications := make(chan *Publication)
	if n.config.WebSocketServer.Service != nil {
		errors := make(chan error)
		unsubscribe := n.config.WebSocketServer.Service.Subscribe(publications, errors, n.config.WebSocketServer.Topic)
		zap.L().Info("Subscribed to events")
		defer unsubscribe()
	}

	zap.L().Info("Event server started",
		zap.String("endpoint", n.address+"/events"),
	)

	for {
		select {

		case publication := <-publications:

			event := &elemental.Event{}
			if err := publication.Decode(event); err != nil {
				zap.L().Error("Unable to decode event", zap.Error(err))
				break
			}

			// Keep a references to all current ready push sessions as it may change at any time, we lost 8h on this one...
			n.sessionsLock.Lock()
			var sessions []*Session
			for _, session := range n.sessions {
				if session.sType == sessionTypeEvent {
					sessions = append(sessions, session)
				}
			}
			n.sessionsLock.Unlock()

			// Dispatch the event to all sessions
			for _, session := range sessions {

				go func(s *Session, evt *elemental.Event) {

					if n.config.WebSocketServer.SessionsHandler != nil {

						ok, err := s.config.WebSocketServer.SessionsHandler.ShouldPush(s, evt)
						if err != nil {
							zap.L().Error("Error while checking authorization", zap.Error(err))
							return
						}

						if !ok {
							return
						}
					}
					// we put back userInfo to nil before sending it clients.
					evt.UserInfo = nil
					s.DirectPush(evt)

				}(session, event.Duplicate())
			}

		case <-n.close:

			n.closeAllSessions()
			zap.L().Info("Push server stopped")
			return
		}
	}
}

// stops the push server
func (n *pushServer) stop() {

	n.close <- true
}
