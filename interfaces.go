// Author: Antoine Mercadal
// See LICENSE file for full LICENSE
// Copyright 2016 Aporeto.

package bahamut

import "github.com/aporeto-inc/elemental"

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

// Authenticator is the interface that must be implemented in order to
// to be used as the Bahamut main Authenticator.
type Authenticator interface {
	IsAuthenticated(*Context) (bool, error)
}

// Authorizer is the interface that must be implemented in order to
// to be used as the Bahamut main Authorizer.
type Authorizer interface {
	IsAuthorized(*Context) (bool, error)
}

// PushSessionsHandler is the interface that must be implemented in order to
// to be used as the Bahamut Push Server handler.
type PushSessionsHandler interface {
	OnPushSessionStart(*PushSession)
	OnPushSessionStop(*PushSession)
	ShouldPush(*PushSession, *elemental.Event) bool
}
