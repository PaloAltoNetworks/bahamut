// Author: Antoine Mercadal
// See LICENSE file for full LICENSE
// Copyright 2016 Aporeto.

package bahamut

import (
	"context"
	"net/http"
	"sync"

	"github.com/aporeto-inc/elemental"
	"github.com/go-zoo/bone"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

type websocketServer struct {
	sessions        map[string]internalWSSession
	multiplexer     *bone.Mux
	config          Config
	processorFinder processorFinderFunc
	sessionsLock    *sync.Mutex
	mainContext     context.Context
}

func newWebsocketServer(config Config, multiplexer *bone.Mux, processorFinder processorFinderFunc) *websocketServer {

	srv := &websocketServer{
		sessions:        map[string]internalWSSession{},
		multiplexer:     multiplexer,
		config:          config,
		sessionsLock:    &sync.Mutex{},
		processorFinder: processorFinder,
	}

	var upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}

	// If push is not completely disabled and dispatching of event is not disabled, we install
	// the websocket routes.
	if !config.WebSocketServer.PushDisabled && !config.WebSocketServer.PushDispatchDisabled {
		srv.multiplexer.Get("/events", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			r = r.WithContext(srv.mainContext)

			session := newWSPushSession(r, config, srv.unregisterSession)
			if err := srv.authSession(session); err != nil {
				writeHTTPResponse(w, makeErrorResponse(r.Context(), elemental.NewResponse(elemental.NewRequest()), err))
				return
			}

			if err := srv.initPushSession(session); err != nil {
				writeHTTPResponse(w, makeErrorResponse(r.Context(), elemental.NewResponse(elemental.NewRequest()), err))
				return
			}

			ws, err := upgrader.Upgrade(w, r, nil)
			if err != nil {
				writeHTTPResponse(w, makeErrorResponse(r.Context(), elemental.NewResponse(elemental.NewRequest()), err))
				return
			}

			srv.registerSession(session)
			session.setConn(ws)
			session.listen()
		}))

		zap.L().Debug("Websocket push handlers installed")
	}

	if !config.WebSocketServer.APIDisabled {
		srv.multiplexer.Get("/wsapi", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			session := newWSAPISession(r, config, srv.unregisterSession, srv.processorFinder, srv.pushEvents)
			if err := srv.authSession(session); err != nil {
				return
			}

			ws, err := upgrader.Upgrade(w, r, nil)
			if err != nil {
				return
			}

			srv.registerSession(session)
			session.setConn(ws)
			session.listen()
		}))

		zap.L().Debug("Websocket api handlers installed")
	}

	return srv
}

func (n *websocketServer) registerSession(session internalWSSession) {

	n.sessionsLock.Lock()
	if session.Identifier() == "" {
		n.sessionsLock.Unlock()
		panic("cannot register websocket session. empty identifier")
	}
	n.sessions[session.Identifier()] = session
	n.sessionsLock.Unlock()

	if handler := n.config.WebSocketServer.PushDispatchHandler; handler != nil {
		if s, ok := session.(PushSession); ok {
			handler.OnPushSessionStart(s)
		}
	}
}

func (n *websocketServer) unregisterSession(session internalWSSession) {

	n.sessionsLock.Lock()
	if session.Identifier() == "" {
		n.sessionsLock.Unlock()
		panic("cannot unregister websocket session. empty identifier")
	}
	delete(n.sessions, session.Identifier())
	n.sessionsLock.Unlock()

	if handler := n.config.WebSocketServer.PushDispatchHandler; handler != nil {
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

		action, err = authenticator.AuthenticateSession(session)
		if err != nil {
			return elemental.NewError("Unauthorized", err.Error(), "bahamut", http.StatusUnauthorized)
		}

		if action == AuthActionKO {
			return elemental.NewError("Unauthorized", "You are not authorized to start a session", "bahamut", http.StatusUnauthorized)
		}

		if action == AuthActionOK {
			break
		}
	}

	return nil
}

func (n *websocketServer) initPushSession(session *wsPushSession) error {

	if n.config.WebSocketServer.PushDispatchHandler == nil {
		return nil
	}

	ok, err := n.config.WebSocketServer.PushDispatchHandler.OnPushSessionInit(session)
	if err != nil {
		return elemental.NewError("Forbidden", err.Error(), "bahamut", http.StatusForbidden)
	}

	if !ok {
		return elemental.NewError("Forbidden", "You are not authorized to initiate a push session", "bahamut", http.StatusForbidden)
	}

	return nil
}

func (n *websocketServer) pushEvents(events ...*elemental.Event) {

	// If we don't have a service or publication is explicitly disabled, we do nothing.
	if n.config.WebSocketServer.Service == nil || n.config.WebSocketServer.PushPublishDisabled {
		return
	}

	var err error

	for _, event := range events {

		if n.config.WebSocketServer.PushPublishHandler != nil {
			var ok bool
			ok, err = n.config.WebSocketServer.PushPublishHandler.ShouldPublish(event)
			if err != nil {
				zap.L().Error("Error while calling ShouldPublish", zap.Error(err))
				continue
			}

			if !ok {
				continue
			}
		}

		publication := NewPublication(n.config.WebSocketServer.Topic)
		if err = publication.Encode(event); err != nil {
			zap.L().Error("Unable to encode event", zap.Error(err))
			break
		}

		for i := 0; i < 3; i++ {
			err = n.config.WebSocketServer.Service.Publish(publication)
			if err != nil {
				zap.L().Warn("Unable to publish event", zap.String("topic", publication.Topic), zap.Stringer("event", event), zap.Error(err))
				continue
			}
			break
		}
	}
}

func (n *websocketServer) start(ctx context.Context) {

	// If dispatching of events is disabled, we sit here
	// until the context is canceled.
	if n.config.WebSocketServer.PushDispatchDisabled {
		<-ctx.Done()
		zap.L().Info("Push server stopped")
		return
	}

	n.mainContext = ctx
	defer func() { n.mainContext = nil }()

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

		case p := <-publications:

			go func(publication *Publication) {

				event := &elemental.Event{}
				if err := publication.Decode(event); err != nil {
					zap.L().Error("Unable to decode event", zap.Error(err))
					return
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

						if n.config.WebSocketServer.PushDispatchHandler != nil {

							ok, err := n.config.WebSocketServer.PushDispatchHandler.ShouldDispatch(s, evt)
							if err != nil {
								zap.L().Error("Error while calling SessionsHandler ShouldPush", zap.Error(err))
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
			}(p)

		case <-ctx.Done():

			n.sessionsLock.Lock()
			for _, session := range n.sessions {
				session.stop()
			}
			n.sessionsLock.Unlock()

			zap.L().Info("Push server stopped")
			return
		}
	}
}
