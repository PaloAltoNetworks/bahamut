// Author: Antoine Mercadal
// See LICENSE file for full LICENSE
// Copyright 2016 Aporeto.

package bahamut

import (
	"context"
	"net/http"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"go.aporeto.io/elemental"
)

func TestProcessorHelpers_checkAuthenticated(t *testing.T) {

	Convey("Given I create a new Bahamut", t, func() {

		auth := &mockAuth{}

		h := http.Header{}
		h.Add("Origin", "http://origin.com")
		ctx := newContext(context.TODO(), &elemental.Request{Headers: h})

		Convey("When I check authentication with no registered authenticator", func() {

			err := CheckAuthentication(nil, ctx)

			Convey("Then it should be authenticated", func() {
				So(err, ShouldBeNil)
			})
		})

		Convey("When I check the authentication with a registered authenticator", func() {

			auth.action = AuthActionOK

			err := CheckAuthentication([]RequestAuthenticator{auth}, ctx)

			Convey("Then it should be authenticated", func() {
				So(err, ShouldBeNil)
			})
		})

		Convey("When I check the authentication with a registered authenticator that returns no", func() {

			auth.action = AuthActionKO
			auth.errored = false

			err := CheckAuthentication([]RequestAuthenticator{auth}, ctx)

			Convey("Then it should not be authenticated", func() {
				So(err, ShouldNotBeNil)
			})

			Convey("Then the http status should be 500", func() {
				So(err.(elemental.Error).Code, ShouldEqual, 401)
			})
		})

		Convey("When I check the authentication with a registered authenticator that returns an error", func() {

			auth.action = AuthActionKO
			auth.errored = true

			err := CheckAuthentication([]RequestAuthenticator{auth}, ctx)

			Convey("Then it should not be authenticated", func() {
				So(err, ShouldNotBeNil)
			})
		})

		Convey("When I check the authentication with two registered authenticators, first one says ok, second errrors", func() {

			auth.action = AuthActionOK
			auth.errored = false

			auth2 := &mockAuth{}
			auth2.errored = true

			err := CheckAuthentication([]RequestAuthenticator{auth, auth2}, ctx)

			Convey("Then it should be authenticated", func() {
				So(err, ShouldBeNil)
			})
		})

		Convey("When I check the authentication with two registered authenticators, first one continue, second errrors", func() {

			auth.action = AuthActionContinue
			auth.errored = false

			auth2 := &mockAuth{}
			auth2.errored = true

			err := CheckAuthentication([]RequestAuthenticator{auth, auth2}, ctx)

			Convey("Then it should not be authenticated", func() {
				So(err, ShouldNotBeNil)
			})
		})
	})
}

func TestProcessorHelpers_checkAuthorized(t *testing.T) {

	Convey("Given I create a new Bahamut", t, func() {

		auth := &mockAuth{}

		h := http.Header{}
		h.Add("Origin", "http://origin.com")

		ctx := newContext(context.TODO(), &elemental.Request{Headers: h})

		Convey("When I check authorization with no registered authorizer", func() {

			err := CheckAuthorization(nil, ctx)

			Convey("Then it should be authorized", func() {
				So(err, ShouldBeNil)
			})
		})

		Convey("When I check the authorization with a registered authorizer", func() {

			auth.action = AuthActionOK
			auth.errored = false

			err := CheckAuthorization([]Authorizer{auth}, ctx)

			Convey("Then it should be authorized", func() {
				So(err, ShouldBeNil)
			})
		})

		Convey("When I check the authorization with a registered authorizer that returns no", func() {

			auth.action = AuthActionKO
			auth.errored = false

			err := CheckAuthorization([]Authorizer{auth}, ctx)

			Convey("Then it should not be authorized", func() {
				So(err, ShouldNotBeNil)
			})

			Convey("Then the http status should be 403", func() {
				So(err.(elemental.Error).Code, ShouldEqual, 403)
			})
		})

		Convey("When I check the authorization with a registered authorizer that returns an error", func() {

			auth.action = AuthActionKO
			auth.errored = true

			err := CheckAuthorization([]Authorizer{auth}, ctx)

			Convey("Then it should not be authorized", func() {
				So(err, ShouldNotBeNil)
			})

			Convey("Then the http status should be 500", func() {
				So(err.(elemental.Error).Code, ShouldEqual, 500)
			})
		})

		Convey("When I check the authorization with two registered authorizers, first one says ok, second errrors", func() {

			auth.action = AuthActionOK
			auth.errored = false

			auth2 := &mockAuth{}
			auth2.errored = true

			err := CheckAuthorization([]Authorizer{auth, auth2}, ctx)

			Convey("Then it should be authenticated", func() {
				So(err, ShouldBeNil)
			})
		})

		Convey("When I check the authorization with two registered authorizers, first one continue, second errrors", func() {

			auth.action = AuthActionContinue
			auth.errored = false

			auth2 := &mockAuth{}
			auth2.errored = true

			err := CheckAuthorization([]Authorizer{auth, auth2}, ctx)

			Convey("Then it should not be authenticated", func() {
				So(err, ShouldNotBeNil)
			})
		})

		Convey("When I check the authorization with two registered authorizers, first one continue, second one continue", func() {

			auth.action = AuthActionContinue
			auth.errored = false

			auth2 := &mockAuth{}
			auth2.action = AuthActionContinue
			auth2.errored = false

			err := CheckAuthorization([]Authorizer{auth, auth2}, ctx)

			Convey("Then it should not be authenticated", func() {
				So(err, ShouldBeNil)
			})
		})
	})
}
