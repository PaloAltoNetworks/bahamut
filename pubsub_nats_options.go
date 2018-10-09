package bahamut

import "crypto/tls"

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
