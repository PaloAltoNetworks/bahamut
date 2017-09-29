package bahamut

import (
	"crypto/tls"
	"crypto/x509"
	"time"

	uuid "github.com/satori/go.uuid"
)

// A PubSubServer is a structure that provides a publish/subscribe mechanism.
type PubSubServer interface {
	Publish(publication *Publication) error
	Subscribe(pubs chan *Publication, errors chan error, topic string, args ...interface{}) func()
	Connect() Waiter
	Disconnect() error
}

// A Pinger is PubSubServer that has a Ping method
type Pinger interface {
	Ping(timeout time.Duration) error
}

// NewNATSPubSubServer returns a PubSubServer backed by NATS.
func NewNATSPubSubServer(natsURL string, clusterID string, clientID string) PubSubServer {

	return NewNATSPubSubServerWithAuth(natsURL, clusterID, clientID, "", "")
}

// NewNATSPubSubServerWithAuth returns a PubSubServer backed by NATS using authentication.
func NewNATSPubSubServerWithAuth(natsURL string, clusterID string, clientID string, username string, password string) PubSubServer {

	if clientID == "" {
		clientID = uuid.NewV4().String()
	}

	if clusterID == "" {
		clusterID = "test-cluster"
	}

	return newNatsPubSub(natsURL, clusterID, clientID, username, password, nil, nil, nil)
}

// NewNATSPubSubServerWithTLSAuth returns a PubSubServer backed by NATS using TLS authentication.
func NewNATSPubSubServerWithTLSAuth(natsURL string, clusterID string, clientID string, username string, password string, rootCAPool *x509.CertPool, clientCAPool *x509.CertPool, clientCerts []tls.Certificate) PubSubServer {

	if clientID == "" {
		clientID = uuid.NewV4().String()
	}

	if clusterID == "" {
		clusterID = "test-cluster"
	}

	return newNatsPubSub(natsURL, clusterID, clientID, username, password, rootCAPool, clientCAPool, clientCerts)
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
