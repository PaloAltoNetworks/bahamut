// Author: Antoine Mercadal
// See LICENSE file for full LICENSE
// Copyright 2016 Aporeto.

package bahamut

import (
	"net/http"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestKakfaInfo_MakeAPIServerConfig(t *testing.T) {

	Convey("Given I create have a new config", t, func() {

		config := MakeAPIServerConfig("addr", "ca.pem", "cert.pem", "key.pem", []*Route{}, func(w http.ResponseWriter, req *http.Request) {}, "addr2", "/h")

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

		Convey("Then the health handler should be set", func() {
			So(config.HealthHandler, ShouldNotBeNil)
		})

		Convey("Then the health listen address should be set", func() {
			So(config.HealthListenAddress, ShouldEqual, "addr2")
		})

		Convey("Then the health endpoint address should be set", func() {
			So(config.HealthEndpoint, ShouldEqual, "/h")
		})
	})
}
