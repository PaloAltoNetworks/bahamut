// Author: Antoine Mercadal
// See LICENSE file for full LICENSE
// Copyright 2016 Aporeto.

package bahamut

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestKakfaInfo_MakeAPIServerConfig(t *testing.T) {

	Convey("Given I create have a new config", t, func() {

		config := MakeAPIServerConfig("addr", "ca.pem", "cert.pem", "key.pem", []*Route{})

		Convey("Then the the address should be set", func() {
			So(config.ListenAddress, ShouldEqual, "addr")
		})

		Convey("Then the the ca info should be set", func() {
			So(config.TLSCAPath, ShouldEqual, "ca.pem")
		})

		Convey("Then the the cert info should be set", func() {
			So(config.TLSCertificatePath, ShouldEqual, "cert.pem")
		})

		Convey("Then the the key info should be set", func() {
			So(config.TLSKeyPath, ShouldEqual, "key.pem")
		})

		Convey("Then enabled flag should be set", func() {
			So(config.enabled, ShouldBeTrue)
		})
	})
}
