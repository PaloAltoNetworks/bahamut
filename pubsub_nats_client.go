package bahamut

import (
	"context"

	nats "github.com/nats-io/nats.go"
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
