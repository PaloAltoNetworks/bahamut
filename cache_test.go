package bahamut

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestCaching_NewGenericCache(t *testing.T) {

	Convey("Given I create a new memory cache", t, func() {

		c := NewMemoryCache()

		Convey("Then the cache should be initialized", func() {
			So(c.(*memoryCache).data, ShouldResemble, map[string]*cacheItem{})
			So(len(c.(*memoryCache).data), ShouldBeZeroValue)
		})

		Convey("When I Get a non cached object", func() {
			smt := c.Get("id-something")

			Convey("Then retrieved object should be nil", func() {
				So(smt, ShouldBeNil)
			})
		})

		Convey("When I Exists a non cached object", func() {
			ex := c.Exists("id-something")

			Convey("Then the result should be false", func() {
				So(ex, ShouldBeFalse)
			})
		})

		Convey("When I Del a non cached object", func() {

			Convey("It should not panic", func() {
				So(func() { c.Del("id-something") }, ShouldNotPanic)
			})
		})

		Convey("When I cache something", func() {

			something := &struct {
				Name        string
				Description string
			}{
				Name:        "something",
				Description: "that's something",
			}

			c.Set("id-something", something)

			Convey("Then the object should be cached", func() {
				So(len(c.(*memoryCache).data), ShouldEqual, 1)
			})

			Convey("When I Get a cached object", func() {
				smt := c.Get("id-something")

				Convey("Then retrieved object should be the same as the original", func() {
					So(smt, ShouldEqual, something)
				})
			})

			Convey("When I Exists a the cached object", func() {
				ex := c.Exists("id-something")

				Convey("Then the result should be true", func() {
					So(ex, ShouldBeTrue)
				})
			})

			Convey("When I Del a non cached object", func() {

				Convey("It should not panic", func() {
					So(func() { c.Del("id-something") }, ShouldNotPanic)

					Convey("When I Exists a the deleted cached object", func() {
						ex := c.Exists("id-something")

						Convey("Then the result should be false", func() {
							So(ex, ShouldBeFalse)
						})
					})
				})
			})
		})
	})
}

func TestCaching_Expiration(t *testing.T) {

	Convey("Given I create a new memory cache", t, func() {

		c := NewMemoryCache()

		Convey("When I set an item that expired after 1sec", func() {
			c.SetWithExpiration("id", "item", 1*time.Second)

			Convey("Then the item should be present", func() {

				So(c.Get("id"), ShouldEqual, "item")

				Convey("When I wait for 2 sec", func() {

					<-time.After(2 * time.Second)

					Convey("Then the item should be gone", func() {
						So(c.Get("id"), ShouldBeNil)
					})
				})
			})
		})
	})
}
