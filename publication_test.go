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
	"testing"
	"time"

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

func TestReply(t *testing.T) {

	threshold := 100 * time.Millisecond
	testCases := []struct {
		description    string
		setup          func() *Publication
		response       *Publication
		expectingReply bool
		shouldError    bool
	}{
		{
			description: "should send publication response to reply channel",
			setup: func() *Publication {
				return &Publication{
					replyCh: make(chan *Publication),
				}
			},
			response:       NewPublication("test topic"),
			expectingReply: true,
			shouldError:    false,
		},
		{
			description: "should return an error if no reply channel is configured for publication",
			setup: func() *Publication {
				return NewPublication("")
			},
			response:    NewPublication("test topic"),
			shouldError: true,
		},
		{
			description: "should return an error if passed in a nil publication",
			setup: func() *Publication {
				return &Publication{
					replyCh: make(chan *Publication),
				}
			},
			response:    nil,
			shouldError: true,
		},
		{
			description: "should return an error if publication has already been responded to",
			setup: func() *Publication {
				return &Publication{
					replyCh: make(chan *Publication),
					replied: true,
				}
			},
			response:    NewPublication("test topic"),
			shouldError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			done := make(chan struct{})
			pub := tc.setup()

			go func() {
				if pub.replyCh != nil {
					select {
					case response := <-pub.replyCh:
						if !tc.expectingReply {
							t.Errorf("was NOT expecting to get a response in reply channel, but got: %+v", response)
						}
					case <-time.After(threshold):
						if tc.expectingReply {
							t.Errorf("expected to get a response in reply channel, but got nothing.")
						}
					}
				}

				close(done)
			}()

			if err := pub.Reply(tc.response); !tc.shouldError && err != nil {
				t.Errorf("error returned when none was expected - error: %+v", err)
			}

			<-done
		})
	}
}
