package push

import (
	"time"

	"golang.org/x/time/rate"
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
	peerTimeout                 time.Duration
	peerTimeoutCheckInterval    time.Duration
	peerPingInterval            time.Duration
	randomizer                  Randomizer
	tokenLimitingBurst          int
	tokenLimitingRPS            rate.Limit
	globalServiceTopic          string
}

func newUpstreamConfig() upstreamConfig {
	return upstreamConfig{
		eventsAPIs:                  map[string]string{},
		latencySampleSize:           20,
		serviceTimeout:              30 * time.Second,
		serviceTimeoutCheckInterval: 5 * time.Second,
		peerTimeout:                 30 * time.Second,
		peerTimeoutCheckInterval:    5 * time.Second,
		peerPingInterval:            10 * time.Second,
		randomizer:                  newRandomizer(),
		tokenLimitingBurst:          2000,
		tokenLimitingRPS:            500,
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
// OptionUpstreamerRegisterEventAPI("serviceA", "events")
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

// OptionUpstreamerPeersTimeout sets for how long a peer ping
// should stay valid after receiving it.
// The default is 30s.
func OptionUpstreamerPeersTimeout(t time.Duration) UpstreamerOption {
	return func(cfg *upstreamConfig) {
		cfg.peerTimeout = t
	}
}

// OptionUpstreamerPeersCheckInterval sets the frequency at which the upstreamer
// will check for outdated peers.
// The default is 5s.
func OptionUpstreamerPeersCheckInterval(t time.Duration) UpstreamerOption {
	return func(cfg *upstreamConfig) {
		cfg.peerTimeoutCheckInterval = t
	}
}

// OptionUpstreamerPeersPingInterval sets how often the upstreamer will
// ping its peers.
// The default is 10s.
func OptionUpstreamerPeersPingInterval(t time.Duration) UpstreamerOption {
	return func(cfg *upstreamConfig) {
		cfg.peerPingInterval = t
	}
}

// OptionUpstreamerTokenRateLimiting configures the per source rate limiting.
// The default is cps:500/burst:2000
func OptionUpstreamerTokenRateLimiting(rps rate.Limit, burst int) UpstreamerOption {
	return func(cfg *upstreamConfig) {
		cfg.tokenLimitingRPS = rps
		cfg.tokenLimitingBurst = burst
		if cfg.tokenLimitingRPS <= 0 {
			panic("rps cannot be <= 0")
		}
		if cfg.tokenLimitingBurst <= 0 {
			panic("burst cannot be <= 0")
		}
	}
}

// OptionUpstreamerGlobalServiceTopic sets the global topic that the gateway
// will use to listen for service pings coming from global services.
func OptionUpstreamerGlobalServiceTopic(topic string) UpstreamerOption {
	return func(cfg *upstreamConfig) {
		cfg.globalServiceTopic = topic
	}
}
