package bahamut

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestLocalPubSub_NewPubSubServer(t *testing.T) {

	Convey("Given I create a new PubSubServer", t, func() {

		ps := newlocalPubSub(nil)

		Convey("Then the PubSubServer should be correctly initialized", func() {
			So(ps.subscribers, ShouldHaveSameTypeAs, map[string][]chan *Publication{})
		})
	})
}

func TestLocalPubSub_ConnectDisconnect(t *testing.T) {

	Convey("Given I create a new PubSubServer", t, func() {

		ps := newlocalPubSub(nil)

		Convey("When I connect", func() {

			connected := ps.Connect().Wait(1 * time.Millisecond)

			Convey("Then call Connect should connect right away", func() {
				So(connected, ShouldBeTrue)
			})
		})

		Convey("Whan I call Disconnect nothing should happen", func() {
			ps.Disconnect()
		})
	})
}

func TestLocalPubSub_RegisterUnregister(t *testing.T) {

	Convey("Given I create a new PubSubServer", t, func() {

		ps := newlocalPubSub(nil)
		ps.Connect()
		defer ps.Disconnect()

		Convey("When I register a channel to a topic", func() {

			c := make(chan *Publication)

			ps.registerSubscriberChannel(c, "topic")
			time.Sleep(30 * time.Millisecond)

			Convey("Then the channel should be correctly registered", func() {
				So(ps.chansForTopic("topic")[0], ShouldEqual, c)
			})

			Convey("When I unregister it", func() {

				ps.unregisterSubscriberChannel(c, "topic")
				time.Sleep(30 * time.Millisecond)

				Convey("Then the channel should be correctly unregistered", func() {
					So(len(ps.chansForTopic("topic")[0]), ShouldEqual, 0)
				})

				Convey("Then the channel should be closed", func() {
					_, ok := <-c
					So(ok, ShouldBeFalse)
				})
			})
		})
	})
}

func TestLocalPubSub_PublishSubscribe(t *testing.T) {

	Convey("Given I create a new PubSubServer", t, func() {

		ps := newlocalPubSub(nil)
		ps.Connect()
		defer ps.Disconnect()

		Convey("When I register a 2 channels to a topic 'topic' and a another one to 'nottopic'", func() {

			c1 := make(chan *Publication)
			c2 := make(chan *Publication)
			c3 := make(chan *Publication)

			u1 := ps.Subscribe(c1, "topic")
			u2 := ps.Subscribe(c2, "topic")
			u3 := ps.Subscribe(c3, "nottopic")
			time.Sleep(30 * time.Millisecond)

			Convey("When Publish somthing", func() {

				publ := NewPublication("topic")
				go ps.Publish(publ)

				time.Sleep(30 * time.Millisecond)

				var ok1, ok2, ok3 bool
			LOOP:
				for {
					select {
					case <-c1:
						ok1 = true
					case <-c2:
						ok2 = true
					case <-c3:
						ok3 = true
					case <-time.After(300 * time.Millisecond):
						break LOOP
					}
				}
				Convey("Then the first two channels should receive the publication", func() {
					So(ok1, ShouldBeTrue)
					So(ok2, ShouldBeTrue)
				})

				Convey("Then the third channel should not receive anything", func() {
					So(ok3, ShouldBeFalse)
				})

				Convey("When I ubsubscribe", func() {
					u1()
					u2()
					u3()

					Convey("Then all channels should be closed", func() {
						_, ok1 := <-c1
						_, ok2 := <-c2
						_, ok3 := <-c3

						So(ok1, ShouldBeFalse)
						So(ok2, ShouldBeFalse)
						So(ok3, ShouldBeFalse)
					})

				})
			})
		})
	})
}
