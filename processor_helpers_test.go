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

		auth := &Auth{}

		h := http.Header{}
		h.Add("Origin", "http://origin.com")
		ctx := &Context{
			Info: &Info{
				Headers: h,
			},
		}

		Convey("When I check authentication with no registered authenticator", func() {

			ok := CheckAuthentication(nil, ctx, nil)

			Convey("Then it should be authenticated", func() {
				So(ok, ShouldBeTrue)
			})
		})

		Convey("When I check the authentication with a registered authenticator", func() {

			s := NewServer(APIServerConfig{Authenticator: auth}, PushServerConfig{})
			auth.authenticated = true
			auth.errored = false

			ok := CheckAuthentication(s.Authenticator(), ctx, nil)

			Convey("Then it should be authenticated", func() {
				So(ok, ShouldBeTrue)
			})
		})

		Convey("When I check the authentication with a registered authenticator that returns no", func() {

			s := NewServer(APIServerConfig{Authenticator: auth}, PushServerConfig{})
			auth.authenticated = false
			auth.errored = false

			w := httptest.NewRecorder()
			ok := CheckAuthentication(s.Authenticator(), ctx, w)

			Convey("Then it should not be authenticated", func() {
				So(ok, ShouldBeFalse)
			})

			Convey("Then the http status should be 500", func() {
				So(w.Code, ShouldEqual, 401)
			})
		})

		Convey("When I check the authentication with a registered authenticator that returns an error", func() {

			s := NewServer(APIServerConfig{Authenticator: auth}, PushServerConfig{})
			auth.authenticated = false
			auth.errored = true

			w := httptest.NewRecorder()
			ok := CheckAuthentication(s.Authenticator(), ctx, w)

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

		auth := &Auth{}

		h := http.Header{}
		h.Add("Origin", "http://origin.com")
		ctx := &Context{
			Info: &Info{
				Headers: h,
			},
		}

		Convey("When I check authorization with no registered authorizer", func() {

			ok := CheckAuthorization(nil, ctx, nil)

			Convey("Then it should be authorized", func() {
				So(ok, ShouldBeTrue)
			})
		})

		Convey("When I check the authorization with a registered authorizer", func() {

			s := NewServer(APIServerConfig{Authorizer: auth}, PushServerConfig{})
			auth.authorized = true
			auth.errored = false

			ok := CheckAuthorization(s.Authorizer(), ctx, nil)

			Convey("Then it should be authorized", func() {
				So(ok, ShouldBeTrue)
			})
		})

		Convey("When I check the authorization with a registered authorizer that returns no", func() {

			s := NewServer(APIServerConfig{Authorizer: auth}, PushServerConfig{})
			auth.authorized = false
			auth.errored = false

			w := httptest.NewRecorder()
			ok := CheckAuthorization(s.Authorizer(), ctx, w)

			Convey("Then it should not be authorized", func() {
				So(ok, ShouldBeFalse)
			})

			Convey("Then the http status should be 500", func() {
				So(w.Code, ShouldEqual, 403)
			})
		})

		Convey("When I check the authorization with a registered authorizer that returns an error", func() {

			s := NewServer(APIServerConfig{Authorizer: auth}, PushServerConfig{})
			auth.authorized = false
			auth.errored = true

			w := httptest.NewRecorder()
			ok := CheckAuthorization(s.Authorizer(), ctx, w)

			Convey("Then it should be authorized", func() {
				So(ok, ShouldBeFalse)
			})

			Convey("Then the http status should be 500", func() {
				So(w.Code, ShouldEqual, 500)
			})
		})
	})
}
