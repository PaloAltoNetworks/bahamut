// Author: Antoine Mercadal
// See LICENSE file for full LICENSE
// Copyright 2016 Aporeto.

package bahamut

import (
	"context"
	"fmt"
	"sync"

	"go.aporeto.io/elemental"

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
	eventsLock         *sync.Mutex
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
		id:                 uuid.NewV4().String(),
		events:             elemental.Events{},
		customMessagesLock: &sync.Mutex{},
		eventsLock:         &sync.Mutex{},
	}
}

// NewContextWithRequest creates a new *Context with the given elemental.Request.
func NewContextWithRequest(req *elemental.Request) *Context {

	ctx := NewContext()
	ctx.Request = req

	return ctx
}

// WithContext returns a shallow copy of the manipulate.Context
// with it's internal context changed to the given context.Context.
func (c *Context) WithContext(ctx context.Context) *Context {

	if ctx == nil {
		panic("nil context")
	}

	c2 := &Context{}
	*c2 = *c
	c2.ctx = ctx

	return c2
}

// Context returns the internal context.Context of the
// manipulate.Context.
func (c *Context) Context() context.Context {

	if c.ctx != nil {
		return c.ctx
	}

	return context.Background()
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

	c.eventsLock.Lock()
	defer c.eventsLock.Unlock()

	c.events = append(c.events, events...)
}

// SetEvents set the full list of Errors in the Context.
func (c *Context) SetEvents(events elemental.Events) {

	c.eventsLock.Lock()
	defer c.eventsLock.Unlock()

	c.events = events
}

// HasEvents returns true if the context has some custom events.
func (c *Context) HasEvents() bool {

	c.eventsLock.Lock()
	defer c.eventsLock.Unlock()

	return len(c.events) > 0
}

// Events returns the current Events.
func (c *Context) Events() elemental.Events {

	c.eventsLock.Lock()
	defer c.eventsLock.Unlock()

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
