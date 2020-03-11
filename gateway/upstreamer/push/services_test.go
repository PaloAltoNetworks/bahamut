package push

import (
	"sort"
	"strings"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func Test_Services(t *testing.T) {

	Convey("Given I create a new service", t, func() {

		srv := newService("mysrv")

		Convey("Then srv should be correctly initialized", func() {
			So(srv.name, ShouldEqual, "mysrv")
			So(srv.endpoints, ShouldResemble, map[string]*endpointInfo{})
		})

		Convey("When I register two endpoints", func() {

			srv.registerEndpoint("1.1.1.1:4443", 0.3)
			srv.registerEndpoint("2.2.2.2:4443", 0.4)

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
				So(eps[0].lastSeen.Round(time.Second), ShouldEqual, time.Now().Round(time.Second))
				So(eps[1].address, ShouldEqual, "2.2.2.2:4443")
				So(eps[1].lastLoad, ShouldEqual, 0.4)
				So(eps[1].lastSeen.Round(time.Second), ShouldEqual, time.Now().Round(time.Second))
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
