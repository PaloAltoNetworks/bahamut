package push

import (
	"context"
	"net"
	"net/http"
	"sync"
	"time"

	"go.aporeto.io/bahamut"
	"go.aporeto.io/bahamut/gateway"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
)

// A Upstreamer listens and update the
// list of the backend services.
type Upstreamer struct {
	pubsub             bahamut.PubSubClient
	apis               map[string][]*endpointInfo
	lock               sync.RWMutex
	serviceStatusTopic string
	config             upstreamConfig
	latencies          sync.Map
}

// NewUpstreamer returns a new push backed upstreamer latency based
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
func (c *Upstreamer) Upstream(req *http.Request) (string, error) {

	identity := getTargetIdentity(req.URL.Path)

	c.lock.RLock()
	defer c.lock.RUnlock()

	l := len(c.apis[identity])

	var n1, n2 int

	switch l {

	case 0:
		return "", nil

	case 1:
		ep := c.apis[identity][0]
		ep.RLock()
		defer ep.RUnlock()

		return ep.address, nil

	case 2:
		n1, n2 = 0, 1

	default:
		n1, n2 = pick(c.config.randomizer, l)
	}

	epi1 := c.apis[identity][n1]
	epi2 := c.apis[identity][n2]

	addresses := [2]string{}
	loads := [2]float64{}
	rls := [2]*rate.Limiter{}

	// BEGIN LOCKED OPERATIONS
	epi1.RLock()
	epi2.RLock()

	addresses[0], addresses[1] = epi1.address, epi2.address
	loads[0], loads[1] = epi1.lastLoad, epi2.lastLoad

	if epi1.limiters != nil && epi1.limiters[identity] != nil {
		rls[0] = epi1.limiters[identity].limiter
	}
	if epi2.limiters != nil && epi2.limiters[identity] != nil {
		rls[1] = epi2.limiters[identity].limiter
	}

	epi1.RUnlock()
	epi2.RUnlock()
	// END LOCKED OPERATIONS

	w := [2]float64{.0, .0}

	// fill our weight from the Feedbackloop
	if ma, ok := c.latencies.Load(addresses[0]); ok {
		if v, err := ma.(*movingAverage).average(); err == nil {
			w[0] = v
		}
	}

	if ma, ok := c.latencies.Load(addresses[1]); ok {
		if v, err := ma.(*movingAverage).average(); err == nil {
			w[1] = v
		}
	}

	// Make sure we got an average for both
	// otherwise default to loads
	if w[0] == 0 || w[1] == 0 {
		w[0] = loads[0]
		w[1] = loads[1]
	}

	// sort
	if w[0] > w[1] {
		addresses[1], addresses[0] = addresses[0], addresses[1]
		loads[1], loads[0] = loads[0], loads[1]
		rls[1], rls[0] = rls[0], rls[1]
		w[1], w[0] = w[0], w[1]
	}

	// Compute cummulative distribution
	w[1] = w[0] + w[1]

	// Given a random choice from 0 to w[1]+1
	draw := float64(c.config.randomizer.Intn(int(w[1]) + 1))

	// routine to extract the endpoint for the given
	// choice index. If it returns false, the object
	// has a rate limiter, and it is currently full.
	check := func(idx uint8) (string, bool) {
		if rls[idx] != nil && !rls[idx].Allow() {
			return "", false
		}
		return addresses[idx], true
	}

	// We pick the fastest/less loaded candidate
	// and get the index of the winner.
	var idx uint8
	if draw <= w[0] {
		idx = 1
	}

	// We check if the winner should be
	// ok to handle the request based on its
	// requested rate limiting. If so, we return
	// it's address.
	addr, ok := check(idx)
	if ok {
		return addr, nil
	}

	// If not, we flip the index.
	if idx == 0 {
		idx = 1
	} else {
		idx = 0
	}

	// And we check if the other endpoint would
	// be ok to handle the request.
	//
	// Note: we may need to make a decision based on the difference
	// of the load between the 2 candidates.
	addr, ok = check(idx)
	if ok {
		return addr, nil
	}

	// If it is sill not ok, we return a 429 error.
	return "", gateway.ErrUpstreamerTooManyRequests
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
					c.latencies.Delete(ep)
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
					c.latencies.Delete(sp.Endpoint)
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

// CollectLatency implement the LatencyBasedUpstreamer interface to add new
// samples into the latencies sync map
func (c *Upstreamer) CollectLatency(address string, responseTime time.Duration) {

	if values, ok := c.latencies.Load(address); ok {
		values.(*movingAverage).insertValue(float64(responseTime.Microseconds()))
	} else {
		c.latencies.Store(address, newMovingAverage(c.config.latencySampleSize))
		c.CollectLatency(address, responseTime)
	}
}
