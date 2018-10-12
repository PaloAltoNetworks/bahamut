package bahamut

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
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
	replier    func(data []byte) []byte
}

type natsPublishConfig struct {
	ctx            context.Context
	replyValidator func(data []byte) error
}

// NATSOptSubscribeQueue sets the NATS subscribe queue.
func NATSOptSubscribeQueue(queueGroup string) PubSubOptSubscribe {
	return func(c interface{}) {
		c.(*natsSubscribeConfig).queueGroup = queueGroup
	}
}

// NATSOptSubscribeReplyer sets the function that will be called to eventually
// reply to a nats Request.
func NATSOptSubscribeReplyer(replier func(data []byte) []byte) PubSubOptSubscribe {
	return func(c interface{}) {
		c.(*natsSubscribeConfig).replier = replier
	}
}

// NATSOptPublishReplyValidator sets the NATS subscribe queue.
func NATSOptPublishReplyValidator(ctx context.Context, validator func(data []byte) error) PubSubOptPublish {
	return func(c interface{}) {
		c.(*natsPublishConfig).replyValidator = validator
		c.(*natsPublishConfig).ctx = ctx
	}
}

// NATSOptPublishRequireAck sets the NATS subscribe queue.
func NATSOptPublishRequireAck(ctx context.Context) PubSubOptPublish {
	return func(c interface{}) {
		c.(*natsPublishConfig).replyValidator = func(data []byte) error {
			if !bytes.Equal(data, ackMessage) {
				return fmt.Errorf("invalid ack: %s", string(data))
			}
			return nil
		}
		c.(*natsPublishConfig).ctx = ctx
	}
}
