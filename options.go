package bahamut

import (
	"crypto/tls"
	"crypto/x509"
	"net/http"
	"time"

	"github.com/aporeto-inc/elemental"
)

// An Option represents a configuration option.
type Option func(*Config)

func createBaseConfig() Config {

	c := Config{}

	// TODO: Basculate to Enabled everywhere when migration
	// is complete, then remove this function
	c.ReSTServer.Disabled = true
	c.PushServer.Disabled = true
	c.PushServer.PublishDisabled = true
	c.PushServer.DispatchDisabled = true
	c.HealthServer.Disabled = true

	return c
}

// OptDisablePanicRecovery disables panic recovery.
func OptDisablePanicRecovery() Option {
	return func(c *Config) {
		c.General.PanicRecoveryDisabled = true
	}
}

// OptRestServer configures the listening address of the server.
//
// listen is the general listening address for the API server as
func OptRestServer(listen string) Option {
	return func(c *Config) {
		c.ReSTServer.Disabled = false
		c.ReSTServer.ListenAddress = listen
	}
}

// OptTimeouts configures the timeouts of the server.
func OptTimeouts(read, write, idle time.Duration) Option {
	return func(c *Config) {
		c.ReSTServer.ReadTimeout = read
		c.ReSTServer.WriteTimeout = write
		c.ReSTServer.IdleTimeout = idle
	}
}

// OptDisableKeepAlive disables http keepalives.
//
// There is a bug in Go <= 1.7 which makes the server eats all available
// fds. Use this option if you are using these versions.
func OptDisableKeepAlive() Option {
	return func(c *Config) {
		c.ReSTServer.DisableKeepalive = true
	}
}

// OptCustomRootHandler configures the custom root (/) handler.
func OptCustomRootHandler(handler http.HandlerFunc) Option {
	return func(c *Config) {
		c.ReSTServer.CustomRootHandlerFunc = handler
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
func OptPushServer(service PubSubServer, topic string) Option {
	return func(c *Config) {
		c.PushServer.Disabled = false
		c.PushServer.Service = service
		c.PushServer.Topic = topic
	}
}

// OptPushDispatchHandler configures the push dispatcher.
//
// DispatchHandler defines the handler that will be used to
// decide if a push event should be dispatch to push sessions.
func OptPushDispatchHandler(dispatchHandler PushDispatchHandler) Option {
	return func(c *Config) {
		c.PushServer.DispatchDisabled = false
		c.PushServer.DispatchHandler = dispatchHandler
	}
}

// OptPushPublishHandler configures the push publisher.
//
// PublishHandler defines the handler that will be used to
// decide if an event should be published.
func OptPushPublishHandler(publishHandler PushPublishHandler) Option {
	return func(c *Config) {
		c.PushServer.PublishDisabled = false
		c.PushServer.PublishHandler = publishHandler
	}
}

// OptHealthServer enables and configures the health server.
//
// ListenAddress is the general listening address for the health server.
// HealthHandler is the type of the function to run to determine the health of the server.
func OptHealthServer(listen string, handler HealthServerFunc) Option {
	return func(c *Config) {
		c.HealthServer.Disabled = false
		c.HealthServer.ListenAddress = listen
		c.HealthServer.HealthHandler = handler
	}
}

// OptHealthServerTimeouts configures the health server timeouts.
func OptHealthServerTimeouts(read, write, idle time.Duration) Option {
	return func(c *Config) {
		c.HealthServer.ReadTimeout = read
		c.HealthServer.WriteTimeout = write
		c.HealthServer.IdleTimeout = idle
	}
}

// OptProfilingLocal configure local goops profiling.
func OptProfilingLocal(listen string) Option {
	return func(c *Config) {
		c.ProfilingServer.Enabled = true
		c.ProfilingServer.Mode = "gops"
		c.ProfilingServer.ListenAddress = listen
	}
}

// OptProfilingGCP configure gcp profiling.
//
// ProjectID is the GCP project to use. When running on gcp, this can be empty.
// servicePrefix can be set to add a prefix to your service name when reporting
// profile to GCP. This allows to differentiate multiple instance
// of an application running in the same project.
func OptProfilingGCP(projectID string, servicePrefix string) Option {
	return func(c *Config) {
		c.ProfilingServer.Enabled = true
		c.ProfilingServer.Mode = "gcp"
		c.ProfilingServer.GCPProjectID = projectID
		c.ProfilingServer.GCPServicePrefix = servicePrefix
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
	return func(c *Config) {
		c.TLS.ServerCertificates = certs
		c.TLS.ServerCertificatesRetrieverFunc = certRetriever
	}
}

// OptMTLS configures the tls client authentication mechanism.
//
// ClientCAPool is the *x509.CertPool to use for the authentifying client.
// AuthType defines the tls authentication mode to use for a secure server.
func OptMTLS(caPool *x509.CertPool, authType tls.ClientAuthType) Option {
	return func(c *Config) {
		c.TLS.ClientCAPool = caPool
		c.TLS.AuthType = authType
	}
}

// OptLetsEncrypt enables and configures the auto letsencrypt certification.
//
// Domains contains the list of white listed domain name to use for
// issuing certificates.
// cache gives the path where to store certificate cache.
// If empty, the default temp folder of the machine will be used.
func OptLetsEncrypt(domains []string, cache string) Option {
	return func(c *Config) {
		c.TLS.EnableLetsEncrypt = true
		c.TLS.LetsEncryptDomainWhiteList = domains
		c.TLS.LetsEncryptCertificateCacheFolder = cache
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
	return func(c *Config) {
		c.Security.RequestAuthenticators = requestAuthenticators
		c.Security.SessionAuthenticators = sessionAuthenticators
	}
}

// OptAuthorizers configures the authorizers.
//
// Authorizers defines the list Authorizers to use to authorize the requests.
// They are executed in order from index 0 to index n. They will return a bahamut.AuthAction to tell if
// the current authorizer grants, denies or let the chain continue. If an error is returned, the
// chain fails immediately.
func OptAuthorizers(authorizers []Authorizer) Option {
	return func(c *Config) {
		c.Security.Authorizers = authorizers
	}
}

// OptAuditer configures the auditor to use to audit the requests.
//
// The Audit() method will be run in a go routine so there is no
// need to deal with it in your implementation.
func OptAuditer(auditer Auditer) Option {
	return func(c *Config) {
		c.Security.Auditer = auditer
	}
}

// OptRateLimiting configures the rate limiting.
func OptRateLimiting(limiter RateLimiter) Option {
	return func(c *Config) {
		c.RateLimiting.RateLimiter = limiter
	}
}

// OptModel configures the elemental Model for the server.
//
// Factory is a function that returns a instance of a model
// according to its identity.
// registry contains each elemental model RelationshipsRegistry for each version.
func OptModel(factory elemental.IdentifiableFactory, registry map[int]elemental.RelationshipsRegistry) Option {
	return func(c *Config) {
		c.Model.IdentifiablesFactory = factory
		c.Model.RelationshipsRegistry = registry
	}
}

// OptReadOnly sets the server in read only mode.
//
// All write operations will return a Locked HTTP Code (423)
// This is useful during maintenance.
// Excluded defines a list of elemental.Identity that will not be affected
// by the read only mode.
func OptReadOnly(excluded []elemental.Identity) Option {
	return func(c *Config) {
		c.Model.ReadOnly = true
		c.Model.ReadOnlyExcludedIdentities = excluded
	}
}

// OptUnmarshallers sets the custom unmarshallers.
//
// Unmarshallers contains a list of custom umarshaller per identity.
// This allows to create custom function to umarshal the payload of a request.
// If none is provided for a particular identity, the standard unmarshal function
// is used.
func OptUnmarshallers(unmarshallers map[elemental.Identity]CustomUmarshaller) Option {
	return func(c *Config) {
		c.Model.Unmarshallers = unmarshallers
	}
}

// OptMockServer enables and configures the mock server.
func OptMockServer(listen string) Option {
	return func(c *Config) {
		c.MockServer.Enabled = true
		c.MockServer.ListenAddress = listen
	}
}

// OptMockServerTimeouts configures the mock server timeouts.
func OptMockServerTimeouts(read, write, idle time.Duration) Option {
	return func(c *Config) {
		c.MockServer.ReadTimeout = read
		c.MockServer.WriteTimeout = write
		c.MockServer.IdleTimeout = idle
	}
}

// OptServiceInfo configures the service basic information.
//
// ServiceName contains the name of the service.
// ServiceVersion contains the version of the service itself.
// Version should contain information relative to the service version.
// like all it's libraries and things like that.
func OptServiceInfo(name string, version string, subversions map[string]interface{}) Option {
	return func(c *Config) {
		c.Meta.ServiceName = name
		c.Meta.ServiceVersion = version
		c.Meta.Version = subversions
	}
}

// OptDisableMetaRoutes disables the meta routing.
func OptDisableMetaRoutes() Option {
	return func(c *Config) {
		c.Meta.DisableMetaRoute = true
	}
}
