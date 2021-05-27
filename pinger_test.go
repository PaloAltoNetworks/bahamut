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

package bahamut

import (
	"fmt"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

type MockPinger struct {
	PingStatus error
}

func (m MockPinger) Ping(_ time.Duration) error {
	return m.PingStatus
}

func Test_RetrieveHealthStatus(t *testing.T) {

	Convey("Given the following pingers", t, func() {
		pingers := map[string]Pinger{
			"p1": MockPinger{PingStatus: nil},
			"p2": MockPinger{PingStatus: fmt.Errorf(PingStatusTimeout)},
			"p3": MockPinger{PingStatus: fmt.Errorf("Another status")},
		}
		results := RetrieveHealthStatus(time.Second, pingers)

		Convey("Then I should have the following status results", func() {
			So(results, ShouldNotBeNil)
		})
	})

	Convey("Given the following pingers", t, func() {
		pingers := map[string]Pinger{
			"p1": MockPinger{PingStatus: nil},
			"p2": MockPinger{PingStatus: nil},
			"p3": MockPinger{PingStatus: nil},
		}
		results := RetrieveHealthStatus(time.Second, pingers)

		Convey("Then I should have the following status results", func() {
			So(results, ShouldBeNil)
		})
	})
}

func Test_stringifyStatus(t *testing.T) {

	Convey("Given the stringifyStatus method", t, func() {
		So(stringifyStatus(nil), ShouldEqual, PingStatusOK)
		So(stringifyStatus(fmt.Errorf(PingStatusTimeout)), ShouldEqual, PingStatusTimeout)
		So(stringifyStatus(fmt.Errorf("Another status")), ShouldEqual, PingStatusError)
	})
}
