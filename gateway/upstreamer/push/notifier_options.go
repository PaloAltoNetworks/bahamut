package push

import (
	"time"

	"go.aporeto.io/elemental"
)

type notifierConfig struct {
	rateLimits       IdentityToAPILimitersRegistry
	privateOverrides map[string]bool
	prefix           string
	pingInterval     time.Duration
}

func newNotifierConfig() notifierConfig {
	return notifierConfig{
		rateLimits:       IdentityToAPILimitersRegistry{},
		pingInterval:     5 * time.Second,
		privateOverrides: map[string]bool{},
	}
}

// A NotifierOption is the kind of option that can be passed
// to the notifier.
type NotifierOption func(*notifierConfig)

// OptionNotifierPingInterval sets the interval between sending
// 2 pings. The default is 5s.
func OptionNotifierPingInterval(interval time.Duration) NotifierOption {
	return func(c *notifierConfig) {
		c.pingInterval = interval
	}
}

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

// OptionNotifierPrefix sets the API prefix that the gateway should
// add to the API routes for that service.
func OptionNotifierPrefix(prefix string) NotifierOption {
	return func(c *notifierConfig) {
		c.prefix = prefix
	}
}

// OptionNotifierPrivateAPIOverrides allows to pass a map of identity to boolean
// that will be used to override the specificaton's "private" flag. This allows
// the service to force a public API to be private (or vice versa).
//
// NOTE: this does not change the internal data in bahamut's server RouteInfo.
// As far as bahamut server is concerned, the route Private flag did not
// change. This is only affecting the gateway.
func OptionNotifierPrivateAPIOverrides(overrides map[elemental.Identity]bool) NotifierOption {
	return func(c *notifierConfig) {
		for k, v := range overrides {
			c.privateOverrides[k.Category] = v
		}
	}
}
