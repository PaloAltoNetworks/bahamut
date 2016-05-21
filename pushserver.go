// Author: Antoine Mercadal
// See LICENSE file for full LICENSE
// Copyright 2016 Aporeto.

package bahamut

import (
	"golang.org/x/net/websocket"

	log "github.com/Sirupsen/logrus"
	"github.com/aporeto-inc/elemental"
	"github.com/go-zoo/bone"
)

type pushServer struct {
	address     string
	sessions    map[string]*pushSession
	register    chan *pushSession
	unregister  chan *pushSession
	events      chan *elemental.Event
	stop        chan bool
	multiplexer *bone.Mux
}

func newPushServer(address string, multiplexer *bone.Mux) *pushServer {

	srv := &pushServer{
		address:     address,
		sessions:    map[string]*pushSession{},
		register:    make(chan *pushSession),
		unregister:  make(chan *pushSession),
		stop:        make(chan bool),
		events:      make(chan *elemental.Event),
		multiplexer: multiplexer,
	}

	srv.multiplexer.Handle("/events", websocket.Handler(srv.handleConnection))

	return srv
}

func (n *pushServer) handleConnection(ws *websocket.Conn) {

	session := newSession(ws, n)
	n.registerSession(session)
	session.listen()
}

func (n *pushServer) registerSession(session *pushSession) {

	n.register <- session
}

func (n *pushServer) unregisterSession(session *pushSession) {

	n.unregister <- session
}

func (n *pushServer) pushEvents(events ...*elemental.Event) {

	for _, e := range events {
		n.events <- e
	}
}

func (n *pushServer) start() {

	log.WithFields(log.Fields{
		"endpoint": n.address + "/events",
	}).Info("starting event server")

	for {
		select {

		case session := <-n.register:

			if _, ok := n.sessions[session.id]; ok {
				break
			}

			n.sessions[session.id] = session

			log.WithFields(log.Fields{
				"total":  len(n.sessions),
				"client": session.socket.RemoteAddr(),
			}).Info("started push session")

		case session := <-n.unregister:

			if _, ok := n.sessions[session.id]; !ok {
				break
			}

			delete(n.sessions, session.id)

			log.WithFields(log.Fields{
				"total":  len(n.sessions),
				"client": session.socket.RemoteAddr(),
			}).Info("closed session")

		case event := <-n.events:

			go func() {
				for _, session := range n.sessions {
					session.events <- event
				}
			}()

		case <-n.stop:

			for _, session := range n.sessions {
				session.close <- true
			}
			n.sessions = map[string]*pushSession{}
			return
		}
	}
}

func (n *pushServer) Stop() {

	n.stop <- true
}
