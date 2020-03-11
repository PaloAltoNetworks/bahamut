package push

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func Test_Options(t *testing.T) {

	c := newUpstreamConfig()

	Convey("Calling OptionExposePrivateAPIs should work", t, func() {
		OptionExposePrivateAPIs(true)(&c)
		So(c.exposePrivateAPIs, ShouldEqual, true)
	})

	Convey("Calling OptionOverrideEndpointsAddresses should work", t, func() {
		OptionOverrideEndpointsAddresses("127.0.0.1:443")(&c)
		So(c.overrideEndpointAddress, ShouldEqual, "127.0.0.1:443")
	})

	Convey("Calling OptionRegisterEventAPI should work", t, func() {
		OptionRegisterEventAPI("srva", "events")(&c)
		OptionRegisterEventAPI("srvb", "hello")(&c)
		So(len(c.eventsAPIs), ShouldEqual, 2)
		So(c.eventsAPIs["srva"], ShouldEqual, "events")
		So(c.eventsAPIs["srvb"], ShouldEqual, "hello")
	})

}
