package bahamut

import "time"

// A PubSubServer is a structure that provides a publish/subscribe mechanism.
type PubSubServer interface {
	Publish(publication *Publication) error
	Subscribe(pubs chan *Publication, errors chan error, topic string) func()
	Connect() Waiter
	Disconnect()
}

// NewKafkaPubSubServer returns a PubSubServer backed by Kafka.
func NewKafkaPubSubServer(services []string) PubSubServer {

	return newKafkaPubSub(services)
}

// NewLocalPubSubServer returns a PubSubServer backed by local channels.
func NewLocalPubSubServer(services []string) PubSubServer {

	return newlocalPubSub(nil)
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
