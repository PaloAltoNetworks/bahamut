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

	nats "github.com/nats-io/go-nats"
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

	Convey("Calling NATSOptConnectRetryInterval should set the connection retry interval", t, func() {
		desiredRetryInternal := 1 * time.Second
		NATSOptConnectRetryInterval(desiredRetryInternal)(n)
		So(n.retryInterval, ShouldEqual, desiredRetryInternal)
	})
}

func TestBahamut_PubSubNatsOptionsSubscribe(t *testing.T) {

	c := natsSubscribeConfig{}

	Convey("Calling NATSOptSubscribeQueue should work", t, func() {
		NATSOptSubscribeQueue("queueGroup")(&c)
		So(c.queueGroup, ShouldEqual, "queueGroup")
	})

	Convey("Calling NATSOptSubscribeReplyer should work", t, func() {
		r := func(msg *nats.Msg) []byte { return nil }
		NATSOptSubscribeReplyer(r)(&c)
		So(c.replier, ShouldEqual, r)
	})
}

func TestBahamut_PubSubNatsOptionsPublish(t *testing.T) {

	c := natsPublishConfig{}

	Convey("Calling NATSOptPublishRequireAck should work", t, func() {
		NATSOptPublishRequireAck(context.TODO())(&c)
		So(c.ctx, ShouldEqual, context.TODO())
		So(c.requestMode, ShouldEqual, requestModeACK)
	})
}
