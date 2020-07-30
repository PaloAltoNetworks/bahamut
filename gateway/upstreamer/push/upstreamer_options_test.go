package push

import (
	"math/rand"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func Test_Options(t *testing.T) {

	c := newUpstreamConfig()

	Convey("Calling OptionExposePrivateAPIs should work", t, func() {
		OptionUpstreamerExposePrivateAPIs(true)(&c)
		So(c.exposePrivateAPIs, ShouldEqual, true)
	})

	Convey("Calling OptionOverrideEndpointsAddresses should work", t, func() {
		OptionUpstreamerOverrideEndpointsAddresses("127.0.0.1:443")(&c)
		So(c.overrideEndpointAddress, ShouldEqual, "127.0.0.1:443")
	})

	Convey("Calling OptionRegisterEventAPI should work", t, func() {
		OptionUpstreamerRegisterEventAPI("srva", "events")(&c)
		OptionUpstreamerRegisterEventAPI("srvb", "hello")(&c)
		So(len(c.eventsAPIs), ShouldEqual, 2)
		So(c.eventsAPIs["srva"], ShouldEqual, "events")
		So(c.eventsAPIs["srvb"], ShouldEqual, "hello")
	})

	Convey("Calling OptionRequiredServices should work", t, func() {
		OptionRequiredServices([]string{"srv1"})(&c)
		So(c.requiredServices, ShouldResemble, []string{"srv1"})
	})

	Convey("Calling OptionServiceTimeout should work", t, func() {
		OptionUpstreamerServiceTimeout(time.Hour, time.Minute)(&c)
		So(c.serviceTimeout, ShouldEqual, time.Hour)
		So(c.serviceTimeoutCheckInterval, ShouldEqual, time.Minute)
	})

	Convey("Calling OptionRandomizer should work", t, func() {
		rn := rand.New(rand.NewSource(time.Now().UnixNano()))
		OptionUpstreamerRandomizer(rn)(&c)
		So(c.randomizer, ShouldResemble, rn)
	})

	Convey("Calling OptionUpstreamerPeersTimeout should work", t, func() {
		OptionUpstreamerPeersTimeout(time.Hour)(&c)
		So(c.peerTimeout, ShouldResemble, time.Hour)
	})

	Convey("Calling OptionUpstreamerPeersCheckInterval should work", t, func() {
		OptionUpstreamerPeersCheckInterval(time.Hour)(&c)
		So(c.peerTimeoutCheckInterval, ShouldResemble, time.Hour)
	})

	Convey("Calling OptionUpstreamerPeersPingInterval should work", t, func() {
		OptionUpstreamerPeersPingInterval(time.Hour)(&c)
		So(c.peerPingInterval, ShouldResemble, time.Hour)
	})
}
