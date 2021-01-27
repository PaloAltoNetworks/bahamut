package push

import "time"

const defaultPriorityLabel = "default"

type notifierConfig struct {
	rateLimits    IdentityToAPILimitersRegistry
	pingInterval  time.Duration
	priorityLabel string
}

func newNotifierConfig() notifierConfig {
	return notifierConfig{
		rateLimits:    IdentityToAPILimitersRegistry{},
		pingInterval:  5 * time.Second,
		priorityLabel: defaultPriorityLabel,
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

// OptionPriorityLabel sets the priority label the service
// will declare to the gateway. The gateway will prioritize routing
// to services that have the same priority label as itself, or
// will fallback to the rest otherwise.
func OptionPriorityLabel(label string) NotifierOption {
	return func(c *notifierConfig) {
		c.priorityLabel = label
	}
}
