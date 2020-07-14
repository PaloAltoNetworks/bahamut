package push

import (
	"sync"
	"time"

	"go.aporeto.io/bahamut"
)

type endpointInfo struct {
	address   string
	frequency time.Duration
	lastSeen  time.Time
	lastLoad  float64

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

func (b *service) registerEndpoint(ep string, load float64) {

	b.endpoints[ep] = &endpointInfo{lastSeen: time.Now(), lastLoad: load, address: ep}
}

func (b *service) pokeEndpoint(ep string, load float64) {

	if epi, ok := b.endpoints[ep]; ok {
		epi.Lock()
		now := time.Now()
		if !epi.lastSeen.IsZero() {
			epi.frequency = now.Sub(epi.lastSeen)
		}
		epi.lastSeen = now
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
