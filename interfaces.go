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

	// RoutesInfo returns the routing information of the server.
	RoutesInfo() map[int][]RouteInfo

	// VersionsInfo returns additional versioning info.
	VersionsInfo() map[string]interface{}

	// PushEndpoint returns the configured push endpoints.
	// If the push server is not active, it will return an
	// empty string.
	PushEndpoint() string

	// Run runs the server using the given context.Context.
	// You can stop the server by canceling the context.
	Run(context.Context)
}

// A Context contains all information about a current operation.
type Context interface {

	// Identifier returns the internal unique identifier of the context.
	Identifier() string

	// Context returns the underlying context.Context.
	Context() context.Context

	// Request returns the underlying *elemental.Request.
	Request() *elemental.Request

	// InputData returns the data sent by the client
	InputData() interface{}

	// SetInputData replaces the current input data.
	SetInputData(interface{})

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

	// SetNext can be use to give the next pagination token.
	SetNext(string)

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

	// outputCookies adds cookies to the response that
	// will be returned to the client.
	AddOutputCookies(cookies ...*http.Cookie)
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

	// OnPushSessionInit is called when a new push session wants to connect.
	// If it returns false, the push session will be considered
	// as forbidden. If it returns an error, it will be returned
	// to the client.
	OnPushSessionInit(PushSession) (bool, error)

	// OnPushSessionStart is called when a new push session starts.
	OnPushSessionStart(PushSession)

	// OnPushSessionStart is called when a new push session terminated.
	OnPushSessionStop(PushSession)

	// ShouldDispatch is called to decide if the given event should be sent to the given session.
	// The last parameter will contain whatever has been returned by SummarizeEvent.
	// It is NOT safe to modify the given *elemental.Event. This would cause
	// race conditions. You can only safely read from it.
	ShouldDispatch(PushSession, *elemental.Event, interface{}) (bool, error)

	// RelatedEventIdentities allows to return a list of related identities
	// associated to the main event identity. This allows to pass filtering
	// in case a push on identity A must also trigger a push on identity B and C.
	RelatedEventIdentities(string) []string

	// SummarizeEvent is called once per event and allows the implementation
	// to return an interface that will be passed to ShouldDispatch.
	// If you need to decode an event to read some information to make a
	// dispatch decision, this is a good place as it will allow you to only
	// do this once.
	SummarizeEvent(event *elemental.Event) (interface{}, error)
}

// CloserNotifierError is an error type that can optionally be returned by the callback, ShouldDispatch(PushSession, *elemental.Event, interface{}),
// defined in the PushDispatchHandler interface. If the callback returns an error that satisfies the CloserNotifierError,
// bahamut will call the `ShouldCloseSocket()` to determine whether the socket connection should be closed. Useful in situations where
// it makes no sense keeping the socket alive anymore (e.g. client sent a push config that contains an unsupported filter comparator)
type CloserNotifierError interface {
	ShouldCloseSocket() (bool, int)
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
	Header(string) string
	PushConfig() *elemental.PushConfig
	SetClaims([]string)
	Claims() []string
	ClaimsMap() map[string]string
	Token() string
	TLSConnectionState() *tls.ConnectionState
	Metadata() interface{}
	SetMetadata(interface{})
	Context() context.Context
	ClientIP() string
	Cookie(string) (*http.Cookie, error)
}

// PushSession is a Push Session
type PushSession interface {
	Session

	DirectPush(...*elemental.Event)
}
