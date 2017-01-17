package bahamut

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestCount_NewCount(t *testing.T) {

	Convey("Given I create a new Count", t, func() {

		c := newCount()

		Convey("Then it should be correctly initialized", func() {
			So(c.Total, ShouldEqual, 0)
			So(c.Current, ShouldEqual, 0)
		})
	})
}

func TestCount_String(t *testing.T) {

	Convey("Given I create a new Count", t, func() {

		c := newCount()
		c.Total = 10
		c.Current = 1

		Convey("When I use the String method", func() {

			s := c.String()

			Convey("Then the string should be correct", func() {
				So(s, ShouldEqual, "<count total:10 current:1>")
			})
		})
	})
}

func TestCount_Dulicate(t *testing.T) {

	Convey("Given I have a count", t, func() {

		c := newCount()
		c.Total = 10
		c.Current = 1

		Convey("When I use call Duplicate", func() {

			c2 := c.Duplicate()

			Convey("Then the duplicated count should be correct", func() {
				So(c.Total, ShouldEqual, c2.Total)
				So(c.Current, ShouldEqual, c2.Current)
			})
		})
	})
}
