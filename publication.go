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
	"errors"
	"sync"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/opentracing/opentracing-go/log"
	"go.aporeto.io/elemental"
)

// ResponseMode represents the response that is expected to be produced by the subscriber
// handling a publication.
type ResponseMode int

func (r ResponseMode) String() string {
	switch r {
	case ResponseModeNone:
		return "ResponseModeNone"
	case ResponseModeACK:
		return "ResponseModeACK"
	case ResponseModePublication:
		return "ResponseModePublication"
	default:
		return "ResponseModeUnknown"
	}
}

const (
	// ResponseModeNone indicates that no response is expected for the received publication
	ResponseModeNone ResponseMode = iota
	// ResponseModeACK indicates that the subscriber should reply back with an ACK
	// as soon as it has received the publication BEFORE it starts processing the
	// publication.
	ResponseModeACK
	// ResponseModePublication indicates that the subscriber should reply back with a
	// Publication AFTER it has finished processing the publication. Obviously, the
	// subscriber should try to respond ASAP as there is a client waiting for a response.
	ResponseModePublication
)

// Publication is a structure that can be published to a PublishServer.
type Publication struct {
	Data         []byte                     `msgpack:"data,omitempty" json:"data,omitempty"`
	Topic        string                     `msgpack:"topic,omitempty" json:"topic,omitempty"`
	Partition    int32                      `msgpack:"partition,omitempty" json:"partition,omitempty"`
	TrackingName string                     `msgpack:"trackingName,omitempty" json:"trackingName,omitempty"`
	TrackingData opentracing.TextMapCarrier `msgpack:"trackingData,omitempty" json:"trackingData,omitempty"`
	Encoding     elemental.EncodingType     `msgpack:"encoding,omitempty" json:"encoding,omitempty"`
	ResponseMode ResponseMode               `msgpack:"responseMode,omitempty" json:"responseMode,omitempty"`

	replyCh  chan *Publication
	replied  bool
	timedOut bool
	mux      sync.Mutex
	span     opentracing.Span
}

// NewPublication returns a new Publication.
func NewPublication(topic string) *Publication {

	return &Publication{
		Topic:        topic,
		TrackingData: opentracing.TextMapCarrier{},
	}
}

// Encode the given object into the publication.
func (p *Publication) Encode(o interface{}) error {
	return p.EncodeWithEncoding(o, elemental.EncodingTypeMSGPACK)
}

// EncodeWithEncoding the given object into the publication using the given encoding.
func (p *Publication) EncodeWithEncoding(o interface{}, encoding elemental.EncodingType) error {

	data, err := elemental.Encode(encoding, o)
	if err != nil {
		return err
	}

	p.Data = data
	p.Encoding = encoding

	if p.span != nil {
		p.span.LogFields(log.Object("payload", string(p.Data)))
	}

	return nil
}

// Decode decodes the data into the given dest.
func (p *Publication) Decode(dest interface{}) error {

	if p.span != nil {
		p.span.LogFields(log.Object("payload", string(p.Data)))
	}

	return elemental.Decode(p.Encoding, p.Data, dest)
}

// StartTracingFromSpan starts a new child opentracing.Span using the given span as parent.
func (p *Publication) StartTracingFromSpan(span opentracing.Span, name string) error {

	tracer := span.Tracer()
	if tracer == nil {
		return nil
	}

	p.span = opentracing.StartSpan(name, opentracing.ChildOf(span.Context()))
	p.span.SetTag("topic", p.Topic)
	p.span.SetTag("partition", p.Partition)

	return tracer.Inject(p.span.Context(), opentracing.TextMap, p.TrackingData)
}

// StartTracing starts a new tracer using wired data if any.
func (p *Publication) StartTracing(tracer opentracing.Tracer, name string) {

	if tracer == nil {
		return
	}

	wireContext, _ := tracer.Extract(opentracing.TextMap, p.TrackingData)

	p.span = opentracing.StartSpan(name, ext.RPCServerOption(wireContext))
	p.span.SetTag("topic", p.Topic)
	p.span.SetTag("partition", p.Partition)

}

// Span returns the current tracking span.
func (p *Publication) Span() opentracing.Span {

	return p.span
}

// Duplicate returns a copy of the publication
func (p *Publication) Duplicate() *Publication {

	pub := NewPublication(p.Topic)
	pub.Data = p.Data
	pub.Partition = p.Partition
	pub.TrackingName = p.TrackingName
	pub.TrackingData = p.TrackingData
	pub.Encoding = p.Encoding
	pub.ResponseMode = p.ResponseMode
	pub.span = p.span

	return pub
}

// Reply will publish the provided publication back to the client. An error is returned if
// the client was not expecting a response or the supplied publication was nil. If you take
// too long to reply to a publication an error may be returned in the errors channel you provided
// in your call to the `Subscribe` method as the client may have given up waiting for your response.
// Reply can only be called once for
func (p *Publication) Reply(response *Publication) error {
	p.mux.Lock()
	defer p.mux.Unlock()

	switch {
	case p.timedOut:
		return errors.New("took too long to reply to publication")
	case response == nil:
		return errors.New("response cannot be nil")
	case p.replied:
		return errors.New("already replied to publication")
	case p.replyCh == nil:
		return errors.New("no response required for publication")
	}

	p.replyCh <- response
	p.replied = true

	return nil
}

func (p *Publication) setExpired() {
	p.mux.Lock()
	defer p.mux.Unlock()
	p.timedOut = true
	p.replyCh = nil
}
