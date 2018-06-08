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
