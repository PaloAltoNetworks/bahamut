// Copyright 2019 Aporeto Inc.
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//     http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package simple

import (
	"fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"go.aporeto.io/bahamut"
)

func TestAuththenticator_NewAuthenticator(t *testing.T) {

	Convey("Given I call NewAuthenticator with two funcs", t, func() {

		f1 := func(bahamut.Context) (bahamut.AuthAction, error) { return bahamut.AuthActionOK, nil }
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

		f1 := func(bahamut.Context) (bahamut.AuthAction, error) { return bahamut.AuthActionOK, nil }

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

		f1 := func(bahamut.Context) (bahamut.AuthAction, error) { return bahamut.AuthActionOK, fmt.Errorf("paf") }

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
