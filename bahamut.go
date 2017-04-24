// Author: Antoine Mercadal
// See LICENSE file for full LICENSE
// Copyright 2016 Aporeto.

package bahamut

import (
	"fmt"
	"os"
	"os/signal"

	"go.uber.org/zap"

	"github.com/aporeto-inc/elemental"
	"github.com/go-zoo/bone"
)

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

type server struct {
	multiplexer *bone.Mux
	processors  map[string]Processor

	apiServer       *apiServer
	pushServer      *pushServer
	healthServer    *healthServer
	profilingServer *profilingServer

	stop chan bool
}

// NewServer returns a new Bahamut Server.
//
// It will use the given apiConfig and pushConfig to initialize the various servers.
func NewServer(config Config) Server {

	mux := bone.New()
	srv := &server{
		multiplexer: mux,
		stop:        make(chan bool),
		processors:  make(map[string]Processor),
	}

	if !config.ReSTServer.Disabled {
		srv.apiServer = newAPIServer(config, mux)
		srv.apiServer.processorFinder = srv.ProcessorForIdentity
		srv.apiServer.pusher = srv.Push
	}

	if !config.WebSocketServer.Disabled {
		srv.pushServer = newPushServer(config, mux)
		srv.pushServer.processorFinder = srv.ProcessorForIdentity
	}

	if !config.HealthServer.Disabled {
		srv.healthServer = newHealthServer(config)
	}

	if !config.ProfilingServer.Disabled {
		srv.profilingServer = newProfilingServer(config)
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

// handleExit handle the interrupt signal an will try
// to cleanly stop all current routines.
func (b *server) handleExit() {

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	<-c

	b.Stop()
}

func (b *server) Start() {

	if b.apiServer != nil {
		go b.apiServer.start()
	}

	if b.pushServer != nil {
		go b.pushServer.start()
	}

	if b.healthServer != nil {
		go b.healthServer.start()
	}

	if b.profilingServer != nil {
		go b.profilingServer.start()
	}

	go b.handleExit()

	<-b.stop
}

func (b *server) Stop() {

	if b.apiServer != nil {
		b.apiServer.stop()
	}

	if b.pushServer != nil {
		b.pushServer.stop()
	}

	if b.healthServer != nil {
		b.healthServer.stop()
	}

	if b.profilingServer != nil {
		b.profilingServer.stop()
	}

	b.stop <- true
}
