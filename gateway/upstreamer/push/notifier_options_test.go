package push

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
	"golang.org/x/time/rate"
)

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

	Convey("Calling OptionNotifierPingInterval should work", t, func() {
		OptionNotifierPingInterval(3 * time.Hour)(&c)
		So(c.pingInterval, ShouldEqual, 3*time.Hour)
	})

	Convey("Calling OptionPriorityLabel should work", t, func() {
		OptionPriorityLabel("coucou")(&c)
		So(c.priorityLabel, ShouldEqual, "coucou")
	})
}
