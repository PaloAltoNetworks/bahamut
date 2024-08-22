package push

import (
	"sort"
	"strings"
	"testing"
	"time"

	// nolint:revive // Allow dot imports for readability in tests
	. "github.com/smartystreets/goconvey/convey"
	"golang.org/x/time/rate"
)

func Test_Services(t *testing.T) {

	Convey("Given I create a new service", t, func() {

		srv := newService("mysrv")

		Convey("Then srv should be correctly initialized", func() {
			So(srv.name, ShouldEqual, "mysrv")
			So(srv.endpoints, ShouldResemble, map[string]*endpointInfo{})
		})

		Convey("When I register two endpoints", func() {

			rls1 := IdentityToAPILimitersRegistry{
				"identity-a": {Limit: 10, Burst: 20},
				"identity-b": {Limit: 11, Burst: 21},
			}
			rls2 := IdentityToAPILimitersRegistry{
				"identity-c": {Limit: 100, Burst: 200},
			}

			srv.registerEndpoint("1.1.1.1:4443", 0.3, rls1)
			srv.registerEndpoint("2.2.2.2:4443", 0.4, rls2)

			Convey("Then they should be registered", func() {

				eps := srv.getEndpoints()

				// sort since it is stored as a map
				sort.Slice(eps, func(i, j int) bool {
					return strings.Compare(eps[i].address, eps[j].address) == -1
				})

				So(srv.hasEndpoint("1.1.1.1:4443"), ShouldBeTrue)
				So(srv.hasEndpoint("2.2.2.2:4443"), ShouldBeTrue)

				So(len(eps), ShouldEqual, 2)
				So(eps[0].address, ShouldEqual, "1.1.1.1:4443")
				So(eps[0].lastLoad, ShouldEqual, 0.3)
				So(eps[0].limiters, ShouldEqual, rls1)
				So(eps[0].lastSeen.Round(time.Second), ShouldEqual, time.Now().Round(time.Second))
				So(eps[0].limiters["identity-a"].limiter, ShouldHaveSameTypeAs, &rate.Limiter{})
				So(eps[0].limiters["identity-a"].limiter.Limit(), ShouldEqual, rate.Limit(10))
				So(eps[0].limiters["identity-a"].limiter.Burst(), ShouldEqual, rate.Limit(20))
				So(eps[0].limiters["identity-b"].limiter, ShouldHaveSameTypeAs, &rate.Limiter{})
				So(eps[0].limiters["identity-b"].limiter.Limit(), ShouldEqual, rate.Limit(11))
				So(eps[0].limiters["identity-b"].limiter.Burst(), ShouldEqual, rate.Limit(21))

				So(eps[1].address, ShouldEqual, "2.2.2.2:4443")
				So(eps[1].lastLoad, ShouldEqual, 0.4)
				So(eps[1].lastSeen.Round(time.Second), ShouldEqual, time.Now().Round(time.Second))
				So(eps[1].limiters, ShouldEqual, rls2)
				So(eps[1].limiters["identity-c"].limiter, ShouldHaveSameTypeAs, &rate.Limiter{})
				So(eps[1].limiters["identity-c"].limiter.Limit(), ShouldEqual, rate.Limit(100))
				So(eps[1].limiters["identity-c"].limiter.Burst(), ShouldEqual, rate.Limit(200))

			})

			Convey("When I poke one endpoint", func() {

				time.Sleep(1500 * time.Millisecond)
				srv.pokeEndpoint("2.2.2.2:4443", 0.6)

				eps := srv.getEndpoints()

				// sort since it is stored as a map
				sort.Slice(eps, func(i, j int) bool {
					return strings.Compare(eps[i].address, eps[j].address) == -1
				})

				Convey("Then the endpoint should be poked", func() {
					So(eps[1].address, ShouldEqual, "2.2.2.2:4443")
					So(eps[1].lastLoad, ShouldEqual, 0.6)
					So(eps[1].lastSeen.Round(time.Second), ShouldEqual, time.Now().Round(time.Second))
				})

				Convey("Then the second one should be outdated", func() {

					outdated := srv.outdatedEndpoints(time.Now().Add(-1 * time.Second))

					So(len(outdated), ShouldEqual, 1)
					So(outdated, ShouldContain, "1.1.1.1:4443")
				})
			})

			Convey("When I unregister the 2 endpoints", func() {

				srv.unregisterEndpoint("1.1.1.1:4443")
				srv.unregisterEndpoint("2.2.2.2:4443")

				Convey("Then all endpoints should be removed", func() {
					So(len(srv.getEndpoints()), ShouldEqual, 0)
				})
			})
		})
	})
}
