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
	"go.aporeto.io/elemental"
)

func TestPublishHandler_NewPublishHandler(t *testing.T) {

	Convey("Given I call NewPublishHandler with one funcs", t, func() {

		f1 := func(*elemental.Event) (bool, error) { return true, nil }

		pub := NewPublishHandler(f1)

		Convey("Then it should be correctly initialized", func() {
			So(pub.shouldPublishFunc, ShouldEqual, f1)

		})
	})
}

func TestPublishHandler_ShouldPublish(t *testing.T) {

	Convey("Given I call NewPublishHandler and a func that says ok", t, func() {

		f1 := func(*elemental.Event) (bool, error) { return true, nil }

		pub := NewPublishHandler(f1)

		Convey("When I call ShouldPublish", func() {

			action, err := pub.ShouldPublish(nil)

			Convey("Then err should be nil", func() {
				So(err, ShouldBeNil)
			})

			Convey("Then action should be OK", func() {
				So(action, ShouldEqual, true)
			})
		})
	})

	Convey("Given I call NewPublishHandler and no func", t, func() {

		pub := NewPublishHandler(nil)

		Convey("When I call ShouldPublish", func() {

			action, err := pub.ShouldPublish(nil)

			Convey("Then err should be nil", func() {
				So(err, ShouldBeNil)
			})

			Convey("Then action should be Continue", func() {
				So(action, ShouldEqual, true)
			})
		})
	})

	Convey("Given I call NewPublishHandler and a func that returns an error", t, func() {

		f1 := func(*elemental.Event) (bool, error) { return false, fmt.Errorf("paf") }

		pub := NewPublishHandler(f1)

		Convey("When I call ShouldPublish", func() {

			action, err := pub.ShouldPublish(nil)

			Convey("Then err should not be nil", func() {
				So(err.Error(), ShouldEqual, "paf")
			})

			Convey("Then action should be KO", func() {
				So(action, ShouldEqual, false)
			})
		})
	})
}
