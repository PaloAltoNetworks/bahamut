package bahamut

import (
	"fmt"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/nats-io/go-nats"
	"github.com/nats-io/go-nats-streaming"
)

type natsPubSub struct {
	natsURL       string
	nc            *nats.Conn
	client        stan.Conn
	retryInterval time.Duration
	clientID      string
	clusterID     string
	password      string
	username      string
}

// newNatsPubSub Initializes the pubsub server.
func newNatsPubSub(natsURL string, clusterID string, clientID string, username string, password string) *natsPubSub {

	return &natsPubSub{
		natsURL:       natsURL,
		retryInterval: 5 * time.Second,
		clientID:      clientID,
		clusterID:     clusterID,
		username:      username,
		password:      password,
	}
}

func (p *natsPubSub) Publish(publication *Publication) error {

	if p.client == nil {
		return fmt.Errorf("Not connected to nats. Messages dropped")
	}

	log.WithFields(logrus.Fields{
		"topic":   publication.Topic,
		"natsURL": p.natsURL,
		"data":    string(publication.data),
	}).Debug("Publishing message in nats")

	return p.client.Publish(publication.Topic, publication.data)
}

func (p *natsPubSub) Subscribe(pubs chan *Publication, errors chan error, topic string, args ...interface{}) func() {

	var queueGroup string
	options := []stan.SubscriptionOption{}

	for i, arg := range args {
		if i == 0 {
			if q, ok := arg.(string); ok {
				queueGroup = q
			} else {
				panic("You must provide a string as queue group name")
			}
			continue
		}

		if opt, ok := arg.(stan.SubscriptionOption); ok {
			options = append(options, opt)
		} else {
			panic("Subsequent arguments must be of type stan.SubscriptionOption")
		}

	}

	var sub stan.Subscription
	var err error

	handler := func(m *stan.Msg) {
		publication := NewPublication(topic)
		publication.data = m.Data
		pubs <- publication
	}

	if queueGroup == "" {
		sub, err = p.client.Subscribe(topic, handler, options...)
	} else {
		sub, err = p.client.QueueSubscribe(topic, queueGroup, handler, options...)
	}

	if err != nil {
		errors <- err
		return func() {}
	}

	return func() { _ = sub.Unsubscribe() }
}

func (p *natsPubSub) Connect() Waiter {

	abort := make(chan bool, 2)
	connected := make(chan bool, 2)

	go func() {

		// First, we create a connection to the nats cluster.
		for p.nc == nil {

			var err error

			if p.username != "" || p.password != "" {
				p.nc, err = nats.Connect(p.natsURL, nats.UserInfo(p.username, p.password))
			} else {
				p.nc, err = nats.Connect(p.natsURL)
			}

			if err == nil {
				break
			}

			log.WithFields(logrus.Fields{
				"url":   p.natsURL,
				"error": err.Error(),
				"retry": p.retryInterval,
			}).Warn("Unable to connect to nats cluster. Retrying.")

			select {
			case <-time.After(p.retryInterval):
			case <-abort:
				connected <- false
				return
			}
		}

		// Then, we open a nats streaming session using the nats connection.
		for p.client == nil {

			var err error
			var client stan.Conn

			if p.nc.IsConnected() {

				client, err = stan.Connect(p.clusterID, p.clientID, stan.NatsConn(p.nc))

				if err == nil && client != nil {
					p.client = client
					break
				}
			}

			log.WithFields(logrus.Fields{
				"url":       p.natsURL,
				"clusterID": p.clusterID,
				"clientID":  p.clientID,
				"retry":     p.retryInterval,
				"error":     err.Error(),
			}).Warn("Unable to connect to nats streaming server. Retrying.")

			select {
			case <-time.After(p.retryInterval):
			case <-abort:
				connected <- false
				return
			}
		}

		connected <- true
	}()

	return connectionWaiter{
		ok:    connected,
		abort: abort,
	}
}

func (p *natsPubSub) Disconnect() error {

	return p.client.Close()
}
