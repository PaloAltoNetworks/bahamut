package push

import (
	"context"
	"net/http"
	"net/url"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
	"go.aporeto.io/bahamut"
)

func TestUpstreamer(t *testing.T) {

	Convey("Given I have a pubsub client and an upstreamer with required services", t, func() {

		pubsub := bahamut.NewLocalPubSubClient()
		if !pubsub.Connect().Wait(time.Second) {
			panic("cannot start pubsub")
		}

		u := NewUpstreamer(
			pubsub,
			"topic",
			OptionOverrideEndpointsAddresses("127.0.0.1"),
			OptionRequiredServices([]string{"srv1"}),
			OptionServiceTimeout(2*time.Second, 1*time.Second),
		)

		Convey("Then the upstreamer should be correct", func() {
			So(u, ShouldNotBeNil)
			So(u.pubsub, ShouldEqual, pubsub)
			So(u.serviceStatusTopic, ShouldEqual, "topic")
			So(u.apis, ShouldResemble, map[string][]*endpointInfo{})
		})

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		ready := u.Start(ctx)

		select {
		case <-time.After(300 * time.Millisecond):
		case <-ready:
			panic("got ready but it should not have been")
		}

		Convey("When I ask for the upstream for /cats", func() {

			upstream, load := u.Upstream(&http.Request{
				URL: &url.URL{Path: "/cats"},
			})

			Convey("Then upstream should be empty", func() {
				So(upstream, ShouldBeEmpty)
				So(load, ShouldEqual, 0.0)
			})
		})

		Convey("When I send a hello ping for srv1", func() {

			sping := &ping{
				Name:     "srv1",
				Endpoint: "1.1.1.1:1",
				Status:   serviceStatusHello,
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

				upstream, load := u.Upstream(&http.Request{
					URL: &url.URL{Path: "/cats"},
				})

				Convey("Then upstream should be correct", func() {
					So(upstream, ShouldEqual, "127.0.0.1:1")
					So(load, ShouldEqual, 0.2)
					So(len(u.apis["cats"]), ShouldEqual, 1)
					So(u.apis["cats"][0].address, ShouldEqual, "127.0.0.1:1")
					So(u.apis["cats"][0].lastLoad, ShouldEqual, 0.2)
					So(u.apis["cats"][0].lastSeen.Round(time.Second), ShouldEqual, time.Now().Add(-time.Second).Round(time.Second))
				})

				Convey("When I wait 2 additional second", func() {

					time.Sleep(2 * time.Second)

					Convey("Then endpoint should have been removed because it is outdated", func() {

						upstream, load := u.Upstream(&http.Request{
							URL: &url.URL{Path: "/cats"},
						})

						Convey("Then upstream should be correct", func() {
							So(upstream, ShouldEqual, "")
							So(load, ShouldEqual, 0)
							So(len(u.apis["cats"]), ShouldEqual, 0)
						})
					})
				})
			})

			Convey("When I send a goodbye ping for srv1", func() {

				sping := &ping{
					Name:     "srv1",
					Endpoint: "1.1.1.1:1",
					Status:   serviceStatusGoodbye,
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

					upstream, load := u.Upstream(&http.Request{
						URL: &url.URL{Path: "/cats"},
					})

					Convey("Then upstream should be correct", func() {
						So(upstream, ShouldEqual, "")
						So(load, ShouldEqual, 0.0)
						So(len(u.apis["cats"]), ShouldEqual, 0)
					})
				})
			})
		})
	})

	Convey("Given I have a pubsub client and an upstreamer with no required services", t, func() {

		pubsub := bahamut.NewLocalPubSubClient()
		if !pubsub.Connect().Wait(time.Second) {
			panic("cannot start pubsub")
		}

		u := NewUpstreamer(
			pubsub,
			"topic",
			OptionOverrideEndpointsAddresses("127.0.0.1"),
			OptionServiceTimeout(2*time.Second, 1*time.Second),
		)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		ready := u.Start(ctx)

		select {
		case <-time.After(300 * time.Millisecond):
			panic("not ready in time")
		case <-ready:

		}
	})
}

func TestUpstreamUpstreamer(t *testing.T) {

	Convey("Given I have an upstreamer with 3 registered apis with different loads", t, func() {

		u := NewUpstreamer(nil, "topic")
		u.apis = map[string][]*endpointInfo{
			"cats": {
				{
					address:  "1.1.1.1:1",
					lastLoad: 0.1,
				},
				{
					address:  "2.2.2.2:1",
					lastLoad: 0.2,
				},
				{
					address:  "3.3.3.3:1",
					lastLoad: 0.9,
				},
			},
		}

		Convey("When I call upstream on /cats", func() {

			upstream, load := u.Upstream(&http.Request{
				URL: &url.URL{Path: "/cats"},
			})

			Convey("Then upstream should be correct", func() {
				So(upstream, ShouldNotBeEmpty)
				So(upstream, ShouldNotEqual, "3.3.3.3:1")
				So(load, ShouldNotEqual, 0)
				So(load, ShouldNotEqual, 0.9)
			})
		})
	})

	Convey("Given I have an upstreamer with 3 registered apis with same loads", t, func() {

		u := NewUpstreamer(nil, "topic")
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

			upstream, load := u.Upstream(&http.Request{
				URL: &url.URL{Path: "/cats"},
			})

			Convey("Then upstream should be correct", func() {
				So(upstream, ShouldNotBeEmpty)
				So(load, ShouldNotEqual, 0)
			})
		})
	})

	Convey("Given I have an upstreamer with not registered api", t, func() {

		u := NewUpstreamer(nil, "topic")
		u.apis = map[string][]*endpointInfo{}

		Convey("When I call upstream on /cats", func() {

			upstream, load := u.Upstream(&http.Request{
				URL: &url.URL{Path: "/cats"},
			})

			Convey("Then upstream should be correct", func() {
				So(upstream, ShouldBeEmpty)
				So(load, ShouldEqual, 0)
			})
		})
	})

	Convey("Given I have an upstreamer with a single registered api", t, func() {

		u := NewUpstreamer(nil, "topic")
		u.apis = map[string][]*endpointInfo{
			"cats": {
				{
					address:  "1.1.1.1:1",
					lastLoad: 0.1,
				},
			},
		}

		Convey("When I call upstream on /cats", func() {

			upstream, load := u.Upstream(&http.Request{
				URL: &url.URL{Path: "/cats"},
			})

			Convey("Then upstream should be correct", func() {
				So(upstream, ShouldEqual, "1.1.1.1:1")
				So(load, ShouldEqual, 0.1)
			})
		})
	})

	Convey("Given I have an upstreamer with 2 registered apis", t, func() {

		u := NewUpstreamer(nil, "topic")
		u.apis = map[string][]*endpointInfo{
			"cats": {
				{
					address:  "1.1.1.1:1",
					lastLoad: 0.2,
				},
				{
					address:  "2.2.2.2:1",
					lastLoad: 0.1,
				},
			},
		}

		Convey("When I call upstream on /cats", func() {

			upstream, load := u.Upstream(&http.Request{
				URL: &url.URL{Path: "/cats"},
			})

			Convey("Then upstream should be correct", func() {
				So(upstream, ShouldEqual, "2.2.2.2:1")
				So(load, ShouldEqual, 0.1)
			})
		})
	})

}
