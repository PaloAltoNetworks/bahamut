// Author: Antoine Mercadal
// See LICENSE file for full LICENSE
// Copyright 2016 Aporeto.

package bahamut

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/go-zoo/bone"
	"github.com/gorilla/websocket"
	"go.aporeto.io/elemental"
	"go.aporeto.io/wsc"
	"go.uber.org/zap"
)

type pushServer struct {
	sessions        map[string]*wsPushSession
	multiplexer     *bone.Mux
	cfg             config
	processorFinder processorFinderFunc
	sessionsLock    sync.RWMutex
	mainContext     context.Context
}

func newPushServer(cfg config, multiplexer *bone.Mux, processorFinder processorFinderFunc) *pushServer {

	srv := &pushServer{
		sessions:        map[string]*wsPushSession{},
		multiplexer:     multiplexer,
		cfg:             cfg,
		sessionsLock:    sync.RWMutex{},
		processorFinder: processorFinder,
	}

	endpoint := cfg.pushServer.endpoint
	if endpoint == "" {
		endpoint = "/events"
	}

	// If push is not completely disabled and dispatching of event is not disabled, we install
	// the websocket routes.
	if cfg.pushServer.enabled && cfg.pushServer.dispatchEnabled {
		srv.multiplexer.Get(endpoint, http.HandlerFunc(srv.handleRequest))
		zap.L().Debug("Websocket push handlers installed")
	}

	return srv
}

func (n *pushServer) registerSession(session *wsPushSession) {

	if n.cfg.healthServer.metricsManager != nil {
		n.cfg.healthServer.metricsManager.RegisterWSConnection()
	}

	if session.Identifier() == "" {
		panic("cannot register websocket session. empty identifier")
	}

	n.sessionsLock.Lock()
	n.sessions[session.Identifier()] = session
	n.sessionsLock.Unlock()

	if handler := n.cfg.pushServer.dispatchHandler; handler != nil {
		handler.OnPushSessionStart(session)
	}
}

func (n *pushServer) unregisterSession(session *wsPushSession) {

	if handler := n.cfg.pushServer.dispatchHandler; handler != nil {
		handler.OnPushSessionStop(session)
	}

	if session.Identifier() == "" {
		panic("cannot unregister websocket session. empty identifier")
	}

	n.sessionsLock.Lock()
	delete(n.sessions, session.Identifier())
	n.sessionsLock.Unlock()

	if n.cfg.healthServer.metricsManager != nil {
		n.cfg.healthServer.metricsManager.UnregisterWSConnection()
	}
}

func (n *pushServer) authSession(session *wsPushSession) error {

	if len(n.cfg.security.sessionAuthenticators) == 0 {
		return nil
	}

	var action AuthAction
	var err error

	for _, authenticator := range n.cfg.security.sessionAuthenticators {

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

	if n.cfg.pushServer.dispatchHandler == nil {
		return nil
	}

	ok, err := n.cfg.pushServer.dispatchHandler.OnPushSessionInit(session)
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
	if n.cfg.pushServer.service == nil || !n.cfg.pushServer.enabled {
		return
	}

	var err error

	for _, event := range events {

		if n.cfg.pushServer.publishHandler != nil {
			var ok bool
			ok, err = n.cfg.pushServer.publishHandler.ShouldPublish(event)
			if err != nil {
				zap.L().Error("Error while calling ShouldPublish", zap.Error(err))
				continue
			}

			if !ok {
				continue
			}
		}

		publication := NewPublication(n.cfg.pushServer.topic)
		if err = publication.Encode(event); err != nil {
			zap.L().Error("Unable to encode event", zap.Error(err))
			break
		}

		for i := 0; i < 3; i++ {
			err = n.cfg.pushServer.service.Publish(publication)
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

	session := newWSPushSession(r, n.cfg, n.unregisterSession)
	session.setTLSConnectionState(r.TLS)
	session.setRemoteAddress(r.RemoteAddr)

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

	conn, err := wsc.Accept(r.Context(), ws, wsc.Config{WriteChanSize: 1024, ReadChanSize: 512})
	if err != nil {
		writeHTTPResponse(w, makeErrorResponse(r.Context(), elemental.NewResponse(elemental.NewRequest()), err))
		return
	}

	session.setConn(conn)

	n.registerSession(session)

	session.listen()
}

func (n *pushServer) start(ctx context.Context) {

	// If dispatching of events is disabled, we sit here
	// until the context is canceled.
	if !n.cfg.pushServer.enabled {
		<-ctx.Done()
		return
	}

	n.mainContext = ctx

	publications := make(chan *Publication, 1000)
	if n.cfg.pushServer.service != nil {
		errors := make(chan error, 1000)
		unsubscribe := n.cfg.pushServer.service.Subscribe(publications, errors, n.cfg.pushServer.topic)
		defer unsubscribe()
	}

	zap.L().Debug("Websocket server started",
		zap.Bool("push-enabled", n.cfg.pushServer.enabled),
		zap.Bool("push-dispatching-enabled", n.cfg.pushServer.dispatchEnabled),
		zap.Bool("push-publish-enabled", n.cfg.pushServer.publishEnabled),
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
				n.sessionsLock.RLock()
				sessions := make([]PushSession, len(n.sessions))
				var i int
				for _, s := range n.sessions {
					sessions[i] = s
					i++
				}
				n.sessionsLock.RUnlock()

				// Dispatch the event to all sessions
				for _, session := range sessions {

					go func(s PushSession, evt *elemental.Event) {

						if n.cfg.pushServer.dispatchHandler != nil {

							ok, err := n.cfg.pushServer.dispatchHandler.ShouldDispatch(s, evt)
							if err != nil {
								zap.L().Error("Error while calling SessionsHandler ShouldPush", zap.Error(err))
								return
							}

							if !ok {
								return
							}
						}

						s.DirectPush(evt)

					}(session, event.Duplicate())
				}
			}(p)

		case <-ctx.Done():
			return
		}
	}
}

func (n *pushServer) stop() {

	// we wait for all session to get cleanly terminated.
	for {
		n.sessionsLock.RLock()
		leftOvers := len(n.sessions)
		n.sessionsLock.RUnlock()

		if leftOvers == 0 {
			break
		}

		time.Sleep(10 * time.Millisecond)
	}

	zap.L().Info("Push server stopped")
}
