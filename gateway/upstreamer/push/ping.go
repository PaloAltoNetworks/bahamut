package push

import (
	"go.aporeto.io/bahamut"
	"golang.org/x/time/rate"
)

type serviceStatus int

const (
	serviceStatusGoodbye serviceStatus = 0
	serviceStatusHello                 = 1
)

// An APILimiter holds the parameters of a *rate.Limiter.
// It is used to announce a desired rate limit for
// inconming requests.
type APILimiter struct {
	// Decodable: must be public
	Limit rate.Limit
	Burst int

	limiter *rate.Limiter
}

// IdentityToAPILimitersRegistry is a map of elemental.Identity Name
// to an AnnouncedRateLimits.
type IdentityToAPILimitersRegistry map[string]*APILimiter

type ping struct {
	// Decodable: must be public
	Name         string
	Endpoint     string
	PushEndpoint string
	Status       serviceStatus
	Routes       map[int][]bahamut.RouteInfo
	Versions     map[string]interface{}
	Load         float64
	APILimiters  IdentityToAPILimitersRegistry
}
