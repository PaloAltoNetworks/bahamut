package bahamut

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestPubsub_NewServer(t *testing.T) {

	Convey("Given I create a new KafkaPubSubServer", t, func() {

		ps := NewKafkaPubSubServer([]string{"123:123"})

		Convey("Then the PubSubServer should be correctly initialized", func() {
			So(ps, ShouldImplement, (*PubSubServer)(nil))
		})
	})

	Convey("Given I create a new localPubSubServer", t, func() {

		ps := NewLocalPubSubServer(nil)

		Convey("Then the PubSubServer should be correctly initialized", func() {
			So(ps, ShouldImplement, (*PubSubServer)(nil))
		})
	})
}
