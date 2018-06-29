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

type bcontext struct {
	claims       []string
	claimsMap    map[string]string
	count        int
	ctx          context.Context
	events       elemental.Events
	eventsLock   *sync.Mutex
	id           string
	inputData    interface{}
	messages     []string
	messagesLock *sync.Mutex
	metadata     map[interface{}]interface{}
	outputData   interface{}
	redirect     string
	request      *elemental.Request
	statusCode   int
}

// NewContext creates a new *Context.
func NewContext(ctx context.Context, request *elemental.Request) Context {
	return newContext(ctx, request)
}

func newContext(ctx context.Context, request *elemental.Request) *bcontext {

	if ctx == nil {
		panic("nil context")
	}

	return &bcontext{
		claims:       nil,
		claimsMap:    map[string]string{},
		ctx:          ctx,
		eventsLock:   &sync.Mutex{},
		id:           uuid.NewV4().String(),
		messagesLock: &sync.Mutex{},
		request:      request,
	}
}

func (c *bcontext) Identifier() string {

	return c.id
}

func (c *bcontext) Context() context.Context {

	if c.ctx != nil {
		return c.ctx
	}

	return context.Background()
}

func (c *bcontext) Request() *elemental.Request {
	return c.request
}

func (c *bcontext) Count() int {
	return c.count
}

func (c *bcontext) SetCount(count int) {
	c.count = count
}

func (c *bcontext) InputData() interface{} {
	return c.inputData
}

func (c *bcontext) SetInputData(data interface{}) {
	c.inputData = data
}

func (c *bcontext) OutputData() interface{} {
	return c.outputData
}

func (c *bcontext) SetOutputData(data interface{}) {
	c.outputData = data
}

func (c *bcontext) StatusCode() int {
	return c.statusCode
}

func (c *bcontext) SetStatusCode(code int) {
	c.statusCode = code
}

func (c *bcontext) Redirect() string {
	return c.redirect
}

func (c *bcontext) SetRedirect(url string) {
	c.redirect = url
}

func (c *bcontext) Metadata(key interface{}) interface{} {

	if c.metadata == nil {
		return nil
	}

	return c.metadata[key]
}

func (c *bcontext) SetMetadata(key, value interface{}) {

	if c.metadata == nil {
		c.metadata = map[interface{}]interface{}{}
	}

	c.metadata[key] = value
}

func (c *bcontext) SetClaims(claims []string) {

	if claims == nil {
		return
	}

	c.claims = claims
	c.claimsMap = claimsToMap(claims)
}

func (c *bcontext) Claims() []string {

	return c.claims
}

func (c *bcontext) ClaimsMap() map[string]string {

	return c.claimsMap
}

func (c *bcontext) EnqueueEvents(events ...*elemental.Event) {

	c.eventsLock.Lock()
	defer c.eventsLock.Unlock()

	c.events = append(c.events, events...)
}

func (c *bcontext) AddMessage(msg string) {
	c.messagesLock.Lock()
	c.messages = append(c.messages, msg)
	c.messagesLock.Unlock()
}

func (c *bcontext) Duplicate() Context {

	c2 := newContext(c.ctx, c.request.Duplicate())

	c2.inputData = c.inputData
	c2.count = c.count
	c2.statusCode = c.statusCode
	c2.outputData = c.outputData
	c2.claims = append(c2.claims, c.claims...)
	c2.messages = append(c2.messages, c.messages...)

	for k, v := range c.claimsMap {
		c2.claimsMap[k] = v
	}

	if c.metadata != nil {
		c2.metadata = map[interface{}]interface{}{}
		for k, v := range c.metadata {
			c2.metadata[k] = v
		}
	}

	return c2
}

func (c *bcontext) String() string {

	return fmt.Sprintf("<context id:%s request:%s totalcount:%d>",
		c.Identifier(),
		c.request,
		c.count,
	)
}
