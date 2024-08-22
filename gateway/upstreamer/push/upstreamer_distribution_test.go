package push

import (
	"net/http"
	"net/url"
	"sync"
	"testing"
	"time"

	// nolint:revive // Allow dot imports for readability in tests
	. "github.com/smartystreets/goconvey/convey"
	"golang.org/x/time/rate"
)

func TestUpstreamUpstreamerDistribution(t *testing.T) {

	Convey("Given I have an upstreamer with 3 registered apis with different loads", t, func() {

		u := NewUpstreamer(nil, "topic", "topic2")
		u.apis = map[string][]*endpointInfo{
			"/cats": {
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

	Convey("Given I have an upstreamer with 1 not loaded/ratelimited and one loaded/not ratelimited", t, func() {

		u := NewUpstreamer(nil, "topic", "topic2")
		u.apis = map[string][]*endpointInfo{
			"/cats": {
				{
					address:  "1.1.1.1:1",
					lastLoad: 10.0,
					limiters: IdentityToAPILimitersRegistry{
						"cats": {limiter: rate.NewLimiter(rate.Limit(1), 1)},
					},
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
				So(counts["1.1.1.1:1"], ShouldAlmostEqual, 0, 10)
				So(counts["3.3.3.3:1"], ShouldAlmostEqual, 2000, 10)
			})
		})
	})

	Convey("Given I have an upstreamer with 1 not loaded/not ratelimited and one loaded/ratelimited", t, func() {

		u := NewUpstreamer(nil, "topic", "topic2")
		u.apis = map[string][]*endpointInfo{
			"/cats": {
				{
					address:  "1.1.1.1:1",
					lastLoad: 10.0,
				},
				{
					address:  "3.3.3.3:1",
					lastLoad: 81.0,
					limiters: IdentityToAPILimitersRegistry{
						"cats": {limiter: rate.NewLimiter(rate.Limit(1), 1)},
					},
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
				So(counts["1.1.1.1:1"], ShouldAlmostEqual, 2000, 10)
				So(counts["3.3.3.3:1"], ShouldAlmostEqual, 0, 10)
			})
		})
	})
}

func TestLatencyBasedUpstreamer(t *testing.T) {

	Convey("Given I have a new latency based upstreamer", t, func() {
		u := NewUpstreamer(nil, "topic", "topic2")
		u.config.latencySampleSize = 2

		Convey("When I there is no entries the average is not available", func() {

			var v float64
			var err error

			if ma, ok := u.latencies.Load("foo"); ok {
				v, err = ma.(movingAverage).average()
			}

			So(v, ShouldEqual, 0)
			So(err, ShouldBeNil)
		})

		Convey("When I add one entry the average is not yet available", func() {
			u.CollectLatency("bar", 1*time.Microsecond)

			var v float64
			var err error

			if ma, ok := u.latencies.Load("bar"); ok {
				v, err = ma.(movingAverage).average()
			}

			So(v, ShouldEqual, 0)
			So(err, ShouldNotBeNil)
		})

		Convey("When I add two entries the average is available", func() {
			u.CollectLatency("bar", 1*time.Microsecond)
			u.CollectLatency("bar", 1*time.Microsecond)

			var v float64
			var err error

			if ma, ok := u.latencies.Load("bar"); ok {
				v, err = ma.(movingAverage).average()
			}

			So(v, ShouldEqual, 1)
			So(err, ShouldBeNil)
		})

		Convey("When I add entries concurently there is no race", func() {

			u := NewUpstreamer(nil, "topic", "topic2")
			u.config.latencySampleSize = 100

			var wg sync.WaitGroup

			inc := func() {
				defer wg.Done()
				u.CollectLatency("bar", 1*time.Microsecond)
			}

			for i := 0; i < 100; i++ {
				wg.Add(1)
				go inc()
			}

			wg.Wait()

			if ma, ok := u.latencies.Load("bar"); ok {
				// As there is no garrantee of the result as the operation can overlap
				// we are not checking the result here. This is just to track races
				_, _ = ma.(movingAverage).average()
			}

		})

		Convey("When I delete an entry a values the average is not available", func() {
			u.latencies.Delete("bar")
			var v float64
			var err error

			if ma, ok := u.latencies.Load("bar"); ok {
				v, err = ma.(movingAverage).average()
			}

			So(v, ShouldEqual, 0)
			So(err, ShouldBeNil)
		})

	})
}
