package simple

import (
	"fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"go.aporeto.io/elemental"
)

func TestPublishHandler_NewPublishHandler(t *testing.T) {

	Convey("Given I call NewPublishHandler with one funcs", t, func() {

		f1 := func(*elemental.Event) (bool, error) { return true, nil }

		pub := NewPublishHandler(f1)

		Convey("Then it should be correctly initialized", func() {
			So(pub.shouldPublishFunc, ShouldEqual, f1)

		})
	})
}

func TestPublishHandler_ShouldPublish(t *testing.T) {

	Convey("Given I call NewPublishHandler and a func that says ok", t, func() {

		f1 := func(*elemental.Event) (bool, error) { return true, nil }

		pub := NewPublishHandler(f1)

		Convey("When I call ShouldPublish", func() {

			action, err := pub.ShouldPublish(nil)

			Convey("Then err should be nil", func() {
				So(err, ShouldBeNil)
			})

			Convey("Then action should be OK", func() {
				So(action, ShouldEqual, true)
			})
		})
	})

	Convey("Given I call NewPublishHandler and no func", t, func() {

		pub := NewPublishHandler(nil)

		Convey("When I call ShouldPublish", func() {

			action, err := pub.ShouldPublish(nil)

			Convey("Then err should be nil", func() {
				So(err, ShouldBeNil)
			})

			Convey("Then action should be Continue", func() {
				So(action, ShouldEqual, true)
			})
		})
	})

	Convey("Given I call NewPublishHandler and a func that returns an error", t, func() {

		f1 := func(*elemental.Event) (bool, error) { return false, fmt.Errorf("paf") }

		pub := NewPublishHandler(f1)

		Convey("When I call ShouldPublish", func() {

			action, err := pub.ShouldPublish(nil)

			Convey("Then err should not be nil", func() {
				So(err.Error(), ShouldEqual, "paf")
			})

			Convey("Then action should be KO", func() {
				So(action, ShouldEqual, false)
			})
		})
	})
}
