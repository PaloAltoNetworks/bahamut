// Author: Antoine Mercadal
// See LICENSE file for full LICENSE
// Copyright 2016 Aporeto.

package bahamut

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestKakfaInfo_MakePushServerConfig(t *testing.T) {

	Convey("Given I create have a new config", t, func() {

		config := MakePushServerConfig([]string{":1234"}, "topic", nil)

		Convey("Then the kafka info should have the address set", func() {
			So(len(config.kafkaAddresses), ShouldEqual, 1)
			So(config.kafkaAddresses[0], ShouldEqual, ":1234")
		})

		Convey("Then the kafka info should have the default topic set", func() {
			So(config.defaultTopic, ShouldEqual, "topic")
		})

		Convey("Then enabled flag should be set", func() {
			So(config.enabled, ShouldBeTrue)
		})
	})

	Convey("Given I create have a new config with an empty address array", t, func() {

		Convey("Then it should panic ", func() {
			So(func() { MakePushServerConfig([]string{}, "topic", nil) }, ShouldPanic)
		})
	})

	Convey("Given I create have a new config with an empty topic", t, func() {

		Convey("Then it should panic ", func() {
			So(func() { MakePushServerConfig([]string{":1234"}, "", nil) }, ShouldPanic)
		})
	})
}

func TestKakfaInfo_String(t *testing.T) {

	Convey("Given I create have a new config with kafka info", t, func() {

		config := MakePushServerConfig([]string{"127.0.0.1:1234", "127.0.0.1:1235"}, "topic", nil)

		Convey("Then the string representation should be correct", func() {
			So(config.String(), ShouldEqual, "<PushServerConfig Addresses: [127.0.0.1:1234 127.0.0.1:1235] DefaultTopic: topic>")
		})
	})
}
