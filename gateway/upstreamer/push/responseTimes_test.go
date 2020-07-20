package push

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestResponseTimes(t *testing.T) {

	Convey("Given I have a new responseTime of size 3", t, func() {
		r := newResponseTimes(2)

		Convey("When I there is no entries the average is not available", func() {
			v, err := r.getResponseTime("foo")
			So(v, ShouldEqual, 0)
			So(err, ShouldNotBeNil)
		})

		Convey("When I add one entry the average is not yet available", func() {
			r.StoreResponseTime("bar", 1*time.Microsecond)
			v, err := r.getResponseTime("bar")
			So(v, ShouldEqual, 0)
			So(err, ShouldNotBeNil)
		})

		Convey("When I add two entries the average is not yet available", func() {
			r.StoreResponseTime("bar", 1*time.Microsecond)
			r.StoreResponseTime("bar", 1*time.Microsecond)
			v, err := r.getResponseTime("bar")
			So(v, ShouldEqual, 1)
			So(err, ShouldBeNil)
		})

		Convey("When I delete an entry a values the average is not available", func() {
			r.deleteResponseTimes("bar")
			v, err := r.getResponseTime("bar")
			So(v, ShouldEqual, 0)
			So(err, ShouldNotBeNil)
		})

	})

}

func TestMovingAverage(t *testing.T) {

	Convey("Given I have a moving average of size 3", t, func() {
		ma := newMovingAverage(3)

		Convey("When I push no values the average is not available", func() {
			v, err := ma.average()
			So(v, ShouldEqual, 0)
			So(err, ShouldNotBeNil)
		})

		Convey("When I push two values the average is not available", func() {
			ma.insertValue(1)
			ma.insertValue(1)
			v, err := ma.average()
			So(v, ShouldEqual, 0)
			So(err, ShouldNotBeNil)
		})

		Convey("When I push a tree values the average calculated", func() {
			ma.insertValue(1)
			ma.insertValue(1)
			ma.insertValue(1)
			v, err := ma.average()
			So(v, ShouldEqual, 1)
			So(err, ShouldBeNil)
		})

		Convey("When I push a four values the average calculated", func() {
			ma.insertValue(1)
			ma.insertValue(1)
			ma.insertValue(1)
			ma.insertValue(1)
			v, err := ma.average()
			So(v, ShouldEqual, 1)
			So(err, ShouldBeNil)
		})

	})
}
