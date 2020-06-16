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
	"net/http"
	"sync"

	"github.com/gofrs/uuid"
	"go.aporeto.io/elemental"
)

type bcontext struct {
	claims                []string
	claimsMap             map[string]string
	count                 int
	ctx                   context.Context
	events                elemental.Events
	eventsLock            *sync.Mutex
	id                    string
	inputData             interface{}
	messages              []string
	messagesLock          *sync.Mutex
	metadata              map[interface{}]interface{}
	next                  string
	outputCookies         []*http.Cookie
	outputData            interface{}
	redirect              string
	request               *elemental.Request
	responseWriter        ResponseWriter
	statusCode            int
	disableOutputDataPush bool
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
		id:           uuid.Must(uuid.NewV4()).String(),
		messagesLock: &sync.Mutex{},
		request:      request,
	}
}

func (c *bcontext) Identifier() string {
	return c.id
}

func (c *bcontext) Context() context.Context {
	return c.ctx
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

func (c *bcontext) SetDisableOutputDataPush(disabled bool) {
	c.disableOutputDataPush = disabled
}

func (c *bcontext) SetOutputData(data interface{}) {

	if c.responseWriter != nil {
		panic("you cannot use SetOutputData after using SetResponseWriter")
	}

	c.outputData = data
}

func (c *bcontext) SetResponseWriter(writer ResponseWriter) {

	if c.outputData != nil {
		panic("you cannot use SetResponseWriter after using SetOutputData")
	}

	c.responseWriter = writer
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

	c.claims = append([]string{}, claims...)
	c.claimsMap = claimsToMap(c.claims)
}

func (c *bcontext) Claims() []string {

	return append([]string{}, c.claims...)
}

func (c *bcontext) ClaimsMap() map[string]string {

	o := make(map[string]string, len(c.claimsMap))

	for k, v := range c.claimsMap {
		o[k] = v
	}

	return o
}

func (c *bcontext) EnqueueEvents(events ...*elemental.Event) {

	c.eventsLock.Lock()
	defer c.eventsLock.Unlock()

	c.events = append(c.events, events...)
}

func (c *bcontext) SetNext(next string) {
	c.next = next
}

func (c *bcontext) AddMessage(msg string) {
	c.messagesLock.Lock()
	c.messages = append(c.messages, msg)
	c.messagesLock.Unlock()
}

func (c *bcontext) AddOutputCookies(cookies ...*http.Cookie) {
	c.outputCookies = append(c.outputCookies, cookies...)
}

func (c *bcontext) Duplicate() Context {

	c2 := newContext(c.ctx, c.request.Duplicate())

	c2.inputData = c.inputData
	c2.count = c.count
	c2.statusCode = c.statusCode
	c2.outputData = c.outputData
	c2.claims = append(c2.claims, c.claims...)
	c2.redirect = c.redirect
	c2.messages = append(c2.messages, c.messages...)
	c2.next = c.next
	c2.outputCookies = append(c2.outputCookies, c.outputCookies...)
	c2.responseWriter = c.responseWriter
	c2.disableOutputDataPush = c.disableOutputDataPush

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
