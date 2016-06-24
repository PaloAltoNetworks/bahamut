// Author: Antoine Mercadal
// See LICENSE file for full LICENSE
// Copyright 2016 Aporeto.

package bahamut

import (
	"fmt"
	"time"

	"gopkg.in/redis.v3"

	"github.com/satori/go.uuid"
	"golang.org/x/net/websocket"
)

const (
	redisSessionEventQueuesKey = "bahamut:sessions:eventqueues"
)

// pushSession represents a client session.
type pushSession struct {
	events      chan string
	id          string
	redisClient *redis.Client
	redisKey    string
	server      *pushServer
	socket      *websocket.Conn
	stop        chan bool
}

func newSession(ws *websocket.Conn, server *pushServer) *pushSession {

	id := uuid.NewV4().String()

	return &pushSession{
		events:      make(chan string),
		id:          id,
		redisClient: server.redisClient,
		redisKey:    fmt.Sprintf("%s:%s", redisSessionEventQueuesKey, id),
		server:      server,
		socket:      ws,
		stop:        make(chan bool, 1),
	}
}

func (s *pushSession) startEventQueueListener() {

	if s.redisClient == nil {
		return
	}

	for {
		select {
		case <-s.stop:
			return

		default:
			resp := s.redisClient.BRPop(5*time.Second, s.redisKey).Val()

			if len(resp) == 2 {
				s.events <- resp[1]
			}
		}
	}
}

func (s *pushSession) stopEventQueueListener() {

	if s.redisClient == nil {
		return
	}

	s.stop <- true
}

func (s *pushSession) terminate() {

	s.server.unregisterSession(s)
	s.stopEventQueueListener()
}

func (s *pushSession) read() {

	for {
		if _, err := s.socket.Read(nil); err != nil {
			s.terminate()
			break
		}
	}
}

func (s *pushSession) listen() {

	defer s.socket.Close()

	go s.read()
	go s.startEventQueueListener()

	for {
		if err := websocket.Message.Send(s.socket, <-s.events); err != nil {
			s.terminate()
			break
		}
	}
}
