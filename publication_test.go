package bahamut

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestPublication_NewPublication(t *testing.T) {

	Convey("Given I create a new Publication", t, func() {

		publication := NewPublication("topic")

		Convey("Then the publication should be correctly initialized", func() {
			So(publication.Topic, ShouldEqual, "topic")
		})
	})
}

func TestPublication_EncodeDecode(t *testing.T) {

	Convey("Given I create a new Publication", t, func() {

		publication := NewPublication("topic")

		Convey("When I encode some object", func() {

			list := NewList()
			list.Name = "l1"
			list.ID = "xxx"

			err := publication.Encode(list)

			Convey("Then err should be nil", func() {
				So(err, ShouldBeNil)
			})

			Convey("Then the publication contains the correct data", func() {
				So(string(publication.Data()), ShouldEqual, "{\"ID\":\"xxx\",\"creationOnly\":\"\",\"description\":\"\",\"name\":\"l1\",\"parentID\":\"\",\"parentType\":\"\",\"readOnly\":\"\"}\n")
			})

			Convey("When I decode the object", func() {

				var l2 *List
				err := publication.Decode(l2)

				Convey("Then err should be nil", func() {
					So(err, ShouldBeNil)
				})

				Convey("Then l2 should ressemble to l1", func() {
					So(l2, ShouldResemble, l2)
				})
			})
		})

		Convey("When I encode some unencodable object", func() {

			list := NewUnmarshalableList()
			list.Name = "l1"
			list.ID = "xxx"

			err := publication.Encode(list)

			Convey("Then err should not be nil", func() {
				So(err, ShouldNotBeNil)
			})

			Convey("Then the publication contains the correct data", func() {
				So(string(publication.Data()), ShouldEqual, "")
			})

			Convey("When I decode the non existing object", func() {

				var l2 *List
				err := publication.Decode(l2)

				Convey("Then err should not be nil", func() {
					So(err, ShouldNotBeNil)
				})
			})
		})
	})
}
