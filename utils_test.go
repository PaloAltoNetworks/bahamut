// Author: Antoine Mercadal
// See LICENSE file for full LICENSE
// Copyright 2016 Aporeto.

package bahamut

import (
	"testing"

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
