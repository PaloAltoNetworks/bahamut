package push

import (
	"context"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
	"go.aporeto.io/bahamut"
	"go.aporeto.io/elemental"
	testmodel "go.aporeto.io/elemental/test/model"
)

func TestNotifier(t *testing.T) {

	Convey("Given I have a pubsub client and a bahamut server", t, func() {

		server := bahamut.New(
			bahamut.OptModel(
				map[int]elemental.ModelManager{
					0: testmodel.Manager(),
				},
			),
		)

		pubsub := bahamut.NewLocalPubSubClient()
		if err := pubsub.Connect(context.Background()); err != nil {
			panic(err)
		}

		errCh := make(chan error, 10)
		pubCh := make(chan *bahamut.Publication, 10)
		defer pubsub.Subscribe(pubCh, errCh, "topic")()

		Convey("When I call NewNotifier", func() {

			limiters := IdentityToAPILimitersRegistry{}

			n := NewNotifier(
				pubsub,
				"topic",
				"srv1",
				"1.1.1.1:1",
				OptionNotifierAnnounceRateLimits(limiters),
				OptionNotifierPingInterval(time.Second),
			)

			Convey("Then n should be correct", func() {
				So(n, ShouldNotBeNil)
				So(n.pubsub, ShouldEqual, pubsub)
				So(n.serviceName, ShouldEqual, "srv1")
				So(n.serviceStatusTopic, ShouldEqual, "topic")
				So(n.endpoint, ShouldEqual, "1.1.1.1:1")
				So(n.limiters, ShouldResemble, limiters)
				So(n.limiters, ShouldNotEqual, limiters)
			})

			Convey("When I call MakeStartHook and call the hook", func() {

				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				hook := n.MakeStartHook(ctx)

				err := hook(server)

				Convey("Then err should be nil", func() {
					So(err, ShouldBeNil)
				})

				var p *bahamut.Publication
				select {
				case <-time.After(300 * time.Millisecond):
				case p = <-pubCh:
				}

				Convey("Then the pubsub should have received a push", func() {

					So(p, ShouldNotBeNil)

					sping := &servicePing{}
					if err := p.Decode(sping); err != nil {
						panic(err)
					}

					So(sping.Name, ShouldEqual, "srv1")
					So(sping.Endpoint, ShouldEqual, "1.1.1.1:1")
					So(sping.Status, ShouldEqual, entityStatusHello)

					Convey("Then I wait 1.5sec and I should get another pusb", func() {

						var p *bahamut.Publication
						select {
						case p = <-pubCh:
						case <-time.After(1500 * time.Millisecond):
						}

						sping := &servicePing{}
						if err := p.Decode(sping); err != nil {
							panic(err)
						}

						So(sping.Name, ShouldEqual, "srv1")
						So(sping.Endpoint, ShouldEqual, "1.1.1.1:1")
						So(sping.Status, ShouldEqual, entityStatusHello)
					})
				})
			})

			Convey("When I call MakeStopHook and call the hook", func() {

				hook := n.MakeStopHook()

				err := hook(server)

				Convey("Then err should be nil", func() {
					So(err, ShouldBeNil)
				})

				var p *bahamut.Publication
				select {
				case <-time.After(300 * time.Millisecond):
				case p = <-pubCh:
				}

				Convey("Then the pubsub should have received a push", func() {

					So(p, ShouldNotBeNil)

					sping := &servicePing{}
					if err := p.Decode(sping); err != nil {
						panic(err)
					}

					So(sping.Name, ShouldEqual, "srv1")
					So(sping.Endpoint, ShouldEqual, "1.1.1.1:1")
					So(sping.Status, ShouldEqual, entityStatusGoodbye)
				})
			})
		})
	})
}
