package push

import (
	"net/http"
	"net/url"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestUpstreamUpstreamerDistribution(t *testing.T) {

	Convey("Given I have an upstreamer with 3 registered apis with different loads", t, func() {

		u := NewUpstreamer(nil, "topic")
		u.apis = map[string][]*endpointInfo{
			"cats": {
				{
					address:  "1.1.1.1:1",
					lastLoad: 10.0,
				},
				{
					address:  "2.2.2.2:1",
					lastLoad: 10.0,
				},
				{
					address:  "3.3.3.3:1",
					lastLoad: 81.0,
				},
			},
		}

		Convey("When I call upstream on /cats 2k times", func() {

			counts := make(map[string]int)

			for i := 0; i <= 2000; i++ {
				upstream, _ := u.Upstream(&http.Request{
					URL: &url.URL{Path: "/cats"},
				})
				counts[upstream]++
			}

			Convey("Then the repoartition should be correct", func() {
				So(counts["1.1.1.1:1"], ShouldAlmostEqual, counts["2.2.2.2:1"], 200)
				So(counts["3.3.3.3:1"], ShouldBeLessThan, counts["1.1.1.1:1"]/2)
			})
		})
	})
}

func TestLatencyBasedUpstreamer(t *testing.T) {

	Convey("Given I have a new latency based upstreamer", t, func() {
		u := NewUpstreamer(nil, "topic")
		u.config.latencySampleSize = 2

		Convey("When I there is no entries the average is not available", func() {
			v, err := u.average("foo")
			So(v, ShouldEqual, 0)
			So(err, ShouldNotBeNil)
		})

		Convey("When I add one entry the average is not yet available", func() {
			u.CollectLatency("bar", 1*time.Microsecond)
			v, err := u.average("bar")
			So(v, ShouldEqual, 0)
			So(err, ShouldNotBeNil)
		})

		Convey("When I add two entries the average is not yet available", func() {
			u.CollectLatency("bar", 1*time.Microsecond)
			u.CollectLatency("bar", 1*time.Microsecond)
			v, err := u.average("bar")
			So(v, ShouldEqual, 1)
			So(err, ShouldBeNil)
		})

		Convey("When I delete an entry a values the average is not available", func() {
			u.delete("bar")
			v, err := u.average("bar")
			So(v, ShouldEqual, 0)
			So(err, ShouldNotBeNil)
		})

	})

}
