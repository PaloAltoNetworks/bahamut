package push

import (
	"context"
	"math/rand"
	"net"
	"net/http"
	"sync"
	"time"

	"go.aporeto.io/bahamut"
	"go.uber.org/zap"
)

// A Upstreamer listens and update the
// list of the backend services.
type Upstreamer struct {
	pubsub             bahamut.PubSubClient
	apis               map[string][]*endpointInfo
	lock               sync.RWMutex
	serviceStatusTopic string
	config             upstreamConfig
}

// NewUpstreamer returns a new push backed upstreamer.
func NewUpstreamer(pubsub bahamut.PubSubClient, serviceStatusTopic string, options ...Option) *Upstreamer {

	cfg := newUpstreamConfig()
	for _, opt := range options {
		opt(&cfg)
	}

	return &Upstreamer{
		pubsub:             pubsub,
		apis:               map[string][]*endpointInfo{},
		serviceStatusTopic: serviceStatusTopic,
		config:             cfg,
	}
}

// Upstream returns the upstream to go for the given path
func (c *Upstreamer) Upstream(req *http.Request) (string, float64) {

	identity := getTargetIdentity(req.URL.Path)

	c.lock.RLock()
	defer c.lock.RUnlock()

	l := len(c.apis[identity])

	var n1, n2 int

	switch l {

	case 0:
		return "", 0.0

	case 1:
		ep := c.apis[identity][0]
		ep.RLock()
		defer ep.RUnlock()

		return ep.address, ep.lastLoad

	case 2:
		n1, n2 = 0, 1

	default:
		n1, n2 = pick(l)
	}

	epi1 := c.apis[identity][n1]
	epi2 := c.apis[identity][n2]

	var address string
	var load float64

	epi1.RLock()
	epi2.RLock()

	switch {

	case l >= c.config.loadBasedBalancerThreshold:
		if c.config.loadBasedBalancerFunc(epi1.lastLoad, epi2.lastLoad) {
			address = epi1.address
			load = epi1.lastLoad
		} else {
			address = epi2.address
			load = epi2.lastLoad
		}

	default:
		if rand.Intn(2) == 0 {
			address = epi1.address
			load = epi1.lastLoad
		} else {
			address = epi2.address
			load = epi2.lastLoad
		}

	}

	epi1.RUnlock()
	epi2.RUnlock()

	return address, load
}

// Start starts for new backend services.
func (c *Upstreamer) Start(ctx context.Context) chan struct{} {

	ready := make(chan struct{})

	go c.listenService(ctx, ready)

	return ready
}

func (c *Upstreamer) listenService(ctx context.Context, ready chan struct{}) {

	var err error

	pubs := make(chan *bahamut.Publication, 1024)
	errs := make(chan error, 1024)

	unsub := c.pubsub.Subscribe(pubs, errs, c.serviceStatusTopic)
	defer unsub()

	var requiredReady int
	var requiredNotifSent bool

	requiredCount := len(c.config.requiredServices)
	requiredServices := map[string]bool{}
	for _, srv := range c.config.requiredServices {
		requiredServices[srv] = false
	}

	if requiredCount == 0 {
		requiredNotifSent = true
		close(ready)
	}

	services := servicesConfig{}

	ticker := time.NewTicker(c.config.serviceTimeoutCheckInterval)
	defer ticker.Stop()

	for {
		select {

		case <-ticker.C:

			since := time.Now().Add(-c.config.serviceTimeout)

			var foundOutdated bool
			for _, srv := range services {
				for _, ep := range srv.outdatedEndpoints(since) {
					foundOutdated = foundOutdated || handleRemoveServicePing(services, ping{Name: srv.name, Endpoint: ep})
					zap.L().Info("Handled outdated service", zap.String("name", srv.name), zap.String("backend", ep))
				}
			}

			if foundOutdated {
				c.lock.Lock()
				c.apis = resyncRoutes(services, c.config.exposePrivateAPIs, c.config.eventsAPIs)
				c.lock.Unlock()
			}

		case pub := <-pubs:

			var sp ping

			if err = pub.Decode(&sp); err != nil {
				zap.L().Error("Unable to decode service ping", zap.Error(err))
				break
			}

			if c.config.overrideEndpointAddress != "" {
				_, p, err := net.SplitHostPort(sp.Endpoint)
				if err == nil {
					sp.Endpoint = c.config.overrideEndpointAddress + ":" + p
				}
			}

			switch sp.Status {
			case serviceStatusHello:

				if handleAddServicePing(services, sp) {
					c.lock.Lock()
					c.apis = resyncRoutes(services, c.config.exposePrivateAPIs, c.config.eventsAPIs)
					c.lock.Unlock()
					zap.L().Debug("Handled service hello", zap.String("name", sp.Name), zap.String("backend", sp.Endpoint))
				}

				if requiredCount > 0 && !requiredNotifSent {

					if r, ok := requiredServices[sp.Name]; ok && !r {
						requiredServices[sp.Name] = true
						requiredReady++
					}

					if requiredReady == requiredCount {
						requiredNotifSent = true
						close(ready)
					}
				}

			case serviceStatusGoodbye:

				if handleRemoveServicePing(services, sp) {
					c.lock.Lock()
					c.apis = resyncRoutes(services, c.config.exposePrivateAPIs, c.config.eventsAPIs)
					c.lock.Unlock()
					zap.L().Debug("Handled service goodbye", zap.String("name", sp.Name), zap.String("backend", sp.Endpoint))
				}
			}

		case err = <-errs:
			if err.Error() == "nats: invalid connection" {
				zap.L().Fatal("Unrecoverable error from pubsub", zap.Error(err))
			}
			zap.L().Error("Received error from pubsub", zap.Error(err))

		case <-ctx.Done():
			return
		}
	}
}
