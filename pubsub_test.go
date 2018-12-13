package bahamut

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestPubsub_NewServer(t *testing.T) {

	Convey("Given I create a new localPubSubServer", t, func() {

		ps := NewLocalPubSubClient()

		Convey("Then the PubSubServer should be correctly initialized", func() {
			So(ps, ShouldImplement, (*PubSubClient)(nil))
		})
	})
}

func TestPubSub_connectionWaiter(t *testing.T) {

	Convey("Given I have a waiter", t, func() {

		w := connectionWaiter{
			abort: make(chan struct{}),
			ok:    make(chan bool),
		}

		Convey("When I call Wait and it returns true", func() {

			go func() { w.ok <- true }()

			ok := w.Wait(10 * time.Second)

			Convey("Then ok should be true", func() {
				So(ok, ShouldBeTrue)
			})
		})

		Convey("When I call Wait and it returns false", func() {

			go func() { w.ok <- false }()

			ok := w.Wait(10 * time.Second)

			Convey("Then ok should be true", func() {
				So(ok, ShouldBeFalse)
			})
		})

		Convey("When I call Wait but it timeouts", func() {

			ok := w.Wait(300 * time.Millisecond)

			Convey("Then ok should be true", func() {
				So(ok, ShouldBeFalse)
			})
		})
	})
}
