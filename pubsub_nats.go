// Copyright 2019 Aporeto Inc.
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//     http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package bahamut

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"time"

	"github.com/gofrs/uuid"
	nats "github.com/nats-io/nats.go"
	"go.aporeto.io/elemental"
	"go.uber.org/zap"
)

// natsClient is an interface for objects that can act as a NATS client
type natsClient interface {
	Publish(subj string, data []byte) error
	RequestWithContext(ctx context.Context, subj string, data []byte) (*nats.Msg, error)
	Subscribe(subj string, cb nats.MsgHandler) (*nats.Subscription, error)
	QueueSubscribe(subj, queue string, cb nats.MsgHandler) (*nats.Subscription, error)
	IsConnected() bool
	IsReconnecting() bool
	Flush() error
	Close()
}

type natsPubSub struct {
	natsURL         string
	client          natsClient
	retryInterval   time.Duration
	clientID        string
	clusterID       string
	password        string
	username        string
	tlsConfig       *tls.Config
	errorHandleFunc func(*nats.Conn, *nats.Subscription, error)
}

// NewNATSPubSubClient returns a new PubSubClient backend by Nats.
func NewNATSPubSubClient(natsURL string, options ...NATSOption) PubSubClient {

	n := &natsPubSub{
		natsURL:       natsURL,
		retryInterval: 5 * time.Second,
		clientID:      uuid.Must(uuid.NewV4()).String(),
		clusterID:     "test-cluster",
	}

	for _, opt := range options {
		opt(n)
	}

	return n
}

func (p *natsPubSub) Publish(publication *Publication, opts ...PubSubOptPublish) error {

	if p.client == nil {
		return errors.New("not connected to nats. messages dropped")
	}

	if publication == nil {
		return errors.New("publication cannot be nil")
	}

	config := natsPublishConfig{}
	for _, opt := range opts {
		opt(&config)
	}

	publication.ResponseMode = config.desiredResponse
	data, err := elemental.Encode(elemental.EncodingTypeMSGPACK, publication)
	if err != nil {
		return fmt.Errorf("unable to encode publication. message dropped: %s", err)
	}

	switch config.desiredResponse {
	case ResponseModeACK, ResponseModePublication:

		msg, err := p.client.RequestWithContext(config.ctx, publication.Topic, data)
		if err != nil {
			// TODO: should return a custom error type here to let the client know
			// that the request failed so client code has sufficient context if
			// it needs to build some kind of retry strategy based on the returned
			// error type.
			return err
		}

		if config.desiredResponse == ResponseModeACK {
			if !bytes.Equal(msg.Data, ackMessage) {
				return fmt.Errorf("invalid ack: %s", string(msg.Data))
			}
		}

		if config.desiredResponse == ResponseModePublication {
			responsePub := NewPublication("")
			if err := elemental.Decode(elemental.EncodingTypeMSGPACK, msg.Data, responsePub); err != nil {
				return err
			}

			config.responseCh <- responsePub
		}

		return nil

	default:
		return p.client.Publish(publication.Topic, data)
	}
}

func (p *natsPubSub) Subscribe(pubs chan *Publication, errch chan error, topic string, opts ...PubSubOptSubscribe) func() {

	config := defaultSubscribeConfig()
	for _, opt := range opts {
		opt(&config)
	}

	var sub *nats.Subscription
	var err error

	responseHandler := func(replyAddr string, pub *Publication) {
		select {
		case r := <-pub.replyCh:
			// no response should be expected for a response, therefore override this in case the caller
			// has set the response mode to something else
			r.ResponseMode = ResponseModeNone
			r.Topic = replyAddr
			if err := p.Publish(r); err != nil {
				errch <- err
			}
		// TODO: the publisher should be able to provide a response deadline for the publication to the
		// subscriber so that the subscriber knows when to give up in the event that processing has taken
		// too long and will no longer be fruitful (e.g. the client that made the request no longer cares
		// about the reply because you took too long to respond).
		case <-time.After(config.replyTimeout):
			pub.setExpired()
			errch <- fmt.Errorf("timed out waiting for response to send to subscriber on NATS subject: %s", replyAddr)
		}
	}

	handler := func(m *nats.Msg) {
		publication := NewPublication(topic)

		if e := elemental.Decode(elemental.EncodingTypeMSGPACK, m.Data, publication); e != nil {
			zap.L().Error("Unable to decode publication envelope. Message dropped.", zap.Error(e))
			return
		}

		if m.Reply != "" {
			switch publication.ResponseMode {
			// `ResponseModeACK` mode responds to the client right away, BEFORE the subscriber has had the opportunity
			// to process the publication. Most consumers do message processing asynchronously and simply need
			// to respond to the publisher with an ACK.
			case ResponseModeACK:
				if err := p.client.Publish(m.Reply, ackMessage); err != nil {
					errch <- err
					return
				}
			// `ResponseModePublication` mode is used in cases when the subscriber needs to do processing on the
			// received publication PRIOR to publishing a response back. In such case, the received publication
			// will have a write-only response channel set up so that the subscriber can send a response publication
			// to whenever it is ready. The subscriber SHOULD attempt to respond ASAP as there is a client waiting
			// for a response.
			case ResponseModePublication:
				publication.replyCh = make(chan *Publication)
				go responseHandler(m.Reply, publication)
			}
		}

		pubs <- publication
	}

	if config.queueGroup == "" {
		sub, err = p.client.Subscribe(topic, handler)
	} else {
		sub, err = p.client.QueueSubscribe(topic, config.queueGroup, handler)
	}

	if err != nil {
		errch <- err
		return func() {}
	}

	return func() { _ = sub.Unsubscribe() }
}

func (p *natsPubSub) Connect(ctx context.Context) error {

	opts := []nats.Option{}

	if p.username != "" || p.password != "" {
		opts = append(opts, nats.UserInfo(p.username, p.password))
	}

	if p.tlsConfig != nil {
		opts = append(opts, nats.Secure(p.tlsConfig))
	}

	for {

		var err error
		if p.client, err = nats.Connect(p.natsURL, opts...); err == nil {
			if p.errorHandleFunc != nil {
				nats.ErrorHandler(p.errorHandleFunc)
			}
			return nil
		}

		select {
		case <-ctx.Done():
			return fmt.Errorf("unable to connect to nats on time. last error: %s", err)
		default:
			time.Sleep(p.retryInterval)
		}
	}
}

func (p *natsPubSub) Disconnect() error {

	if err := p.client.Flush(); err != nil {
		return err
	}

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
