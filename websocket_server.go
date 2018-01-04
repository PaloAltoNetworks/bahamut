// Author: Antoine Mercadal
// See LICENSE file for full LICENSE
// Copyright 2016 Aporeto.

package bahamut

import (
	"net/http"
	"sync"

	"github.com/aporeto-inc/elemental"
	"github.com/go-zoo/bone"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

type websocketServer struct {
	sessions        map[string]internalWSSession
	close           chan struct{}
	multiplexer     *bone.Mux
	config          Config
	processorFinder processorFinderFunc
	sessionsLock    *sync.Mutex
}

func newWebsocketServer(config Config, multiplexer *bone.Mux, processorFinder processorFinderFunc) *websocketServer {

	srv := &websocketServer{
		sessions:        map[string]internalWSSession{},
		close:           make(chan struct{}, 1),
		multiplexer:     multiplexer,
		config:          config,
		sessionsLock:    &sync.Mutex{},
		processorFinder: processorFinder,
	}

	// TODO: this is also a hack for compat with
	// golang/x/net/websocket.
	var upgrader = websocket.Upgrader{
		ReadBufferSize:  32 << 20,
		WriteBufferSize: 32 << 20,
		// ReadBufferSize:  1024,
		// WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool { return true },
	}

	if !config.WebSocketServer.PushDisabled {
		srv.multiplexer.Handle("/events", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			request, err := elemental.NewRequestFromHTTPRequest(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			session := newWSPushSession(r, config, srv.unregisterSession)
			if err = srv.authSession(session); err != nil {
				writeHTTPError(w, request, err)
				return
			}

			if err = srv.initPushSession(session); err != nil {
				writeHTTPError(w, request, err)
				return
			}

			ws, err := upgrader.Upgrade(w, r, nil)
			if err != nil {
				writeHTTPError(w, request, err)
				return
			}

			srv.registerSession(session)
			session.setSocket(ws)
			session.listen()
		}))

		zap.L().Debug("Websocket push handlers installed")
	}

	if !config.WebSocketServer.APIDisabled {
		srv.multiplexer.Handle("/wsapi", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			session := newWSAPISession(r, config, srv.unregisterSession, srv.processorFinder, srv.pushEvents)
			if err := srv.authSession(session); err != nil {
				return
			}

			ws, err := upgrader.Upgrade(w, r, nil)
			if err != nil {
				return
			}

			// TODO: this is here for backward compat.
			// we should remvove this when all enforcers
			// are switched to at least manipulate 2.x
			ws.WriteJSON(&elemental.Response{ // nolint: errcheck
				StatusCode: http.StatusOK,
			})
			// END OF HACK

			srv.registerSession(session)
			session.setSocket(ws)
			session.listen()
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

func (n *websocketServer) authSession(session internalWSSession) error {

	if len(n.config.Security.SessionAuthenticators) == 0 {
		return nil
	}

	var action AuthAction
	var err error

	for _, authenticator := range n.config.Security.SessionAuthenticators {

		if action, err = authenticator.AuthenticateSession(session); action == AuthActionKO || err != nil {
			return elemental.NewError("Unauthorized", err.Error(), "bahamut", http.StatusUnauthorized)
		}

		if action == AuthActionOK {
			break
		}
	}

	return nil
}

func (n *websocketServer) initPushSession(session *wsPushSession) error {

	if n.config.WebSocketServer.SessionsHandler == nil {
		return nil
	}

	ok, err := n.config.WebSocketServer.SessionsHandler.OnPushSessionInit(session)
	if err != nil {
		return elemental.NewError("Forbidden", err.Error(), "bahamut", http.StatusForbidden)
	}

	if !ok {
		return elemental.NewError("Forbidden", "Rejected and refused to provide a reason", "bahamut", http.StatusForbidden)
	}

	return nil
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

	close(n.close)
}
