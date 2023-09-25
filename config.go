// Copyright 2019 Aporeto Inc.
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//     http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package bahamut

import (
	"crypto/tls"
	"crypto/x509"
	"log"
	"net"
	"net/http"
	"time"

	opentracing "github.com/opentracing/opentracing-go"
	"go.aporeto.io/elemental"
	"golang.org/x/time/rate"
)

// HealthServerFunc is the type used by the Health Server to check the health of the server.
type HealthServerFunc func() error

// HealthStatFunc is the type used by the Health Server to return additional custom health info.
type HealthStatFunc func(http.ResponseWriter, *http.Request)

// TraceCleaner is the type of function that can be used to clean a trace data
// before it is sent to OpenTracing server. You can use this to strip passwords
// or other sensitive data.
type TraceCleaner func(elemental.Identity, []byte) []byte

// An IdentifiableRetriever is the type of function you can use to perform transparent
// patch support using elemental.SparseIdentifiable.
// If this is set in the configuration, the handler for PATCH method will use
// this function to retrieve the target identifiable, will apply the patch and
// treat the request as a standard update.
type IdentifiableRetriever func(*elemental.Request) (elemental.Identifiable, error)

type apiRateLimit struct {
	limiter   *rate.Limiter
	condition func(*elemental.Request) bool
}

// A config represents the configuration of Bahamut.
type config struct {
	opentracing struct {
		tracer             opentracing.Tracer
		excludedIdentities map[string]struct{}
		onlyErrors         bool
		traceCleaner       TraceCleaner
	}
	hooks struct {
		postStart        func(Server) error
		preStop          func(Server) error
		errorTransformer func(error) error
	}
	rateLimiting struct {
		rateLimiter     *rate.Limiter
		apiRateLimiters map[elemental.Identity]apiRateLimit
	}
	security struct {
		auditer               Auditer
		corsController        CORSPolicyController
		requestAuthenticators []RequestAuthenticator
		sessionAuthenticators []SessionAuthenticator
		authorizers           []Authorizer
	}
	pushServer struct {
		service                   PubSubClient
		dispatchHandler           PushDispatchHandler
		publishHandler            PushPublishHandler
		topic                     string
		endpoint                  string
		enabled                   bool
		subjectHierarchiesEnabled bool
		publishEnabled            bool
		dispatchEnabled           bool
	}
	meta struct {
		version          map[string]any
		serviceName      string
		serviceVersion   string
		disableMetaRoute bool
	}
	profilingServer struct {
		listenAddress string
		enabled       bool
	}
	model struct {
		modelManagers              map[int]elemental.ModelManager
		unmarshallers              map[elemental.Identity]CustomUmarshaller
		marshallers                map[elemental.Identity]CustomMarshaller
		retriever                  IdentifiableRetriever
		readOnlyExcludedIdentities []elemental.Identity
		readOnly                   bool
	}
	tls struct {
		clientCAPool                    *x509.CertPool
		serverCertificatesRetrieverFunc func(*tls.ClientHelloInfo) (*tls.Certificate, error)
		peerCertificateVerifyFunc       func([][]byte, [][]*x509.Certificate) error
		serverCertificates              []tls.Certificate
		nextProtos                      []string
		authType                        tls.ClientAuthType
		disableSessionTicket            bool
	}
	healthServer struct {
		metricsManager MetricsManager
		healthHandler  HealthServerFunc
		customStats    map[string]HealthStatFunc
		listenAddress  string
		readTimeout    time.Duration
		writeTimeout   time.Duration
		idleTimeout    time.Duration
		enabled        bool
	}
	restServer struct {
		customListener        net.Listener
		customRootHandlerFunc http.HandlerFunc
		httpLogger            *log.Logger
		apiPrefix             string
		customRoutePrefix     string
		listenAddress         string
		maxConnection         int
		idleTimeout           time.Duration
		writeTimeout          time.Duration
		readTimeout           time.Duration
		enabled               bool
		disableKeepalive      bool
		disableCompression    bool
	}
	general struct{ panicRecoveryDisabled bool }
}
