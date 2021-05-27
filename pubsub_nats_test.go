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
	natsserver "github.com/nats-io/nats-server/v2/test"
	nats "github.com/nats-io/nats.go"
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
					// should never be called in this case as passing in the NATSOptPublishRequireAck
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
					// should never be called in this case as passing in the NATSOptPublishRequireAck
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
		description string
		// the publication that will be sent to the publication channel
		expectedPublication *Publication
		// the error that will be sent to the errors channel
		expectedError error
		// this callback is called right after the call to Subscribe has been made. this is opportunity for you
		// to setup mocks and mimic client behaviour by making actual publications to the test topic
		setup func(t *testing.T, pub *Publication, client PubSubClient)
		// this callback is called in the event that a publication was sent to the configured publication channel.
		// you use this as an opportunity to reply back to the publication using the `Reply` method
		replier              func(t *testing.T, pub *Publication, client PubSubClient) bool
		subscribeOptions     []PubSubOptSubscribe
		natsOptionsGenerator func() ([]NATSOption, func())
	}{
		{
			description: "should successfully subscribe to topic and receive a publication in provided channel",
			setup: func(t *testing.T, pub *Publication, _ PubSubClient) {
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
			setup: func(t *testing.T, _ *Publication, _ PubSubClient) {
				if err := nc.Publish(subscribeTopic, []byte("this cannot be decoded into a publication")); err != nil {
					t.Fatalf("test setup failed - could not publish publication - error: %+v", err)
					return
				}

			},
			expectedPublication: nil,
			subscribeOptions:    nil,
		},
		{
			description: "should receive error in errors channel if responding back with an ACK fails",
			setup: func(t *testing.T, _ *Publication, client PubSubClient) {

				pc, ok := client.(*natsPubSub)
				if !ok {
					t.Fatalf("test setup failed - could not assert `PubSubClient` to `*natsPubSub`")
					return
				}

				ctrl := gomock.NewController(t)
				mockClient := mocks.NewMockNATSClient(ctrl)
				// hack for fault injection: we override the NATS client we are using to a mock client so we can cause
				// a failure when the message handler attempts to respond back to the request with an ACK.
				pc.client = mockClient
				mockClient.
					EXPECT().
					Publish(gomock.Any(), ackMessage).
					// notice how this is returning an error
					Return(errors.New("whoops, failed to publish ACK response")).
					Times(1)

				// act as a client - publish a message expecting to get back an ACK
				pub := NewPublication(subscribeTopic)
				pub.ResponseMode = ResponseModeACK
				data, err := elemental.Encode(elemental.EncodingTypeMSGPACK, pub)
				if err != nil {
					t.Fatalf("test setup failed - could not encode publication - error: %+v", err)
					return
				}

				// our request should fail now (w/ a timeout) as we have injected a deliberate fault (see above)
				// when the Subscriber attempts to respond back to our request with an ACK response
				response, err := nc.Request(subscribeTopic, data, threshold)
				if err == nil {
					t.Fatalf("test setup failed - expected to get an error here, but got a response back instead: \"%+v\". "+
						"Hint: test structure has likely been messed up.", *response)
					return
				}
			},
			// we expect to get an error because when Subscriber should fail to Publish back its ACK response to the client's request
			// it will send the error it gets back from its call to Publish to the configured error channel.
			expectedError:       errors.New(""),
			expectedPublication: nil,
			subscribeOptions:    nil,
		},
		{
			description: "should be able to respond back to a publication manually",
			setup: func(t *testing.T, pub *Publication, _ PubSubClient) {

				// act as the client making the request:
				//   - make a request expecting to get back an Publication as a response by setting the
				//     response mode to ResponseModePublication
				pub.ResponseMode = ResponseModePublication
				data, err := elemental.Encode(elemental.EncodingTypeMSGPACK, pub)
				if err != nil {
					t.Fatalf("test setup failed - could not encode publication - error: %+v", err)
					return
				}

				response, err := nc.Request(subscribeTopic, data, threshold)
				if err != nil {
					t.Fatalf("test setup failed - request was not successful - error: %+v", err)
					return
				}

				// the received response should be a publication
				responsePub := NewPublication("")
				if err := elemental.Decode(elemental.EncodingTypeMSGPACK, response.Data, responsePub); err != nil {
					t.Fatalf("test failed - the response to the request could not be decoded into a *Publication - error: %+v", err)
					return
				}

			},
			replier: func(t *testing.T, pub *Publication, _ PubSubClient) bool {

				response := &Publication{
					Data: []byte("some response"),
				}

				if err := pub.Reply(response); err != nil {
					t.Errorf("test failed - could not reply to publication - err: %+v", err)
				}

				return false
			},
			expectedPublication: &Publication{
				Topic: subscribeTopic,
				Data:  []byte("message"),
			},
		},
		{
			description: "should receive error in errors channel if publishing manual response fails for whatever reason",
			setup: func(t *testing.T, pub *Publication, _ PubSubClient) {
				// high level overview of this test:
				//   - make a request expecting to get back a publication
				//   - subscriber gets the publication
				//   - subscriber responds to the publication via the Reply call
				//   - the Reply call fails internally for some reason because the call the Publish call fails
				//   - this should result in an error being sent to the configured errors channel

				// act as a client - publish a message expecting to get back a Publication response
				pub.ResponseMode = ResponseModePublication
				data, err := elemental.Encode(elemental.EncodingTypeMSGPACK, pub)
				if err != nil {
					t.Fatalf("test setup failed - could not encode publication - error: %+v", err)
					return
				}

				// our request should fail now (w/ a timeout) as we have injected a deliberate fault (see replier below)
				// when the Subscriber attempts to respond back to our request with a publication response
				response, err := nc.Request(subscribeTopic, data, threshold)
				if err == nil {
					t.Fatalf("test setup failed - expected to get an error here, but got a response back instead: \"%+v\". "+
						"Hint: test structure has likely been messed up.", *response)
					return
				}
			},
			expectedError: errors.New(""),
			expectedPublication: &Publication{
				Topic: subscribeTopic,
				Data:  []byte("message"),
			}, replier: func(t *testing.T, pub *Publication, client PubSubClient) bool {

				pc, ok := client.(*natsPubSub)
				if !ok {
					t.Fatalf("test setup failed - could not assert `PubSubClient` to `*natsPubSub`")
					return false
				}

				replyMessage := &Publication{
					Data:         []byte("some custom reply"),
					ResponseMode: ResponseModeNone,
				}

				ctrl := gomock.NewController(t)
				mockClient := mocks.NewMockNATSClient(ctrl)
				// hack for fault injection: we override the NATS client we are using to a mock client so we can cause
				// a failure when the message handler attempts to respond back to the request with a publication.
				pc.client = mockClient
				mockClient.
					EXPECT().
					Publish(gomock.Any(), gomock.Any()).
					// notice how this is returning an error
					Return(errors.New("whoops, failed to publish publication response")).
					Times(1)

				// stronger assertion
				replyMessage.ResponseMode = ResponseModeACK
				if err := pub.Reply(replyMessage); err != nil {
					t.Errorf("test failed - could not reply to publication - err: %+v", err)
				}

				return true

			},
		},
		{
			description: "should receive error in errors channel if you take too long to respond to publication",
			setup: func(t *testing.T, pub *Publication, _ PubSubClient) {
				// high level overview of this test:
				//   - make a request expecting to get back a publication
				//   - subscriber gets the publication
				//   - subscriber responds to the publication via the Reply call
				//   - the Reply call will fail because the replyCh will be set to nil by the message
				//     handler as the deadline will be reached.
				//   - this should result in an error being sent to the configured errors channel

				// act as a client - publish a message expecting to get back a Publication response
				pub.ResponseMode = ResponseModePublication
				data, err := elemental.Encode(elemental.EncodingTypeMSGPACK, pub)
				if err != nil {
					t.Fatalf("test setup failed - could not encode publication - error: %+v", err)
					return
				}

				// our request should fail as the subscriber is not going to be able to respond in time due to an artificial
				// deadline (0 nanoseconds) - see the subscribe options for this test
				response, err := nc.Request(subscribeTopic, data, threshold)
				if err == nil {
					t.Fatalf("test setup failed - expected to get an error here, but got a response back instead: \"%+v\". "+
						"Hint: test structure has likely been messed up.", *response)
					return
				}
			},
			// we expect to get an error back, because we take longer than the configured timeout to respond to the publication
			expectedError: errors.New(""),
			subscribeOptions: []PubSubOptSubscribe{
				// we deliberately set a low timeout here!
				NATSOptSubscribeReplyTimeout(100 * time.Millisecond),
			},
			expectedPublication: &Publication{
				Topic: subscribeTopic,
				Data:  []byte("message"),
			}, replier: func(t *testing.T, pub *Publication, client PubSubClient) bool {

				response := &Publication{
					Data: []byte("some response"),
				}

				// simulate some work, just long enough that we will reach our configured timeout
				<-time.After(1 * time.Second)

				// this should fail now because we took too long to call Reply
				if err := pub.Reply(response); err == nil {
					t.Error("test failed - expected to receive an error, but got nothing")
				}

				// we are expecting to get an error as a result of failing to respond back to the
				// publication in time, so we get the test suite to read from all the channels again
				// by returning true. This is so we can assert that an error was sent to the errors
				// channel.
				return true
			},
		},
		{
			description: "should respond back with an ACK message to all publications that expect an ACK response",
			setup: func(t *testing.T, pub *Publication, _ PubSubClient) {

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
			setup:            func(t *testing.T, pub *Publication, client PubSubClient) {},
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
			// note: we prepend the NATSOption natsOptClient option to allow test cases to override the actual client being used
			// (e.g. if they want to provide a mock client instead)
			natOpts = append([]NATSOption{natsOptClient(nc)}, natOpts...)
			ps := NewNATSPubSubClient(
				fmt.Sprintf("%s:%d", serverAddr.IP, serverAddr.Port),
				natOpts...,
			)
			done := make(chan struct{})
			publications := make(chan *Publication)
			errors := make(chan error)
			go func() {

			Loop:
				for {
					select {
					case pub := <-publications:

						if test.expectedPublication == nil {
							t.Errorf("did not expect to receive any publications, but received publication - \"%+v\"", pub)
							return
						}

						if pub.Topic != test.expectedPublication.Topic {
							t.Errorf("expected publication with topic \"%s\", but received publication with topic \"%s\"", test.expectedPublication.Topic, pub.Topic)
						}

						if !bytes.Equal(pub.Data, test.expectedPublication.Data) {
							t.Errorf("expected publication with data: \"%+v\", but received publication with data: \"%+v\"", pub.Data, test.expectedPublication.Data)
						}

						if test.replier == nil {
							break Loop
						}

						// if there is a replier func, we don't want to break out of the select loop just yet because maybe the call to Reply fails
						// in which case the test scenario may want to validate that an error is sent to the errors channel
						if readChannelsAgain := test.replier(t, pub, ps); !readChannelsAgain {
							break Loop
						}

					case err := <-errors:

						if test.expectedError == nil {
							t.Errorf("received an unexpected error - err: \"%+v\"", err)
							return
						}

						if actualErrType := reflect.TypeOf(test.expectedError); actualErrType != reflect.TypeOf(err) {
							t.Errorf("received error of type \"%+v\", when an error of type \"%+v\" was expected", actualErrType, reflect.TypeOf(test.expectedError))
						}

						break Loop

					case <-time.After(threshold):

						if test.expectedPublication != nil {
							t.Errorf("timed out expecting to receive a publication: \"%+v\"", test.expectedPublication)
						}

						if test.expectedError != nil {
							t.Errorf("timed out expecting to receive an error: \"%+v\"", test.expectedError)
						}

						break Loop
					}
				}

				close(done)
			}()

			unsub := ps.Subscribe(publications, errors, subscribeTopic, test.subscribeOptions...)
			defer unsub()
			test.setup(t, test.expectedPublication, ps)

			// wait for assertions to finish
			<-done
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
