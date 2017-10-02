// Author: Antoine Mercadal
// See LICENSE file for full LICENSE
// Copyright 2016 Aporeto.

package bahamut

import (
	"crypto/tls"
	"crypto/x509"
	"net/http"
	"time"

	"github.com/aporeto-inc/elemental"
)

// HealthServerFunc is the type used by the Health Server to check the health of the server
type HealthServerFunc func() error

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

		// DisableKeepalive controls if the ReSTServer should have keepalive activated or not.
		// There is a bug in Go <= 1.7 which makes the server eats all available fd, so DisableKeepalive should
		// be set to true if you are using those versions.
		DisableKeepalive bool

		// Disabled controls if the ReSTServer should be disabled.
		Disabled bool

		// CustomRootHandlerFunc defines a custom handler func for / API.
		CustomRootHandlerFunc http.HandlerFunc
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

		// Disabled defines if the the entire websocket server should be disabled.
		// If you set this to true, other options has no effect.
		Disabled bool

		// APIDisabled defines if the ReST API system should be disabled.
		APIDisabled bool

		// PushDisabled defines if the Push system should be disabled.
		PushDisabled bool
	}

	// HealthServer contains the configuration for the Health Server.
	HealthServer struct {

		// ListenAddress is the general listening address for the health server.
		ListenAddress string

		// HealthHandler is the type of the function to run to determine the health of the server.
		HealthHandler HealthServerFunc

		// ReadTimeout defines the read http timeout.
		ReadTimeout time.Duration

		// WriteTimeout defines the write http timeout.
		WriteTimeout time.Duration

		// WriteTimeout defines the idle http timeout.
		IdleTimeout time.Duration

		// Disabled defines if the health server should be disabled.
		Disabled bool
	}

	// ProfilingServer contains information about profiling server.
	ProfilingServer struct {

		// ListenAddress is the general listening address for the profiling server.
		ListenAddress string

		// Disabled defines if the profiling server should be disabled.
		Disabled bool
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

		// EnableLetsEncrypt defines if the server should get a certificate from letsencrypt automagically.
		EnableLetsEncrypt bool

		// LetsEncryptDomainWhiteList contains the list of white listed domain name to use for
		// issuing certificates.
		LetsEncryptDomainWhiteList []string

		// LetsEncryptCertificateCacheFolder gives the path where to store certificate cache.
		// If empty, the default temp folder of the machine will be used.
		LetsEncryptCertificateCacheFolder string
	}

	// Security contains the Authenticator and Authorizer.
	Security struct {

		// RequestAuthenticator is the RequestAuthenticator to use to authenticate the requests.
		RequestAuthenticator RequestAuthenticator

		// SessionAuthenticator defines the SessionAuthenticator that will be used to
		// initially authentify a websocket connection.
		SessionAuthenticator SessionAuthenticator

		// Authorizer is the Authorizer to use to authorize the requests.
		Authorizer Authorizer

		// Auditer is the Auditer to use to audit the requests.
		Auditer Auditer
	}

	RateLimiting struct {

		// RateLimiter is the RateLimiter to use eventually limit the rate of some calls.
		RateLimiter RateLimiter
	}

	// Model contains the model configuration.
	Model struct {

		// IdentifiablesFactory is a function that returns a instance of a model
		// according to its identity.
		IdentifiablesFactory elemental.IdentifiableFactory

		// RelationshipsRegistry contains each elemental model RelationshipsRegistry for each version.
		RelationshipsRegistry map[int]elemental.RelationshipsRegistry

		// If ReadOnly is set to true, all write operations will return a Locked HTTP Code (423)
		// This is useful during maintenance.
		ReadOnly bool

		// If ReadOnly is aset to true, this will bypass the readonly mode for the set identities.
		ReadOnlyExcludedIdentities []elemental.Identity
	}
}
