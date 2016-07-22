package bahamut

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestMultiCastBoolChannel_NewMultiCastBoolChannel(t *testing.T) {

	Convey("Given create a new MultiCastBoolChannel", t, func() {

		mc := NewMultiCastBooleanChannel()

		Convey("Then the MultiCastBoolChannel should be correctly initialized", func() {
			So(mc.channels, ShouldHaveSameTypeAs, []chan bool{})
		})
	})
}

func TestMultiCastBoolChannel_RegisterUnregister(t *testing.T) {

	Convey("Given create a new MultiCastBoolChannel", t, func() {

		mc := NewMultiCastBooleanChannel()

		Convey("when I register a channel", func() {
			c1 := make(chan bool)

			mc.Register(c1)

			Convey("Then the channel should be registered", func() {
				So(len(mc.channels), ShouldEqual, 1)
			})

			Convey("When I unregister the channel", func() {

				mc.Unregister(c1)

				Convey("Then the channel should be unregistered", func() {
					So(len(mc.channels), ShouldEqual, 0)
				})
			})
		})
	})
}

func TestMultiCastBoolChannel_Send(t *testing.T) {

	Convey("Given create a new MultiCastBoolChannel and I register 3 channels", t, func() {

		mc := NewMultiCastBooleanChannel()

		c1 := make(chan bool)
		c2 := make(chan bool)
		c3 := make(chan bool)

		mc.Register(c1)
		mc.Register(c2)
		mc.Register(c3)

		Convey("When send something", func() {

			go mc.Send(true)

			Convey("Then all three channels should receive the value", func() {
				So(<-c1, ShouldBeTrue)
				So(<-c2, ShouldBeTrue)
				So(<-c3, ShouldBeTrue)
			})
		})
	})
}
