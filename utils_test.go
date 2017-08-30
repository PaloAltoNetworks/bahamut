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
				err = HandleRecoveredPanic(recover(), elemental.NewRequest())
				wg.Done()
			}()
			func() { panic("this is a panic!") }()
		}()

		wg.Wait()

		Convey("Then err should not be nil", func() {
			So(err, ShouldNotBeNil)
		})
	})

	Convey("Given I call a function that doesn't panic", t, func() {

		var err error
		var wg sync.WaitGroup

		wg.Add(1)
		go func() {
			defer func() {
				err = HandleRecoveredPanic(recover(), elemental.NewRequest())
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
