// Author: Antoine Mercadal
// See LICENSE file for full LICENSE
// Copyright 2016 Aporeto.

package bahamut

import (
	"net/http"
	"testing"

	"github.com/aporeto-inc/elemental"
	. "github.com/smartystreets/goconvey/convey"
)

func TestProcessorHelpers_checkAuthenticated(t *testing.T) {

	Convey("Given I create a new Bahamut", t, func() {

		auth := &MockAuth{}

		h := http.Header{}
		h.Add("Origin", "http://origin.com")
		ctx := &Context{
			Request: &elemental.Request{
				Headers: h,
			},
		}

		Convey("When I check authentication with no registered authenticator", func() {

			err := CheckAuthentication(nil, ctx)

			Convey("Then it should be authenticated", func() {
				So(err, ShouldBeNil)
			})
		})

		Convey("When I check the authentication with a registered authenticator", func() {

			auth.ExpectedAuthenticatedResult = true

			err := CheckAuthentication(auth, ctx)

			Convey("Then it should be authenticated", func() {
				So(err, ShouldBeNil)
			})
		})

		Convey("When I check the authentication with a registered authenticator that returns no", func() {

			auth.ExpectedAuthenticatedResult = false
			auth.DefinedHasError = false

			err := CheckAuthentication(auth, ctx)

			Convey("Then it should not be authenticated", func() {
				So(err, ShouldNotBeNil)
			})

			Convey("Then the http status should be 500", func() {
				So(err.(elemental.Error).Code, ShouldEqual, 401)
			})
		})

		Convey("When I check the authentication with a registered authenticator that returns an error", func() {

			auth.ExpectedAuthenticatedResult = false
			auth.DefinedHasError = true

			err := CheckAuthentication(auth, ctx)

			Convey("Then it should be authenticated", func() {
				So(err, ShouldNotBeNil)
			})
		})
	})
}

func TestProcessorHelpers_checkAuthorized(t *testing.T) {

	Convey("Given I create a new Bahamut", t, func() {

		auth := &MockAuth{}

		h := http.Header{}
		h.Add("Origin", "http://origin.com")
		ctx := &Context{
			Request: &elemental.Request{
				Headers: h,
			},
		}

		Convey("When I check authorization with no registered authorizer", func() {

			err := CheckAuthorization(nil, ctx)

			Convey("Then it should be authorized", func() {
				So(err, ShouldBeNil)
			})
		})

		Convey("When I check the authorization with a registered authorizer", func() {

			auth.ExpectedAuthorizedResult = true
			auth.DefinedHasError = false

			err := CheckAuthorization(auth, ctx)

			Convey("Then it should be authorized", func() {
				So(err, ShouldBeNil)
			})
		})

		Convey("When I check the authorization with a registered authorizer that returns no", func() {

			auth.ExpectedAuthorizedResult = false
			auth.DefinedHasError = false

			err := CheckAuthorization(auth, ctx)

			Convey("Then it should not be authorized", func() {
				So(err, ShouldNotBeNil)
			})

			Convey("Then the http status should be 403", func() {
				So(err.(elemental.Error).Code, ShouldEqual, 403)
			})
		})

		Convey("When I check the authorization with a registered authorizer that returns an error", func() {

			auth.ExpectedAuthorizedResult = false
			auth.DefinedHasError = true

			err := CheckAuthorization(auth, ctx)

			Convey("Then it should not be authorized", func() {
				So(err, ShouldNotBeNil)
			})

			Convey("Then the http status should be 500", func() {
				So(err.(elemental.Error).Code, ShouldEqual, 500)
			})
		})
	})
}
