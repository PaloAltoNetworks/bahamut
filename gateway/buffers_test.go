package gateway

import (
	"testing"

	// nolint:revive // Allow dot imports for readability in tests
	. "github.com/smartystreets/goconvey/convey"
)

func TestBufferPool(t *testing.T) {

	Convey("Given I create a new buffer pool", t, func() {

		bp := newPool(42)
		Convey("Then bp should be correct", func() {
			So(bp, ShouldNotBeNil)
		})

		buff := bp.Get()

		Convey("Then the size of a buffer should be correct", func() {
			So(cap(buff), ShouldEqual, 42)
		})

		bp.Put(buff)
	})
}
