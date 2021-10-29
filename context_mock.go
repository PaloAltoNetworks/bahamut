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

	"github.com/gofrs/uuid"
	"go.aporeto.io/elemental"
)

type MockContext struct {
	MockClaims                []string
	MockClaimsMap             map[string]string
	MockCount                 int
	MockCtx                   context.Context
	MockEvents                elemental.Events
	MockID                    string
	MockInputData             interface{}
	MockMessages              []string
	MockMetadata              map[interface{}]interface{}
	MockNext                  string
	MockOutputCookies         []*http.Cookie
	MockOutputData            interface{}
	MockRedirect              string
	MockRequest               *elemental.Request
	MockResponseWriter        ResponseWriter
	MockStatusCode            int
	MockDisableOutputDataPush bool
}

func NewMockContext(ctx context.Context) *MockContext {
	return &MockContext{
		MockClaimsMap: map[string]string{},
		MockCtx:       ctx,
		MockID:        uuid.Must(uuid.NewV4()).String(),
	}
}
func (c *MockContext) Identifier() string {
	return c.MockID
}

func (c *MockContext) Context() context.Context {
	return c.MockCtx
}

func (c *MockContext) Request() *elemental.Request {
	return c.MockRequest
}

func (c *MockContext) Count() int {
	return c.MockCount
}

func (c *MockContext) SetCount(count int) {
	c.MockCount = count
}

func (c *MockContext) InputData() interface{} {
	return c.MockInputData
}

func (c *MockContext) SetInputData(data interface{}) {
	c.MockInputData = data
}

func (c *MockContext) OutputData() interface{} {
	return c.MockOutputData
}

func (c *MockContext) SetDisableOutputDataPush(disabled bool) {
	c.MockDisableOutputDataPush = disabled
}

func (c *MockContext) SetOutputData(data interface{}) {
	c.MockOutputData = data
}

func (c *MockContext) SetResponseWriter(writer ResponseWriter) {
	c.MockResponseWriter = writer
}

func (c *MockContext) StatusCode() int {
	return c.MockStatusCode
}

func (c *MockContext) SetStatusCode(code int) {
	c.MockStatusCode = code
}

func (c *MockContext) Redirect() string {
	return c.MockRedirect
}

func (c *MockContext) SetRedirect(url string) {
	c.MockRedirect = url
}

func (c *MockContext) Metadata(key interface{}) interface{} {

	if c.MockMetadata == nil {
		return nil
	}

	return c.MockMetadata[key]
}

func (c *MockContext) SetMetadata(key, value interface{}) {

	if c.MockMetadata == nil {
		c.MockMetadata = map[interface{}]interface{}{}
	}

	c.MockMetadata[key] = value
}

func (c *MockContext) SetClaims(claims []string) {

	if claims == nil {
		return
	}

	c.MockClaims = append([]string{}, claims...)
	c.MockClaimsMap = claimsToMap(c.MockClaims)
}

func (c *MockContext) Claims() []string {

	return append([]string{}, c.MockClaims...)
}

func (c *MockContext) ClaimsMap() map[string]string {

	o := make(map[string]string, len(c.MockClaimsMap))

	for k, v := range c.MockClaimsMap {
		o[k] = v
	}

	return o
}

func (c *MockContext) EnqueueEvents(events ...*elemental.Event) {

	c.MockEvents = append(c.MockEvents, events...)
}

func (c *MockContext) SetNext(next string) {
	c.MockNext = next
}

func (c *MockContext) AddMessage(msg string) {
	c.MockMessages = append(c.MockMessages, msg)
}

func (c *MockContext) AddOutputCookies(cookies ...*http.Cookie) {
	c.MockOutputCookies = append(c.MockOutputCookies, cookies...)
}

func (c *MockContext) Duplicate() Context {

	c2 := NewMockContext(c.MockCtx)

	if c.MockRequest != nil {
		c2.MockRequest = c.MockRequest.Duplicate()
	}

	c2.MockInputData = c.MockInputData
	c2.MockCount = c.MockCount
	c2.MockStatusCode = c.MockStatusCode
	c2.MockOutputData = c.MockOutputData
	c2.MockClaims = append(c2.MockClaims, c.MockClaims...)
	c2.MockRedirect = c.MockRedirect
	c2.MockMessages = append(c2.MockMessages, c.MockMessages...)
	c2.MockNext = c.MockNext
	c2.MockOutputCookies = append(c2.MockOutputCookies, c.MockOutputCookies...)
	c2.MockResponseWriter = c.MockResponseWriter
	c2.MockDisableOutputDataPush = c.MockDisableOutputDataPush

	for k, v := range c.MockClaimsMap {
		c2.MockClaimsMap[k] = v
	}

	if c.MockMetadata != nil {
		c2.MockMetadata = map[interface{}]interface{}{}
		for k, v := range c.MockMetadata {
			c2.MockMetadata[k] = v
		}
	}

	return c2
}
