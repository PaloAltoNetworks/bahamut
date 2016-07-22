package pubsub

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestPubsub_NewServer(t *testing.T) {

	Convey("Given I create a new pubsub.Server", t, func() {

		ps := NewServer([]string{"123:123"})

		Convey("Then the PubSubServer should be correctly initialized", func() {
			So(ps, ShouldImplement, (*Server)(nil))
		})
	})
}
