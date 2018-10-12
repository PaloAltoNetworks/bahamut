package bahamut

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"

	"github.com/nats-io/go-nats"
)

// A NATSOption represents an option to the pubsub backed by nats
type NATSOption func(*natsPubSub)

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

var ackMessage = []byte("ack")

type natsSubscribeConfig struct {
	queueGroup string
	replier    func(msg *nats.Msg) []byte
}

type natsPublishConfig struct {
	ctx            context.Context
	replyValidator func(msg *nats.Msg) error
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
		c.(*natsPublishConfig).replyValidator = func(msg *nats.Msg) error {
			if !bytes.Equal(msg.Data, ackMessage) {
				return fmt.Errorf("invalid ack: %s", string(msg.Data))
			}
			return nil
		}
		c.(*natsPublishConfig).ctx = ctx
	}
}
