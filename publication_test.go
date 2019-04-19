package bahamut

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"go.aporeto.io/elemental"
	testmodel "go.aporeto.io/elemental/test/model"
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

		tracer := &mockTracer{}
		publication := NewPublication("topic")
		publication.StartTracing(tracer, "test")

		Convey("When I encode some object using JSON encoding", func() {

			list := testmodel.NewList()
			list.Name = "l1"
			list.ID = "xxx"

			err := publication.EncodeWithEncoding(list, elemental.EncodingTypeJSON)

			Convey("Then err should be nil", func() {
				So(err, ShouldBeNil)
			})

			Convey("Then the publication contains the correct data", func() {
				d, _ := elemental.Encode(elemental.EncodingTypeJSON, list)
				So(publication.Data, ShouldResemble, d)
			})

			Convey("When I decode the object", func() {

				l2 := testmodel.NewList()
				err := publication.Decode(l2)

				Convey("Then err should be nil", func() {
					So(err, ShouldBeNil)
				})

				Convey("Then l2 should resemble to l1", func() {
					So(l2, ShouldResemble, list)
				})
			})
		})

		Convey("When I encode some object using MSGPACK encoding", func() {

			list := testmodel.NewList()
			list.Name = "l1"
			list.ID = "xxx"

			err := publication.EncodeWithEncoding(list, elemental.EncodingTypeMSGPACK)

			Convey("Then err should be nil", func() {
				So(err, ShouldBeNil)
			})

			Convey("Then the publication contains the correct data", func() {
				d, _ := elemental.Encode(elemental.EncodingTypeMSGPACK, list)
				So(publication.Data, ShouldResemble, d)
			})

			Convey("When I decode the object", func() {

				l2 := testmodel.NewList()
				err := publication.Decode(l2)

				Convey("Then err should be nil", func() {
					So(err, ShouldBeNil)
				})

				Convey("Then l2 should resemble to l1", func() {
					So(l2, ShouldResemble, list)
				})
			})
		})

		Convey("When I encode some unencodable object", func() {

			list := testmodel.NewUnmarshalableList()
			list.Name = "l1"
			list.ID = "xxx"

			err := publication.EncodeWithEncoding(list, elemental.EncodingTypeJSON)

			Convey("Then err should not be nil", func() {
				So(err, ShouldNotBeNil)
			})

			Convey("Then the publication contains the correct data", func() {
				So(string(publication.Data), ShouldEqual, "")
			})

			Convey("When I decode the non existing object", func() {

				l2 := testmodel.NewList()
				err := publication.Decode(l2)

				Convey("Then err should not be nil", func() {
					So(err, ShouldNotBeNil)
				})
			})
		})
	})
}

func TestPublicationTracing(t *testing.T) {

	Convey("Given I have no tracer", t, func() {

		publication := NewPublication("topic")

		Convey("When I call StartTracing", func() {

			publication.StartTracing(nil, "test")

			Convey("Then the span should be correct", func() {
				So(publication.Span(), ShouldBeNil)
			})
		})

		Convey("When I call StartTracingFromSpan", func() {

			span := &mockSpan{}
			err := publication.StartTracingFromSpan(span, "test")

			Convey("Then err should be nil", func() {
				So(err, ShouldBeNil)
			})

			Convey("Then the span should be correct", func() {
				So(publication.Span(), ShouldBeNil)
			})
		})
	})

	Convey("Given I have a tracer", t, func() {

		tracer := &mockTracer{}
		publication := NewPublication("topic")

		Convey("When I call StartTracing", func() {

			publication.StartTracing(tracer, "test")

			Convey("Then the span should be correct", func() {
				So(publication.Span(), ShouldNotBeNil)
			})
		})

		Convey("When I call StartTracingFromSpan", func() {

			span := newMockSpan(tracer)
			err := publication.StartTracingFromSpan(span, "test")

			Convey("Then err should be nil", func() {
				So(err, ShouldBeNil)
			})

			Convey("Then the span should be correct", func() {
				So(publication.Span(), ShouldNotBeNil)
			})
		})
	})
}

func TestPublication_Duplicate(t *testing.T) {

	Convey("Given I have a publication", t, func() {

		pub := NewPublication("topic")
		pub.Data = []byte("data")
		pub.Partition = 12
		pub.TrackingName = "TrackingName"

		Convey("When I call duplicate", func() {

			dup := pub.Duplicate()

			Convey("Then the copy should be correct", func() {
				So(dup, ShouldNotEqual, pub)
				So(dup.Data, ShouldResemble, pub.Data)
				So(dup.Partition, ShouldEqual, pub.Partition)
				So(dup.TrackingName, ShouldEqual, pub.TrackingName)
				So(dup.Topic, ShouldEqual, pub.Topic)
				So(dup.Encoding, ShouldEqual, pub.Encoding)
			})
		})
	})
}
