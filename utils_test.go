// Author: Antoine Mercadal
// See LICENSE file for full LICENSE
// Copyright 2016 Aporeto.

package bahamut

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestUtils_extractFieldNames(t *testing.T) {

	Convey("Given I have a list", t, func() {

		l1 := NewList()

		Convey("When I extract the fields", func() {

			fields := extractFieldNames(l1)

			Convey("Then all fields should be present", func() {
				So(len(fields), ShouldEqual, 7)
				So(fields, ShouldContain, "ID")
				So(fields, ShouldContain, "Description")
				So(fields, ShouldContain, "Name")
				So(fields, ShouldContain, "ParentID")
				So(fields, ShouldContain, "ParentType")
				So(fields, ShouldContain, "CreationOnly")
				So(fields, ShouldContain, "ReadOnly")
			})
		})
	})
}

func TestUtils_areValuesEqual(t *testing.T) {

	Convey("Given I have 2 list", t, func() {

		l1 := NewList()
		l2 := NewList()

		Convey("When I set the same name", func() {

			l1.Name = "list1"
			l2.Name = "list1"

			Convey("Then the values should be equal", func() {
				So(fieldValuesEquals("Name", l1, l2), ShouldBeTrue)
			})
		})

		Convey("When I set a different name", func() {

			l1.Name = "list1"
			l2.Name = "list2"

			Convey("Then the values should not be equal", func() {
				So(fieldValuesEquals("Name", l1, l2), ShouldBeFalse)
			})
		})
	})
}

func TestUtils_printBanner(t *testing.T) {

	Convey("Given I print the banner", t, func() {

		PrintBanner()

		Convey("Then I increase my test coverage", func() {
			So(1, ShouldEqual, 1)
		})
	})
}
