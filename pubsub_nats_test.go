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
	"context"
	"crypto/tls"
	"errors"
	"reflect"
	"testing"
	"time"

	"go.aporeto.io/elemental"

	"github.com/nats-io/go-nats"

	"github.com/gofrs/uuid"
	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"
	"go.aporeto.io/bahamut/mocks"
)

func TestNats_NewNATSPubSubClient(t *testing.T) {

	Convey("Given I create a new PubSubServer with no option", t, func() {

		ps := NewNATSPubSubClient("nats://localhost:4222").(*natsPubSub)

		Convey("Then the PubSubServer should be correctly initialized", func() {
			So(ps.natsURL, ShouldEqual, "nats://localhost:4222")
			So(ps.clusterID, ShouldEqual, "test-cluster")
			So(ps.clientID, ShouldNotBeEmpty)
			So(ps.username, ShouldEqual, "")
			So(ps.password, ShouldEqual, "")
			So(ps.retryInterval, ShouldBeGreaterThan, 0)
			So(ps.tlsConfig, ShouldEqual, nil)
			// verify that client id is a proper V4 UUID
			id, err := uuid.FromString(ps.clientID)
			So(err, ShouldBeNil)
			So(id.Version(), ShouldEqual, uuid.V4)
		})
	})

	Convey("Given I create a new PubSubServer with all options", t, func() {

		tlsconfig := &tls.Config{}

		ps := NewNATSPubSubClient(
			"nats://localhost:4222",
			NATSOptClusterID("cid"),
			NATSOptClientID("id"),
			NATSOptCredentials("username", "password"),
			NATSOptTLS(tlsconfig),
			NATSOptConnectRetryInterval(500*time.Millisecond),
		).(*natsPubSub)

		Convey("Then the PubSubServer should be correctly initialized", func() {
			So(ps.natsURL, ShouldEqual, "nats://localhost:4222")
			So(ps.clusterID, ShouldEqual, "cid")
			So(ps.clientID, ShouldEqual, "id")
			So(ps.username, ShouldEqual, "username")
			So(ps.password, ShouldEqual, "password")
			So(ps.tlsConfig, ShouldEqual, tlsconfig)
			So(ps.retryInterval, ShouldEqual, 500*time.Millisecond)
		})
	})
}

func TestPublish(t *testing.T) {

	natsURL := "nats://localhost:4222"

	tests := []struct {
		description     string
		setup           func(t *testing.T, mockClient *mocks.MockNATSClient, pub *Publication)
		expectedErrType error
		publication     *Publication
		natsOptions     []NATSOption
		publishOptions  []PubSubOptPublish
	}{
		{
			description: "should successfully publish publication",
			publication: NewPublication("test topic"),
			setup: func(t *testing.T, mockClient *mocks.MockNATSClient, pub *Publication) {

				mockClient.
					EXPECT().
					Publish(pub.Topic, gomock.Any()).
					Return(nil).
					Times(1)

				mockClient.
					EXPECT().
					RequestWithContext(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(0)
			},
			expectedErrType: nil,
			natsOptions:     []NATSOption{},
			publishOptions:  []PubSubOptPublish{},
		},
		{
			description: "should return a NoClientError error if no NATS client had been connected",
			publication: NewPublication("test topic"),
			setup: func(t *testing.T, mockClient *mocks.MockNATSClient, pub *Publication) {
				mockClient.
					EXPECT().
					Publish(pub.Topic, gomock.Any()).
					// should never be called in this case!
					Times(0)
			},
			expectedErrType: &NoClientError{},
			natsOptions: []NATSOption{
				// notice how we pass a nil client explicitly to simulate the failure scenario
				// desired by this test
				NATSOptClient(nil),
			},
			publishOptions: []PubSubOptPublish{},
		},
		{
			description: "should return an EncodingError error if the publication fails to get encoded",
			// pass in a nil publication to cause an EncodingError!
			publication: nil,
			setup: func(t *testing.T, mockClient *mocks.MockNATSClient, pub *Publication) {
				mockClient.
					EXPECT().
					Publish(gomock.Any(), gomock.Any()).
					// should never be called in this case!
					Times(0)
			},
			expectedErrType: &EncodingError{},
			natsOptions:     []NATSOption{},
			publishOptions:  []PubSubOptPublish{},
		},
		{
			description: "should be able to provide a custom reply validator using the NATSOptPublishRequireAck option",
			// pass in a nil publication to cause an EncodingError!
			publication: NewPublication("test topic"),
			setup: func(t *testing.T, mockClient *mocks.MockNATSClient, pub *Publication) {
				mockClient.
					EXPECT().
					Publish(gomock.Any(), gomock.Any()).
					// should never be called in this case as passing in the NATSOptPublishReplyValidator
					// will cause Publish to use the Request-Reply pattern that is synchronous (i.e. will not
					// return until a response is returned or we timeout waiting for one)
					// See Request-Reply: https://nats.io/documentation/writing_applications/publishing/
					Times(0)

				expectedData, err := elemental.Encode(elemental.EncodingTypeMSGPACK, pub)
				if err != nil {
					t.Fatalf("test setup failed - could not encode publication - error: %+v", err)
					return
				}

				mockClient.
					EXPECT().
					RequestWithContext(gomock.Any(), pub.Topic, expectedData).
					Return(&nats.Msg{
						Data: []byte("helloworld"),
					}, nil).
					Times(1)
			},
			expectedErrType: nil,
			natsOptions:     []NATSOption{},
			publishOptions: []PubSubOptPublish{
				// this is a friendly validator, it always passes successfully!
				NATSOptPublishReplyValidator(context.Background(), func(_ *nats.Msg) error {
					return nil
				}),
			},
		},
		{
			description: "should return an error if custom response validator fails response message validation",
			// pass in a nil publication to cause an EncodingError!
			publication: NewPublication("test topic"),
			setup: func(t *testing.T, mockClient *mocks.MockNATSClient, pub *Publication) {
				mockClient.
					EXPECT().
					Publish(gomock.Any(), gomock.Any()).
					// should never be called in this case as passing in the NATSOptPublishReplyValidator
					// will cause Publish to use the Request-Reply pattern that is synchronous (i.e. will not
					// return until a response is returned or we timeout waiting for one)
					// See Request-Reply: https://nats.io/documentation/writing_applications/publishing/
					Times(0)

				expectedData, err := elemental.Encode(elemental.EncodingTypeMSGPACK, pub)
				if err != nil {
					t.Fatalf("test setup failed - could not encode publication - error: %+v", err)
					return
				}

				mockClient.
					EXPECT().
					RequestWithContext(gomock.Any(), pub.Topic, expectedData).
					Return(&nats.Msg{
						Data: []byte("helloworld"),
					}, nil).
					Times(1)
			},
			expectedErrType: errors.New(""),
			natsOptions:     []NATSOption{},
			publishOptions: []PubSubOptPublish{
				// this is an un-friendly validator, it always fails!
				NATSOptPublishReplyValidator(context.Background(), func(_ *nats.Msg) error {
					return errors.New("message validation failed :-(")
				}),
			},
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockNATSClient := mocks.NewMockNATSClient(ctrl)
			test.setup(t, mockNATSClient, test.publication)

			// note: we prepend the NATSOption client option to use our mock client just in case the
			// test case wishes to override this option (e.g. to provide a nil client)
			test.natsOptions = append([]NATSOption{NATSOptClient(mockNATSClient)}, test.natsOptions...)
			ps := NewNATSPubSubClient(
				natsURL,
				test.natsOptions...,
			)

			if actualErrType := reflect.TypeOf(ps.Publish(test.publication, test.publishOptions...)); actualErrType != reflect.TypeOf(test.expectedErrType) {
				t.Errorf("Call to publish returned error of type \"%+v\", when an error of type \"%+v\" was expected", actualErrType, reflect.TypeOf(test.expectedErrType))
			}
		})
	}
}
