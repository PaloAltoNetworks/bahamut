// Author: Antoine Mercadal
// See LICENSE file for full LICENSE
// Copyright 2016 Aporeto.

package bahamut

import (
	"encoding/json"
	"strings"

	"github.com/aporeto-inc/bahamut/multistop"
	"github.com/aporeto-inc/bahamut/pubsub"
	"github.com/aporeto-inc/elemental"
	"github.com/satori/go.uuid"
	"golang.org/x/net/websocket"

	log "github.com/Sirupsen/logrus"
)

// PushSession represents a client session.
type PushSession struct {
	id        string
	server    *pushServer
	socket    *websocket.Conn
	out       chan string
	UserInfo  interface{}
	multicast *multistop.MultiStop
}

func newPushSession(ws *websocket.Conn, server *pushServer) *PushSession {

	return &PushSession{
		id:        uuid.NewV4().String(),
		server:    server,
		socket:    ws,
		out:       make(chan string, 1024),
		multicast: multistop.NewMultiStop(),
	}
}

// Identifier returns the identifier of the push session
func (s *PushSession) Identifier() string {

	return s.id
}

// continuously read data from the websocket
func (s *PushSession) read() {

	for {
		var data []byte
		if err := websocket.Message.Receive(s.socket, &data); err != nil {
			s.close()
			return
		}
	}
}

func (s *PushSession) write() {

	stopCh := make(chan bool)
	s.multicast.Register(stopCh)
	defer s.multicast.Unregister(stopCh)

	for {
		select {
		case data := <-s.out:
			if err := websocket.Message.Send(s.socket, data); err != nil {
				go s.close()
			}
		case <-stopCh:
			return
		}
	}
}

// send given bytes to the websocket
func (s *PushSession) send(message string) error {

	if s.server.config.SessionsHandler != nil {

		var event *elemental.Event
		if err := json.NewDecoder(strings.NewReader(message)).Decode(&event); err != nil {
			log.WithFields(log.Fields{
				"session": s,
				"message": message,
				"package": "bahamut",
			}).Error("Unable to decode event.")
			return err
		}

		if !s.server.config.SessionsHandler.ShouldPush(s, event) {
			return nil
		}
	}

	select {
	case s.out <- message:
	default:
	}

	return nil
}

// force close the current socket
func (s *PushSession) close() {

	s.multicast.Send(true)
}

// listens to events, either from kafka or from local events.
func (s *PushSession) listen() {

	publications := make(chan *pubsub.Publication)
	unsubscribe := s.server.config.Service.Subscribe(publications, s.server.config.Topic)

	stopCh := make(chan bool)
	s.multicast.Register(stopCh)

	defer s.multicast.Unregister(stopCh)
	defer s.server.unregisterSession(s)
	defer s.socket.Close()
	defer unsubscribe()

	go s.read()
	go s.write()

	for {
		select {
		case message := <-publications:
			s.send(string(message.Data()))
		case <-stopCh:
			return
		}
	}
}
