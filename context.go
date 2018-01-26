// Author: Antoine Mercadal
// See LICENSE file for full LICENSE
// Copyright 2016 Aporeto.

package bahamut

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/aporeto-inc/elemental"

	uuid "github.com/satori/go.uuid"
)

// A Context contains all information about a current operation.
//
// It contains various Info like the Headers, the current parent identity and ID
// (if any) for a given ReST call, the children identity, and other things like that.
// It also contains information about Pagination, as well as elemental.Idenfiable (or list Idenfiables)
// the user sent through the request.
type Context struct {

	// Info contains various request related information.
	Request *elemental.Request

	// CountTotal contains various information about the counting of objects.
	CountTotal int

	// InputData contains the data sent by the client. It can be either a single *elemental.Identifiable
	// or a []*elemental.Identifiable.
	InputData interface{}

	// OutputData contains the information that you want to send back to the user. You will
	// mostly need to set this in your processors.
	OutputData interface{}

	// StatusCode contains the HTTP status code to return.
	// Bahamut will try to guess it, but you can set it yourself.
	StatusCode int

	// Redirect will be used to redirect a request if set.
	Redirect string

	// Metadata is contains random user defined metadata.
	Metadata map[string]interface{}

	ctx                context.Context
	customMessages     []string
	customMessagesLock *sync.Mutex
	events             elemental.Events
	id                 string
	claims             []string
	claimsMap          map[string]string
}

// NewContext creates a new *Context.
func NewContext() *Context {

	return &Context{
		Request:            elemental.NewRequest(),
		Metadata:           map[string]interface{}{},
		claims:             []string{},
		claimsMap:          map[string]string{},
		id:                 uuid.Must(uuid.NewV4()).String(),
		events:             elemental.Events{},
		customMessagesLock: &sync.Mutex{},
		ctx:                context.Background(),
	}
}

// NewContextWithRequest creates a new *Context with the given elemental.Request.
func NewContextWithRequest(req *elemental.Request) *Context {

	ctx := NewContext()
	ctx.Request = req

	return ctx
}

// SetClaims implements elemental.ClaimsHolder
func (c *Context) SetClaims(claims []string) {

	if claims == nil {
		return
	}

	c.claims = claims
	c.claimsMap = claimsToMap(claims)
}

// GetClaims implements elemental.ClaimsHolder
func (c *Context) GetClaims() []string {

	return c.claims
}

// GetClaimsMap returns a list of claims as map.
func (c *Context) GetClaimsMap() map[string]string {

	return c.claimsMap
}

// Identifier returns the unique identifier of the context.
func (c *Context) Identifier() string {

	return c.id
}

// EnqueueEvents enqueues the given event to the Context.
//
// Bahamut will automatically generate events on the currently processed object.
// But if your processor creates other objects alongside with the main one and you want to
// send a push to the user, then you can use this method.
//
// The events you enqueue using EnqueueEvents will be sent in order to the enqueueing, and
// *before* the main object related event.
func (c *Context) EnqueueEvents(events ...*elemental.Event) {

	c.events = append(c.events, events...)
}

// SetEvents set the full list of Errors in the Context.
func (c *Context) SetEvents(events elemental.Events) {

	c.events = events
}

// HasEvents returns true if the context has some custom events.
func (c *Context) HasEvents() bool {

	return len(c.events) > 0
}

// Events returns the current Events.
func (c *Context) Events() elemental.Events {

	return c.events
}

// AddMessage adds a new message that will be sent in the response.
func (c *Context) AddMessage(msg string) {
	c.customMessagesLock.Lock()
	c.customMessages = append(c.customMessages, msg)
	c.customMessagesLock.Unlock()
}

func (c *Context) messages() []string {
	c.customMessagesLock.Lock()
	defer c.customMessagesLock.Unlock()

	return c.customMessages
}

// Duplicate duplicates the context.
func (c *Context) Duplicate() *Context {

	ctx := NewContext()

	ctx.CountTotal = c.CountTotal
	ctx.StatusCode = c.StatusCode
	ctx.InputData = c.InputData
	ctx.OutputData = c.OutputData
	ctx.Request = c.Request.Duplicate()
	ctx.claims = append(ctx.claims, c.claims...)
	ctx.customMessages = append(ctx.customMessages, c.customMessages...)

	for k, v := range c.claimsMap {
		ctx.claimsMap[k] = v
	}

	for k, v := range c.Metadata {
		ctx.Metadata[k] = v
	}

	return ctx
}

func (c *Context) String() string {

	return fmt.Sprintf("<context id:%s request:%s totalcount:%d>",
		c.Identifier(),
		c.Request,
		c.CountTotal,
	)
}

// Done implements the context.Context interface.
func (c *Context) Done() <-chan struct{} { return c.ctx.Done() }

// Err implements the context.Context interface.
func (c *Context) Err() error { return c.ctx.Err() }

// Deadline implements the context.Context interface.
func (c *Context) Deadline() (time.Time, bool) { return c.ctx.Deadline() }

// Value implements the context.Context interface.
func (c *Context) Value(key interface{}) interface{} { return c.ctx.Value(key) }
