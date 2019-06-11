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
		description             string
		setup                   func(t *testing.T, mockClient *mocks.MockNATSClient, pub *Publication)
		expectedErrType         error
		publication             *Publication
		natsOptions             []NATSOption
		publishOptionsGenerator func(t *testing.T) ([]PubSubOptPublish, func())
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
		},
		{
			description: "should send received publication response to the configured response channel via the NATSOptRespondToChannel option",
			publication: NewPublication("test topic"),
			setup: func(t *testing.T, mockClient *mocks.MockNATSClient, pub *Publication) {

				mockClient.
					EXPECT().
					Publish(gomock.Any(), gomock.Any()).
					Times(0)

				// note: passing in the NATSOptRespondToChannel option should set the publication response mode
				// to ReplyWithPublication before encoding the publication
				pub.ResponseMode = ResponseModePublication
				expectedPublishData, err := elemental.Encode(elemental.EncodingTypeMSGPACK, pub)
				if err != nil {
					t.Fatalf("test setup failed - could not encode publication - error: %+v", err)
					return
				}

				responsePayload, err := elemental.Encode(elemental.EncodingTypeMSGPACK, &Publication{Data: []byte("hello world!")})
				if err != nil {
					t.Fatalf("test setup failed - could not encode publication - error: %+v", err)
					return
				}

				mockClient.
					EXPECT().
					RequestWithContext(gomock.Any(), pub.Topic, expectedPublishData).
					Return(&nats.Msg{
						Data: responsePayload,
					}, nil).
					Times(1)
			},
			publishOptionsGenerator: func(t *testing.T) ([]PubSubOptPublish, func()) {

				respCh := make(chan *Publication, 1)
				callback := func() {
					select {
					case response := <-respCh:
						if !bytes.Equal(response.Data, []byte("hello world!")) {
							t.Errorf("received a response, but data did not match. Hint: test setup may be broken")
						}
					default:
						t.Errorf("expected to receive a response in channel, but got nothing!")
					}
				}

				return []PubSubOptPublish{
					NATSOptRespondToChannel(context.Background(), respCh),
				}, callback
			},
			expectedErrType: nil,
			natsOptions:     []NATSOption{},
		},
		{
			description: "should return an error if the response could not be decoded into a Publication when using the NATSOptRespondToChannel option",
			publication: NewPublication("test topic"),
			setup: func(t *testing.T, mockClient *mocks.MockNATSClient, pub *Publication) {

				mockClient.
					EXPECT().
					Publish(gomock.Any(), gomock.Any()).
					Times(0)

				// note: passing in the NATSOptRespondToChannel option should set the publication response mode
				// to ReplyWithPublication before encoding the publication
				pub.ResponseMode = ResponseModePublication
				expectedPublishData, err := elemental.Encode(elemental.EncodingTypeMSGPACK, pub)
				if err != nil {
					t.Fatalf("test setup failed - could not encode publication - error: %+v", err)
					return
				}

				mockClient.
					EXPECT().
					RequestWithContext(gomock.Any(), pub.Topic, expectedPublishData).
					Return(&nats.Msg{
						Data: []byte("this cannot be decoded into a publication!"),
					}, nil).
					Times(1)

			},
			publishOptionsGenerator: func(t *testing.T) ([]PubSubOptPublish, func()) {

				respCh := make(chan *Publication, 1)
				callback := func() {
					select {
					case response := <-respCh:
						t.Errorf("did not expect to receive a response in channel, but received - %+v", response)
					default:
					}
				}

				return []PubSubOptPublish{
					NATSOptRespondToChannel(context.Background(), respCh),
				}, callback
			},
			expectedErrType: errors.New(""),
			natsOptions:     []NATSOption{},
		},
		{
			description: "should return an error if no NATS client had been connected",
			publication: NewPublication("test topic"),
			setup: func(t *testing.T, mockClient *mocks.MockNATSClient, pub *Publication) {
				mockClient.
					EXPECT().
					Publish(pub.Topic, gomock.Any()).
					// should never be called in this case!
					Times(0)
			},
			expectedErrType: errors.New(""),
			natsOptions: []NATSOption{
				// notice how we pass a nil client explicitly to simulate the failure scenario
				// desired by this test
				natsOptClient(nil),
			},
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

				// note: passing in the NATSOptPublishRequireAck option should set the publication response mode
				// to ACK before encoding the publication
				pub.ResponseMode = ResponseModeACK
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
			publishOptionsGenerator: func(t *testing.T) ([]PubSubOptPublish, func()) {
				return []PubSubOptPublish{
					NATSOptPublishRequireAck(context.Background()),
				}, func() {}
			},
		},
		{
			description: "should return an error if response is not a valid ACK response when using the NATSOptPublishRequireAck option",
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

				// note: passing in the NATSOptPublishRequireAck option should set the publication response mode
				// to ACK before encoding the publication
				pub.ResponseMode = ResponseModeACK
				expectedData, err := elemental.Encode(elemental.EncodingTypeMSGPACK, pub)
				if err != nil {
					t.Fatalf("test setup failed - could not encode publication - error: %+v", err)
					return
				}

				// notice how the response is not a valid ACK response!
				mockClient.
					EXPECT().
					RequestWithContext(gomock.Any(), pub.Topic, expectedData).
					Return(&nats.Msg{
						Data: []byte("this is not a valid ACK response!"),
					}, nil).
					Times(1)
			},
			expectedErrType: errors.New(""),
			natsOptions:     []NATSOption{},
			publishOptionsGenerator: func(t *testing.T) ([]PubSubOptPublish, func()) {
				return []PubSubOptPublish{
					NATSOptPublishRequireAck(context.Background()),
				}, func() {}
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

			var pubOpts []PubSubOptPublish
			if test.publishOptionsGenerator != nil {
				var callback func()
				pubOpts, callback = test.publishOptionsGenerator(t)
				defer callback()
			}

			pubErr := ps.Publish(test.publication, pubOpts...)
			if actualErrType := reflect.TypeOf(pubErr); actualErrType != reflect.TypeOf(test.expectedErrType) {
				t.Errorf("Call to publish returned error of type \"%+v\", when an error of type \"%+v\" was expected", actualErrType, reflect.TypeOf(test.expectedErrType))
				t.Logf("Received error: %v - Expected error: %v", pubErr, test.expectedErrType)
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

	var tests = []struct {
		description          string
		expectedPublication  *Publication
		expectedError        error
		setup                func(t *testing.T, pub *Publication)
		subscribeOptions     []PubSubOptSubscribe
		natsOptionsGenerator func() ([]NATSOption, func())
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

				pub.ResponseMode = ResponseModeACK
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
			description:      "should receive an error in errors channel if subscribing fails for any reason",
			expectedError:    nats.ErrConnectionClosed,
			subscribeOptions: []PubSubOptSubscribe{},
			setup:            func(t *testing.T, pub *Publication) {},
			natsOptionsGenerator: func() ([]NATSOption, func()) {

				// we use a mock client in this test as we want to simulate a failure when `Subscribe` is called
				ctrl := gomock.NewController(t)
				callback := func() {
					ctrl.Finish()
				}

				mockClient := mocks.NewMockNATSClient(ctrl)
				mockClient.
					EXPECT().
					Subscribe(subscribeTopic, gomock.Any()).
					Return(nil, nats.ErrInvalidConnection).
					Times(1)

				// we haven't configured a queueGroup so QueueSubscribe should never be called!
				mockClient.
					EXPECT().
					QueueSubscribe(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(0)

				return []NATSOption{
					natsOptClient(mockClient),
				}, callback
			},
			expectedPublication: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {

			var natOpts []NATSOption
			if test.natsOptionsGenerator != nil {
				var cleanup func()
				natOpts, cleanup = test.natsOptionsGenerator()
				defer cleanup()
			}
			// note: we prepend the NATSOption client option to allow test cases to override the actual client being used
			// (e.g. if they want to provide a mock client instead)
			natOpts = append([]NATSOption{natsOptClient(nc)}, natOpts...)
			ps := NewNATSPubSubClient(
				fmt.Sprintf("%s:%d", serverAddr.IP, serverAddr.Port),
				natOpts...,
			)
			publications := make(chan *Publication)
			// Why the buffered channel for errors?
			// Because otherwise if the call to Subscribe or QueueSubscribe fails then an error will be written to the error channel,
			// which would block indefinitely as there are no active readers on that channel yet.
			errors := make(chan error, 1)
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

				if test.expectedError == nil {
					t.Fatalf("received an unexpected error - err: \"%+v\"", err)
				}

				if actualErrType := reflect.TypeOf(test.expectedError); actualErrType != reflect.TypeOf(err) {
					t.Errorf("received error of type \"%+v\", when an error of type \"%+v\" was expected", actualErrType, reflect.TypeOf(test.expectedError))
				}

			case <-time.After(threshold):

				if test.expectedPublication != nil {
					t.Errorf("timed out expecting to receive a publication: %+v", test.expectedPublication)
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
