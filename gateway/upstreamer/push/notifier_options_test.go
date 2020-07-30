package push

import (
	"testing"

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

}
