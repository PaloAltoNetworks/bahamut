package simple

import (
	"fmt"
	"testing"

	"go.aporeto.io/bahamut"
	. "github.com/smartystreets/goconvey/convey"
)

func TestAuththenticator_NewAuthenticator(t *testing.T) {

	Convey("Given I call NewAuthenticator with two funcs", t, func() {

		f1 := func(*bahamut.Context) (bahamut.AuthAction, error) { return bahamut.AuthActionOK, nil }
		f2 := func(bahamut.Session) (bahamut.AuthAction, error) { return bahamut.AuthActionOK, nil }

		auth := NewAuthenticator(f1, f2)

		Convey("Then it should be correctly initialized", func() {
			So(auth.customAuthRequestFunc, ShouldEqual, f1)
			So(auth.customAuthSessionFunc, ShouldEqual, f2)
		})
	})
}

func TestAuththenticator_AuthenticateRequest(t *testing.T) {

	Convey("Given I call NewAuthenticator and a func that says ok", t, func() {

		f1 := func(*bahamut.Context) (bahamut.AuthAction, error) { return bahamut.AuthActionOK, nil }

		auth := NewAuthenticator(f1, nil)

		Convey("When I call AuthenticateRequest", func() {

			action, err := auth.AuthenticateRequest(nil)

			Convey("Then err should be nil", func() {
				So(err, ShouldBeNil)
			})

			Convey("Then action should be OK", func() {
				So(action, ShouldEqual, bahamut.AuthActionOK)
			})
		})
	})

	Convey("Given I call NewAuthenticator and no func", t, func() {

		auth := NewAuthenticator(nil, nil)

		Convey("When I call AuthenticateRequest", func() {

			action, err := auth.AuthenticateRequest(nil)

			Convey("Then err should be nil", func() {
				So(err, ShouldBeNil)
			})

			Convey("Then action should be Continue", func() {
				So(action, ShouldEqual, bahamut.AuthActionContinue)
			})
		})
	})

	Convey("Given I call NewAuthenticator and a func that returns an error", t, func() {

		f1 := func(*bahamut.Context) (bahamut.AuthAction, error) { return bahamut.AuthActionOK, fmt.Errorf("paf") }

		auth := NewAuthenticator(f1, nil)

		Convey("When I call AuthenticateRequest", func() {

			action, err := auth.AuthenticateRequest(nil)

			Convey("Then err should not be nil", func() {
				So(err.Error(), ShouldEqual, "paf")
			})

			Convey("Then action should be KO", func() {
				So(action, ShouldEqual, bahamut.AuthActionKO)
			})
		})
	})
}

func TestAuththenticator_AuthenticateSession(t *testing.T) {

	Convey("Given I call NewAuthenticator and a func that says ok", t, func() {

		f1 := func(bahamut.Session) (bahamut.AuthAction, error) { return bahamut.AuthActionOK, nil }

		auth := NewAuthenticator(nil, f1)

		Convey("When I call AuthenticateSession", func() {

			action, err := auth.AuthenticateSession(nil)

			Convey("Then err should be nil", func() {
				So(err, ShouldBeNil)
			})

			Convey("Then action should be OK", func() {
				So(action, ShouldEqual, bahamut.AuthActionOK)
			})
		})
	})

	Convey("Given I call NewAuthenticator and no func", t, func() {

		auth := NewAuthenticator(nil, nil)

		Convey("When I call AuthenticateSession", func() {

			action, err := auth.AuthenticateSession(nil)

			Convey("Then err should be nil", func() {
				So(err, ShouldBeNil)
			})

			Convey("Then action should be Continue", func() {
				So(action, ShouldEqual, bahamut.AuthActionContinue)
			})
		})
	})

	Convey("Given I call NewAuthenticator and a func that returns an error", t, func() {

		f1 := func(bahamut.Session) (bahamut.AuthAction, error) { return bahamut.AuthActionOK, fmt.Errorf("paf") }

		auth := NewAuthenticator(nil, f1)

		Convey("When I call AuthenticateSession", func() {

			action, err := auth.AuthenticateSession(nil)

			Convey("Then err should not be nil", func() {
				So(err.Error(), ShouldEqual, "paf")
			})

			Convey("Then action should be KO", func() {
				So(action, ShouldEqual, bahamut.AuthActionKO)
			})
		})
	})
}
