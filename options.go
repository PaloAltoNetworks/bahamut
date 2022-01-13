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
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	opentracing "github.com/opentracing/opentracing-go"
	"go.aporeto.io/elemental"
	"golang.org/x/time/rate"
)

// An Option represents a configuration option.
type Option func(*config)

// OptDisablePanicRecovery disables panic recovery.
func OptDisablePanicRecovery() Option {
	return func(c *config) {
		c.general.panicRecoveryDisabled = true
	}
}

// OptRestServer configures the listening address of the server.
//
// listen is the general listening address for the API server as
func OptRestServer(listen string) Option {
	return func(c *config) {
		c.restServer.enabled = true
		c.restServer.listenAddress = listen
	}
}

// OptCustomListener allows to set a custom listener for the api server..
func OptCustomListener(listener net.Listener) Option {
	return func(c *config) {
		c.restServer.customListener = listener
	}
}

// OptMaxConnection sets the maximum number of concurrent
// connection to the server. 0, which is the default, means
// no limit.
func OptMaxConnection(n int) Option {
	return func(c *config) {
		c.restServer.maxConnection = n
	}
}

// OptTimeouts configures the timeouts of the server.
func OptTimeouts(read, write, idle time.Duration) Option {
	return func(c *config) {
		c.restServer.readTimeout = read
		c.restServer.writeTimeout = write
		c.restServer.idleTimeout = idle
	}
}

// OptDisableKeepAlive disables http keepalives.
//
// There is a bug in Go <= 1.7 which makes the server eats all available
// fds. Use this option if you are using these versions.
func OptDisableKeepAlive() Option {
	return func(c *config) {
		c.restServer.disableKeepalive = true
	}
}

// OptDisableCompression disables HTTP gzip compression.
//
// This can be useful when your servers run behind
// a proxy for instance.
func OptDisableCompression() Option {
	return func(c *config) {
		c.restServer.disableCompression = true
	}
}

// OptCustomRootHandler configures the custom root (/) handler.
func OptCustomRootHandler(handler http.HandlerFunc) Option {
	return func(c *config) {
		c.restServer.customRootHandlerFunc = handler
	}
}

// OptHTTPLogger sets the logger to be used internally
// by the underlying Go HTTP server.
func OptHTTPLogger(l *log.Logger) Option {
	return func(c *config) {
		c.restServer.httpLogger = l
	}
}

// OptEnableCustomRoutePathPrefix enables custom routes in the server that
// start with the given prefix. A user must also provide an API
// prefix in this case and the two must not overlap. Otherwise,
// the configuration will be rejected.
func OptEnableCustomRoutePathPrefix(prefix string) Option {
	return func(c *config) {
		u, err := url.ParseRequestURI(path.Clean(prefix))

		if err != nil {
			panic(fmt.Sprintf("Invalid custom route prefix provided: %s error: %s", prefix, err))
		}
		if u.Host != "" || u.Scheme != "" {
			panic(fmt.Sprintf("Custom route prefix must not include host or scheme: host: %s, scheme: %s", u.Host, u.Scheme))
		}

		c.restServer.customRoutePrefix = u.Path
	}
}

// OptEnableAPIPathPrefix enables a path prefix for all API resources
// other than root. This enables customization of the paths of the API
// endpoints.
func OptEnableAPIPathPrefix(prefix string) Option {
	return func(c *config) {
		u, err := url.ParseRequestURI(path.Clean(prefix))

		if err != nil {
			panic(fmt.Sprintf("Invalid API prefix provided: %s error: %s", prefix, err))
		}
		if u.Host != "" || u.Scheme != "" {
			panic(fmt.Sprintf("API route prefix must not include host or scheme"))
		}

		c.restServer.apiPrefix = u.Path
	}
}

// OptPushServer enables and configures the push server.
//
// Service defines the pubsub server to use.
// Topic defines the default notification topic to use.
// DispatchHandler defines the handler that will be used to
// decide if a push event should be dispatch to push sessions.
// PublishHandler defines the handler that will be used to
// decide if an event should be published.
func OptPushServer(service PubSubClient, topic string) Option {
	return func(c *config) {
		c.pushServer.enabled = true
		c.pushServer.service = service
		c.pushServer.topic = topic
	}
}

// OptPushServerEnableSubjectHierarchies will cause the push server to push to specific subject hierarchies under the configured
// pub/sub topic you have chosen for your push server. This option has no effect if OptPushServer is not set.
//
// For example:
//
//   If the push server topic has been set to "global-events" and the server is about to push a "create" event w/ an identity
//   value of "apples", enabling this option, would cause the push server to target a new publication to the subject
//   "global-events.apples.create", INSTEAD OF "global-events". Consequently, as a result of this, any upstream push
//   servers that are interested in receiving all events you publish to this topic would need to utilize subject wildcards.
//
//   See: https://docs.nats.io/nats-concepts/subjects#wildcards for more details.
func OptPushServerEnableSubjectHierarchies() Option {
	return func(c *config) {
		c.pushServer.subjectHierarchiesEnabled = true
	}
}

// OptPushEndpoint sets the endpoint to use for websocket channel.
//
// If unset, it fallsback to the default which is /events. This option
// has not effect if OptPushServer is not set.
func OptPushEndpoint(endpoint string) Option {
	return func(c *config) {
		c.pushServer.endpoint = endpoint
	}
}

// OptPushDispatchHandler configures the push dispatcher.
//
// DispatchHandler defines the handler that will be used to
// decide if a push event should be dispatch to push sessions.
func OptPushDispatchHandler(dispatchHandler PushDispatchHandler) Option {
	return func(c *config) {
		c.pushServer.dispatchEnabled = true
		c.pushServer.dispatchHandler = dispatchHandler
	}
}

// OptPushPublishHandler configures the push publisher.
//
// PublishHandler defines the handler that will be used to
// decide if an event should be published.
func OptPushPublishHandler(publishHandler PushPublishHandler) Option {
	return func(c *config) {
		c.pushServer.publishEnabled = true
		c.pushServer.publishHandler = publishHandler
	}
}

// OptHealthServer enables and configures the health server.
//
// ListenAddress is the general listening address for the health server.
// HealthHandler is the type of the function to run to determine the health of the server.
func OptHealthServer(listen string, handler HealthServerFunc) Option {
	return func(c *config) {
		c.healthServer.enabled = true
		c.healthServer.listenAddress = listen
		c.healthServer.healthHandler = handler
	}
}

// OptHealthServerMetricsManager sets the MetricManager in the health server.
//
// This option has no effect if the health server is not enabled.
func OptHealthServerMetricsManager(manager MetricsManager) Option {
	return func(c *config) {
		c.healthServer.metricsManager = manager
	}
}

// OptHealthCustomStats configures additional stats handler.
//
// The healt server must be enabled using OptHealthServer or this option
// will have no effect. Parameter handlers is a map where the key
// will be used as the path in the health server. They must not start
// with an `_`, contain any `/`, be empty or the function will panic. If key
// contains a nil function, it will also panic.
func OptHealthCustomStats(handlers map[string]HealthStatFunc) Option {

	for k, f := range handlers {

		if k == "" {
			panic("key must not be empty")
		}

		if strings.HasPrefix(k, "_") {
			panic(fmt.Sprintf("key '%s' must not start with an '_'", k))
		}

		if strings.Contains(k, "/") {
			panic(fmt.Sprintf("key '%s' must not contain with any '/'", k))
		}

		if f == nil {
			panic(fmt.Sprintf("stat function for key '%s' must not be nil", k))
		}
	}

	return func(c *config) {
		c.healthServer.customStats = handlers
	}
}

// OptHealthServerTimeouts configures the health server timeouts.
func OptHealthServerTimeouts(read, write, idle time.Duration) Option {
	return func(c *config) {
		c.healthServer.readTimeout = read
		c.healthServer.writeTimeout = write
		c.healthServer.idleTimeout = idle
	}
}

// OptProfilingLocal configure local goops profiling.
func OptProfilingLocal(listen string) Option {
	return func(c *config) {
		c.profilingServer.enabled = true
		c.profilingServer.listenAddress = listen
	}
}

// OptTLS configures server TLS.
//
// ServerCertificates are the TLS certficates to use for the secure api server.
// If you set ServerCertificatesRetrieverFunc, the value of ServerCertificates will be ignored.
// ServerCertificatesRetrieverFunc is standard tls GetCertifcate function to use to
// retrieve the server certificates dynamically.
// - If you set this, the value of ServerCertificates will be ignored.
// - If EnableLetsEncrypt is set, this will be ignored
func OptTLS(certs []tls.Certificate, certRetriever func(*tls.ClientHelloInfo) (*tls.Certificate, error)) Option {
	return func(c *config) {
		c.tls.serverCertificates = certs
		c.tls.serverCertificatesRetrieverFunc = certRetriever
	}
}

// OptTLSNextProtos configures server TLS next protocols.
//
// You can use it to set it to []string{'h2'} for instance to
// enable http2
func OptTLSNextProtos(nextProtos []string) Option {
	return func(c *config) {
		c.tls.nextProtos = nextProtos
	}
}

// OptMTLS configures the tls client authentication mechanism.
//
// ClientCAPool is the *x509.CertPool to use for the authentifying client.
// AuthType defines the tls authentication mode to use for a secure server.
func OptMTLS(caPool *x509.CertPool, authType tls.ClientAuthType) Option {
	return func(c *config) {
		c.tls.clientCAPool = caPool
		c.tls.authType = authType
	}
}

// OptMTLSVerifyPeerCertificate configures the optionnal function to
// to perform custom peer certificate verification.
func OptMTLSVerifyPeerCertificate(f func([][]byte, [][]*x509.Certificate) error) Option {
	return func(c *config) {
		c.tls.peerCertificateVerifyFunc = f
	}
}

// OptTLSDisableSessionTicket controls if the TLS session tickets should
// be disabled.
func OptTLSDisableSessionTicket(disabled bool) Option {
	return func(c *config) {
		c.tls.disableSessionTicket = disabled
	}
}

// OptAuthenticators configures the authenticators.
//
// RequestAuthenticators defines the list the RequestAuthenticator to use to authenticate the requests.
// They are executed in order from index 0 to index n. They will return a bahamut.AuthAction to tell if
// the current request authenticator grants, denies or let the chain continue. If an error is returned, the
// chain fails immediately.
// SessionAuthenticators defines the list of SessionAuthenticator that will be used to
// initially authentify a websocket connection.
// They are executed in order from index 0 to index n.They will return a bahamut.AuthAction to tell if
// the current session authenticator grants, denies or let the chain continue. If an error is returned, the
// chain fails immediately.
func OptAuthenticators(requestAuthenticators []RequestAuthenticator, sessionAuthenticators []SessionAuthenticator) Option {
	return func(c *config) {
		c.security.requestAuthenticators = requestAuthenticators
		c.security.sessionAuthenticators = sessionAuthenticators
	}
}

// OptAuthorizers configures the authorizers.
//
// Authorizers defines the list Authorizers to use to authorize the requests.
// They are executed in order from index 0 to index n. They will return a bahamut.AuthAction to tell if
// the current authorizer grants, denies or let the chain continue. If an error is returned, the
// chain fails immediately.
func OptAuthorizers(authorizers []Authorizer) Option {
	return func(c *config) {
		c.security.authorizers = authorizers
	}
}

// OptAuditer configures the auditor to use to audit the requests.
//
// The Audit() method will be run in a go routine so there is no
// need to deal with it in your implementation.
func OptAuditer(auditer Auditer) Option {
	return func(c *config) {
		c.security.auditer = auditer
	}
}

// OptCORSAccessControl configures CORS access control policy.
//
// By default, no CORS headers are injected by bahamut.
// You can use NewDefaultCORSAccessControler to get a sensible
// default policy.
func OptCORSAccessControl(controller CORSPolicyController) Option {
	return func(c *config) {
		c.security.corsController = controller
	}
}

// OptRateLimiting configures the global rate limiting.
func OptRateLimiting(limit float64, burst int) Option {
	return func(c *config) {
		c.rateLimiting.rateLimiter = rate.NewLimiter(rate.Limit(limit), burst)
	}
}

// OptAPIRateLimiting configures the per-api rate limiting.
// The optional parameter condition is a function that can be provided
// to decide if the rate limiter should apply based on the custom computation
// iof the incoming request.
func OptAPIRateLimiting(identity elemental.Identity, limit float64, burst int, condition func(*elemental.Request) bool) Option {
	return func(c *config) {
		if c.rateLimiting.apiRateLimiters == nil {
			c.rateLimiting.apiRateLimiters = map[elemental.Identity]apiRateLimit{}
		}

		c.rateLimiting.apiRateLimiters[identity] = apiRateLimit{
			limiter:   rate.NewLimiter(rate.Limit(limit), burst),
			condition: condition,
		}
	}
}

// OptModel configures the elemental Model for the server.
//
// modelManagers is a map of version to elemental.ModelManager.
// according to its identity.
func OptModel(modelManagers map[int]elemental.ModelManager) Option {
	return func(c *config) {
		c.model.modelManagers = modelManagers
	}
}

// OptReadOnly sets the server in read only mode.
//
// All write operations will return a Locked HTTP Code (423)
// This is useful during maintenance.
// Excluded defines a list of elemental.Identity that will not be affected
// by the read only mode.
func OptReadOnly(excluded []elemental.Identity) Option {
	return func(c *config) {
		c.model.readOnly = true
		c.model.readOnlyExcludedIdentities = excluded
	}
}

// OptUnmarshallers sets the custom unmarshallers.
//
// Unmarshallers contains a list of custom umarshaller per identity.
// This allows to create custom function to umarshal the payload of a request.
// If none is provided for a particular identity, the standard unmarshal function
// is used.
func OptUnmarshallers(unmarshallers map[elemental.Identity]CustomUmarshaller) Option {
	return func(c *config) {
		c.model.unmarshallers = unmarshallers
	}
}

// OptMarshallers sets the custom marshallers.
//
// Marshallers contains a list of custom marshaller per identity.
// This allows to create custom function to marshal the payload of a response.
// If none is provided for a particular identity, the standard unmarshal function
// is used.
func OptMarshallers(marshallers map[elemental.Identity]CustomMarshaller) Option {
	return func(c *config) {
		c.model.marshallers = marshallers
	}
}

// OptServiceInfo configures the service basic information.
//
// ServiceName contains the name of the service.
// ServiceVersion contains the version of the service itself.
// Version should contain information relative to the service version.
// like all it's libraries and things like that.
func OptServiceInfo(name string, version string, subversions map[string]interface{}) Option {
	return func(c *config) {
		c.meta.serviceName = name
		c.meta.serviceVersion = version
		c.meta.version = subversions
	}
}

// OptDisableMetaRoutes disables the meta routing.
func OptDisableMetaRoutes() Option {
	return func(c *config) {
		c.meta.disableMetaRoute = true
	}
}

// OptOpentracingTracer sets the opentracing.Tracer to use.
func OptOpentracingTracer(tracer opentracing.Tracer) Option {
	return func(c *config) {
		c.opentracing.tracer = tracer
	}
}

// OptOpentracingExcludedIdentities excludes the given identity from being traced.
func OptOpentracingExcludedIdentities(identities []elemental.Identity) Option {
	return func(c *config) {
		c.opentracing.excludedIdentities = map[string]struct{}{}
		for _, i := range identities {
			c.opentracing.excludedIdentities[i.Name] = struct{}{}
		}
	}
}

// OptPostStartHook registers a function that will be executed right after the server is started.
func OptPostStartHook(hook func(Server) error) Option {
	return func(c *config) {
		c.hooks.postStart = hook
	}
}

// OptPreStopHook registers a function that will be executed just bbefore the server is stopped.
func OptPreStopHook(hook func(Server) error) Option {
	return func(c *config) {
		c.hooks.preStop = hook
	}
}

// OptTraceCleaner registers a trace cleaner that will be called to
// let a chance to clean up various sensitive information before
// sending the trace to the OpenTracing server.
func OptTraceCleaner(cleaner TraceCleaner) Option {
	return func(c *config) {
		c.opentracing.traceCleaner = cleaner
	}
}

// OptIdentifiableRetriever sets the IdentifiableRetriever tha will be used to perform transparent
// patch support using elemental.SparseIdentifiable. When set, the handler for PATCH method will use
// this function to retrieve the target identifiable, will apply the patch and
// treat the request as a standard elemental update operation.
func OptIdentifiableRetriever(f IdentifiableRetriever) Option {
	return func(c *config) {
		c.model.retriever = f
	}
}

// OptErrorTransformer sets the error transformer func to use. If non
// nil, this will be called to eventually transform the error before
// converting it to the elemental.Errors that will be returned to the client.
// If the function return nil, the original error will be used.
func OptErrorTransformer(f func(error) error) Option {
	return func(c *config) {
		c.hooks.errorTransformer = f
	}
}
