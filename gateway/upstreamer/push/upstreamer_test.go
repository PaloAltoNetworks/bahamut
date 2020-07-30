package push

import (
	"context"
	"net/http"
	"net/url"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
	"go.aporeto.io/bahamut"
	"go.aporeto.io/bahamut/gateway"
	"golang.org/x/time/rate"
)

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

		Convey("When I send a hello ping for srv1", func() {

			sping := &servicePing{
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
				Versions: map[string]interface{}{
					"hello": "hey",
				},
				Load: 0.2,
			}

			pub := bahamut.NewPublication("topic")
			if err := pub.Encode(sping); err != nil {
				panic(err)
			}

			if err := pubsub.Publish(pub); err != nil {
				panic(err)
			}

			time.Sleep(time.Second)

			select {
			case <-time.After(300 * time.Millisecond):
				panic("not ready but it should have been")
			case <-ready:
			}

			Convey("When I ask for the upstream for /cats", func() {

				upstream, err := u.Upstream(&http.Request{
					URL: &url.URL{Path: "/cats"},
				})

				Convey("Then upstream should be correct", func() {
					So(err, ShouldBeNil)
					So(upstream, ShouldEqual, "127.0.0.1:1")
					So(len(u.apis["cats"]), ShouldEqual, 1)
					So(u.apis["cats"][0].address, ShouldEqual, "127.0.0.1:1")
					So(u.apis["cats"][0].lastLoad, ShouldEqual, 0.2)
					So(u.apis["cats"][0].lastSeen.Round(time.Second), ShouldEqual, time.Now().Add(-time.Second).Round(time.Second))
				})

				Convey("When I wait 2 additional second", func() {

					time.Sleep(2 * time.Second)

					Convey("Then endpoint should have been removed because it is outdated", func() {

						upstream, err := u.Upstream(&http.Request{
							URL: &url.URL{Path: "/cats"},
						})

						Convey("Then upstream should be correct", func() {
							So(err, ShouldBeNil)
							So(upstream, ShouldEqual, "")
							So(len(u.apis["cats"]), ShouldEqual, 0)
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
						So(len(u.apis["cats"]), ShouldEqual, 0)
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
			"cats": {
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
			"cats": {
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
			"cats": {
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
			"cats": {
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
			"cats": {
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
			"cats": {
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
			"cats": {
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
			"cats": {
				{
					address:           "2.2.2.2:1",
					lastLoad:          10.0,
					lastLimiterAdjust: now.Add(-time.Hour),
					limiters: IdentityToAPILimitersRegistry{
						"cats": &APILimiter{
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
				So(u.apis["cats"][0].limiters["cats"].limiter.Limit(), ShouldAlmostEqual, rate.Limit(1.0))
				So(u.apis["cats"][0].limiters["cats"].limiter.Burst(), ShouldEqual, 3)
				So(u.apis["cats"][0].lastLimiterAdjust.Round(time.Second), ShouldEqual, now.Round(time.Second))

				So(u.apis["cats"][1].limiters["cats"].limiter.Limit(), ShouldAlmostEqual, rate.Limit(10.0))
				So(u.apis["cats"][1].limiters["cats"].limiter.Burst(), ShouldEqual, 30)
				So(u.apis["cats"][1].lastLimiterAdjust.Round(time.Second), ShouldEqual, now.Round(time.Second))
			})
		})
	})
}
