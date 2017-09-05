package bahamut

import (
	"bytes"
	"encoding/json"

	"github.com/opentracing/opentracing-go/ext"
	"github.com/opentracing/opentracing-go/log"

	opentracing "github.com/opentracing/opentracing-go"
)

// Publication is a structure that can be published to a PublishServer.
type Publication struct {
	Data         json.RawMessage            `json:"data,omitempty"`
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
	if err := json.NewEncoder(buffer).Encode(o); err != nil {
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

	return json.NewDecoder(bytes.NewReader(p.Data)).Decode(&dest)
}

// StartTracingFromSpan starts a new child opentracing.Span using the given span as parent.
func (p *Publication) StartTracingFromSpan(span opentracing.Span, name string) error {

	tracer := opentracing.GlobalTracer()
	if tracer == nil {
		return nil
	}

	p.span = opentracing.StartSpan(name, opentracing.ChildOf(span.Context()))
	p.populateSpan()

	return tracer.Inject(p.span.Context(), opentracing.TextMap, p.TrackingData)
}

// StartTracing starts a new tracer using wired data if any.
func (p *Publication) StartTracing(name string) {

	tracer := opentracing.GlobalTracer()
	if tracer == nil {
		return
	}

	wireContext, _ := tracer.Extract(opentracing.TextMap, opentracing.TextMapCarrier(p.TrackingData))

	p.span = opentracing.StartSpan(name, ext.RPCServerOption(wireContext))
	p.populateSpan()
}

// Span returns the current tracking span.
func (p *Publication) Span() opentracing.Span {

	return p.span
}

// SetSpanTag sets the tag of the inner span if any.
func (p *Publication) SetSpanTag(key string, value interface{}) {

	if p.span == nil {
		return
	}

	p.span.SetTag(key, value)
}

// SetSpanLogs sets the logs of the inner span if any.
func (p *Publication) SetSpanLogs(fields ...log.Field) {

	if p.span == nil {
		return
	}

	p.span.LogFields(fields...)
}

// FinishTracing will finish the publication tracing.
func (p *Publication) FinishTracing() {

	if p.span == nil {
		return
	}

	p.span.Finish()
}

func (p *Publication) populateSpan() {

	if p.span == nil {
		return
	}

	p.span.SetTag("topic", p.Topic)
	p.span.SetTag("partition", p.Partition)
}
