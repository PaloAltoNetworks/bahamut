package bahamut

import (
	"context"
	"errors"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestJob_RunJob(t *testing.T) {

	Convey("Given I have a context and a job func to run", t, func() {

		var called int

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		j := func() error {
			called++
			return nil
		}

		Convey("When I call RunJob", func() {

			interrupted, err := RunJob(ctx, j)

			Convey("Then interrupted should be false", func() {
				So(interrupted, ShouldBeFalse)
				So(err, ShouldBeNil)
				So(called, ShouldEqual, 1)

			})
		})
	})

	Convey("Given I have a context and a job func to run that returns an error", t, func() {

		var called int

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		j := func() error {
			called++
			return errors.New("oops")
		}

		Convey("When I call RunJob", func() {

			interrupted, err := RunJob(ctx, j)

			Convey("Then interrupted should be false", func() {
				So(interrupted, ShouldBeFalse)
				So(err, ShouldNotBeNil)
				So(called, ShouldEqual, 1)
			})
		})
	})

	Convey("Given I have a context and a job func to run that I cancel", t, func() {

		var called int

		ctx, cancel := context.WithCancel(context.Background())

		j := func() error {
			time.Sleep(2 * time.Second)
			called++
			return errors.New("oops")
		}

		Convey("When I call RunJob", func() {

			var interrupted bool
			var err error

			go func() { interrupted, err = RunJob(ctx, j) }()
			time.Sleep(300 * time.Millisecond)
			cancel()

			Convey("Then interrupted should be false", func() {
				So(interrupted, ShouldBeTrue)
				So(err, ShouldBeNil)
				So(called, ShouldEqual, 0)
			})
		})
	})
}
