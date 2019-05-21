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
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"reflect"
	"testing"
	"time"

	"github.com/gofrs/uuid"
	"github.com/golang/mock/gomock"
	"github.com/nats-io/go-nats"
	natsserver "github.com/nats-io/nats-server/test"
	. "github.com/smartystreets/goconvey/convey"
	"go.aporeto.io/bahamut/mocks"
	"go.aporeto.io/elemental"
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
				natsOptClient(nil),
			},
			publishOptions: []PubSubOptPublish{},
		},
		{
			description: "should return an error if passed in a nil publication",
			publication: nil,
			setup: func(t *testing.T, mockClient *mocks.MockNATSClient, pub *Publication) {
				mockClient.
					EXPECT().
					Publish(gomock.Any(), gomock.Any()).
					// should never be called in this case!
					Times(0)
			},
			expectedErrType: errors.New(""),
			natsOptions:     []NATSOption{},
			publishOptions:  []PubSubOptPublish{},
		},
		{
			description: "should return an error if Publish returns an error",
			publication: NewPublication("test topic"),
			setup: func(t *testing.T, mockClient *mocks.MockNATSClient, pub *Publication) {
				mockClient.
					EXPECT().
					Publish(pub.Topic, gomock.Any()).
					Return(errors.New("failed to publish")).
					Times(1)
			},
			expectedErrType: errors.New(""),
			natsOptions:     []NATSOption{},
			publishOptions:  []PubSubOptPublish{},
		},
		{
			description: "should be able to provide a custom reply validator using the NATSOptPublishRequireAck option",
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
		{
			description: "should return an error if RequestWithContext returns an error",
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

				// notice how we are returning an error here!
				mockClient.
					EXPECT().
					RequestWithContext(gomock.Any(), pub.Topic, expectedData).
					Return(nil, errors.New("darn, failed to get a response")).
					Times(1)
			},
			expectedErrType: errors.New(""),
			natsOptions:     []NATSOption{},
			publishOptions: []PubSubOptPublish{
				// this is a friendly validator, it always passes successfully!
				NATSOptPublishReplyValidator(context.Background(), func(_ *nats.Msg) error {
					return nil
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
			test.natsOptions = append([]NATSOption{natsOptClient(mockNATSClient)}, test.natsOptions...)
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

func TestSubscribe(t *testing.T) {

	srv := natsserver.RunDefaultServer()
	defer srv.Shutdown()
	nc := newDefaultConnection(t)
	defer nc.Close()

	threshold := 500 * time.Millisecond
	subscribeTopic := "test-topic"
	serverAddr := srv.Addr().(*net.TCPAddr)
	ps := NewNATSPubSubClient(
		fmt.Sprintf("%s:%d", serverAddr.IP, serverAddr.Port),
		natsOptClient(nc),
	)

	var tests = []struct {
		description         string
		expectedPublication *Publication
		setup               func(t *testing.T, pub *Publication)
		subscribeOptions    []PubSubOptSubscribe
	}{
		{
			description: "should successfully subscribe to topic and receive a publication in provided channel",
			setup: func(t *testing.T, pub *Publication) {
				data, err := elemental.Encode(elemental.EncodingTypeMSGPACK, pub)
				if err != nil {
					t.Fatalf("test setup failed - could not encode publication - error: %+v", err)
					return
				}

				if err := nc.Publish(subscribeTopic, data); err != nil {
					t.Fatalf("test setup failed - could not publish publication - error: %+v", err)
					return
				}

			},
			expectedPublication: &Publication{
				Topic: subscribeTopic,
				Data:  []byte("message"),
			},
			subscribeOptions: nil,
		},
		{
			description: "should NOT receive anything in publication channel that cannot be decoded into a publication",
			setup: func(t *testing.T, pub *Publication) {
				if err := nc.Publish(subscribeTopic, []byte("this cannot be decoded into a publication")); err != nil {
					t.Fatalf("test setup failed - could not publish publication - error: %+v", err)
					return
				}

			},
			expectedPublication: nil,
			subscribeOptions:    nil,
		},
		{
			description: "should respond back with an ACK message to all publications that expect an ACK response",
			setup: func(t *testing.T, pub *Publication) {
				data, err := elemental.Encode(elemental.EncodingTypeMSGPACK, pub)
				if err != nil {
					t.Fatalf("test setup failed - could not encode publication - error: %+v", err)
					return
				}

				// publish a message expecting a response
				msg, err := nc.Request(subscribeTopic, data, threshold)
				if err != nil {
					t.Fatalf("test setup failed to publish/receive message - error: %+v", err)
					return
				}

				// validate that the received response is an ACK message
				if !bytes.Equal(msg.Data, ackMessage) {
					t.Errorf("expected response message data to be \"%s\", but received \"%s\"", ackMessage, msg.Data)
				}

			},
			expectedPublication: &Publication{
				Topic: subscribeTopic,
				Data:  []byte("message"),
			},
			subscribeOptions: nil,
		},
		{
			description: "should be able to set a custom response using the NATSOptSubscribeReplyer option to all publications that expect a response",
			subscribeOptions: []PubSubOptSubscribe{
				NATSOptSubscribeReplyer(func(_ *nats.Msg) []byte {
					return []byte("custom response message")
				}),
			},
			setup: func(t *testing.T, pub *Publication) {
				data, err := elemental.Encode(elemental.EncodingTypeMSGPACK, pub)
				if err != nil {
					t.Fatalf("test setup failed - could not encode publication - error: %+v", err)
					return
				}

				// publish a message expecting a (custom) response
				msg, err := nc.Request(subscribeTopic, data, threshold)
				if err != nil {
					t.Fatalf("test setup failed to publish/receive message - error: %+v", err)
					return
				}

				// validate that the received response is the custom response message we specified using the NATSOptSubscribeReplyer option
				if !bytes.Equal(msg.Data, []byte("custom response message")) {
					t.Errorf("expected response message data to be \"%s\", but received \"%s\"", ackMessage, msg.Data)
				}

			},
			expectedPublication: &Publication{
				Topic: subscribeTopic,
				Data:  []byte("message"),
			},
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {

			publications := make(chan *Publication)
			errors := make(chan error)
			unsub := ps.Subscribe(publications, errors, subscribeTopic, test.subscribeOptions...)
			defer unsub()
			test.setup(t, test.expectedPublication)

			select {
			case pub := <-publications:

				if test.expectedPublication == nil {
					t.Fatalf("did not expect to receive any publications, but received publication - \"%+v\"", pub)
				}

				if pub.Topic != test.expectedPublication.Topic {
					t.Errorf("expected publication with topic \"%s\", but received publication with topic \"%s\"", test.expectedPublication.Topic, pub.Topic)
				}

				if !bytes.Equal(pub.Data, test.expectedPublication.Data) {
					t.Errorf("expected publication with data: \"%+v\", but received publication with data: \"%+v\"", pub.Data, test.expectedPublication.Data)
				}

			case err := <-errors:
				t.Fatalf("received unexpected error - err: \"%+v\"", err)
			case <-time.After(threshold):
				if test.expectedPublication != nil {
					t.Fatalf("timed out waiting to receive a publication: %+v", test.expectedPublication)
				}
			}
		})
	}
}

func newDefaultConnection(t *testing.T) *nats.Conn {
	url := fmt.Sprintf("nats://127.0.0.1:%d", nats.DefaultPort)
	nc, err := nats.Connect(url)
	if err != nil {
		t.Fatalf("failed to create default connection: %v\n", err)
		return nil
	}
	return nc
}
