package bahamut

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"time"

	"github.com/nats-io/go-nats"
	"go.uber.org/zap"
)

type natsPubSub struct {
	natsURL        string
	client         *nats.Conn
	retryInterval  time.Duration
	publishTimeout time.Duration
	retryNumber    int
	clientID       string
	clusterID      string
	password       string
	username       string
	clientCerts    []tls.Certificate
	rootCAPool     *x509.CertPool
	clientCAPool   *x509.CertPool
}

// newNatsPubSub Initializes the pubsub server.
func newNatsPubSub(natsURL string, clusterID string, clientID string, username string, password string, rootCAPool *x509.CertPool, clientCAPool *x509.CertPool, clientCerts []tls.Certificate) *natsPubSub {

	return &natsPubSub{
		natsURL:        natsURL,
		retryInterval:  5 * time.Second,
		publishTimeout: 8 * time.Second,
		retryNumber:    5,
		clientID:       clientID,
		clusterID:      clusterID,
		username:       username,
		password:       password,
		clientCerts:    clientCerts,
		rootCAPool:     rootCAPool,
		clientCAPool:   clientCAPool,
	}
}

func (p *natsPubSub) Publish(publication *Publication) error {

	if p.client == nil {
		return fmt.Errorf("not connected to nats. messages dropped")
	}

	data, err := json.Marshal(publication)
	if err != nil {
		return fmt.Errorf("unable to encode publication. message dropped: %s", err)
	}

	zap.L().Debug("Publishing message in nats",
		zap.String("topic", publication.Topic),
		zap.String("natsURL", p.natsURL),
		zap.ByteString("data", publication.Data),
	)

	return p.client.Publish(publication.Topic, data)
}

func (p *natsPubSub) Subscribe(pubs chan *Publication, errors chan error, topic string, args ...interface{}) func() {

	var queueGroup string

	for i, arg := range args {
		if i == 0 {
			if q, ok := arg.(string); ok {
				queueGroup = q
			} else {
				panic("You must provide a string as queue group name")
			}
			continue
		}
	}

	var sub *nats.Subscription
	var err error

	handler := func(m *nats.Msg) {
		publication := NewPublication(topic)
		if e := json.Unmarshal(m.Data, publication); e != nil {
			zap.L().Error("Unable to decode publication envelope. Message dropped.", zap.Error(e))
			return
		}
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

	return func() { _ = sub.Unsubscribe() }
}

func (p *natsPubSub) Connect() Waiter {

	abort := make(chan struct{})
	connected := make(chan bool)

	go func() {

		// First, we create a connection to the nats cluster.
		for p.client == nil {

			var err error

			tlsConfig := &tls.Config{
				Certificates: p.clientCerts,
				RootCAs:      p.rootCAPool,
				ClientCAs:    p.clientCAPool,
			}

			if p.username != "" || p.password != "" {
				p.client, err = nats.Connect(p.natsURL, nats.UserInfo(p.username, p.password), nats.Secure(tlsConfig))
			} else {
				p.client, err = nats.Connect(p.natsURL, nats.Secure(tlsConfig))
			}

			if err == nil {
				break
			}

			zap.L().Warn("Unable to connect to nats cluster. Retrying",
				zap.String("url", p.natsURL),
				zap.Duration("retry", p.retryInterval),
				zap.Error(err),
			)

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

	p.client.Close()
	return nil
}

func (p *natsPubSub) Ping(timeout time.Duration) error {

	errChannel := make(chan error)

	go func() {
		if p.client.IsConnected() {
			errChannel <- nil
		} else if p.client.IsReconnecting() {
			errChannel <- fmt.Errorf("reconnecting")
		} else {
			errChannel <- fmt.Errorf("connection closed")
		}
	}()

	select {
	case <-time.After(timeout):
		return fmt.Errorf("connection timeout")
	case err := <-errChannel:
		return err
	}
}
