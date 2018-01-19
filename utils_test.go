// Author: Antoine Mercadal
// See LICENSE file for full LICENSE
// Copyright 2016 Aporeto.

package bahamut

import (
	"sync"
	"testing"

	"github.com/aporeto-inc/elemental"
	. "github.com/smartystreets/goconvey/convey"
)

func TestUtils_printBanner(t *testing.T) {

	Convey("Given I print the banner", t, func() {

		PrintBanner()

		Convey("Then I increase my test coverage", func() {
			So(1, ShouldEqual, 1)
		})
	})
}

func TestUtils_RecoverFromPanic(t *testing.T) {

	Convey("Given I call a function that panics", t, func() {

		var err error
		var wg sync.WaitGroup

		wg.Add(1)
		go func() {
			defer func() {
				err = handleRecoveredPanic(recover(), elemental.NewRequest(), true)
				wg.Done()
			}()
			panic("this is a panic!")
		}()

		wg.Wait()

		Convey("Then err should not be nil", func() {
			So(err, ShouldNotBeNil)
		})
	})

	Convey("Given I call a function that panics and I don't want to recover", t, func() {

		f := func() {
			defer func() {
				handleRecoveredPanic(recover(), elemental.NewRequest(), false) // nolint
			}()
			func() { panic("this is a panic!") }()
		}

		Convey("Then err should not be nil", func() {
			So(f, ShouldPanic)
		})
	})

	Convey("Given I call a function that doesn't panic", t, func() {

		var err error
		var wg sync.WaitGroup

		wg.Add(1)
		go func() {
			defer func() {
				err = handleRecoveredPanic(recover(), elemental.NewRequest(), true)
				wg.Done()
			}()
			func() {}()
		}()

		wg.Wait()

		Convey("Then err should be nil", func() {
			So(err, ShouldBeNil)
		})
	})
}

func TestUtils_claimsToMap(t *testing.T) {

	Convey("Given I have a claims list", t, func() {

		claims := []string{"a=b", "c=d"}

		Convey("When I call claimsToMap", func() {

			out := claimsToMap(claims)

			Convey("Then the maps should be correct", func() {
				So(len(out), ShouldEqual, 2)
				So(out["a"], ShouldEqual, "b")
				So(out["c"], ShouldEqual, "d")
			})
		})
	})

	Convey("Given I have a claims with bad claims", t, func() {

		claims := []string{"a=b", "c"}

		Convey("When I call claimsToMap", func() {

			Convey("Then it should should panic", func() {
				So(func() { claimsToMap(claims) }, ShouldPanic)
			})
		})
	})
}
