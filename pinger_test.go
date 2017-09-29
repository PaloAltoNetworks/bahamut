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
			So(len(results), ShouldEqual, len(pingers))
			So(results["p1"], ShouldEqual, PingStatusOK)
			So(results["p2"], ShouldEqual, PingStatusTimeout)
			So(results["p3"], ShouldEqual, PingStatusError)
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
