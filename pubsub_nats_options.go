package bahamut

import "crypto/tls"

// A NATSOption represents an option to the pubsub backed by nats
type NATSOption func(*natsPubSub)

// NATSOptionCredentials sets the username and password to use to connect to nats.
func NATSOptionCredentials(username string, password string) NATSOption {
	return func(n *natsPubSub) {
		n.username = username
		n.password = password
	}
}

// NATSOptionClusterID sets the clusterID to use to connect to nats.
func NATSOptionClusterID(clusterID string) NATSOption {
	return func(n *natsPubSub) {
		n.clusterID = clusterID
	}
}

// NATSOptionClientID sets the client ID to use to connect to nats.
func NATSOptionClientID(clientID string) NATSOption {
	return func(n *natsPubSub) {
		n.clientID = clientID
	}
}

// NATSOptionTLS sets the tls config to use to connect nats.
func NATSOptionTLS(tlsConfig *tls.Config) NATSOption {
	return func(n *natsPubSub) {
		n.tlsConfig = tlsConfig
	}
}
