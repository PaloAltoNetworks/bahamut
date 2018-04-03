// Author: Antoine Mercadal
// See LICENSE file for full LICENSE
// Copyright 2016 Aporeto.

package bahamut

import (
	"context"
	"net/http"
	"sync"

	"github.com/aporeto-inc/addedeffect/wsc"
	"github.com/aporeto-inc/elemental"
	"github.com/go-zoo/bone"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

type pushServer struct {
	sessions        map[string]*wsPushSession
	multiplexer     *bone.Mux
	config          Config
	processorFinder processorFinderFunc
	sessionsLock    *sync.Mutex
	mainContext     context.Context
}

func newPushServer(config Config, multiplexer *bone.Mux, processorFinder processorFinderFunc) *pushServer {

	srv := &pushServer{
		sessions:        map[string]*wsPushSession{},
		multiplexer:     multiplexer,
		config:          config,
		sessionsLock:    &sync.Mutex{},
		processorFinder: processorFinder,
	}

	// If push is not completely disabled and dispatching of event is not disabled, we install
	// the websocket routes.
	if !config.PushServer.Disabled && !config.PushServer.DispatchDisabled {
		srv.multiplexer.Get("/events", http.HandlerFunc(srv.handleRequest))
		zap.L().Debug("Websocket push handlers installed")
	}

	return srv
}

func (n *pushServer) registerSession(session *wsPushSession) {

	n.sessionsLock.Lock()
	if session.Identifier() == "" {
		n.sessionsLock.Unlock()
		panic("cannot register websocket session. empty identifier")
	}
	n.sessions[session.Identifier()] = session
	n.sessionsLock.Unlock()

	if handler := n.config.PushServer.DispatchHandler; handler != nil {
		handler.OnPushSessionStart(session)
	}
}

func (n *pushServer) unregisterSession(session *wsPushSession) {

	n.sessionsLock.Lock()
	if session.Identifier() == "" {
		n.sessionsLock.Unlock()
		panic("cannot unregister websocket session. empty identifier")
	}
	delete(n.sessions, session.Identifier())
	n.sessionsLock.Unlock()

	if handler := n.config.PushServer.DispatchHandler; handler != nil {
		handler.OnPushSessionStop(session)
	}
}

func (n *pushServer) authSession(session *wsPushSession) error {

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

func (n *pushServer) initPushSession(session *wsPushSession) error {

	if n.config.PushServer.DispatchHandler == nil {
		return nil
	}

	ok, err := n.config.PushServer.DispatchHandler.OnPushSessionInit(session)
	if err != nil {
		return elemental.NewError("Forbidden", err.Error(), "bahamut", http.StatusForbidden)
	}

	if !ok {
		return elemental.NewError("Forbidden", "You are not authorized to initiate a push session", "bahamut", http.StatusForbidden)
	}

	return nil
}

func (n *pushServer) pushEvents(events ...*elemental.Event) {

	// If we don't have a service or publication is explicitly disabled, we do nothing.
	if n.config.PushServer.Service == nil || n.config.PushServer.PublishDisabled {
		return
	}

	var err error

	for _, event := range events {

		if n.config.PushServer.PublishHandler != nil {
			var ok bool
			ok, err = n.config.PushServer.PublishHandler.ShouldPublish(event)
			if err != nil {
				zap.L().Error("Error while calling ShouldPublish", zap.Error(err))
				continue
			}

			if !ok {
				continue
			}
		}

		publication := NewPublication(n.config.PushServer.Topic)
		if err = publication.Encode(event); err != nil {
			zap.L().Error("Unable to encode event", zap.Error(err))
			break
		}

		for i := 0; i < 3; i++ {
			err = n.config.PushServer.Service.Publish(publication)
			if err != nil {
				zap.L().Warn("Unable to publish event", zap.String("topic", publication.Topic), zap.Stringer("event", event), zap.Error(err))
				continue
			}
			break
		}
	}
}

func (n *pushServer) handleRequest(w http.ResponseWriter, r *http.Request) {

	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}

	r = r.WithContext(n.mainContext)

	session := newWSPushSession(r, n.config, n.unregisterSession)
	if err := n.authSession(session); err != nil {
		writeHTTPResponse(w, makeErrorResponse(r.Context(), elemental.NewResponse(elemental.NewRequest()), err))
		return
	}

	if err := n.initPushSession(session); err != nil {
		writeHTTPResponse(w, makeErrorResponse(r.Context(), elemental.NewResponse(elemental.NewRequest()), err))
		return
	}

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		writeHTTPResponse(w, makeErrorResponse(r.Context(), elemental.NewResponse(elemental.NewRequest()), err))
		return
	}

	conn, err := wsc.Accept(r.Context(), ws, wsc.Config{})
	if err != nil {
		writeHTTPResponse(w, makeErrorResponse(r.Context(), elemental.NewResponse(elemental.NewRequest()), err))
		return
	}

	n.registerSession(session)
	session.setConn(conn)
	session.listen()
}

func (n *pushServer) start(ctx context.Context) {

	// If dispatching of events is disabled, we sit here
	// until the context is canceled.
	if n.config.PushServer.DispatchDisabled {
		<-ctx.Done()
		zap.L().Info("Push server stopped")
		return
	}

	n.mainContext = ctx
	defer func() { n.mainContext = nil }()

	publications := make(chan *Publication)
	if n.config.PushServer.Service != nil {
		errors := make(chan error)
		unsubscribe := n.config.PushServer.Service.Subscribe(publications, errors, n.config.PushServer.Topic)
		defer unsubscribe()
	}

	zap.L().Debug("Websocket server started",
		zap.Bool("push-enabled", !n.config.PushServer.Disabled),
		zap.Bool("push-dispatching-enabled", !n.config.PushServer.DispatchDisabled),
		zap.Bool("push-publish-enabled", !n.config.PushServer.PublishDisabled),
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
				for _, s := range n.sessions {
					sessions = append(sessions, s)
				}
				n.sessionsLock.Unlock()

				// Dispatch the event to all sessions
				for _, session := range sessions {

					go func(s PushSession, evt *elemental.Event) {

						if n.config.PushServer.DispatchHandler != nil {

							ok, err := n.config.PushServer.DispatchHandler.ShouldDispatch(s, evt)
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
				session.close(websocket.CloseGoingAway)
			}
			n.sessionsLock.Unlock()

			zap.L().Info("Push server stopped")
			return
		}
	}
}
