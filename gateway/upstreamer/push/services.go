package push

import (
	"sync"
	"time"

	"go.aporeto.io/bahamut"
	"golang.org/x/time/rate"
)

type endpointInfo struct {
	address  string
	lastSeen time.Time
	lastLoad float64
	limiters IdentityToAPILimitersRegistry

	sync.RWMutex
}

type servicesConfig map[string]*service

type service struct {
	name      string
	routes    map[int][]bahamut.RouteInfo
	versions  map[string]interface{}
	endpoints map[string]*endpointInfo
}

// newService returns a new proxy info from the given string.
func newService(name string) *service {
	return &service{
		name:      name,
		endpoints: map[string]*endpointInfo{},
	}
}

func (b *service) getEndpoints() []*endpointInfo {

	out := make([]*endpointInfo, len(b.endpoints))
	var i int
	for _, v := range b.endpoints {
		out[i] = v
		i++
	}

	return out
}

func (b *service) hasEndpoint(ep string) bool {

	_, ok := b.endpoints[ep]

	return ok
}

func (b *service) registerEndpoint(address string, load float64, apilimiters IdentityToAPILimitersRegistry) {

	if apilimiters == nil {
		apilimiters = IdentityToAPILimitersRegistry{}
	}

	// Instantiate all the actual rate limiters using the values
	// announced by the service.
	for _, l := range apilimiters {
		l.limiter = rate.NewLimiter(l.Limit, l.Burst)
	}

	b.endpoints[address] = &endpointInfo{
		lastSeen: time.Now(),
		lastLoad: load,
		address:  address,
		limiters: apilimiters,
	}
}

func (b *service) pokeEndpoint(ep string, load float64) {

	if epi, ok := b.endpoints[ep]; ok {
		epi.Lock()
		epi.lastSeen = time.Now()
		epi.lastLoad = load
		epi.Unlock()
	}
}

func (b *service) outdatedEndpoints(since time.Time) []string {

	var out []string

	for ep, epi := range b.endpoints {
		epi.RLock()
		if epi.lastSeen.Before(since) {
			out = append(out, ep)
		}
		epi.RUnlock()
	}

	return out
}

func (b *service) unregisterEndpoint(ep string) {

	delete(b.endpoints, ep)
}
