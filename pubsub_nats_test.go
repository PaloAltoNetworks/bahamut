package bahamut

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestNats_NewPubSubServer(t *testing.T) {

	Convey("Given I create a new PubSubServer", t, func() {

		ps := newNatsPubSub("nats://localhost:4222", "cid", "id")

		Convey("Then the PubSubServer should be correctly initialized", func() {
			So(ps.natsURL, ShouldEqual, "nats://localhost:4222")
			So(ps.clusterID, ShouldEqual, "cid")
			So(ps.clientID, ShouldEqual, "id")
		})
	})
}
