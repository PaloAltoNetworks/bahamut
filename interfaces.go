// Author: Antoine Mercadal
// See LICENSE file for full LICENSE
// Copyright 2016 Aporeto.

package bahamut

import (
	"net/http"

	"github.com/aporeto-inc/elemental"
)

type processorFinder func(identity elemental.Identity) (Processor, error)

type eventPusher func(...*elemental.Event)

// Server is the interface of a bahamut server.
type Server interface {

	// RegisterProcessor registers a new Processor for a particular Identity.
	RegisterProcessor(Processor, elemental.Identity) error

	// UnregisterProcessor unregisters a registered Processor for a particular identity.
	UnregisterProcessor(elemental.Identity) error

	// ProcessorForIdentity returns the registered Processor for a particular identity.
	ProcessorForIdentity(elemental.Identity) (Processor, error)

	// ProcessorsCount returns the number of registered processors.
	ProcessorsCount() int

	// Push pushes the given events to all active sessions.
	// It will use the PubSubServer configured in the pushConfig.
	Push(...*elemental.Event)

	// Start starts the Bahamut server.
	Start()

	// Stop stops the Bahamut server.
	Stop()
}

// Processor is the interface for a Processor Unit
type Processor interface{}

// RetrieveManyProcessor is the interface a processor must implement
// in order to be able to manage OperationRetrieveMany.
type RetrieveManyProcessor interface {
	ProcessRetrieveMany(*Context) error
}

// RetrieveProcessor is the interface a processor must implement
// in order to be able to manage OperationRetrieve.
type RetrieveProcessor interface {
	ProcessRetrieve(*Context) error
}

// CreateProcessor is the interface a processor must implement
// in order to be able to manage OperationCreate.
type CreateProcessor interface {
	ProcessCreate(*Context) error
}

// UpdateProcessor is the interface a processor must implement
// in order to be able to manage OperationUpdate.
type UpdateProcessor interface {
	ProcessUpdate(*Context) error
}

// DeleteProcessor is the interface a processor must implement
// in order to be able to manage OperationDelete.
type DeleteProcessor interface {
	ProcessDelete(*Context) error
}

// PatchProcessor is the interface a processor must implement
// in order to be able to manage OperationPatch.
type PatchProcessor interface {
	ProcessPatch(*Context) error
}

// InfoProcessor is the interface a processor must implement
// in order to be able to manage OperationInfo.
type InfoProcessor interface {
	ProcessInfo(*Context) error
}

// RequestAuthenticator is the interface that must be implemented in order to
// to be used as the Bahamut main Authenticator.
type RequestAuthenticator interface {
	AuthenticateRequest(*elemental.Request, elemental.ClaimsHolder) (bool, error)
}

// SessionAuthenticator is the interface that must be implemented in order to
// be used as the initial Web socket session Authenticator.
type SessionAuthenticator interface {
	AuthenticateSession(elemental.SessionHolder, elemental.SpanHolder) (bool, error)
}

// Authorizer is the interface that must be implemented in order to
// to be used as the Bahamut main Authorizer.
type Authorizer interface {
	IsAuthorized(*Context) (bool, error)
}

// PushSessionsHandler is the interface that must be implemented in order to
// to be used as the Bahamut Push Server handler.
type PushSessionsHandler interface {
	OnPushSessionStart(*Session)
	OnPushSessionStop(*Session)
	ShouldPush(*Session, *elemental.Event) (bool, error)
}

// Auditer is the interface an object must implement in order to handle
// audit traces.
type Auditer interface {
	Audit(*Context, error)
}

// A RateLimiter is the interface an object must implement in order to
// limit the rate of the incoming requests.
type RateLimiter interface {
	RateLimit(*http.Request) (bool, error)
}
