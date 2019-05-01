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
	"time"
)

// PubSubOptPublish is the type of option that can use in PubSubClient.Publish.
type PubSubOptPublish func(interface{})

// PubSubOptSubscribe is the type of option that can use in PubSubClient.Subscribe.
type PubSubOptSubscribe func(interface{})

// A PubSubClient is a structure that provides a publish/subscribe mechanism.
type PubSubClient interface {
	Publish(publication *Publication, opts ...PubSubOptPublish) error
	Subscribe(pubs chan *Publication, errors chan error, topic string, opts ...PubSubOptSubscribe) func()
	Connect() Waiter
	Disconnect() error
}

// A Waiter is the interface returned by Server.Connect
// that you can use to wait for the connection.
type Waiter interface {
	Wait(time.Duration) bool
}

// A connectionWaiter is the Waiter for the PubSub Server connection
type connectionWaiter struct {
	ok    chan bool
	abort chan struct{}
}

// Wait waits at most for the given timeout for the connection.
func (w connectionWaiter) Wait(timeout time.Duration) bool {

	select {
	case status := <-w.ok:
		return status
	case <-time.After(timeout):
		close(w.abort)
		return false
	}
}
