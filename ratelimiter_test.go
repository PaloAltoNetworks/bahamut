package bahamut

import (
	"net/http"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestRateLimiter_NewRateLimiter(t *testing.T) {

	Convey("Given I create a new rate limiter", t, func() {

		rl := NewRateLimiter(10).(*basicRateLimiter)

		Convey("Then the rate limiter should be correctly configured", func() {
			So(rl, ShouldNotBeNil)
			So(rl.cache, ShouldNotBeNil)
			So(rl.rps, ShouldEqual, 10)
		})
	})
}

func TestRateLimiter_requestIP(t *testing.T) {

	Convey("Given I create a new rate limiter", t, func() {

		rl := NewRateLimiter(10).(*basicRateLimiter)

		Convey("When I call requestIP on a request with a valid IP", func() {

			req := &http.Request{
				RemoteAddr: "1.2.3.4:1234",
			}

			ip, err := rl.requestIP(req)

			Convey("Then the ip should be correct", func() {
				So(ip, ShouldEqual, "1.2.3.4")
			})

			Convey("Then err should be nil", func() {
				So(err, ShouldBeNil)
			})
		})

		Convey("When I call requestIP on a request with a invalid IP", func() {

			req := &http.Request{
				RemoteAddr: "1.2.3.4",
			}

			ip, err := rl.requestIP(req)

			Convey("Then the ip should be empty", func() {
				So(ip, ShouldBeEmpty)
			})

			Convey("Then err should not be nil", func() {
				So(err, ShouldNotBeNil)
			})
		})
	})
}

func TestRateLimiter_RateLimit(t *testing.T) {

	Convey("Given I create a new rate limiter", t, func() {

		rl := NewRateLimiter(10).(*basicRateLimiter)

		Convey("When I call rate limit on a new IP", func() {

			req := &http.Request{
				RemoteAddr: "1.2.3.4:1234",
			}

			limited, err := rl.RateLimit(req)

			Convey("Then limited should false", func() {
				So(limited, ShouldBeFalse)
			})

			Convey("Then err should be nil", func() {
				So(err, ShouldBeNil)
			})
		})
	})

	Convey("Given I create a new rate limiter", t, func() {

		rl := NewRateLimiter(10).(*basicRateLimiter)

		Convey("When I call rate limit on an IP that is abusing", func() {

			req := &http.Request{
				RemoteAddr: "1.2.3.4:1234",
			}
			for i := 0; i <= 10; i++ {
				rl.RateLimit(req) //  nolint
			}

			limited, err := rl.RateLimit(req)

			Convey("Then limited should false", func() {
				So(limited, ShouldBeTrue)
			})

			Convey("Then err should be nil", func() {
				So(err, ShouldBeNil)
			})

			Convey("When I wait a bit for the cache to expire", func() {
				time.Sleep(1500 * time.Millisecond)

				limited, err := rl.RateLimit(req)

				Convey("Then limited should false", func() {
					So(limited, ShouldBeFalse)
				})

				Convey("Then err should be nil", func() {
					So(err, ShouldBeNil)
				})
			})
		})
	})

	Convey("Given I create a new rate limiter", t, func() {

		rl := NewRateLimiter(10).(*basicRateLimiter)

		Convey("When I call rate limit on a bad IP", func() {

			req := &http.Request{
				RemoteAddr: "1.2.3.",
			}

			limited, err := rl.RateLimit(req)

			Convey("Then limited should false", func() {
				So(limited, ShouldBeFalse)
			})

			Convey("Then err should not be nil", func() {
				So(err, ShouldNotBeNil)
			})
		})
	})
}

func TestRateLimiter_NewRateLimiterWithBan(t *testing.T) {

	Convey("Given I create a new ban rate limiter", t, func() {

		rl := NewRateLimiterWithBan(10, 2*time.Second).(*rateLimiterWithBan)

		Convey("Then the rate limiter should be correctly configured", func() {
			So(rl, ShouldNotBeNil)
			So(rl.cache, ShouldNotBeNil)
			So(rl.rps, ShouldEqual, 10)
			So(rl.banTime, ShouldEqual, 2*time.Second)
			So(rl.banCache, ShouldNotBeNil)
		})
	})
}

func TestRateLimiter_RateLimit_Ban(t *testing.T) {

	Convey("Given I create a new rate limiter", t, func() {

		rl := NewRateLimiterWithBan(10, 4*time.Second).(*rateLimiterWithBan)

		Convey("When I call rate limit on a new IP", func() {

			req := &http.Request{
				RemoteAddr: "1.2.3.4:1234",
			}

			limited, err := rl.RateLimit(req)

			Convey("Then limited should false", func() {
				So(limited, ShouldBeFalse)
			})

			Convey("Then err should be nil", func() {
				So(err, ShouldBeNil)
			})
		})
	})

	Convey("Given I create a new rate limiter", t, func() {

		rl := NewRateLimiterWithBan(10, 5*time.Second).(*rateLimiterWithBan)

		Convey("When I call rate limit on an IP that is abusing", func() {

			req := &http.Request{
				RemoteAddr: "1.2.3.4:1234",
			}
			for i := 0; i <= 10; i++ {
				rl.RateLimit(req) //  nolint
			}

			limited, err := rl.RateLimit(req)

			Convey("Then limited should false", func() {
				So(limited, ShouldBeTrue)
			})

			Convey("Then err should be nil", func() {
				So(err, ShouldBeNil)
			})

			Convey("When I wait a bit for the cache to expire", func() {

				time.Sleep(2 * time.Second)

				limited, err := rl.RateLimit(req)

				Convey("Then limited should true", func() {
					So(limited, ShouldBeTrue)
				})

				Convey("Then err should be nil", func() {
					So(err, ShouldBeNil)
				})

				Convey("When I wait another bit for the cache to expire", func() {

					time.Sleep(2 * time.Second)

					limited, err := rl.RateLimit(req)

					Convey("Then limited should false", func() {
						So(limited, ShouldBeTrue)
					})

					Convey("Then err should be nil", func() {
						So(err, ShouldBeNil)
					})
				})
			})
		})
	})

	Convey("Given I create a new rate limiter", t, func() {

		rl := NewRateLimiterWithBan(10, 5*time.Second).(*rateLimiterWithBan)

		Convey("When I call rate limit on a bad IP", func() {

			req := &http.Request{
				RemoteAddr: "1.2.3.",
			}

			limited, err := rl.RateLimit(req)

			Convey("Then limited should false", func() {
				So(limited, ShouldBeFalse)
			})

			Convey("Then err should not be nil", func() {
				So(err, ShouldNotBeNil)
			})
		})
	})
}
