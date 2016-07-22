package pubsub

import "time"

// A PublisherSubscriber is a structure that provides a publish/subscribe mechanism.
type PublisherSubscriber interface {
	Publish(publication *Publication) error
	Subscribe(c chan *Publication, topic string) func()
	Connect() Waiter
	Disconnect()
}

// NewService returns a PublisherSubscriber.
func NewService(services []string) PublisherSubscriber {

	return newKafkaPubSub(services)
}

// A Waiter is the interface returned by Server.Connect
// that you can use to wait for the connection.
type Waiter interface {
	Wait(time.Duration) bool
}

// A connectionWaiter is the Waiter for the PubSub Server connection
type connectionWaiter struct {
	ok    chan bool
	abort chan bool
}

// Wait waits at most for the given timeout for the connection.
func (w connectionWaiter) Wait(timeout time.Duration) bool {

	select {
	case status := <-w.ok:
		return status
	case <-time.After(timeout):
		w.abort <- true
		return false
	}
}
