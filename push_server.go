// Author: Antoine Mercadal
// See LICENSE file for full LICENSE
// Copyright 2016 Aporeto.

package bahamut

import (
	"bytes"
	"encoding/json"

	"golang.org/x/net/websocket"

	"github.com/Shopify/sarama"
	log "github.com/Sirupsen/logrus"
	"github.com/aporeto-inc/elemental"
	"github.com/go-zoo/bone"
)

type pushServer struct {
	address       string
	sessions      map[string]*PushSession
	register      chan *PushSession
	unregister    chan *PushSession
	events        chan *elemental.Event
	close         chan bool
	multiplexer   *bone.Mux
	config        PushServerConfig
	kafkaProducer sarama.SyncProducer
}

func newPushServer(config PushServerConfig, multiplexer *bone.Mux) *pushServer {

	srv := &pushServer{
		sessions:    map[string]*PushSession{},
		register:    make(chan *PushSession),
		unregister:  make(chan *PushSession),
		close:       make(chan bool),
		events:      make(chan *elemental.Event),
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

	session := newPushSession(ws, n)
	n.registerSession(session)
	session.listen()
}

// push a new event. If the global push system is available, it will be used.
// otherwise, only local sessions will receive the push
func (n *pushServer) pushEvents(events ...*elemental.Event) {

	for _, e := range events {
		n.events <- e
	}
}

// starts the push server
func (n *pushServer) start() {

	log.WithFields(log.Fields{
		"endpoint": n.address + "/events",
	}).Info("starting event server")

	if n.config.HasKafka() {
		n.kafkaProducer = n.config.makeProducer()

		defer n.kafkaProducer.Close()

		log.WithFields(log.Fields{
			"producer": n.kafkaProducer,
		}).Info("global push system is active")
	} else {
		log.Warn("global push system is inactive")
	}

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

			if handler := n.config.sessionsHandler; handler != nil {
				handler.OnPushSessionStart(session)
			}

		case session := <-n.unregister:

			if _, ok := n.sessions[session.id]; !ok {
				break
			}

			delete(n.sessions, session.id)

			log.WithFields(log.Fields{
				"total":  len(n.sessions),
				"client": session.socket.RemoteAddr(),
			}).Info("closed session")

			if handler := n.config.sessionsHandler; handler != nil {
				handler.OnPushSessionStop(session)
			}

		case event := <-n.events:

			buffer := &bytes.Buffer{}
			if err := json.NewEncoder(buffer).Encode(event); err != nil {
				log.WithFields(log.Fields{
					"data": event,
				}).Error("unable to encode event data")
				return
			}

			if n.kafkaProducer != nil {
				message := &sarama.ProducerMessage{
					Topic: n.config.DefaultTopic,
					Key:   sarama.StringEncoder("namespace=default"),
					Value: sarama.ByteEncoder(buffer.Bytes()),
				}

				n.kafkaProducer.SendMessage(message)
			} else {
				for _, session := range n.sessions {
					session.events <- buffer.String()
				}
			}

		case <-n.close:

			for _, session := range n.sessions {
				session.close()
			}
			n.sessions = map[string]*PushSession{}

			return
		}
	}
}

// stops the push server
func (n *pushServer) stop() {

	n.close <- true
}
