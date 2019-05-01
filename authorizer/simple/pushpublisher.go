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

package simple

import (
	"go.aporeto.io/elemental"
)

// CustomShouldPublishFunc is the type of function that can be used
// to decide is an event should be published.
type CustomShouldPublishFunc func(*elemental.Event) (bool, error)

// A PublishHandler handles publish decisions.
type PublishHandler struct {
	shouldPublishFunc CustomShouldPublishFunc
}

// NewPublishHandler returns a new PushSessionsHandler. If shouldPublishFunc is nil
// the publisher will dispatch all events.
func NewPublishHandler(shouldPublishFunc CustomShouldPublishFunc) *PublishHandler {

	return &PublishHandler{
		shouldPublishFunc: shouldPublishFunc,
	}
}

// ShouldPublish is part of the bahamut.PushPublishHandler interface
func (g *PublishHandler) ShouldPublish(event *elemental.Event) (bool, error) {

	if g.shouldPublishFunc == nil {
		return true, nil
	}

	return g.shouldPublishFunc(event)
}
