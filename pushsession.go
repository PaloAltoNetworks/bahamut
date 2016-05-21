// Author: Antoine Mercadal
// See LICENSE file for full LICENSE
// Copyright 2016 Aporeto.

package bahamut

import (
	"github.com/aporeto-inc/cid/materia/elemental"
	"github.com/satori/go.uuid"
	"golang.org/x/net/websocket"
)

// pushSession represents a client session.
type pushSession struct {
	id     string
	socket *websocket.Conn
	events chan *elemental.Event
	close  chan bool
	server *pushServer
}

func newSession(ws *websocket.Conn, server *pushServer) *pushSession {

	return &pushSession{
		id:     uuid.NewV4().String(),
		socket: ws,
		events: make(chan *elemental.Event),
		close:  make(chan bool, 1),
		server: server,
	}
}

func (s *pushSession) read() {

	for {
		if _, err := s.socket.Read(nil); err != nil {
			s.server.unregisterSession(s)
			break
		}
	}
}

func (s *pushSession) listen() {

	defer s.socket.Close()

	go s.read()

	for {
		select {
		case event := <-s.events:

			err := websocket.JSON.Send(s.socket, event)

			if err != nil {
				s.server.unregisterSession(s)
			}

		case <-s.close:
			s.socket.Close()
			return
		}
	}
}
