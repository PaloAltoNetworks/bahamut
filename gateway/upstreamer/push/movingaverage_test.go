package push

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestMovingAverage(t *testing.T) {

	Convey("Given I have a moving average of size 3", t, func() {
		ma := newMovingAverage(3)

		Convey("When I push no values the average is not available", func() {
			v, err := ma.average()
			So(v, ShouldEqual, 0)
			So(err, ShouldNotBeNil)
		})

		Convey("When I push two values the average is not available", func() {
			ma = ma.append(1)
			ma = ma.append(1)
			v, err := ma.average()
			So(v, ShouldEqual, 0)
			So(err, ShouldNotBeNil)
		})

		Convey("When I push a tree values the average calculated", func() {
			ma = ma.append(1)
			ma = ma.append(1)
			ma = ma.append(1)
			v, err := ma.average()
			So(v, ShouldEqual, 1)
			So(err, ShouldBeNil)
		})

		Convey("When I push a four values the average calculated", func() {
			ma = ma.append(1)
			ma = ma.append(1)
			ma = ma.append(1)
			ma = ma.append(1)
			v, err := ma.average()
			So(v, ShouldEqual, 1)
			So(err, ShouldBeNil)
		})

	})
}
