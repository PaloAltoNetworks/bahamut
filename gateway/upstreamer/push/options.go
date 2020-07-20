package push

import (
	"math/rand"
	"sync"
	"time"
)

// An Option represents a configuration option
// for the Upstreamer.
type Option func(*upstreamConfig)

type upstreamConfig struct {
	overrideEndpointAddress     string
	exposePrivateAPIs           bool
	eventsAPIs                  map[string]string
	responseTimeSamples         int
	requiredServices            []string
	serviceTimeout              time.Duration
	serviceTimeoutCheckInterval time.Duration
	randomizer                  *Randomize
}

func newUpstreamConfig() upstreamConfig {
	return upstreamConfig{
		eventsAPIs:                  map[string]string{},
		responseTimeSamples:         100,
		serviceTimeout:              30 * time.Second,
		serviceTimeoutCheckInterval: 5 * time.Second,
		randomizer:                  NewRandomizer(rand.New(rand.NewSource(time.Now().UnixNano()))),
	}
}

// A Randomizer reprensents an interface to randomize
type Randomizer interface {
	Intn(int) int
	Shuffle(n int, swap func(i, j int))
}

// Randomize is a struct holding a randomizer
type Randomize struct {
	lock sync.Mutex
	Randomizer
}

// NewRandomizer return a new Randomizer
func NewRandomizer(r Randomizer) *Randomize {
	return &Randomize{
		Randomizer: r,
	}
}

// OptionRandomizer set a custom Randomizer
// that must implement the following functions
// - Intn(int) int
// - Shuffle(n int, swap func(i, j int))
func OptionRandomizer(randomizer Randomizer) Option {
	return func(cfg *upstreamConfig) {
		cfg.randomizer = NewRandomizer(randomizer)
	}
}

// OptionExposePrivateAPIs configures the Upstreamer to expose
// the private APIs.
func OptionExposePrivateAPIs(enabled bool) Option {
	return func(cfg *upstreamConfig) {
		cfg.exposePrivateAPIs = enabled
	}
}

// OptionResponseTimeSamples configures the size of
// the response time moving average sampling.
// Default is 100.
func OptionResponseTimeSamples(samples int) Option {
	return func(cfg *upstreamConfig) {
		if samples <= 0 {
			panic("responseTimeSamples must be greater than 0")
		}
		cfg.responseTimeSamples = samples
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
