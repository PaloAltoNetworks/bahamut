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
	"testing"
	"time"

	nats "github.com/nats-io/nats.go"
	. "github.com/smartystreets/goconvey/convey"
)

func TestBahamut_NATSOption(t *testing.T) {

	n := &natsPubSub{}

	Convey("Calling NATSOptCredentials should work", t, func() {
		NATSOptCredentials("user", "pass")(n)
		So(n.username, ShouldEqual, "user")
		So(n.password, ShouldEqual, "pass")
	})

	Convey("Calling NATSOptClusterID should work", t, func() {
		NATSOptClusterID("cid")(n)
		So(n.clusterID, ShouldEqual, "cid")
	})

	Convey("Calling NATSOptClientID should work", t, func() {
		NATSOptClientID("cid")(n)
		So(n.clientID, ShouldEqual, "cid")
	})

	Convey("Calling NATSOptTLS should work", t, func() {
		tlscfg := &tls.Config{}
		NATSOptTLS(tlscfg)(n)
		So(n.tlsConfig, ShouldEqual, tlscfg)
	})

	Convey("Calling NATSErrorHandler should work", t, func() {
		f := func(*nats.Conn, *nats.Subscription, error) {}
		NATSErrorHandler(f)(n)
		So(n.errorHandleFunc, ShouldEqual, f)
	})
}

func TestBahamut_PubSubNatsOptionsSubscribe(t *testing.T) {

	c := natsSubscribeConfig{}

	Convey("Calling NATSOptSubscribeQueue should work", t, func() {
		NATSOptSubscribeQueue("queueGroup")(&c)
		So(c.queueGroup, ShouldEqual, "queueGroup")
	})

	Convey("Calling NATSOptSubscribeReplyTimeout should set the timeout", t, func() {
		duration := 15 * time.Second
		NATSOptSubscribeReplyTimeout(duration)(&c)
		So(c.replyTimeout, ShouldEqual, duration)
	})

}

func TestBahamut_PubSubNatsOptionsPublish(t *testing.T) {

	Convey("Setup", t, func() {

		c := natsPublishConfig{}

		Convey("Calling NATSOptPublishRequireAck should work", func() {
			NATSOptPublishRequireAck(context.TODO())(&c)
			So(c.ctx, ShouldResemble, context.TODO())
			So(c.desiredResponse, ShouldEqual, ResponseModeACK)
		})

		Convey("Calling NATSOptPublishRequireAck should panic if requestMode has already been set", func() {
			c.desiredResponse = ResponseModePublication
			So(func() {
				NATSOptPublishRequireAck(context.TODO())(&c)
			}, ShouldPanic)
		})

		Convey("Calling NATSOptPublishRequireAck should panic if supplied context is nil", func() {
			So(func() {
				// nolint - note: ignoring linter feedback as we are trying to cause a panic intentionally by passing in a `nil` context
				NATSOptPublishRequireAck(nil)(&c)
			}, ShouldPanic)
		})

		Convey("Calling NATSOptRespondToChannel should work", func() {
			respCh := make(chan *Publication)
			NATSOptRespondToChannel(context.TODO(), respCh)(&c)
			So(c.ctx, ShouldResemble, context.TODO())
			So(c.responseCh, ShouldEqual, respCh)
			So(c.desiredResponse, ShouldEqual, ResponseModePublication)
		})

		Convey("Calling NATSOptRespondToChannel should panic if requestMode has already been set", func() {
			c.desiredResponse = ResponseModeACK
			So(func() {
				NATSOptRespondToChannel(context.TODO(), make(chan *Publication))(&c)
			}, ShouldPanic)
		})

		Convey("Calling NATSOptRespondToChannel should panic if supplied response channel is nil", func() {
			So(func() {
				NATSOptRespondToChannel(context.TODO(), nil)(&c)
			}, ShouldPanic)
		})

		Convey("Calling NATSOptRespondToChannel should panic if supplied context is nil", func() {
			c.desiredResponse = ResponseModeACK
			So(func() {
				// nolint - note: ignoring linter feedback as we are trying to cause a panic intentionally by passing in a `nil` context
				NATSOptRespondToChannel(nil, make(chan *Publication))(&c)
			}, ShouldPanic)
		})
	})
}
