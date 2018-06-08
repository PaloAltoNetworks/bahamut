package bahamut

import (
	"net/http"
	"reflect"
	"sync"
	"time"

	"github.com/aporeto-inc/elemental"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
)

// A counter is a concurency safe count holder.
type counter struct {
	count int

	sync.Mutex
}

func (c *counter) Add(i int) {

	c.Lock()
	defer c.Unlock()

	c.count += i
}

func (c *counter) Value() int {

	c.Lock()
	defer c.Unlock()

	return c.count
}

// A stopper is a concurency state holder.
type stopper struct {
	stopped bool
	done    chan struct{}

	sync.Mutex
}

func newStopper() *stopper {

	return &stopper{
		done: make(chan struct{}, 1),
	}
}

func (s *stopper) Stop() {

	s.Lock()
	defer s.Unlock()

	select {
	case s.done <- struct{}{}:
	default:
	}

	s.stopped = true
}

func (s *stopper) isStopped() bool {

	s.Lock()
	defer s.Unlock()

	return s.stopped
}

func (s *stopper) Done() chan struct{} {

	return s.done
}

// A mockWebsocket is mock implementation of a internalWebsocket.
type mockWebsocket struct {
	writeErr error
	closeErr error
	outData  chan interface{}
	inData   chan interface{}

	sync.Mutex
}

func newMockWebsocket() *mockWebsocket {

	return &mockWebsocket{
		outData: make(chan interface{}),
		inData:  make(chan interface{}),
	}
}

func (s *mockWebsocket) setWriteErr(err error) {

	s.Lock()
	defer s.Unlock()

	s.writeErr = err
}

func (s *mockWebsocket) setCloseErr(err error) {

	s.Lock()
	defer s.Unlock()

	s.closeErr = err
}

func (s *mockWebsocket) setNextRead(i interface{}) {

	go func() { s.inData <- i }()
}

func (s *mockWebsocket) getLastWrite() <-chan interface{} {

	return s.outData
}

func (s *mockWebsocket) ReadJSON(data interface{}) error {

	s.Lock()
	defer s.Unlock()

	d := <-s.inData

	if err, ok := d.(error); ok {
		return err
	}

	reflect.ValueOf(data).Elem().Set(reflect.ValueOf(d))

	return nil
}

func (s *mockWebsocket) WriteJSON(data interface{}) error {

	s.Lock()
	defer s.Unlock()

	if s.writeErr != nil {
		return s.writeErr
	}

	s.outData <- data

	return nil
}

func (s *mockWebsocket) Close() error {

	s.Lock()
	defer s.Unlock()

	return s.closeErr
}

// A mockAuditer is a mockable auditer
type mockAuditer struct {
	nbCalls int

	sync.Mutex
}

func (p *mockAuditer) Audit(*Context, error) {

	p.Lock()
	p.nbCalls++
	p.Unlock()
}

func (p *mockAuditer) GetCallCount() int {

	<-time.After(300 * time.Millisecond) // wait for the go routine running the auditer to be done.

	p.Lock()
	defer p.Unlock()

	return p.nbCalls
}

// A mockAuth is a mockable Authorizer or Authenticator.
type mockAuth struct {
	action  AuthAction
	errored bool
	err     error
}

func (a *mockAuth) AuthenticateRequest(ctx *Context) (AuthAction, error) {

	if a.errored {
		if a.err == nil {
			a.err = elemental.NewError("Error", "This is an error.", "bahamut-test", http.StatusInternalServerError)
		}
		return AuthActionKO, a.err
	}

	return a.action, nil
}

func (a *mockAuth) IsAuthorized(ctx *Context) (AuthAction, error) {

	if a.errored {
		if a.err == nil {
			a.err = elemental.NewError("Error", "This is an error.", "bahamut-test", http.StatusInternalServerError)
		}
		return AuthActionKO, a.err
	}

	return a.action, nil
}

// A mockEmptyProcessor is an empty process implementation.
type mockEmptyProcessor struct{}

// A mockProcessor is an mockable Processor.
type mockProcessor struct {
	err    error
	output interface{}
	events []*elemental.Event
}

func (p *mockProcessor) ProcessRetrieveMany(ctx *Context) error {

	ctx.OutputData = p.output
	ctx.EnqueueEvents(p.events...)

	return p.err
}

func (p *mockProcessor) ProcessRetrieve(ctx *Context) error {

	ctx.OutputData = p.output
	ctx.EnqueueEvents(p.events...)

	return p.err
}

func (p *mockProcessor) ProcessCreate(ctx *Context) error {

	ctx.OutputData = p.output
	ctx.EnqueueEvents(p.events...)

	return p.err
}

func (p *mockProcessor) ProcessUpdate(ctx *Context) error {

	ctx.OutputData = p.output
	ctx.EnqueueEvents(p.events...)

	return p.err
}

func (p *mockProcessor) ProcessDelete(ctx *Context) error {

	ctx.OutputData = p.output
	ctx.EnqueueEvents(p.events...)

	return p.err
}

func (p *mockProcessor) ProcessPatch(ctx *Context) error {

	ctx.OutputData = p.output
	ctx.EnqueueEvents(p.events...)

	return p.err
}

func (p *mockProcessor) ProcessInfo(ctx *Context) error {

	ctx.OutputData = p.output
	ctx.EnqueueEvents(p.events...)

	return p.err
}

// A mockPusher is a mockable implementation of a Pusher.
type mockPusher struct {
	events []*elemental.Event
	sync.Mutex
}

func (f *mockPusher) Push(evt ...*elemental.Event) {

	f.Lock()
	defer f.Unlock()

	f.events = append(f.events, evt...)
}

// A mockSpanContext is a mockable opentracing.SpanContext
type mockSpanContext struct {
}

func (t *mockSpanContext) ForeachBaggageItem(handler func(k, v string) bool) {}

// A mockTracer is a mockable opentracing.Tracer
type mockTracer struct {
	currentSpan *mockSpan
	injected    interface{}
}

func (t *mockTracer) StartSpan(string, ...opentracing.StartSpanOption) opentracing.Span {

	if t.currentSpan == nil {
		t.currentSpan = newMockSpan(t)
	}

	return t.currentSpan
}

func (t *mockTracer) Inject(span opentracing.SpanContext, format interface{}, carrier interface{}) error {
	t.injected = carrier
	return nil
}

func (t *mockTracer) Extract(interface{}, interface{}) (opentracing.SpanContext, error) {

	return &mockSpanContext{}, nil
}

// A mockSpan is a mockable opentracing.Span
type mockSpan struct {
	finished bool
	tracer   opentracing.Tracer
	tags     map[string]interface{}
	fields   []log.Field
}

func newMockSpan(tracer opentracing.Tracer) *mockSpan {
	return &mockSpan{
		tracer: tracer,
		tags:   map[string]interface{}{},
		fields: []log.Field{},
	}
}

func (s *mockSpan) Finish() {

	s.finished = true
}

func (s *mockSpan) FinishWithOptions(opts opentracing.FinishOptions) {
	s.finished = true
}

func (s *mockSpan) Context() opentracing.SpanContext {
	return &mockSpanContext{}
}

func (s *mockSpan) SetOperationName(operationName string) opentracing.Span {
	return s
}

func (s *mockSpan) SetTag(key string, value interface{}) opentracing.Span {

	s.tags[key] = value

	return s
}

func (s *mockSpan) LogFields(fields ...log.Field) {

	s.fields = append(s.fields, fields...)
}

func (s *mockSpan) LogKV(alternatingKeyValues ...interface{}) {

}

func (s *mockSpan) SetBaggageItem(restrictedKey, value string) opentracing.Span {
	return s
}

func (s *mockSpan) BaggageItem(restrictedKey string) string {
	return ""

}
func (s *mockSpan) Tracer() opentracing.Tracer {
	return s.tracer
}

func (s *mockSpan) LogEvent(event string)                                 {}
func (s *mockSpan) LogEventWithPayload(event string, payload interface{}) {}
func (s *mockSpan) Log(data opentracing.LogData)                          {}
