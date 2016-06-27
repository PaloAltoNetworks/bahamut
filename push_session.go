// Author: Antoine Mercadal
// See LICENSE file for full LICENSE
// Copyright 2016 Aporeto.

package bahamut

import (
	"github.com/Shopify/sarama"
	"github.com/satori/go.uuid"
	"golang.org/x/net/websocket"

	log "github.com/Sirupsen/logrus"
)

// PushSession represents a client session.
type PushSession struct {
	events           chan string
	id               string
	pushServerConfig *PushServerConfig
	server           *pushServer
	socket           *websocket.Conn
	stop             chan bool
	out              chan []byte
}

func newPushSession(ws *websocket.Conn, server *pushServer) *PushSession {

	return &PushSession{
		events:           make(chan string),
		id:               uuid.NewV4().String(),
		pushServerConfig: server.config,
		server:           server,
		socket:           ws,
		stop:             make(chan bool, 1),
		out:              make(chan []byte),
	}
}

// continuously read data from the websocket
func (s *PushSession) read() {

	for {
		var data []byte
		if err := websocket.Message.Receive(s.socket, &data); err != nil {
			s.stop <- true
			break
		}
	}
}

func (s *PushSession) write() {

	for {
		select {
		case data := <-s.out:
			if err := websocket.Message.Send(s.socket, data); err != nil {
				s.stop <- true
				return
			}
		case <-s.stop:
			s.stop <- true
			return
		}
	}
}

// send given bytes to the websocket
func (s *PushSession) send(message []byte) {

	s.out <- message
}

// force close the current socket
func (s *PushSession) close() {

	s.stop <- true
}

// listens to events, either from kafka or from local events.
func (s *PushSession) listen() {

	defer s.socket.Close()

	go s.read()
	go s.write()

	if s.pushServerConfig != nil {
		s.listenToKafkaMessages()
	} else {
		s.listenToLocalMessages()
	}
}

// continuously listens for new kafka messages
func (s *PushSession) listenToKafkaMessages() error {

	consumer := s.pushServerConfig.makeConsumer()
	defer consumer.Close()

	parititionConsumer, err := consumer.ConsumePartition(s.pushServerConfig.DefaultTopic, 0, sarama.OffsetNewest)
	if err != nil {
		log.WithFields(log.Fields{
			"session": s,
			"error":   err,
		}).Error("unable to consume topic")
		return err
	}
	defer parititionConsumer.Close()

	for {
		select {
		case message := <-parititionConsumer.Messages():
			s.send(message.Value)
		case <-s.stop:
			s.server.unregisterSession(s)
			return nil
		}
	}
}

// continuously listens for new local messages
func (s *PushSession) listenToLocalMessages() error {

	for {
		select {
		case message := <-s.events:
			s.send([]byte(message))
		case <-s.stop:
			s.server.unregisterSession(s)
			return nil
		}
	}
}
