// Copyright 2019 Aporeto Inc.
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//     http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package bahamut

import (
	"context"
	"fmt"
	"net/http"
	"strings"
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
	publications    chan *Publication
}

func newPushServer(cfg config, multiplexer *bone.Mux, processorFinder processorFinderFunc) *pushServer {

	srv := &pushServer{
		sessions:        map[string]*wsPushSession{},
		multiplexer:     multiplexer,
		cfg:             cfg,
		sessionsLock:    sync.RWMutex{},
		processorFinder: processorFinder,
		publications:    make(chan *Publication, 24000),
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

	session.cancel()

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

		// if subject hierarchies are enabled, for the benefit of subscribers interested in specific identities, we utilize
		// a subject hierarchy here to publish to a specific subject under the configured topic of the push server.
		//
		// for example:
		//
		//   if the push server topic has been set to "global-events" and the server is about to push a "create" event w/ an identity
		//   value of "apples", enabling this option, would cause the push server to target a new publication to the subject
		//   "global-events.apples.create", INSTEAD OF "global-events".
		//
		//   consequently, clients interested in receiving events pertaining to the "apples" resource can then subscribe
		//   on that specific topic, as opposed to ignoring events they don't care about. For clients interested in receiving
		//   ALL events published to "global-events", they can utilize NATS wildcards and subscribe to "global-events.>"
		//   ('>' targets all hierarchies) or "global-events.*" ('*' matching a single token).
		//
		// more details: https://docs.nats.io/nats-concepts/subjects#subject-hierarchies
		topic := n.cfg.pushServer.topic
		if n.cfg.pushServer.subjectHierarchiesEnabled {
			topic = fmt.Sprintf("%s.%s.%s", topic, event.Identity, event.Type)
		}

		publication := NewPublication(topic)
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

	readEncodingType, writeEncodingType, err := elemental.EncodingFromHeaders(r.Header)
	if err != nil {
		writeHTTPResponse(
			w,
			makeErrorResponse(
				r.Context(),
				elemental.NewResponse(elemental.NewRequest()),
				err,
				nil,
				nil,
			),
			r.Header.Get("origin"),
			n.cfg.security.accessControl,
		)
	}

	session := newWSPushSession(r, n.cfg, n.unregisterSession, readEncodingType, writeEncodingType)
	session.setTLSConnectionState(r.TLS)

	var clientIP string
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		clientIP = ip
	} else if ip := r.Header.Get("X-Real-IP"); ip != "" {
		clientIP = ip
	} else {
		clientIP = r.RemoteAddr
	}
	session.setRemoteAddress(clientIP)
	session.cookies = r.Cookies()

	if err := n.authSession(session); err != nil {
		writeHTTPResponse(
			w,
			makeErrorResponse(
				r.Context(),
				elemental.NewResponse(elemental.NewRequest()),
				err,
				nil,
				nil,
			),
			r.Header.Get("origin"),
			n.cfg.security.accessControl,
		)
		return
	}

	if err := n.initPushSession(session); err != nil {
		writeHTTPResponse(
			w,
			makeErrorResponse(
				r.Context(),
				elemental.NewResponse(elemental.NewRequest()),
				err,
				nil,
				nil,
			),
			r.Header.Get("origin"),
			n.cfg.security.accessControl,
		)
		return
	}

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		writeHTTPResponse(
			w,
			makeErrorResponse(
				r.Context(),
				elemental.NewResponse(elemental.NewRequest()),
				err,
				nil,
				nil,
			),
			r.Header.Get("origin"),
			n.cfg.security.accessControl,
		)
		return
	}

	conn, err := wsc.Accept(r.Context(), ws, wsc.Config{WriteChanSize: 64, ReadChanSize: 16})
	if err != nil {
		writeHTTPResponse(
			w,
			makeErrorResponse(
				r.Context(),
				elemental.NewResponse(elemental.NewRequest()),
				err,
				nil,
				nil,
			),
			r.Header.Get("origin"),
			n.cfg.security.accessControl,
		)
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

	if n.cfg.pushServer.service != nil {
		errors := make(chan error, 24000)
		subTopic := n.cfg.pushServer.topic
		// backwards compatibility: if the push server is using subject hierarchies when publishing events, we must by default
		// listen to all child subjects of the configured topic via a wildcard '>'.
		//
		// see: https://docs.nats.io/nats-concepts/subjects#wildcards for more details.
		//
		// TODO: in the future, support to subscribing to specific subjects and/or wildcards may be added.
		if n.cfg.pushServer.subjectHierarchiesEnabled {
			subTopic = fmt.Sprintf("%s.>", subTopic)
		}

		defer n.cfg.pushServer.service.Subscribe(n.publications, errors, subTopic)()
	}

	zap.L().Debug("Websocket server started",
		zap.Bool("push-enabled", n.cfg.pushServer.enabled),
		zap.Bool("push-dispatching-enabled", n.cfg.pushServer.dispatchEnabled),
		zap.Bool("push-publish-enabled", n.cfg.pushServer.publishEnabled),
	)

	for {
		select {

		case p := <-n.publications:

			go func(publication *Publication) {

				event := &elemental.Event{}
				if err := publication.Decode(event); err != nil {
					zap.L().Error("Unable to decode event",
						zap.Stringer("event", event),
						zap.Error(err),
					)
					return
				}

				// We prepare the event data in both json and msgpack
				// once for all.
				dataMSGPACK, dataJSON, err := prepareEventData(event)
				if err != nil {
					zap.L().Error("Unable to prepare event encoding",
						zap.Stringer("event", event),
						zap.Error(err),
					)
					return
				}

				// We prepate the event summary if needed
				var eventSummary interface{}
				if n.cfg.pushServer.dispatchHandler != nil {
					eventSummary, err = n.cfg.pushServer.dispatchHandler.SummarizeEvent(event)
					if err != nil {
						zap.L().Error("Unable to summary event",
							zap.Stringer("event", event),
							zap.Error(err),
						)
						return
					}
				}

				// Keep a references to all current ready push sessions as it may change at any time, we lost 8h on this one...
				n.sessionsLock.RLock()
				sessions := make([]*wsPushSession, len(n.sessions))
				var i int
				for _, s := range n.sessions {
					sessions[i] = s
					i++
				}
				n.sessionsLock.RUnlock()

				// Dispatch the event to all sessions
				for _, session := range sessions {

					// Client sent an invalid push config, this is a noop as it makes no sense to continue processing;
					// wait until they send another message that is valid.
					if session.inErrorState() {
						continue
					}

					// If event happened before session, we don't send it.
					if event.Timestamp.Before(session.startTime) {
						continue
					}

					// If the event identity (or related identities) are filtered out
					// we don't send it.
					if f := session.currentPushConfig(); f != nil {

						identities := []string{event.Identity}
						if n.cfg.pushServer.dispatchHandler != nil {
							identities = append(identities, n.cfg.pushServer.dispatchHandler.RelatedEventIdentities(event.Identity)...)
						}

						var ok bool
						for _, identity := range identities {
							if !f.IsFilteredOut(identity, event.Type) {
								ok = true
								break
							}
						}

						if !ok {
							continue
						}
					}

					if n.cfg.pushServer.dispatchHandler != nil {
						dispatch, err := n.cfg.pushServer.dispatchHandler.ShouldDispatch(session, event, eventSummary)
						if err != nil {
							// temp before we move to error wrapping
							if err != context.Canceled && !strings.Contains(err.Error(), "context canceled") {
								zap.L().Error("Error while calling dispatchHandler.ShouldDispatch", zap.Error(err))
							}

							continue
						}

						if !dispatch {
							continue
						}
					}

					switch session.encodingWrite {
					case elemental.EncodingTypeMSGPACK:
						session.send(dataMSGPACK)
					case elemental.EncodingTypeJSON:
						session.send(dataJSON)
					}
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

func prepareEventData(event *elemental.Event) (msgpack []byte, json []byte, err error) {

	eventCopy := event.Duplicate()

	switch event.GetEncoding() {

	case elemental.EncodingTypeMSGPACK:

		msgpack, err = elemental.Encode(elemental.EncodingTypeMSGPACK, event)
		if err != nil {
			return nil, nil, fmt.Errorf("unable to encode original msgpack event: %s", err)
		}

		if err = eventCopy.Convert(elemental.EncodingTypeJSON); err != nil {
			return nil, nil, fmt.Errorf("unable to convert original msgpack encoding to json: %s", err)
		}

		json, err = elemental.Encode(elemental.EncodingTypeJSON, eventCopy)
		if err != nil {
			return nil, nil, fmt.Errorf("unable to encode json version of original msgpack event: %s", err)
		}

	case elemental.EncodingTypeJSON:

		json, err = elemental.Encode(elemental.EncodingTypeJSON, event)
		if err != nil {
			return nil, nil, fmt.Errorf("unable to encode original json event: %s", err)
		}

		if err = eventCopy.Convert(elemental.EncodingTypeMSGPACK); err != nil {
			return nil, nil, fmt.Errorf("unable to convert original json encoding to msgpack: %s", err)
		}

		msgpack, err = elemental.Encode(elemental.EncodingTypeMSGPACK, eventCopy)
		if err != nil {
			return nil, nil, fmt.Errorf("unable to encode msgpack version of original json event: %s", err)
		}
	}

	return msgpack, json, nil
}
