// Author: Antoine Mercadal
// See LICENSE file for full LICENSE
// Copyright 2016 Aporeto.

package bahamut

import (
	"bytes"
	"encoding/json"
	"fmt"

	"golang.org/x/net/websocket"
	"gopkg.in/redis.v3"

	log "github.com/Sirupsen/logrus"
	"github.com/aporeto-inc/elemental"
	"github.com/go-zoo/bone"
)

const (
	redisGlobalSessionsKey = "bahamut:global:sessions"
)

type pushServer struct {
	address     string
	sessions    map[string]*pushSession
	register    chan *pushSession
	unregister  chan *pushSession
	events      chan *elemental.Event
	stop        chan bool
	multiplexer *bone.Mux
	redisClient *redis.Client
}

func newPushServer(address string, multiplexer *bone.Mux, redisClient *redis.Client) *pushServer {

	srv := &pushServer{
		address:     address,
		sessions:    map[string]*pushSession{},
		register:    make(chan *pushSession),
		unregister:  make(chan *pushSession),
		stop:        make(chan bool),
		events:      make(chan *elemental.Event),
		multiplexer: multiplexer,
		redisClient: redisClient,
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

	if n.redisClient == nil {
		for _, e := range events {
			n.events <- e
		}

		return
	}

	sessionKeys := n.globalSessionIDs()

	// TODO: add a hook here to decide if we should publish the event here or not.

	n.redisClient.Pipelined(func(pipeline *redis.Pipeline) error {
		for _, sKey := range sessionKeys {

			for _, event := range events {

				eventQueueKey := fmt.Sprintf("%s:%s", redisSessionEventQueuesKey, sKey)

				buffer := &bytes.Buffer{}
				if err := json.NewEncoder(buffer).Encode(event); err != nil {
					log.WithFields(log.Fields{
						"redis": n.redisClient,
						"event": event,
					}).Error("unable to encode event.")
				}

				pipeline.LPush(eventQueueKey, buffer.String())
			}
		}
		return nil
	})
}

func (n *pushServer) createRedisSession(session *pushSession) {

	if n.redisClient == nil {
		return
	}

	// add the session to the global registry
	n.redisClient.SAdd(redisGlobalSessionsKey, session.id)

	log.WithFields(log.Fields{
		"redis":      n.redisClient,
		"session.id": session.id,
	}).Debug("session added to redis.")
}

func (n *pushServer) deleteRedisSession(session *pushSession) {

	if n.redisClient == nil {
		return
	}

	n.redisClient.Pipelined(func(pipeline *redis.Pipeline) error {
		pipeline.SRem(redisGlobalSessionsKey, session.id)
		pipeline.Del(session.redisKey)
		return nil
	})

	log.WithFields(log.Fields{
		"redis":      n.redisClient,
		"session.id": session.id,
	}).Debug("session deleted from redis.")
}

func (n *pushServer) globalSessionIDs() []string {

	return n.redisClient.SMembers(redisGlobalSessionsKey).Val()
}

func (n *pushServer) redisFlushLocalSessions() {

	if n.redisClient == nil {
		return
	}

	n.redisClient.Pipelined(func(pipeline *redis.Pipeline) error {
		for _, session := range n.sessions {
			pipeline.SRem(redisGlobalSessionsKey, session.id)
			pipeline.Del(session.redisKey)
		}
		return nil
	})

	log.WithFields(log.Fields{
		"redis": n.redisClient,
	}).Debug("sessions flushed from redis.")
}

func (n *pushServer) start() {

	if n.redisClient != nil {
		_, err := n.redisClient.Ping().Result()
		if err != nil {
			log.WithFields(log.Fields{
				"redis": n.redisClient,
			}).Warn("unable to contact redis: unclustered push system.")
		} else {
			log.WithFields(log.Fields{
				"redis": n.redisClient,
			}).Info("connected to redis: clustered push system.")

			n.redisClient.ConfigSet("notify-keyspace-events", "KEA")
		}
	}

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
			n.createRedisSession(session)

			log.WithFields(log.Fields{
				"total":  len(n.sessions),
				"client": session.socket.RemoteAddr(),
			}).Info("started push session")

		case session := <-n.unregister:

			if _, ok := n.sessions[session.id]; !ok {
				break
			}

			n.deleteRedisSession(session)
			delete(n.sessions, session.id)

			log.WithFields(log.Fields{
				"total":  len(n.sessions),
				"client": session.socket.RemoteAddr(),
			}).Info("closed session")

		case event := <-n.events:
			go func() {
				for _, session := range n.sessions {
					buffer := &bytes.Buffer{}
					if err := json.NewEncoder(buffer).Encode(event); err != nil {
						log.WithFields(log.Fields{
							"data": event,
						}).Error("unable to encode event data")
						return
					}

					session.events <- buffer.String()
				}
			}()

		case <-n.stop:

			n.redisFlushLocalSessions()

			for _, session := range n.sessions {
				session.socket.Close()
			}
			n.sessions = map[string]*pushSession{}

			return
		}
	}
}

func (n *pushServer) Stop() {

	n.redisFlushLocalSessions()
	n.stop <- true
}
