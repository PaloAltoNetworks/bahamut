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
)

// PubSubOptPublish is the type of option that can use in PubSubClient.Publish.
type PubSubOptPublish func(interface{})

// PubSubOptSubscribe is the type of option that can use in PubSubClient.Subscribe.
type PubSubOptSubscribe func(interface{})

// A PubSubClient is a structure that provides a publish/subscribe mechanism.
type PubSubClient interface {
	Publish(publication *Publication, opts ...PubSubOptPublish) error
	Subscribe(pubs chan *Publication, errors chan error, topic string, opts ...PubSubOptSubscribe) func()
	Connect(ctx context.Context) error
	Disconnect() error
}
