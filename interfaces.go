// Author: Antoine Mercadal
// See LICENSE file for full LICENSE
// Copyright 2016 Aporeto.

package bahamut

import "github.com/aporeto-inc/elemental"

// Processor is the interface for a Processor Unit
type Processor interface{}

// RetrieveManyProcessor is an interface.
type RetrieveManyProcessor interface {
	ProcessRetrieveMany(*Context) error
}

// RetrieveProcessor is an interface.
type RetrieveProcessor interface {
	ProcessRetrieve(*Context) error
}

// CreateProcessor is an interface.
type CreateProcessor interface {
	ProcessCreate(*Context) error
}

// UpdateProcessor is an interface.
type UpdateProcessor interface {
	ProcessUpdate(*Context) error
}

// DeleteProcessor is an interface.
type DeleteProcessor interface {
	ProcessDelete(*Context) error
}

// PatchProcessor is an interface.
type PatchProcessor interface {
	ProcessPatch(*Context) error
}

// InfoProcessor is an interface.
type InfoProcessor interface {
	ProcessInfo(*Context) error
}

// Authenticator is the interface used to verify the authentication
type Authenticator interface {
	IsAuthenticated(*Context) (bool, error)
}

// Authorizer is the interface used to verify the permission
type Authorizer interface {
	IsAuthorized(*Context) (bool, error)
}

// PushSessionsHandler is the interface used to handle sessions lyfecycle
type PushSessionsHandler interface {
	OnPushSessionStart(*PushSession)
	OnPushSessionStop(*PushSession)
	ShouldPush(*PushSession, *elemental.Event) bool
}
