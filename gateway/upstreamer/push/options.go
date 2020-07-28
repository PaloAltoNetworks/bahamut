package push

import (
	"time"
)

// An Option represents a configuration option
// for the Upstreamer.
type Option func(*upstreamConfig)

type upstreamConfig struct {
	overrideEndpointAddress     string
	exposePrivateAPIs           bool
	eventsAPIs                  map[string]string
	latencySampleSize           int
	requiredServices            []string
	serviceTimeout              time.Duration
	serviceTimeoutCheckInterval time.Duration
	randomizer                  Randomizer
}

func newUpstreamConfig() upstreamConfig {
	return upstreamConfig{
		eventsAPIs:                  map[string]string{},
		latencySampleSize:           20,
		serviceTimeout:              30 * time.Second,
		serviceTimeoutCheckInterval: 5 * time.Second,
		randomizer:                  newRandomizer(),
	}
}

// OptionRandomizer set a custom Randomizer
// that must implement the Randomizer interface
// and be safe for concurrent use by multiple goroutines.
func OptionRandomizer(randomizer Randomizer) Option {
	return func(cfg *upstreamConfig) {
		cfg.randomizer = randomizer
	}
}

// OptionExposePrivateAPIs configures the Upstreamer to expose
// the private APIs.
func OptionExposePrivateAPIs(enabled bool) Option {
	return func(cfg *upstreamConfig) {
		cfg.exposePrivateAPIs = enabled
	}
}

// OptionLatencySampleSize configures the size of
// the response time moving average sampling.
// Default is 20.
func OptionLatencySampleSize(samples int) Option {
	return func(cfg *upstreamConfig) {
		if samples <= 0 {
			panic("OptionLatencySampleSize must be greater than 0")
		}
		cfg.latencySampleSize = samples
	}
}

// OptionOverrideEndpointsAddresses configures the Upstreamer
// to always ignore what IP address the services are reporting
// and always use the provided address.
func OptionOverrideEndpointsAddresses(override string) Option {
	return func(cfg *upstreamConfig) {
		cfg.overrideEndpointAddress = override
	}
}

// OptionRegisterEventAPI registers an event API for the given serviceName
// on the given endpoint.
// For instance is serviceA exposes an event API on /events, you can use
//
//      OptionRegisterEventAPI("serviceA", "events")
func OptionRegisterEventAPI(serviceName string, eventEndpoint string) Option {
	return func(cfg *upstreamConfig) {
		cfg.eventsAPIs[serviceName] = eventEndpoint
	}
}

// OptionRequiredServices sets the list of services
// that must be ready before starting the upstreamer.
func OptionRequiredServices(required []string) Option {
	return func(cfg *upstreamConfig) {
		cfg.requiredServices = required
	}
}

// OptionServiceTimeout sets the time to wait for the upstream
// to consider a service that did not ping to be outdated and removed
// in the case no goodbye was sent. Default is 30s.
// The check interval parameters defines how often the upstream
// will check for outdated services. The default is 5s.
func OptionServiceTimeout(timeout time.Duration, checkInterval time.Duration) Option {
	return func(cfg *upstreamConfig) {
		cfg.serviceTimeout = timeout
		cfg.serviceTimeoutCheckInterval = checkInterval
	}
}

type notifierConfig struct {
	rateLimits IdentityToAPILimitersRegistry
}

func newNotifierConfig() notifierConfig {
	return notifierConfig{
		rateLimits: IdentityToAPILimitersRegistry{},
	}
}

// A NotifierOption is the kind of option that can be passed
// to the notifier.
type NotifierOption func(*notifierConfig)

// OptionNotifierAnnounceRateLimits can be used to set a IdentityToAPILimitersRegistry
// to tell the gateways to instantiate some rate limiters for the current
// instance of the service.
//
// It is not guaranteed that the gateway will honor the request.
func OptionNotifierAnnounceRateLimits(rls IdentityToAPILimitersRegistry) NotifierOption {
	return func(c *notifierConfig) {
		c.rateLimits = make(IdentityToAPILimitersRegistry, len(rls))
		for k, v := range rls {
			c.rateLimits[k] = v
		}
	}
}
