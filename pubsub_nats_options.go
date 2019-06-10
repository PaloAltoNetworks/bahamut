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
	"context"
	"crypto/tls"
	"fmt"
	"time"

	"github.com/nats-io/go-nats"
)

type requestMode int

const (
	requestModeNoReply requestMode = iota
	requestModeACK
	requestModePublication
)

func (m requestMode) String() string {
	switch m {
	case requestModeNoReply:
		return "requestModeNoReply"
	case requestModeACK:
		return "requestModeACK"
	case requestModePublication:
		return "requestModePublication"
	default:
		return "unknown"
	}
}

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

var ackMessage = []byte("ack")

type natsSubscribeConfig struct {
	queueGroup string
	replier    func(msg *nats.Msg) []byte
}

type natsPublishConfig struct {
	ctx         context.Context
	requestMode requestMode
	responseCh  chan *Publication
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
//
// This option CANNOT be combined with NATSOptPublishRequireAck
func NATSOptRespondToChannel(ctx context.Context, resp chan *Publication) PubSubOptPublish {
	return func(c interface{}) {
		config := c.(*natsPublishConfig)
		if config.requestMode != requestModeNoReply {
			panic(fmt.Sprintf("illegal option: request mode has already been set to %s", config.requestMode))
		}
		config.ctx = ctx
		config.responseCh = resp
		config.requestMode = requestModePublication
	}
}

// NATSOptPublishRequireAck is a helper to require a ack in the limit
// of the given context.Context. If the other side is bahamut.PubSubClient
// using the Subscribe method, then it will automatically send back the expected
// ack. If you are using a custom replyer with NATSOptSubscribeReplyer, you MUST
// reply `ack` or implement your own logic to handle the reply by using
// the option NATSOptPublishReplyValidator.
//
// This option CANNOT be combined with NATSOptRespondToChannel
func NATSOptPublishRequireAck(ctx context.Context) PubSubOptPublish {
	return func(c interface{}) {
		config := c.(*natsPublishConfig)
		if config.requestMode != requestModeNoReply {
			panic(fmt.Sprintf("illegal option: request mode has already been set to %s", config.requestMode))
		}
		config.ctx = ctx
		config.requestMode = requestModeACK
	}
}
