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
	register        chan *PushSession
	unregister      chan *PushSession
	events          chan *elemental.Event
	close           chan bool
	multiplexer     *bone.Mux
	config          Config
	processorFinder processorFinder
	sessionsLock    *sync.Mutex
}

func newPushServer(config Config, multiplexer *bone.Mux) *pushServer {

	srv := &pushServer{
		sessions:     map[string]*PushSession{},
		register:     make(chan *PushSession),
		unregister:   make(chan *PushSession),
		close:        make(chan bool, 2),
		events:       make(chan *elemental.Event, 1024),
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

	n.register <- session
}

// adds a new push session to unregister from the push server
func (n *pushServer) unregisterSession(session *PushSession) {

	n.unregister <- session
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

	if handler := n.config.WebSocketServer.SessionsHandler; handler != nil {
		ok, err := handler.IsAuthenticated(session)
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

	for _, e := range events {
		select {
		case n.events <- e:
		default:
		}
	}
}

func (n *pushServer) closeAllSessions() {

	for _, session := range n.sessions {
		session.close()
	}
	n.sessions = map[string]*PushSession{}
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

		case session := <-n.register:

			n.sessions[session.id] = session

			log.WithFields(logrus.Fields{
				"total":  len(n.sessions),
				"client": session.socket.RemoteAddr(),
			}).Info("Push session started.")

			if handler := n.config.WebSocketServer.SessionsHandler; handler != nil {
				handler.OnPushSessionStart(session)
			}

		case session := <-n.unregister:

			delete(n.sessions, session.id)

			log.WithFields(logrus.Fields{
				"total":  len(n.sessions),
				"client": session.socket.RemoteAddr(),
			}).Info("Push session closed.")

			if handler := n.config.WebSocketServer.SessionsHandler; handler != nil {
				handler.OnPushSessionStop(session)
			}

		case event := <-n.events:

			if n.config.WebSocketServer.Service == nil {
				break
			}

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
				break
			}

		case publication := <-publications:

			event := &elemental.Event{}
			if err := publication.Decode(event); err != nil {
				log.WithField("error", err.Error()).Error("Unable to decode event.")
				break
			}

			// Keep a references to all current ready push sessions as it may change at any time, we lost 8h on this one...
			var sessions []*PushSession
			for _, session := range n.sessions {
				if session.sType == pushSessionTypeEvent && session.isReady() {
					sessions = append(sessions, session)
				}
			}

			// Dispatch the event to all sessions
			for _, session := range sessions {
				go func(s *PushSession, evt *elemental.Event) { s.events <- evt }(session, event)
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
