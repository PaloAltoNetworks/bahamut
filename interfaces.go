// Author: Antoine Mercadal
// See LICENSE file for full LICENSE
// Copyright 2016 Aporeto.

package bahamut

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"

	"go.aporeto.io/elemental"
)

type processorFinderFunc func(identity elemental.Identity) (Processor, error)

type eventPusherFunc func(...*elemental.Event)

// AuthAction is the type of action an Authenticator or an Authorizer can return.
type AuthAction int

const (

	// AuthActionOK means the authenticator/authorizer takes the responsibility
	// to grant the request. The execution in the chain will
	// stop and will be considered as a success.
	AuthActionOK AuthAction = iota

	// AuthActionKO means the authenticator/authorizer takes the responsibility
	// to reject the request. The execution in the chain will
	// stop and will be considered as a success.
	AuthActionKO

	// AuthActionContinue means the authenticator/authorizer does not take
	// any responsabolity and let the chain continue.
	// If the last authenticator in the chain returns AuthActionContinue,
	// Then the request will be considered as a success.
	AuthActionContinue
)

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
	// It will use the PubSubClient configured in the pushConfig.
	Push(...*elemental.Event)

	// Run runs the server using the given context.Context.
	// You can stop the server by canceling the context.
	Run(context.Context)

	// Start starts the Bahamut server.
	// This will install signal handler to handle
	// graceful interruption.
	//
	// Deprecated: Start is deprecation. use Run.
	Start()
}

// A Context contains all information about a current operation.
//
// It contains various Info like the Headers, the current parent identity and ID
// (if any) for a given ReST call, the children identity, and other things like that.
// It also contains information about Pagination, as well as elemental.Idenfiable (or list Idenfiables)
// the user sent through the request.
type Context interface {

	// Identifier returns the internal unique identifier of the context.
	Identifier() string

	// Context returns the underlying context.Context.
	Context() context.Context

	// Request returns the underlying elemental.Request.
	Request() *elemental.Request

	// InputData returns the data sent by the client
	InputData() interface{}

	// OutputData returns the current output data.
	OutputData() interface{}

	// SetOutputData sets the data that will be returned to the client.
	SetOutputData(interface{})

	// Set count sets the count.
	SetCount(int)

	// Count returns the current count.
	Count() int

	// SetRedirect sets the redirect URL.
	SetRedirect(string)

	// Redirect returns the current value for redirection
	Redirect() string

	// SetStatusCode sets the status code that will be returned to the client.
	SetStatusCode(int)

	// StatusCode returns the current status code.
	StatusCode() int

	// AddMessage adds a custom message that will be sent as repponse header.
	AddMessage(string)

	// SetClaims sets the claims.
	SetClaims(claims []string)

	// Claims returns the list of claims.
	Claims() []string

	// Claims returns claims in a map.
	ClaimsMap() map[string]string

	// Duplicate creates a copy of the Context.
	Duplicate() Context

	// WithInputData creates a copy of the context using the given input data.
	WithInputData(data interface{}) Context

	// EnqueueEvents enqueues the given event to the Context.
	//
	// Bahamut will automatically generate events on the currently processed object.
	// But if your processor creates other objects alongside with the main one and you want to
	// send a push to the user, then you can use this method.
	//
	// The events you enqueue using EnqueueEvents will be sent in order to the enqueueing, and
	// *before* the main object related event.
	EnqueueEvents(...*elemental.Event)

	// SetMetadata sets opaque metadata that can be reteieved by Metadata().
	SetMetadata(key, value interface{})

	// Metadata returns the opaque data set by using SetMetadata().
	Metadata(key interface{}) interface{}

	fmt.Stringer
}

// Processor is the interface for a Processor Unit
type Processor interface{}

// RetrieveManyProcessor is the interface a processor must implement
// in order to be able to manage OperationRetrieveMany.
type RetrieveManyProcessor interface {
	ProcessRetrieveMany(Context) error
}

// RetrieveProcessor is the interface a processor must implement
// in order to be able to manage OperationRetrieve.
type RetrieveProcessor interface {
	ProcessRetrieve(Context) error
}

// CreateProcessor is the interface a processor must implement
// in order to be able to manage OperationCreate.
type CreateProcessor interface {
	ProcessCreate(Context) error
}

// UpdateProcessor is the interface a processor must implement
// in order to be able to manage OperationUpdate.
type UpdateProcessor interface {
	ProcessUpdate(Context) error
}

// DeleteProcessor is the interface a processor must implement
// in order to be able to manage OperationDelete.
type DeleteProcessor interface {
	ProcessDelete(Context) error
}

// PatchProcessor is the interface a processor must implement
// in order to be able to manage OperationPatch.
type PatchProcessor interface {
	ProcessPatch(Context) error
}

// InfoProcessor is the interface a processor must implement
// in order to be able to manage OperationInfo.
type InfoProcessor interface {
	ProcessInfo(Context) error
}

// RequestAuthenticator is the interface that must be implemented in order to
// to be used as the Bahamut Authenticator.
type RequestAuthenticator interface {
	AuthenticateRequest(Context) (AuthAction, error)
}

// SessionAuthenticator is the interface that must be implemented in order to
// be used as the initial Web socket session Authenticator.
type SessionAuthenticator interface {
	AuthenticateSession(Session) (AuthAction, error)
}

// Authorizer is the interface that must be implemented in order to
// to be used as the Bahamut Authorizer.
type Authorizer interface {
	IsAuthorized(Context) (AuthAction, error)
}

// PushDispatchHandler is the interface that must be implemented in order to
// to be used as the Bahamut Push Dispatch handler.
type PushDispatchHandler interface {
	OnPushSessionInit(PushSession) (bool, error)
	OnPushSessionStart(PushSession)
	OnPushSessionStop(PushSession)
	ShouldDispatch(PushSession, *elemental.Event) (bool, error)
}

// PushPublishHandler is the interface that must be implemented in order to
// to be used as the Bahamut Push Publish handler.
type PushPublishHandler interface {
	ShouldPublish(*elemental.Event) (bool, error)
}

// Auditer is the interface an object must implement in order to handle
// audit traces.
type Auditer interface {
	Audit(Context, error)
}

// A RateLimiter is the interface an object must implement in order to
// limit the rate of the incoming requests.
type RateLimiter interface {
	RateLimit(*http.Request) (bool, error)
}

// Session is the interface of a generic websocket session.
type Session interface {
	Identifier() string
	Parameter(string) string
	SetClaims([]string)
	Claims() []string
	ClaimsMap() map[string]string
	Token() string
	TLSConnectionState() *tls.ConnectionState
	Metadata() interface{}
	SetMetadata(interface{})
	Context() context.Context
}

// PushSession is a Push Session
type PushSession interface {
	Session

	DirectPush(...*elemental.Event)
}
