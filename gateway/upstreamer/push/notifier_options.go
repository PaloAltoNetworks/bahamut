package push

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
