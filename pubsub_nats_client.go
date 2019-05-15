package bahamut

import (
	"context"

	nats "github.com/nats-io/go-nats"
)

// NATSClient is an interface for objects that can act as a NATS client
type NATSClient interface {
	Publish(subj string, data []byte) error
	RequestWithContext(ctx context.Context, subj string, data []byte) (*nats.Msg, error)
	Subscribe(subj string, cb nats.MsgHandler) (*nats.Subscription, error)
	QueueSubscribe(subj, queue string, cb nats.MsgHandler) (*nats.Subscription, error)
	IsConnected() bool
	IsReconnecting() bool
	Flush() error
	Close()
}
