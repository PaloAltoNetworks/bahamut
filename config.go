// Author: Antoine Mercadal
// See LICENSE file for full LICENSE
// Copyright 2016 Aporeto.

package bahamut

import (
	"crypto/tls"
	"crypto/x509"
	"time"

	"github.com/aporeto-inc/elemental"
)

// A Config represents the configuration of Bahamut.
type Config struct {

	// ReSTServer contains the configuration for the ReST Server.
	ReSTServer struct {

		// ListenAddress is the general listening address for the API server as
		// well as the PushServer.
		ListenAddress string

		// ReadTimeout defines the read http timeout.
		ReadTimeout time.Duration

		// WriteTimeout defines the write http timeout.
		WriteTimeout time.Duration

		// WriteTimeout defines the idle http timeout.
		IdleTimeout time.Duration

		// Disabled controls if the ReSTServer should be disabled.
		Disabled bool
	}

	// WebSocketServer contains the configuration for the WebSocket Server.
	WebSocketServer struct {

		// Service defines the pubsub server to use.
		Service PubSubServer

		// Topic defines the default notification topic to use.
		Topic string

		// SessionsHandler defines the handler that will be used to
		// manage push session lifecycle.
		SessionsHandler PushSessionsHandler

		// Disabled defines if the ReST API system should be disabled.
		Disabled bool
	}

	// Profiling contains information about profiling server.
	Profiling struct {

		// Enabled defines if the profiling server should be created.
		Enabled bool

		// ListenAddress is the custom listening address to use.
		// It will be only used if Profiling.Enabled is set to true.
		ListenAddress string
	}

	// TLS contains the TLS configuration.
	TLS struct {

		// RootCAPool is the *x509.CertPool to use for the secure bahamut api server.
		RootCAPool *x509.CertPool

		// ClientCAPool is the *x509.CertPool to use for the authentifying client.
		ClientCAPool *x509.CertPool

		// ServerCertificates are the TLS certficates to use for the secure api server.
		ServerCertificates []tls.Certificate

		// AuthType defines the tls authentication mode to use for a secure server.
		AuthType tls.ClientAuthType
	}

	// Security contains the Authenticator and Authorizer.
	Security struct {

		// Authenticator is the Authenticator to use to authenticate the requests.
		Authenticator Authenticator

		// Authorizer is the Authorizer to use to authorize the requests.
		Authorizer Authorizer
	}

	// Model contains the model configuration.
	Model struct {

		// IdentifiablesFactory is a function that returns a instance of a model
		// according to its identity.
		IdentifiablesFactory func(identity string) elemental.Identifiable

		// RelationshipsRegistry contains the elemental model RelationshipsRegistry.
		RelationshipsRegistry elemental.RelationshipsRegistry
	}
}
