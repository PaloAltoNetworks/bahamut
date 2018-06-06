package bahamut

import (
	"crypto/tls"
	"crypto/x509"
	"time"

	uuid "github.com/satori/go.uuid"
)

// A PubSubClient is a structure that provides a publish/subscribe mechanism.
type PubSubClient interface {
	Publish(publication *Publication) error
	Subscribe(pubs chan *Publication, errors chan error, topic string, args ...interface{}) func()
	Connect() Waiter
	Disconnect() error
}

// NewNATSPubSubClient returns a PubSubClient backed by NATS.
func NewNATSPubSubClient(natsURL string, clusterID string, clientID string) PubSubClient {

	return NewNATSPubSubClientWithAuth(natsURL, clusterID, clientID, "", "")
}

// NewNATSPubSubClientWithAuth returns a PubSubClient backed by NATS using authentication.
func NewNATSPubSubClientWithAuth(natsURL string, clusterID string, clientID string, username string, password string) PubSubClient {

	if clientID == "" {
		clientID = uuid.NewV4().String()
	}

	if clusterID == "" {
		clusterID = "test-cluster"
	}

	return newNatsPubSub(natsURL, clusterID, clientID, username, password, nil, nil, nil)
}

// NewNATSPubSubClientWithTLSAuth returns a PubSubClient backed by NATS using TLS authentication.
func NewNATSPubSubClientWithTLSAuth(natsURL string, clusterID string, clientID string, username string, password string, rootCAPool *x509.CertPool, clientCAPool *x509.CertPool, clientCerts []tls.Certificate) PubSubClient {

	if clientID == "" {
		clientID = uuid.NewV4().String()
	}

	if clusterID == "" {
		clusterID = "test-cluster"
	}

	return newNatsPubSub(natsURL, clusterID, clientID, username, password, rootCAPool, clientCAPool, clientCerts)
}

// NewLocalPubSubClient returns a PubSubClient backed by local channels.
func NewLocalPubSubClient(services []string) PubSubClient {

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
