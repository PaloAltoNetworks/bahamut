package push

import (
	"time"
)

// An UpstreamerOption represents a configuration option
// for the Upstreamer.
type UpstreamerOption func(*upstreamConfig)

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

// OptionUpstreamerExposePrivateAPIs configures the Upstreamer to expose
// the private APIs.
func OptionUpstreamerExposePrivateAPIs(enabled bool) UpstreamerOption {
	return func(cfg *upstreamConfig) {
		cfg.exposePrivateAPIs = enabled
	}
}

// OptionUpstreamerOverrideEndpointsAddresses configures the Upstreamer
// to always ignore what IP address the services are reporting
// and always use the provided address.
func OptionUpstreamerOverrideEndpointsAddresses(override string) UpstreamerOption {
	return func(cfg *upstreamConfig) {
		cfg.overrideEndpointAddress = override
	}
}

// OptionUpstreamerRegisterEventAPI registers an event API for the given serviceName
// on the given endpoint.
// For instance is serviceA exposes an event API on /events, you can use
//
//      OptionUpstreamerRegisterEventAPI("serviceA", "events")
func OptionUpstreamerRegisterEventAPI(serviceName string, eventEndpoint string) UpstreamerOption {
	return func(cfg *upstreamConfig) {
		cfg.eventsAPIs[serviceName] = eventEndpoint
	}
}

// OptionRequiredServices sets the list of services
// that must be ready before starting the upstreamer.
func OptionRequiredServices(required []string) UpstreamerOption {
	return func(cfg *upstreamConfig) {
		cfg.requiredServices = required
	}
}

// OptionUpstreamerServiceTimeout sets the time to wait for the upstream
// to consider a service that did not ping to be outdated and removed
// in the case no goodbye was sent. Default is 30s.
// The check interval parameters defines how often the upstream
// will check for outdated services. The default is 5s.
func OptionUpstreamerServiceTimeout(timeout time.Duration, checkInterval time.Duration) UpstreamerOption {
	return func(cfg *upstreamConfig) {
		cfg.serviceTimeout = timeout
		cfg.serviceTimeoutCheckInterval = checkInterval
	}
}

// OptionUpstreamerRandomizer set a custom Randomizer
// that must implement the Randomizer interface
// and be safe for concurrent use by multiple goroutines.
func OptionUpstreamerRandomizer(randomizer Randomizer) UpstreamerOption {
	return func(cfg *upstreamConfig) {
		cfg.randomizer = randomizer
	}
}
