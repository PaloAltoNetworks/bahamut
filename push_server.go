// Author: Antoine Mercadal
// See LICENSE file for full LICENSE
// Copyright 2016 Aporeto.

package bahamut

import (
	"golang.org/x/net/websocket"

	"github.com/aporeto-inc/elemental"
	"github.com/go-zoo/bone"

	log "github.com/Sirupsen/logrus"
)

type pushServer struct {
	address     string
	sessions    map[string]*PushSession
	register    chan *PushSession
	unregister  chan *PushSession
	events      chan *elemental.Event
	close       chan bool
	multiplexer *bone.Mux
	config      PushServerConfig
}

func newPushServer(config PushServerConfig, multiplexer *bone.Mux) *pushServer {

	srv := &pushServer{
		sessions:    map[string]*PushSession{},
		register:    make(chan *PushSession),
		unregister:  make(chan *PushSession),
		close:       make(chan bool, 2),
		events:      make(chan *elemental.Event, 1024),
		multiplexer: multiplexer,
		config:      config,
	}

	srv.multiplexer.Handle("/events", websocket.Handler(srv.handleConnection))

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

// unpublish all local sessions from the global registry system
func (n *pushServer) handleConnection(ws *websocket.Conn) {

	if handler := n.config.SessionsHandler; handler != nil {
		ok, err := handler.IsAuthenticated(ws)
		if err != nil {
			log.WithFields(log.Fields{
				"error":   err,
				"package": "bahamut",
			}).Error("Error during checking authentication.")
		}

		if !ok {
			ws.Close()
			return
		}
	}

	session := newPushSession(ws, n)
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

	log.WithFields(log.Fields{
		"endpoint": n.address + "/events",
		"package":  "bahamut",
	}).Info("Starting event server.")

	for {
		select {

		case session := <-n.register:

			if _, ok := n.sessions[session.id]; ok {
				break
			}

			n.sessions[session.id] = session

			log.WithFields(log.Fields{
				"total":   len(n.sessions),
				"client":  session.socket.RemoteAddr(),
				"package": "bahamut",
			}).Info("Push session started.")

			if handler := n.config.SessionsHandler; handler != nil {
				handler.OnPushSessionStart(session)
			}

		case session := <-n.unregister:

			if _, ok := n.sessions[session.id]; !ok {
				break
			}

			delete(n.sessions, session.id)

			log.WithFields(log.Fields{
				"total":   len(n.sessions),
				"client":  session.socket.RemoteAddr(),
				"package": "bahamut",
			}).Info("Push session closed.")

			if handler := n.config.SessionsHandler; handler != nil {
				handler.OnPushSessionStop(session)
			}

		case event := <-n.events:

			if n.config.Service != nil {
				publication := NewPublication(n.config.Topic)
				if err := publication.Encode(event); err != nil {
					log.WithFields(log.Fields{
						"topic":   publication.Topic,
						"event":   event,
						"error":   err,
						"package": "bahamut",
					}).Error("Unable to encode ervent. Message dropped.")
				}
				err := n.config.Service.Publish(publication)
				if err != nil {
					log.WithFields(log.Fields{
						"topic":   publication.Topic,
						"event":   event,
						"error":   err,
						"package": "bahamut",
					}).Warn("Unable to publish. Message dropped.")
				}
			}

		case <-n.close:
			log.WithFields(log.Fields{
				"package": "bahamut",
			}).Info("Stopping push server.")

			n.closeAllSessions()
			return
		}
	}
}

// stops the push server
func (n *pushServer) stop() {

	n.close <- true
}
