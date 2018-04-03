// Author: Antoine Mercadal
// See LICENSE file for full LICENSE
// Copyright 2016 Aporeto.

package bahamut

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"go.uber.org/zap"

	"github.com/aporeto-inc/elemental"
	"github.com/go-zoo/bone"
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
	tracingServer   *tracingServer
	mockServer      *mockServer
}

// NewServer returns a new Bahamut Server.
func NewServer(config Config) Server {

	if config.Model.Unmarshallers == nil {
		config.Model.Unmarshallers = map[elemental.Identity]CustomUmarshaller{}
	}

	mux := bone.New()
	srv := &server{
		multiplexer: mux,
		processors:  make(map[string]Processor),
	}

	if !config.ReSTServer.Disabled {
		srv.restServer = newRestServer(config, mux, srv.ProcessorForIdentity, srv.Push)
	}

	if !config.PushServer.Disabled {
		srv.pushServer = newPushServer(config, mux, srv.ProcessorForIdentity)
	}

	if !config.HealthServer.Disabled {
		srv.healthServer = newHealthServer(config)
	}

	if config.ProfilingServer.Enabled {
		srv.profilingServer = newProfilingServer(config)
	}

	if config.TracingServer.Enabled {
		srv.tracingServer = newTracingServer(config)
	}

	if config.MockServer.Enabled {
		srv.mockServer = newMockServer(config)
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

	if b.tracingServer != nil {
		go b.tracingServer.start(ctx)
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

	if b.profilingServer != nil {
		b.profilingServer.stop()
	}

	if b.tracingServer != nil {
		b.tracingServer.stop()
	}

	if b.healthServer != nil {
		b.healthServer.stop()
	}

	if b.mockServer != nil {
		b.mockServer.stop()
	}

	if b.restServer != nil {
		b.restServer.stop()
	}
}
