// Author: Antoine Mercadal
// See LICENSE file for full LICENSE
// Copyright 2016 Aporeto.

package bahamut

import (
	"net/http"
	"sync"

	"golang.org/x/net/websocket"

	"github.com/Sirupsen/logrus"
	"github.com/aporeto-inc/elemental"
	"github.com/go-zoo/bone"
)

type pushServer struct {
	address         string
	sessions        map[string]*PushSession
	close           chan bool
	multiplexer     *bone.Mux
	config          Config
	processorFinder processorFinder
	sessionsLock    *sync.Mutex
}

func newPushServer(config Config, multiplexer *bone.Mux) *pushServer {

	srv := &pushServer{
		sessions:     map[string]*PushSession{},
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
func (n *pushServer) registerSession(session *PushSession) {

	n.sessionsLock.Lock()
	defer n.sessionsLock.Unlock()

	n.sessions[session.id] = session

	if handler := n.config.WebSocketServer.SessionsHandler; handler != nil {
		handler.OnPushSessionStart(session)
	}
}

// adds a new push session to unregister from the push server
func (n *pushServer) unregisterSession(session *PushSession) {

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

func (n *pushServer) runSession(ws *websocket.Conn, session *PushSession) {

	if n.config.Security.SessionAuthenticator != nil {

		ok, err := n.config.Security.SessionAuthenticator.AuthenticateSession(session)

		if err != nil {
			log.WithError(err).Error("Error during checking authentication.")
		}
		if !ok {
			if session.sType == pushSessionTypeAPI {
				response := elemental.NewResponse()
				writeWebSocketError(ws, response, elemental.NewError("Unauthorized", "You are not authorized to access this api", "bahamut", http.StatusUnauthorized))
			}
			ws.Close()
			return
		}
	}

	// Send the first hello message.
	if session.sType == pushSessionTypeAPI {
		response := elemental.NewResponse()
		response.StatusCode = http.StatusOK
		websocket.JSON.Send(ws, response)
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
			log.WithField("error", err.Error()).Error("Unable to encode event. Message dropped.")
			break
		}

		if err := n.config.WebSocketServer.Service.Publish(publication); err != nil {
			log.WithFields(logrus.Fields{
				"topic": publication.Topic,
				"event": event.String(),
				"error": err.Error(),
			}).Warn("Unable to publish. Message dropped.")
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

	log.WithField("endpoint", n.address+"/events").Info("Starting event server.")

	publications := make(chan *Publication)
	if n.config.WebSocketServer.Service != nil {
		errors := make(chan error)
		unsubscribe := n.config.WebSocketServer.Service.Subscribe(publications, errors, n.config.WebSocketServer.Topic)
		log.Info("Subscribed to events")
		defer unsubscribe()
	}

	for {
		select {

		case publication := <-publications:

			event := &elemental.Event{}
			if err := publication.Decode(event); err != nil {
				log.WithField("error", err.Error()).Error("Unable to decode event.")
				break
			}

			// Keep a references to all current ready push sessions as it may change at any time, we lost 8h on this one...
			n.sessionsLock.Lock()
			var sessions []*PushSession
			for _, session := range n.sessions {
				if session.sType == pushSessionTypeEvent {
					sessions = append(sessions, session)
				}
			}
			n.sessionsLock.Unlock()

			// Dispatch the event to all sessions
			for _, session := range sessions {

				go func(s *PushSession, evt *elemental.Event) {

					if n.config.WebSocketServer.SessionsHandler != nil {

						ok, err := s.config.WebSocketServer.SessionsHandler.ShouldPush(s, evt)
						if err != nil {
							log.WithError(err).Error("Error while checking authorization.")
							return
						}

						if !ok {
							return
						}
					}

					s.DirectPush(evt)

				}(session, event.Duplicate())
			}

		case <-n.close:

			log.Info("Stopping push server.")
			n.closeAllSessions()
			return
		}
	}
}

// stops the push server
func (n *pushServer) stop() {

	n.close <- true
}
