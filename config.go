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
	General struct {
		PanicRecoveryDisabled bool
	}

	ReSTServer struct {
		ListenAddress         string
		ReadTimeout           time.Duration
		WriteTimeout          time.Duration
		IdleTimeout           time.Duration
		DisableKeepalive      bool
		Disabled              bool
		CustomRootHandlerFunc http.HandlerFunc
	}

	PushServer struct {
		Service          PubSubServer
		Topic            string
		DispatchHandler  PushDispatchHandler
		PublishHandler   PushPublishHandler
		Disabled         bool
		PublishDisabled  bool
		DispatchDisabled bool
	}

	HealthServer struct {
		ListenAddress string
		HealthHandler HealthServerFunc
		ReadTimeout   time.Duration
		WriteTimeout  time.Duration
		IdleTimeout   time.Duration
		Disabled      bool
	}

	ProfilingServer struct {
		ListenAddress    string
		Enabled          bool
		Mode             string
		GCPProjectID     string
		GCPServicePrefix string
	}

	TLS struct {
		ClientCAPool                      *x509.CertPool
		AuthType                          tls.ClientAuthType
		ServerCertificates                []tls.Certificate
		ServerCertificatesRetrieverFunc   func(*tls.ClientHelloInfo) (*tls.Certificate, error)
		EnableLetsEncrypt                 bool
		LetsEncryptDomainWhiteList        []string
		LetsEncryptCertificateCacheFolder string

		RootCAPool *x509.CertPool // REMOVE ME
	}

	Security struct {
		RequestAuthenticators []RequestAuthenticator
		SessionAuthenticators []SessionAuthenticator
		Authorizers           []Authorizer
		Auditer               Auditer
	}

	RateLimiting struct {
		RateLimiter RateLimiter
	}

	Model struct {
		IdentifiablesFactory       elemental.IdentifiableFactory
		RelationshipsRegistry      map[int]elemental.RelationshipsRegistry
		ReadOnly                   bool
		ReadOnlyExcludedIdentities []elemental.Identity
		Unmarshallers              map[elemental.Identity]CustomUmarshaller
	}

	MockServer struct {
		ListenAddress string
		ReadTimeout   time.Duration
		WriteTimeout  time.Duration
		IdleTimeout   time.Duration
		Enabled       bool
	}

	Meta struct {
		ServiceName      string
		ServiceVersion   string
		Version          map[string]interface{}
		DisableMetaRoute bool
	}
}
