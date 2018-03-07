package bahamut

import (
	"bytes"

	"github.com/opentracing/opentracing-go/ext"
	"github.com/opentracing/opentracing-go/log"

	jsoniter "github.com/json-iterator/go"
	opentracing "github.com/opentracing/opentracing-go"
)

// Publication is a structure that can be published to a PublishServer.
type Publication struct {
	Data         jsoniter.RawMessage        `json:"data,omitempty"`
	Topic        string                     `json:"topic,omitempty"`
	Partition    int32                      `json:"partition,omitempty"`
	TrackingName string                     `json:"trackingName,omitempty"`
	TrackingData opentracing.TextMapCarrier `json:"trackingData,omitempty"`

	span opentracing.Span
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

	buffer := &bytes.Buffer{}
	if err := jsoniter.NewEncoder(buffer).Encode(o); err != nil {
		return err
	}

	p.Data = buffer.Bytes()

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

	return jsoniter.NewDecoder(bytes.NewReader(p.Data)).Decode(&dest)
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

	wireContext, _ := tracer.Extract(opentracing.TextMap, opentracing.TextMapCarrier(p.TrackingData))

	p.span = opentracing.StartSpan(name, ext.RPCServerOption(wireContext))
	p.span.SetTag("topic", p.Topic)
	p.span.SetTag("partition", p.Partition)

}

// Span returns the current tracking span.
func (p *Publication) Span() opentracing.Span {

	return p.span
}
