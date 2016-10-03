// Author: Antoine Mercadal
// See LICENSE file for full LICENSE
// Copyright 2016 Aporeto.

package bahamut

import (
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestProcessorHelpers_checkAuthenticated(t *testing.T) {

	Convey("Given I create a new Bahamut", t, func() {

		b := NewBahamut(APIServerConfig{}, PushServerConfig{})
		auth := &Auth{}

		h := http.Header{}
		h.Add("Origin", "http://origin.com")
		ctx := &Context{
			Info: &Info{
				Headers: h,
			},
		}

		Convey("When I check authentication with no registered authenticator", func() {

			ok := CheckAuthentication(ctx, nil)

			Convey("Then it should be authenticated", func() {
				So(ok, ShouldBeTrue)
			})
		})

		Convey("When I check the authentication with a registered authenticator", func() {

			auth.authenticated = true
			auth.errored = false
			b.SetAuthenticator(auth)

			ok := CheckAuthentication(ctx, nil)

			Convey("Then it should be authenticated", func() {
				So(ok, ShouldBeTrue)
			})
		})

		Convey("When I check the authentication with a registered authenticator that returns no", func() {

			auth.authenticated = false
			auth.errored = false
			b.SetAuthenticator(auth)

			w := httptest.NewRecorder()
			ok := CheckAuthentication(ctx, w)

			Convey("Then it should not be authenticated", func() {
				So(ok, ShouldBeFalse)
			})

			Convey("Then the http status should be 500", func() {
				So(w.Code, ShouldEqual, 401)
			})
		})

		Convey("When I check the authentication with a registered authenticator that returns an error", func() {

			auth.authenticated = false
			auth.errored = true
			b.SetAuthenticator(auth)

			w := httptest.NewRecorder()
			ok := CheckAuthentication(ctx, w)

			Convey("Then it should be authenticated", func() {
				So(ok, ShouldBeFalse)
			})

			Convey("Then the http status should be 500", func() {
				So(w.Code, ShouldEqual, 500)
			})
		})
	})
}

func TestProcessorHelpers_checkAuthorized(t *testing.T) {

	Convey("Given I create a new Bahamut", t, func() {

		b := NewBahamut(APIServerConfig{}, PushServerConfig{})
		auth := &Auth{}

		h := http.Header{}
		h.Add("Origin", "http://origin.com")
		ctx := &Context{
			Info: &Info{
				Headers: h,
			},
		}

		Convey("When I check authorization with no registered authorizer", func() {

			ok := CheckAuthorization(ctx, nil)

			Convey("Then it should be authorized", func() {
				So(ok, ShouldBeTrue)
			})
		})

		Convey("When I check the authorization with a registered authorizer", func() {

			auth.authorized = true
			auth.errored = false
			b.SetAuthorizer(auth)

			ok := CheckAuthorization(ctx, nil)

			Convey("Then it should be authorized", func() {
				So(ok, ShouldBeTrue)
			})
		})

		Convey("When I check the authorization with a registered authorizer that returns no", func() {

			auth.authorized = false
			auth.errored = false
			b.SetAuthorizer(auth)

			w := httptest.NewRecorder()
			ok := CheckAuthorization(ctx, w)

			Convey("Then it should not be authorized", func() {
				So(ok, ShouldBeFalse)
			})

			Convey("Then the http status should be 500", func() {
				So(w.Code, ShouldEqual, 403)
			})
		})

		Convey("When I check the authorization with a registered authorizer that returns an error", func() {

			auth.authorized = false
			auth.errored = true
			b.SetAuthorizer(auth)

			w := httptest.NewRecorder()
			ok := CheckAuthorization(ctx, w)

			Convey("Then it should be authorized", func() {
				So(ok, ShouldBeFalse)
			})

			Convey("Then the http status should be 500", func() {
				So(w.Code, ShouldEqual, 500)
			})
		})
	})
}
