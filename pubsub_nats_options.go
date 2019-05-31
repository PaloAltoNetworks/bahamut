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
	"fmt"
	"time"

	nats "github.com/nats-io/go-nats"
)

// A NATSOption represents an option to the pubsub backed by nats
type NATSOption func(*natsPubSub)

// NATSOptConnectRetryInterval sets the connection retry interval
func NATSOptConnectRetryInterval(interval time.Duration) NATSOption {
	return func(n *natsPubSub) {
		n.retryInterval = interval
	}
}

// NATSOptCredentials sets the username and password to use to connect to nats.
func NATSOptCredentials(username string, password string) NATSOption {
	return func(n *natsPubSub) {
		n.username = username
		n.password = password
	}
}

// NATSOptClusterID sets the clusterID to use to connect to nats.
func NATSOptClusterID(clusterID string) NATSOption {
	return func(n *natsPubSub) {
		n.clusterID = clusterID
	}
}

// NATSOptClientID sets the client ID to use to connect to nats.
func NATSOptClientID(clientID string) NATSOption {
	return func(n *natsPubSub) {
		n.clientID = clientID
	}
}

// NATSOptTLS sets the tls config to use to connect nats.
func NATSOptTLS(tlsConfig *tls.Config) NATSOption {
	return func(n *natsPubSub) {
		n.tlsConfig = tlsConfig
	}
}

// natsOptClient sets the NATS client that will be used
// This is useful for unit testing as you can pass in a mocked NATS client
func natsOptClient(client natsClient) NATSOption {
	return func(n *natsPubSub) {
		n.client = client
	}
}

var (
	ackRequest = []byte("ack-request")
	ackMessage = []byte("ack")
)

type natsSubscribeConfig struct {
	queueGroup string
	replier    func(msg *nats.Msg) []byte
}

type natsPublishConfig struct {
	ctx            context.Context
	replyValidator func(msg *nats.Msg) error

	sendAckReq     bool
	useRequestMode bool
	responseCh     chan *Publication
}

// NATSOptSubscribeQueue sets the NATS subscriber queue group.
// In short, this allows to ensure only one subscriber in the
// queue group with the same name will receive the publication.
//
// See: https://nats.io/documentation/concepts/nats-queueing/
func NATSOptSubscribeQueue(queueGroup string) PubSubOptSubscribe {
	return func(c interface{}) {
		c.(*natsSubscribeConfig).queueGroup = queueGroup
	}
}

// NATSOptSubscribeReplyer sets the function that will be called to eventually
// reply to a nats Request.
//
// This is advanced option. You will not receive a
// bahamut.Publication, but the raw nats.Msg received, and you must
// reply with a direct payload []byte. If the reply cannot be sent, the subscription
// channel will not receive anything and it is considered as a terminal error for that
// message.
//
// See: https://nats.io/documentation/concepts/nats-req-rep/
func NATSOptSubscribeReplyer(replier func(msg *nats.Msg) []byte) PubSubOptSubscribe {
	return func(c interface{}) {
		c.(*natsSubscribeConfig).replier = replier
	}
}

// NATSOptRespondToChannel will send the *Publication received to the provided channel.
//
// This is an advanced option which is useful in situations where you want to block until
// you receive a response. The context parameter allows you to provide a deadline on how long
// you should wait before considering the request as a failure:
//
//		myCtx, _ := context.WithTimeout(context.Background(), 5*time.Second)
//		respCh := make(chan *Publication)
// 		publishOption := NATSOptRespondToChannel(myCtx, respCh)
func NATSOptRespondToChannel(ctx context.Context, resp chan *Publication) PubSubOptPublish {
	return func(c interface{}) {
		c.(*natsPublishConfig).useRequestMode = true
		c.(*natsPublishConfig).ctx = ctx
		c.(*natsPublishConfig).responseCh = resp
	}
}

// DEPRECATED: use NATSOptPublishRequireAck and/or NATSOptRespondToChannel instead;
// this option will be removed in a future release
//
// NATSOptPublishReplyValidator sets the function that will be called to validate
// a request reply.
//
// This is advanced option. You will not receive a
// bahamut.Publication, but the raw nats.Msg.Data received as response. If you return
// an error, the entire Publish process will be considered errored, and the caller
// will receive you error as is.
//
// See: https://nats.io/documentation/concepts/nats-req-rep/
func NATSOptPublishReplyValidator(ctx context.Context, validator func(msg *nats.Msg) error) PubSubOptPublish {
	return func(c interface{}) {
		c.(*natsPublishConfig).useRequestMode = true
		c.(*natsPublishConfig).replyValidator = validator
		c.(*natsPublishConfig).ctx = ctx
	}
}

// NATSOptPublishRequireAck is a helper to require a ack in the limit
// of the given context.Context. If the other side is bahamut.PubSubClient
// using the Subscribe method, then it will automatically send back the expected
// ack. If you are using a custom replyer with NATSOptSubscribeReplyer, you MUST
// reply `ack` or implement your own logic to handle the reply by using
// the option NATSOptPublishReplyValidator.
func NATSOptPublishRequireAck(ctx context.Context) PubSubOptPublish {
	return func(c interface{}) {
		c.(*natsPublishConfig).sendAckReq = true
		c.(*natsPublishConfig).replyValidator = ackValidator
		c.(*natsPublishConfig).ctx = ctx
	}
}

func ackValidator(msg *nats.Msg) error {
	if !bytes.Equal(msg.Data, ackMessage) {
		return fmt.Errorf("invalid ack: %s", string(msg.Data))
	}
	return nil
}
