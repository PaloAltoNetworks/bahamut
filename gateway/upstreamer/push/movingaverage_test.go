package push

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestMovingAverage(t *testing.T) {

	Convey("Given I have a moving average of size 3", t, func() {
		ma := NewMovingAverage(3)

		Convey("When I push two values the average is 0", func() {
			ma.Add(1)
			ma.Add(1)
			So(ma.Average(), ShouldEqual, 0)
		})

		Convey("When I push a tree values the average calculated", func() {
			ma.Add(1)
			ma.Add(1)
			ma.Add(1)
			So(ma.Average(), ShouldEqual, 1)
		})

	})
}
