// Author: Antoine Mercadal
// See LICENSE file for full LICENSE
// Copyright 2016 Aporeto.

package bahamut

import (
	"net/http"
	"sync"

	"go.uber.org/zap"

	"github.com/aporeto-inc/elemental"
	"github.com/go-zoo/bone"
	"golang.org/x/net/websocket"
)

type websocketServer struct {
	sessions        map[string]internalWSSession
	close           chan bool
	multiplexer     *bone.Mux
	config          Config
	processorFinder processorFinderFunc
	sessionsLock    *sync.Mutex
}

func newWebsocketServer(config Config, multiplexer *bone.Mux, processorFinder processorFinderFunc) *websocketServer {

	srv := &websocketServer{
		sessions:        map[string]internalWSSession{},
		close:           make(chan bool, 2),
		multiplexer:     multiplexer,
		config:          config,
		sessionsLock:    &sync.Mutex{},
		processorFinder: processorFinder,
	}

	if !config.WebSocketServer.PushDisabled {
		srv.multiplexer.Handle("/events", websocket.Handler(func(ws *websocket.Conn) {
			srv.handleSession(ws, newWSPushSession(ws, srv.config, srv.unregisterSession))
		}))
		zap.L().Debug("Websocket push handlers installed")
	}

	if !config.WebSocketServer.APIDisabled {
		srv.multiplexer.Handle("/wsapi", websocket.Handler(func(ws *websocket.Conn) {
			srv.handleSession(ws, newWSAPISession(ws, srv.config, srv.unregisterSession, srv.processorFinder, srv.pushEvents))
		}))
		zap.L().Debug("Websocket api handlers installed")
	}

	return srv
}

func (n *websocketServer) registerSession(session internalWSSession) {

	n.sessionsLock.Lock()
	n.sessions[session.Identifier()] = session
	n.sessionsLock.Unlock()

	if handler := n.config.WebSocketServer.SessionsHandler; handler != nil {
		if s, ok := session.(PushSession); ok {
			handler.OnPushSessionStart(s)
		}
	}
}

func (n *websocketServer) unregisterSession(session internalWSSession) {

	n.sessionsLock.Lock()
	delete(n.sessions, session.Identifier())
	n.sessionsLock.Unlock()

	if handler := n.config.WebSocketServer.SessionsHandler; handler != nil {
		if s, ok := session.(PushSession); ok {
			handler.OnPushSessionStop(s)
		}
	}
}

func (n *websocketServer) handleSession(ws *websocket.Conn, session internalWSSession) {

	session.setRemoteAddress(ws.Request().RemoteAddr)
	session.setTLSConnectionState(ws.Request().TLS)

	if len(n.config.Security.SessionAuthenticators) != 0 {

		var action AuthAction
		var err error
		for _, authenticator := range n.config.Security.SessionAuthenticators {

			if action, err = authenticator.AuthenticateSession(session); action == AuthActionKO || err != nil {
				ws.Close() // nolint: errcheck
				return
			}

			if action == AuthActionOK {
				break
			}
		}
	}

	switch s := session.(type) {
	case *wsAPISession:
		response := elemental.NewResponse()
		response.StatusCode = http.StatusOK
		if err := websocket.JSON.Send(ws, response); err != nil {
			ws.Close() // nolint: errcheck
			return
		}

	case *wsPushSession:
		if handler := n.config.WebSocketServer.SessionsHandler; handler != nil {
			if ok, err := handler.OnPushSessionInit(s); !ok || err != nil {
				ws.Close() // nolint: errcheck
				return
			}
		}
	}

	n.registerSession(session)
	session.listen()
}

func (n *websocketServer) pushEvents(events ...*elemental.Event) {

	if n.config.WebSocketServer.Service == nil {
		return
	}

	for _, event := range events {

		ok, err := n.config.WebSocketServer.SessionsHandler.ShouldPublish(event)
		if err != nil {
			zap.L().Error("Error while calling ShouldPublish. Event will not be published.", zap.Error(err))
		}

		if !ok {
			continue
		}

		publication := NewPublication(n.config.WebSocketServer.Topic)
		if err = publication.Encode(event); err != nil {
			zap.L().Error("Unable to encode event. Message dropped", zap.Error(err))
			break
		}

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

func (n *websocketServer) start() {

	publications := make(chan *Publication)
	if n.config.WebSocketServer.Service != nil {
		errors := make(chan error)
		unsubscribe := n.config.WebSocketServer.Service.Subscribe(publications, errors, n.config.WebSocketServer.Topic)
		defer unsubscribe()
	}

	zap.L().Debug("Websocket server started",
		zap.Bool("api-enabled", !n.config.WebSocketServer.APIDisabled),
		zap.Bool("push-enabled", !n.config.WebSocketServer.PushDisabled),
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
			var sessions []PushSession
			for _, session := range n.sessions {
				if s, ok := session.(PushSession); ok {
					sessions = append(sessions, s)
				}
			}
			n.sessionsLock.Unlock()

			// Dispatch the event to all sessions
			for _, session := range sessions {

				go func(s PushSession, evt *elemental.Event) {

					if n.config.WebSocketServer.SessionsHandler != nil {

						ok, err := n.config.WebSocketServer.SessionsHandler.ShouldPush(s, evt)
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

			n.sessionsLock.Lock()
			for _, session := range n.sessions {
				session.close()
			}
			n.sessionsLock.Unlock()

			zap.L().Info("Push server stopped")
			return
		}
	}
}

func (n *websocketServer) stop() {

	n.close <- true
}
