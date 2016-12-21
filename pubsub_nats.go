package bahamut

import (
	"fmt"
	"time"

	"github.com/nats-io/go-nats-streaming"

	log "github.com/Sirupsen/logrus"
)

type natsPubSub struct {
	natsURL       string
	client        stan.Conn
	retryInterval time.Duration
	clientID      string
	clusterID     string
}

// newNatsPubSub Initializes the pubsub server.
func newNatsPubSub(natsURL string, clusterID string, clientID string) *natsPubSub {

	return &natsPubSub{
		natsURL:       natsURL,
		retryInterval: 5 * time.Second,
		clientID:      clientID,
		clusterID:     clusterID,
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

	return func() { sub.Unsubscribe() }
}

func (p *natsPubSub) Connect() Waiter {

	abort := make(chan bool, 2)
	connected := make(chan bool, 2)

	go func() {

		for p.client == nil {

			var err error
			p.client, err = stan.Connect(p.clusterID, p.clientID, stan.NatsURL(p.natsURL))
			if err == nil {
				break
			}

			log.WithFields(log.Fields{
				"url":     p.natsURL,
				"package": "bahamut",
				"retryIn": p.retryInterval,
				"error":   err.Error(),
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

	p.client.Close()
}
