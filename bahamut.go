// Author: Antoine Mercadal
// See LICENSE file for full LICENSE
// Copyright 2016 Aporeto.

package bahamut

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/aporeto-inc/elemental"
	"github.com/go-zoo/bone"
	"go.uber.org/zap"
)

// CustomUmarshaller is the type of function use to create custom unmarshalling.
type CustomUmarshaller func(*elemental.Request) (elemental.Identifiable, error)

// RegisterProcessorOrDie will register the given Processor for the given
// Identity and will exit in case of errors. This is just a helper for
// Server.RegisterProcessor function.
func RegisterProcessorOrDie(server Server, processor Processor, identity elemental.Identity) {

	if server == nil {
		zap.L().Fatal("No bahamut set. You must create a bahamut server first")
	}

	if err := server.RegisterProcessor(processor, identity); err != nil {
		zap.L().Fatal("Duplicate identity registration", zap.Error(err))
	}
}

// InstallSIGINTHandler installs signal handlers for graceful shutdown.
func InstallSIGINTHandler(cancelFunc context.CancelFunc) {

	signalCh := make(chan os.Signal, 1)
	signal.Reset(os.Interrupt)
	signal.Notify(signalCh, os.Interrupt)
	go func() {
		<-signalCh
		cancelFunc()
		signal.Stop(signalCh)
		close(signalCh)
	}()
}

type server struct {
	multiplexer *bone.Mux
	processors  map[string]Processor

	restServer      *restServer
	pushServer      *pushServer
	healthServer    *healthServer
	profilingServer *profilingServer
	mockServer      *mockServer
}

// New returns a new bahamut Server configured with
// the given options.
func New(options ...Option) Server {

	c := config{}
	for _, opt := range options {
		opt(&c)
	}

	return NewServer(c)
}

// NewServer returns a new Bahamut Server.
func NewServer(cfg config) Server {

	if cfg.model.unmarshallers == nil {
		cfg.model.unmarshallers = map[elemental.Identity]CustomUmarshaller{}
	}

	mux := bone.New()
	srv := &server{
		multiplexer: mux,
		processors:  make(map[string]Processor),
	}

	if cfg.restServer.enabled {
		srv.restServer = newRestServer(cfg, mux, srv.ProcessorForIdentity, srv.Push)
	}

	if cfg.pushServer.enabled {
		srv.pushServer = newPushServer(cfg, mux, srv.ProcessorForIdentity)
	}

	if cfg.healthServer.enabled {
		srv.healthServer = newHealthServer(cfg)
	}

	if cfg.profilingServer.enabled {
		srv.profilingServer = newProfilingServer(cfg)
	}

	if cfg.mockServer.enabled {
		srv.mockServer = newMockServer(cfg)
	}

	return srv
}

func (b *server) RegisterProcessor(processor Processor, identity elemental.Identity) error {

	if _, ok := b.processors[identity.Name]; ok {
		return fmt.Errorf("identity %s already has a registered processor", identity)
	}

	b.processors[identity.Name] = processor

	return nil
}

func (b *server) UnregisterProcessor(identity elemental.Identity) error {

	if _, ok := b.processors[identity.Name]; !ok {
		return fmt.Errorf("no registered processor for identity %s", identity)
	}

	delete(b.processors, identity.Name)

	return nil
}

func (b *server) ProcessorForIdentity(identity elemental.Identity) (Processor, error) {

	if _, ok := b.processors[identity.Name]; !ok {
		return nil, fmt.Errorf("no registered processor for identity %s", identity)
	}

	return b.processors[identity.Name], nil
}

func (b *server) ProcessorsCount() int {

	return len(b.processors)
}

func (b *server) Push(events ...*elemental.Event) {

	if b.pushServer == nil {
		return
	}

	b.pushServer.pushEvents(events...)
}

func (b *server) Start() {

	zap.L().Warn("Bahamut: deprecated: Server.Start is deprecated. Use Server.Run")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	InstallSIGINTHandler(cancel)
	b.Run(ctx)
}

func (b *server) Run(ctx context.Context) {

	if b.profilingServer != nil {
		go b.profilingServer.start(ctx)
	}

	if b.healthServer != nil {
		go b.healthServer.start(ctx)
	}

	if b.mockServer != nil {
		go b.mockServer.start(ctx)
	}

	if b.restServer != nil {
		go b.restServer.start(ctx)
	}

	if b.pushServer != nil {
		go b.pushServer.start(ctx)
	}

	<-ctx.Done()

	if b.restServer != nil {
		b.restServer.stop()
	}

	if b.pushServer != nil {
		b.pushServer.stop()
	}

	if b.profilingServer != nil {
		b.profilingServer.stop()
	}

	if b.healthServer != nil {
		b.healthServer.stop()
	}

	if b.mockServer != nil {
		b.mockServer.stop()
	}

}
