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

func (m MockPinger) Ping(timeout time.Duration) error {
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
