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

// pushSession represents a client session.
type pushSession struct {
	events    chan string
	id        string
	kafkaInfo *KafkaInfo
	server    *pushServer
	socket    *websocket.Conn
	stop      chan bool
}

func newSession(ws *websocket.Conn, server *pushServer) *pushSession {

	return &pushSession{
		events:    make(chan string),
		id:        uuid.NewV4().String(),
		kafkaInfo: server.kafkaInfo,
		server:    server,
		socket:    ws,
		stop:      make(chan bool, 1),
	}
}

// continuously read data from the websocket
func (s *pushSession) read() {

	for {
		if _, err := s.socket.Read(nil); err != nil {
			s.close()
			break
		}
	}
}

// send given bytes to the websocket
func (s *pushSession) send(message []byte) error {

	return websocket.Message.Send(s.socket, message)
}

// force close the current socket
func (s *pushSession) close() {

	s.stop <- true
}

// listens to events, either from kafka or from local events.
func (s *pushSession) listen() {

	defer s.socket.Close()
	go s.read()

	if s.kafkaInfo != nil {
		s.listenToKafkaMessages()
	} else {
		s.listenToLocalMessages()
	}
}

// continuously listens for new kafka messages
func (s *pushSession) listenToKafkaMessages() {

	kafkaConsumer := s.kafkaInfo.makeConsumer()
	defer kafkaConsumer.Close()

	parititionConsumer, err := kafkaConsumer.ConsumePartition(s.kafkaInfo.Topic, 0, sarama.OffsetNewest)
	if err != nil {
		log.WithFields(log.Fields{
			"session":  s,
			"consumer": kafkaConsumer,
			"error":    err,
		}).Error("unable to comsume topic")
		return
	}
	defer parititionConsumer.Close()

	for {
		select {
		case message := <-parititionConsumer.Messages():
			if err := s.send(message.Value); err != nil {
				return
			}
		case <-s.stop:
			return
		}
	}
}

// continuously listens for new local messages
func (s *pushSession) listenToLocalMessages() {

	for {
		select {
		case message := <-s.events:
			if err := s.send([]byte(message)); err != nil {
				return
			}
		case <-s.stop:
			return
		}
	}
}
