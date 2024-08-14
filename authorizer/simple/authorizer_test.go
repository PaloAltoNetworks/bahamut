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

	// nolint:revive // Allow dot imports for readability in tests
	. "github.com/smartystreets/goconvey/convey"
	"go.aporeto.io/bahamut"
)

func TestAuthorizer_NewAuthorizer(t *testing.T) {

	Convey("Given I call NewAuthorizer with one funcs", t, func() {

		f1 := func(bahamut.Context) (bahamut.AuthAction, error) { return bahamut.AuthActionOK, nil }

		auth := NewAuthorizer(f1)

		Convey("Then it should be correctly initialized", func() {
			So(auth.customAuthFunc, ShouldEqual, f1)

		})
	})
}

func TestAuthorizer_IsAuthorized(t *testing.T) {

	Convey("Given I call NewAuthorizer and a func that says ok", t, func() {

		f1 := func(bahamut.Context) (bahamut.AuthAction, error) { return bahamut.AuthActionOK, nil }

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

		f1 := func(bahamut.Context) (bahamut.AuthAction, error) { return bahamut.AuthActionOK, fmt.Errorf("paf") }

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
