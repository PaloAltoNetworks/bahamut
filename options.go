package bahamut

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/opentracing/opentracing-go"

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

// OptCustomRootHandler configures the custom root (/) handler.
func OptCustomRootHandler(handler http.HandlerFunc) Option {
	return func(c *config) {
		c.restServer.customRootHandlerFunc = handler
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
		c.profilingServer.mode = "gops"
		c.profilingServer.listenAddress = listen
	}
}

// OptProfilingGCP configure gcp profiling.
//
// ProjectID is the GCP project to use. When running on gcp, this can be empty.
// servicePrefix can be set to add a prefix to your service name when reporting
// profile to GCP. This allows to differentiate multiple instance
// of an application running in the same project.
func OptProfilingGCP(projectID string, servicePrefix string) Option {
	return func(c *config) {
		c.profilingServer.enabled = true
		c.profilingServer.mode = "gcp"
		c.profilingServer.gcpProjectID = projectID
		c.profilingServer.gcpServicePrefix = servicePrefix
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

// OptRateLimiting configures the rate limiting.
func OptRateLimiting(limit float64, burst int) Option {
	return func(c *config) {
		c.rateLimiting.rateLimiter = rate.NewLimiter(rate.Limit(limit), burst)
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
