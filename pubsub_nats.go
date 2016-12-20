package bahamut

import (
	"fmt"
	"time"

	"github.com/nats-io/go-nats"

	log "github.com/Sirupsen/logrus"
)

type natsPubSub struct {
	natsURL       string
	client        *nats.Conn
	retryInterval time.Duration
}

// newNatsPubSub Initializes the pubsub server.
func newNatsPubSub(natsURL string) *natsPubSub {

	return &natsPubSub{
		natsURL:       natsURL,
		retryInterval: 5 * time.Second,
	}
}

func (p *natsPubSub) Publish(publication *Publication) error {

	if p.client == nil {
		return fmt.Errorf("Not connected to nats. Messages dropped.")
	}

	return p.client.Publish(publication.Topic, publication.data)
}

func (p *natsPubSub) Subscribe(pubs chan *Publication, errors chan error, topic string, args ...interface{}) func() {

	var queueGroup string
	if len(args) == 1 {

		if q, ok := args[0].(string); ok {
			queueGroup = q
		} else {
			panic("You must provide a string as queue group name")
		}
	}

	var sub *nats.Subscription
	var err error

	handler := func(m *nats.Msg) {
		publication := NewPublication(topic)
		publication.data = m.Data
		pubs <- publication
	}

	if queueGroup == "" {
		sub, err = p.client.Subscribe(topic, handler)
	} else {
		sub, err = p.client.QueueSubscribe(topic, queueGroup, handler)
	}

	if err != nil {
		errors <- err
		return func() {}
	}

	return func() { sub.Unsubscribe() }
}

func (p *natsPubSub) Connect() Waiter {

	abort := make(chan bool, 2)
	connected := make(chan bool, 2)

	go func() {

		for p.client == nil {

			var err error
			p.client, err = nats.Connect(p.natsURL)
			if err == nil {
				break
			}

			log.WithFields(log.Fields{
				"url":     p.natsURL,
				"package": "bahamut",
				"retryIn": p.retryInterval,
			}).Warn("Unable to connect to nats server. retrying in 5 seconds.")

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

func (p *natsPubSub) Disconnect() {
}
