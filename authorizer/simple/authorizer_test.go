package simple

import (
	"fmt"
	"testing"

	"go.aporeto.io/bahamut"

	. "github.com/smartystreets/goconvey/convey"
)

func TestAuthorizer_NewAuthorizer(t *testing.T) {

	Convey("Given I call NewAuthorizer with one funcs", t, func() {

		f1 := func(*bahamut.Context) (bahamut.AuthAction, error) { return bahamut.AuthActionOK, nil }

		auth := NewAuthorizer(f1)

		Convey("Then it should be correctly initialized", func() {
			So(auth.customAuthFunc, ShouldEqual, f1)

		})
	})
}

func TestAuthorizer_IsAuthorized(t *testing.T) {

	Convey("Given I call NewAuthorizer and a func that says ok", t, func() {

		f1 := func(*bahamut.Context) (bahamut.AuthAction, error) { return bahamut.AuthActionOK, nil }

		auth := NewAuthorizer(f1)

		Convey("When I call IsAuthorized", func() {

			action, err := auth.IsAuthorized(nil)

			Convey("Then err should be nil", func() {
				So(err, ShouldBeNil)
			})

			Convey("Then action should be OK", func() {
				So(action, ShouldEqual, bahamut.AuthActionOK)
			})
		})
	})

	Convey("Given I call NewAuthorizer and no func", t, func() {

		auth := NewAuthorizer(nil)

		Convey("When I call IsAuthorized", func() {

			action, err := auth.IsAuthorized(nil)

			Convey("Then err should be nil", func() {
				So(err, ShouldBeNil)
			})

			Convey("Then action should be Continue", func() {
				So(action, ShouldEqual, bahamut.AuthActionContinue)
			})
		})
	})

	Convey("Given I call NewAuthorizer and a func that returns an error", t, func() {

		f1 := func(*bahamut.Context) (bahamut.AuthAction, error) { return bahamut.AuthActionOK, fmt.Errorf("paf") }

		auth := NewAuthorizer(f1)

		Convey("When I call IsAuthorized", func() {

			action, err := auth.IsAuthorized(nil)

			Convey("Then err should not be nil", func() {
				So(err.Error(), ShouldEqual, "paf")
			})

			Convey("Then action should be KO", func() {
				So(action, ShouldEqual, bahamut.AuthActionKO)
			})
		})
	})
}
