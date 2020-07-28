package push

import (
	"math/rand"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
	"golang.org/x/time/rate"
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

	Convey("Calling OptionRandomizer should work", t, func() {
		rn := rand.New(rand.NewSource(time.Now().UnixNano()))
		OptionRandomizer(rn)(&c)
		So(c.randomizer, ShouldResemble, rn)
	})

}

func Test_NotiferOptions(t *testing.T) {

	c := newNotifierConfig()

	Convey("Calling OptionNotifierAnnounceRateLimits should work", t, func() {
		rls := IdentityToAPILimitersRegistry{
			"a": {Limit: rate.Limit(1), Burst: 2},
		}
		OptionNotifierAnnounceRateLimits(rls)(&c)
		So(c.rateLimits, ShouldResemble, rls)
		So(c.rateLimits, ShouldNotEqual, rls)
	})

}
