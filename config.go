// Author: Antoine Mercadal
// See LICENSE file for full LICENSE
// Copyright 2016 Aporeto.

package bahamut

import (
	"crypto/tls"
	"crypto/x509"
	"net/http"
	"time"

	"go.aporeto.io/elemental"
)

// HealthServerFunc is the type used by the Health Server to check the health of the server.
type HealthServerFunc func() error

// HealthStatFunc is the type used by the Health Server to return additional custom health info.
type HealthStatFunc func(http.ResponseWriter, *http.Request)

// A config represents the configuration of Bahamut.
type config struct {
	general struct {
		panicRecoveryDisabled bool
	}

	restServer struct {
		listenAddress         string
		readTimeout           time.Duration
		writeTimeout          time.Duration
		idleTimeout           time.Duration
		disableKeepalive      bool
		enabled               bool
		customRootHandlerFunc http.HandlerFunc
	}

	pushServer struct {
		service         PubSubClient
		topic           string
		dispatchHandler PushDispatchHandler
		publishHandler  PushPublishHandler
		enabled         bool
		publishEnabled  bool
		dispatchEnabled bool
	}

	healthServer struct {
		listenAddress string
		healthHandler HealthServerFunc
		readTimeout   time.Duration
		writeTimeout  time.Duration
		idleTimeout   time.Duration
		enabled       bool
		customStats   map[string]HealthStatFunc
	}

	profilingServer struct {
		listenAddress    string
		enabled          bool
		mode             string
		gcpProjectID     string
		gcpServicePrefix string
	}

	tls struct {
		clientCAPool                    *x509.CertPool
		authType                        tls.ClientAuthType
		serverCertificates              []tls.Certificate
		serverCertificatesRetrieverFunc func(*tls.ClientHelloInfo) (*tls.Certificate, error)
	}

	security struct {
		requestAuthenticators []RequestAuthenticator
		sessionAuthenticators []SessionAuthenticator
		authorizers           []Authorizer
		auditer               Auditer
	}

	rateLimiting struct {
		rateLimiter RateLimiter
	}

	model struct {
		modelManagers              map[int]elemental.ModelManager
		readOnly                   bool
		readOnlyExcludedIdentities []elemental.Identity
		unmarshallers              map[elemental.Identity]CustomUmarshaller
	}

	meta struct {
		serviceName      string
		serviceVersion   string
		version          map[string]interface{}
		disableMetaRoute bool
	}
}
