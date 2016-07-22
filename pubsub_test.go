package bahamut

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestPubsub_NewServer(t *testing.T) {

	Convey("Given I create a new pubsub.Server", t, func() {

		ps := NewPubSub([]string{"123:123"})

		Convey("Then the PubSubServer should be correctly initialized", func() {
			So(ps, ShouldImplement, (*PublisherSubscriber)(nil))
		})
	})
}
