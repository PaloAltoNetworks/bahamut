package push

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	// nolint:revive // Allow dot imports for readability in tests
	. "github.com/smartystreets/goconvey/convey"
	"go.aporeto.io/bahamut"
	"go.aporeto.io/bahamut/gateway"
	"golang.org/x/time/rate"
)

type errorPubSubClient struct {
	sync.Mutex
	publishError    error
	connectError    error
	disconnectError error
	pubs            chan *bahamut.Publication
	errs            chan error
}

func (p *errorPubSubClient) Publish(publication *bahamut.Publication, opts ...bahamut.PubSubOptPublish) error {
	return p.publishError
}

func (p *errorPubSubClient) Subscribe(pubs chan *bahamut.Publication, errors chan error, topic string, opts ...bahamut.PubSubOptSubscribe) func() {
	p.Lock()
	p.pubs = pubs
	p.errs = errors
	p.Unlock()
	return func() {}
}

func (p *errorPubSubClient) Connect(ctx context.Context) error {
	return p.connectError
}

func (p *errorPubSubClient) Disconnect() error {
	return p.disconnectError
}

func TestUpstreamer(t *testing.T) {

	Convey("Given I have a pubsub client and an upstreamer with required services", t, func() {

		pubsub := bahamut.NewLocalPubSubClient()
		if err := pubsub.Connect(context.Background()); err != nil {
			panic(err)
		}

		u := NewUpstreamer(
			pubsub,
			"topic",
			"topic2",
			OptionUpstreamerOverrideEndpointsAddresses("127.0.0.1"),
			OptionRequiredServices([]string{"srv1"}),
			OptionUpstreamerServiceTimeout(2*time.Second, 1*time.Second),
		)

		Convey("Then the upstreamer should be correct", func() {
			So(u, ShouldNotBeNil)
			So(u.pubsub, ShouldEqual, pubsub)
			So(u.serviceStatusTopic, ShouldEqual, "topic")
			So(u.apis, ShouldResemble, map[string][]*endpointInfo{})
		})

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		ready, _ := u.Start(ctx)

		select {
		case <-time.After(300 * time.Millisecond):
		case <-ready:
			panic("got ready but it should not have been")
		}

		Convey("When I ask for the upstream for /cats", func() {

			upstream, err := u.Upstream(&http.Request{
				URL: &url.URL{Path: "/cats"},
			})

			Convey("Then upstream should be empty", func() {
				So(err, ShouldBeNil)
				So(upstream, ShouldBeEmpty)
			})
		})

		Convey("When I ask for the upstream for /_prefix/cats", func() {

			upstream, err := u.Upstream(&http.Request{
				URL: &url.URL{Path: "/_pefix/cats"},
			})

			Convey("Then upstream should be empty", func() {
				So(err, ShouldBeNil)
				So(upstream, ShouldBeEmpty)
			})
		})

		Convey("When I send a hello ping for non prefixed srv1 and a prefixed one", func() {

			sping1 := &servicePing{
				Name:     "srv1",
				Endpoint: "1.1.1.1:1",
				Status:   entityStatusHello,
				Routes: map[int][]bahamut.RouteInfo{
					0: {
						{
							Identity: "cats",
							URL:      "/cats",
							Verbs:    []string{http.MethodGet},
							Private:  false,
						},
						{
							Identity: "kittens",
							URL:      "/kittens",
							Verbs:    []string{http.MethodDelete},
							Private:  true,
						},
					},
				},
				Versions: map[string]any{
					"hello": "hey",
				},
				Load: 0.2,
			}

			pub1 := bahamut.NewPublication("topic")
			if err := pub1.Encode(sping1); err != nil {
				panic(err)
			}
			if err := pubsub.Publish(pub1); err != nil {
				panic(err)
			}

			sping2 := &servicePing{
				Name:     "srv1",
				Prefix:   "prefix",
				Endpoint: "2.2.2.2:2",
				Status:   entityStatusHello,
				Routes: map[int][]bahamut.RouteInfo{
					0: {
						{
							Identity: "cats",
							URL:      "/cats",
							Verbs:    []string{http.MethodGet},
							Private:  false,
						},
						{
							Identity: "kittens",
							URL:      "/kittens",
							Verbs:    []string{http.MethodDelete},
							Private:  true,
						},
					},
				},
				Versions: map[string]any{
					"hello": "hey",
				},
				Load: 0.2,
			}

			pub2 := bahamut.NewPublication("topic")
			if err := pub2.Encode(sping2); err != nil {
				panic(err)
			}
			if err := pubsub.Publish(pub2); err != nil {
				panic(err)
			}

			time.Sleep(time.Second)

			select {
			case <-time.After(300 * time.Millisecond):
				panic("not ready but it should have been")
			case <-ready:
			}

			Convey("when I ask for the apis", func() {

				Convey("/cats", func() {

					upstream, err := u.Upstream(&http.Request{
						URL: &url.URL{Path: "/cats"},
					})

					Convey("Then upstream should be correct", func() {
						So(err, ShouldBeNil)
						So(upstream, ShouldEqual, "127.0.0.1:1")
						So(len(u.apis["/cats"]), ShouldEqual, 1)
						So(u.apis["/cats"][0].address, ShouldEqual, "127.0.0.1:1")
						So(u.apis["/cats"][0].lastLoad, ShouldEqual, 0.2)
						So(u.apis["/cats"][0].lastSeen.Round(time.Second), ShouldEqual, time.Now().Add(-time.Second).Round(time.Second))
					})

					Convey("When I wait 2 seconds", func() {

						time.Sleep(3 * time.Second)

						Convey("Then /cats should have been removed because it is outdated", func() {

							upstream, err := u.Upstream(&http.Request{
								URL: &url.URL{Path: "/cats"},
							})

							Convey("Then upstream should be correct", func() {
								So(err, ShouldBeNil)
								So(upstream, ShouldEqual, "")
								So(len(u.apis["/cats"]), ShouldEqual, 0)
							})
						})
					})
				})

				Convey("/_prefix/cats", func() {

					upstream, err := u.Upstream(&http.Request{
						URL: &url.URL{Path: "/_prefix/cats"},
					})

					Convey("Then upstream should be correct", func() {
						So(err, ShouldBeNil)
						So(upstream, ShouldEqual, "127.0.0.1:2")
						So(len(u.apis["prefix/cats"]), ShouldEqual, 1)
						So(u.apis["prefix/cats"][0].address, ShouldEqual, "127.0.0.1:2")
						So(u.apis["prefix/cats"][0].lastLoad, ShouldEqual, 0.2)
						So(u.apis["prefix/cats"][0].lastSeen.Round(time.Second), ShouldEqual, time.Now().Add(-time.Second).Round(time.Second))
					})

					Convey("When I wait 2 seconds", func() {

						time.Sleep(3 * time.Second)

						Convey("Then /_prefix/cats should have been removed because it is outdated", func() {

							upstream, err := u.Upstream(&http.Request{
								URL: &url.URL{Path: "/_prefix/cats"},
							})

							Convey("Then upstream should be correct", func() {
								So(err, ShouldBeNil)
								So(upstream, ShouldEqual, "")
								So(len(u.apis["prefix/cats"]), ShouldEqual, 0)
							})
						})
					})
				})
			})

			Convey("When I send a goodbye ping for srv1", func() {

				sping := &servicePing{
					Name:     "srv1",
					Endpoint: "1.1.1.1:1",
					Status:   entityStatusGoodbye,
				}

				pub := bahamut.NewPublication("topic")
				if err := pub.Encode(sping); err != nil {
					panic(err)
				}
				if err := pubsub.Publish(pub); err != nil {
					panic(err)
				}
				time.Sleep(time.Second)

				Convey("When I ask for the upstream for /cats", func() {

					upstream, err := u.Upstream(&http.Request{
						URL: &url.URL{Path: "/cats"},
					})

					Convey("Then upstream should be correct", func() {
						So(err, ShouldBeNil)
						So(upstream, ShouldEqual, "")
						So(len(u.apis["/cats"]), ShouldEqual, 0)
						So(len(u.apis["prefix/cats"]), ShouldEqual, 1)
					})
				})

				Convey("When I ask for the upstream for /_prefix/cats", func() {

					upstream, err := u.Upstream(&http.Request{
						URL: &url.URL{Path: "/_prefix/cats"},
					})

					Convey("Then upstream should be correct", func() {
						So(err, ShouldBeNil)
						So(upstream, ShouldEqual, "127.0.0.1:2")
						So(len(u.apis["/cats"]), ShouldEqual, 0)
						So(len(u.apis["prefix/cats"]), ShouldEqual, 1)
					})
				})
			})
		})
	})

	Convey("Given I have a pubsub client and an upstreamer with no required services", t, func() {

		pubsub := bahamut.NewLocalPubSubClient()
		if err := pubsub.Connect(context.Background()); err != nil {
			panic(err)
		}

		u := NewUpstreamer(
			pubsub,
			"topic",
			"topic2",
			OptionUpstreamerOverrideEndpointsAddresses("127.0.0.1"),
			OptionUpstreamerServiceTimeout(2*time.Second, 1*time.Second),
		)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		ready, _ := u.Start(ctx)

		select {
		case <-time.After(300 * time.Millisecond):
			panic("not ready in time")
		case <-ready:

		}
	})
}

func TestGlobalServiceTopic(t *testing.T) {

	Convey("Given I have an upstreamer", t, func() {

		pubsub := bahamut.NewLocalPubSubClient()
		if err := pubsub.Connect(context.Background()); err != nil {
			panic(err)
		}

		u := NewUpstreamer(
			pubsub,
			"local-topic",
			"topic2",
			OptionUpstreamerGlobalServiceTopic("global-topic"),
		)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		ready, _ := u.Start(ctx)

		select {
		case <-ready:
		case <-time.After(300 * time.Millisecond):
			panic("got not ready but it should not have been")
		}

		sendHelloPing := func(endpoint string, topic string) {
			sping := &servicePing{
				Name:     "srv1",
				Endpoint: endpoint,
				Status:   entityStatusHello,
				Routes: map[int][]bahamut.RouteInfo{
					0: {
						{
							Identity: "cats",
							URL:      "/cats",
							Verbs:    []string{http.MethodGet},
							Private:  false,
						},
					},
				},
				Versions: map[string]any{
					"hello": "hey",
				},
				Load: 0.2,
			}

			pub := bahamut.NewPublication(topic)
			if err := pub.Encode(sping); err != nil {
				panic(err)
			}

			if err := pubsub.Publish(pub); err != nil {
				panic(err)
			}
		}

		Convey("When send a local ping", func() {

			sendHelloPing("1.1.1.1:1", "local-topic")
			time.Sleep(time.Second)

			Convey("Then upstream should be correct", func() {
				u.lock.Lock()
				So(len(u.apis["/cats"]), ShouldEqual, 1)
				u.lock.Unlock()
			})

			Convey("When send a global ping", func() {

				sendHelloPing("2.2.2.2:1", "global-topic")
				time.Sleep(time.Second)

				Convey("Then upstream should be correct", func() {
					u.lock.Lock()
					So(len(u.apis["/cats"]), ShouldEqual, 2)
					u.lock.Unlock()
				})
			})
		})
	})
}

type deterministicRandom struct {
	value int
}

func (d deterministicRandom) Intn(int) int {
	return d.value
}

func (d deterministicRandom) Shuffle(n int, swap func(i, j int)) {
	return
}

func TestUpstreamUpstreamer(t *testing.T) {

	// use a deterministic randomizer for tests

	opt := OptionUpstreamerRandomizer(deterministicRandom{
		value: 1,
	})

	Convey("Given I have an upstreamer with 3 registered apis with different loads", t, func() {

		u := NewUpstreamer(nil, "topic", "topic2", opt)
		u.apis = map[string][]*endpointInfo{
			"/cats": {
				{
					address:  "1.1.1.1:1",
					lastLoad: 10.0,
				},
				{
					address:  "2.2.2.2:1",
					lastLoad: 10.0,
				},
				{
					address:  "3.3.3.3:1",
					lastLoad: 81.0,
				},
			},
		}

		Convey("When I call upstream on /cats", func() {

			upstream, err := u.Upstream(&http.Request{
				URL: &url.URL{Path: "/cats"},
			})

			Convey("Then upstream should be correct", func() {
				So(err, ShouldBeNil)
				So(upstream, ShouldNotBeEmpty)
				So(upstream, ShouldNotEqual, "3.3.3.3:1")
			})
		})
	})

	Convey("Given I have an upstreamer with 3 registered apis with same loads", t, func() {

		u := NewUpstreamer(nil, "topic", "topic2")
		u.apis = map[string][]*endpointInfo{
			"/cats": {
				{
					address:  "1.1.1.1:1",
					lastLoad: 0.1,
				},
				{
					address:  "2.2.2.2:1",
					lastLoad: 0.1,
				},
				{
					address:  "3.3.3.3:1",
					lastLoad: 0.1,
				},
			},
		}

		Convey("When I call upstream on /cats", func() {

			upstream, err := u.Upstream(&http.Request{
				URL: &url.URL{Path: "/cats"},
			})

			Convey("Then upstream should be correct", func() {
				So(err, ShouldBeNil)
				So(upstream, ShouldNotBeEmpty)
			})
		})
	})

	Convey("Given I have an upstreamer with not registered api", t, func() {

		u := NewUpstreamer(nil, "topic", "topic2")
		u.apis = map[string][]*endpointInfo{}

		Convey("When I call upstream on /cats", func() {

			upstream, err := u.Upstream(&http.Request{
				URL: &url.URL{Path: "/cats"},
			})

			Convey("Then upstream should be correct", func() {
				So(err, ShouldBeNil)
				So(upstream, ShouldBeEmpty)
			})
		})
	})

	Convey("Given I have an upstreamer with a single registered api", t, func() {

		u := NewUpstreamer(nil, "topic", "topic2")
		u.apis = map[string][]*endpointInfo{
			"/cats": {
				{
					address:  "1.1.1.1:1",
					lastLoad: 0.1,
				},
			},
		}

		Convey("When I call upstream on /cats", func() {

			upstream, err := u.Upstream(&http.Request{
				URL: &url.URL{Path: "/cats"},
			})

			Convey("Then upstream should be correct", func() {
				So(err, ShouldBeNil)
				So(upstream, ShouldEqual, "1.1.1.1:1")
			})
		})
	})

	Convey("Given I have an upstreamer with 2 registered apis", t, func() {

		u := NewUpstreamer(nil, "topic", "topic2", opt)
		u.apis = map[string][]*endpointInfo{
			"/cats": {
				{
					address:  "2.2.2.2:1",
					lastLoad: 3.0,
				},
				{
					address:  "1.1.1.1:1",
					lastLoad: 2.0,
				},
			},
		}

		Convey("When I call upstream on /cats", func() {

			upstream, err := u.Upstream(&http.Request{
				URL: &url.URL{Path: "/cats"},
			})

			Convey("Then upstream should be correct", func() {
				So(err, ShouldBeNil)
				So(upstream, ShouldEqual, "2.2.2.2:1")
			})
		})
	})

	Convey("Given I have an upstreamer with 2 registered apis both over used", t, func() {

		u := NewUpstreamer(nil, "topic", "topic2", opt)
		u.apis = map[string][]*endpointInfo{
			"/cats": {
				{
					address:  "2.2.2.2:1",
					lastLoad: 3.0,
					limiters: IdentityToAPILimitersRegistry{
						"cats": &APILimiter{
							limiter: rate.NewLimiter(rate.Limit(0), 0),
						},
					},
				},
				{
					address:  "1.1.1.1:1",
					lastLoad: 2.0,
					limiters: IdentityToAPILimitersRegistry{
						"cats": &APILimiter{
							limiter: rate.NewLimiter(rate.Limit(0), 0),
						},
					},
				},
			},
		}

		Convey("When I call upstream on /cats", func() {

			upstream, err := u.Upstream(&http.Request{
				URL: &url.URL{Path: "/cats"},
			})

			Convey("Then upstream should be correct", func() {
				So(err, ShouldNotBeNil)
				So(err, ShouldEqual, gateway.ErrUpstreamerTooManyRequests)
				So(upstream, ShouldEqual, "")
			})
		})
	})

	Convey("Given I have an upstreamer with 2 registered apis both the least loaded over used", t, func() {

		u := NewUpstreamer(nil, "topic", "topic2", opt)
		u.apis = map[string][]*endpointInfo{
			"/cats": {
				{
					address:  "2.2.2.2:1",
					lastLoad: 1.0,
					limiters: IdentityToAPILimitersRegistry{
						"cats": &APILimiter{
							limiter: rate.NewLimiter(rate.Limit(0), 0),
						},
					},
				},
				{
					address:  "1.1.1.1:1",
					lastLoad: 10.0,
				},
			},
		}

		Convey("When I call upstream on /cats", func() {

			upstream, err := u.Upstream(&http.Request{
				URL: &url.URL{Path: "/cats"},
			})

			Convey("Then upstream should be correct", func() {
				So(err, ShouldBeNil)
				So(upstream, ShouldEqual, "1.1.1.1:1")
			})
		})
	})

	Convey("Given I have an upstreamer with 2 registered apis both the most loaded over used", t, func() {

		u := NewUpstreamer(nil, "topic", "topic2", opt)
		u.apis = map[string][]*endpointInfo{
			"/cats": {
				{
					address:  "2.2.2.2:1",
					lastLoad: 10.0,
					limiters: IdentityToAPILimitersRegistry{
						"cats": &APILimiter{
							limiter: rate.NewLimiter(rate.Limit(0), 0),
						},
					},
				},
				{
					address:  "1.1.1.1:1",
					lastLoad: 1.0,
				},
			},
		}

		Convey("When I call upstream on /cats", func() {

			upstream, err := u.Upstream(&http.Request{
				URL: &url.URL{Path: "/cats"},
			})

			Convey("Then upstream should be correct", func() {
				So(err, ShouldBeNil)
				So(upstream, ShouldEqual, "1.1.1.1:1")
			})
		})
	})

	Convey("Given I have an upstreamer with 2 registered apis both with rate limiters that needs update", t, func() {

		now := time.Now()
		u := NewUpstreamer(nil, "topic", "topic2", opt)
		u.apis = map[string][]*endpointInfo{
			"/cats": {
				{
					address:           "2.2.2.2:1",
					lastLoad:          10.0,
					lastLimiterAdjust: now.Add(-time.Hour),
					limiters: IdentityToAPILimitersRegistry{
						"cats": &APILimiter{
							Limit:   rate.Limit(10.0),
							Burst:   30,
							limiter: rate.NewLimiter(rate.Limit(10.0), 30),
						},
					},
				},
				{
					address:           "1.1.1.1:1",
					lastLimiterAdjust: now.Add(-time.Hour),
					lastLoad:          1.0,
					limiters: IdentityToAPILimitersRegistry{
						"cats": &APILimiter{
							Limit:   rate.Limit(100.0),
							Burst:   300,
							limiter: rate.NewLimiter(rate.Limit(100.0), 300),
						},
					},
				},
			},
		}
		u.lastPeerChangeDate.Store(now)
		u.peersCount = 9 // + 1

		Convey("When I call upstream on /cats", func() {

			_, err := u.Upstream(&http.Request{
				URL: &url.URL{Path: "/cats"},
			})

			Convey("Then upstream should be correct", func() {
				So(err, ShouldBeNil)
				So(u.apis["/cats"][0].limiters["cats"].limiter.Limit(), ShouldAlmostEqual, rate.Limit(1.0))
				So(u.apis["/cats"][0].limiters["cats"].limiter.Burst(), ShouldEqual, 3)
				So(u.apis["/cats"][0].lastLimiterAdjust.Round(time.Second), ShouldEqual, now.Round(time.Second))

				So(u.apis["/cats"][1].limiters["cats"].limiter.Limit(), ShouldAlmostEqual, rate.Limit(10.0))
				So(u.apis["/cats"][1].limiters["cats"].limiter.Burst(), ShouldEqual, 30)
				So(u.apis["/cats"][1].lastLimiterAdjust.Round(time.Second), ShouldEqual, now.Round(time.Second))
			})
		})
	})
}

func TestUpstreamPeers(t *testing.T) {

	Convey("An upstreamer should send the hello and goodbye pings correctly", t, func() {

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		pubsub := bahamut.NewLocalPubSubClient()
		_ = pubsub.Connect(context.Background())

		u := NewUpstreamer(pubsub, "serviceStatusTopic", "peerStatusTopic", OptionUpstreamerPeersPingInterval(time.Second))

		pubs := make(chan *bahamut.Publication, 1024)
		errs := make(chan error, 1024)
		unsub := pubsub.Subscribe(pubs, errs, "peerStatusTopic")
		defer unsub()

		u.Start(ctx)

		var ping peerPing
		select {
		case p := <-pubs:
			_ = p.Decode(&ping)
		case <-time.After(2 * time.Second):
			panic("no pub in time")
		}

		So(ping.Status, ShouldEqual, entityStatusHello)

		// Wait one second to see if we receive another push
		select {
		case p := <-pubs:
			_ = p.Decode(&ping)
		case <-time.After(2 * time.Second):
			panic("no pub in time")
		}

		So(ping.Status, ShouldEqual, entityStatusHello)

		// Stop the upstreamer
		cancel()

		select {
		case p := <-pubs:
			_ = p.Decode(&ping)
		case <-time.After(2 * time.Second):
			panic("no pub in time")
		}

		So(ping.Status, ShouldEqual, entityStatusGoodbye)
	})

	Convey("An upstreamer should handle receiving an error", t, func() {

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		pubsub := &errorPubSubClient{}

		u := NewUpstreamer(pubsub, "serviceStatusTopic", "peerStatusTopic", OptionUpstreamerPeersPingInterval(time.Second))

		u.Start(ctx)

		<-time.After(300 * time.Millisecond)
		pubsub.Lock()
		pubsub.errs <- fmt.Errorf("bam")
		pubsub.Unlock()
		<-time.After(300 * time.Millisecond)

		pubsub.Lock()
		So(len(pubsub.errs), ShouldEqual, 0)
		pubsub.Unlock()
	})

	Convey("An upstreamer should manage the hello and goodbye pings correctly", t, func() {

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		pubsub := bahamut.NewLocalPubSubClient()
		_ = pubsub.Connect(context.Background())

		u := NewUpstreamer(
			pubsub,
			"serviceStatusTopic",
			"peerStatusTopic",
			OptionUpstreamerPeersPingInterval(time.Second),
			OptionUpstreamerPeersCheckInterval(time.Second),
			OptionUpstreamerPeersTimeout(2*time.Second),
		)

		u.Start(ctx)

		time.Sleep(300 * time.Millisecond)

		fakeHello1 := bahamut.NewPublication("peerStatusTopic")
		if err := fakeHello1.Encode(peerPing{RuntimeID: "id1", Status: entityStatusHello}); err != nil {
			panic(err)
		}
		fakeGoodbye1 := bahamut.NewPublication("peerStatusTopic")
		if err := fakeGoodbye1.Encode(peerPing{RuntimeID: "id1", Status: entityStatusGoodbye}); err != nil {
			panic(err)
		}

		fakeHello2 := bahamut.NewPublication("peerStatusTopic")
		if err := fakeHello2.Encode(peerPing{RuntimeID: "id2", Status: entityStatusHello}); err != nil {
			panic(err)
		}

		// Send first hello from one peer
		if err := pubsub.Publish(fakeHello1); err != nil {
			panic(err)
		}
		time.Sleep(300 * time.Millisecond)
		So(atomic.LoadInt64(&u.peersCount), ShouldEqual, 1)
		So(u.lastRateSet.Load().(*rateSet), ShouldEqual, emptyRateSet)
		limit, burst, _ := u.ExtractRates(nil)
		So(limit, ShouldResemble, rate.Limit(500/2))
		So(burst, ShouldResemble, 2000/2)
		So(u.lastRateSet.Load().(*rateSet), ShouldResemble, &rateSet{limit: rate.Limit(500 / 2), burst: 2000 / 2})

		// Send second hello from same peer
		if err := pubsub.Publish(fakeHello1); err != nil {
			panic(err)
		}
		time.Sleep(300 * time.Millisecond)
		So(atomic.LoadInt64(&u.peersCount), ShouldEqual, 1)
		So(u.lastRateSet.Load().(*rateSet), ShouldNotEqual, emptyRateSet)
		limit, burst, _ = u.ExtractRates(nil)
		So(limit, ShouldResemble, rate.Limit(500/2))
		So(burst, ShouldResemble, 2000/2)
		So(u.lastRateSet.Load().(*rateSet), ShouldResemble, &rateSet{limit: rate.Limit(500 / 2), burst: 2000 / 2})

		// Send first hello from another peer
		if err := pubsub.Publish(fakeHello2); err != nil {
			panic(err)
		}
		time.Sleep(300 * time.Millisecond)
		So(atomic.LoadInt64(&u.peersCount), ShouldEqual, 2)
		So(u.lastRateSet.Load().(*rateSet), ShouldEqual, emptyRateSet)
		limit, burst, _ = u.ExtractRates(nil)
		So(limit, ShouldResemble, rate.Limit(500.0/3.0))
		So(burst, ShouldResemble, 2000/3)
		So(u.lastRateSet.Load().(*rateSet), ShouldResemble, &rateSet{limit: rate.Limit(500.0 / 3.0), burst: 2000 / 3})

		// Send goodbye from first peer
		if err := pubsub.Publish(fakeGoodbye1); err != nil {
			panic(err)
		}
		time.Sleep(300 * time.Millisecond)
		So(atomic.LoadInt64(&u.peersCount), ShouldEqual, 1)
		So(u.lastRateSet.Load().(*rateSet), ShouldEqual, emptyRateSet)
		limit, burst, _ = u.ExtractRates(nil)
		So(limit, ShouldResemble, rate.Limit(500.0/2.0))
		So(burst, ShouldResemble, 2000/2)
		So(u.lastRateSet.Load().(*rateSet), ShouldResemble, &rateSet{limit: rate.Limit(500.0 / 2.0), burst: 2000 / 2})

		// Send another goodbye from first peer (should not happen but should not cause issue)
		if err := pubsub.Publish(fakeGoodbye1); err != nil {
			panic(err)
		}
		time.Sleep(300 * time.Millisecond)
		So(atomic.LoadInt64(&u.peersCount), ShouldEqual, 1)

		// Now let the second peer timeout
		time.Sleep(3 * time.Second)
		So(atomic.LoadInt64(&u.peersCount), ShouldEqual, 0)
		So(u.lastRateSet.Load().(*rateSet), ShouldEqual, emptyRateSet)
		limit, burst, _ = u.ExtractRates(nil)
		So(limit, ShouldResemble, rate.Limit(500.0))
		So(burst, ShouldResemble, 2000)
		So(u.lastRateSet.Load().(*rateSet), ShouldResemble, &rateSet{limit: rate.Limit(500.0), burst: 2000})
	})
}
