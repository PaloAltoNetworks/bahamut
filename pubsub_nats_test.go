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
	"crypto/tls"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestNats_NewPubSubServer(t *testing.T) {

	Convey("Given I create a new PubSubServer with no option", t, func() {

		ps := NewNATSPubSubClient("nats://localhost:4222").(*natsPubSub)

		Convey("Then the PubSubServer should be correctly initialized", func() {
			So(ps.natsURL, ShouldEqual, "nats://localhost:4222")
			So(ps.clusterID, ShouldEqual, "test-cluster")
			So(ps.clientID, ShouldNotBeEmpty)
			So(ps.username, ShouldEqual, "")
			So(ps.password, ShouldEqual, "")
			So(ps.tlsConfig, ShouldEqual, nil)
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
		).(*natsPubSub)

		Convey("Then the PubSubServer should be correctly initialized", func() {
			So(ps.natsURL, ShouldEqual, "nats://localhost:4222")
			So(ps.clusterID, ShouldEqual, "cid")
			So(ps.clientID, ShouldEqual, "id")
			So(ps.username, ShouldEqual, "username")
			So(ps.password, ShouldEqual, "password")
			So(ps.tlsConfig, ShouldEqual, tlsconfig)
		})
	})
}
