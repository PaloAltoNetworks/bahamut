package bahamut

import (
	"time"
)

// A PubSubClient is a structure that provides a publish/subscribe mechanism.
type PubSubClient interface {
	Publish(publication *Publication) error
	Subscribe(pubs chan *Publication, errors chan error, topic string, args ...interface{}) func()
	Connect() Waiter
	Disconnect() error
}

// A Waiter is the interface returned by Server.Connect
// that you can use to wait for the connection.
type Waiter interface {
	Wait(time.Duration) bool
}

// A connectionWaiter is the Waiter for the PubSub Server connection
type connectionWaiter struct {
	ok    chan bool
	abort chan struct{}
}

// Wait waits at most for the given timeout for the connection.
func (w connectionWaiter) Wait(timeout time.Duration) bool {

	select {
	case status := <-w.ok:
		return status
	case <-time.After(timeout):
		close(w.abort)
		return false
	}
}
