package push

import (
	"testing"
	"time"

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

	Convey("Calling OptionRequiredServices should work", t, func() {
		OptionRequiredServices([]string{"srv1"})(&c)
		So(c.requiredServices, ShouldResemble, []string{"srv1"})
	})

	Convey("Calling OptionServiceTimeout should work", t, func() {
		OptionServiceTimeout(time.Hour, time.Minute)(&c)
		So(c.serviceTimeout, ShouldEqual, time.Hour)
		So(c.serviceTimeoutCheckInterval, ShouldEqual, time.Minute)
	})

	Convey("Calling OptionLoadBasedBalancerThreshold should work", t, func() {
		OptionLoadBasedBalancerThreshold(5)(&c)
		So(c.loadBasedBalancerThreshold, ShouldEqual, 5)
	})

	Convey("Calling OptionLoadBasedBalancerFunc should work", t, func() {
		var fn LoadBasedBalancerFunc = func(a, b float64) bool { return a > b }
		OptionLoadBasedBalancerFunc(fn)(&c)
		So(c.loadBasedBalancerFunc, ShouldEqual, fn)
	})

	Convey("Calling OptionLoadBasedBalancerFunc without func should panic", t, func() {
		So(func() { OptionLoadBasedBalancerFunc(nil)(&c) }, ShouldPanicWith, "LoadBasedBalancerFunc must not be nil")
	})

}
