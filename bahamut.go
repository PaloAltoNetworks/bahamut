// Author: Antoine Mercadal
// See LICENSE file for full LICENSE
// Copyright 2016 Aporeto.

package bahamut

import (
	"fmt"
	"os"
	"os/signal"

	"github.com/aporeto-inc/elemental"
	"github.com/go-zoo/bone"

	log "github.com/Sirupsen/logrus"
)

// RegisterProcessorOrDie will register the given Processor for the given
// Identity and will exit in case of errors. This is just a helper for
// Server.RegisterProcessor function.
func RegisterProcessorOrDie(server Server, processor Processor, identity elemental.Identity) {

	if server == nil {
		log.WithFields(log.Fields{
			"package": "bahamut",
		}).Fatal("Not bahamut set. You must create a bahamut server first.")
	}

	if err := server.RegisterProcessor(processor, identity); err != nil {
		log.WithFields(log.Fields{
			"package": "bahamut",
			"error":   err.Error(),
		}).Fatal("Duplicate identity registration.")
	}
}

type server struct {
	multiplexer *bone.Mux
	processors  map[string]Processor

	apiServer  *apiServer
	pushServer *pushServer

	stop chan bool
}

// NewServer returns a new Bahamut Server.
//
// It will use the given apiConfig and pushConfig to initialize the various servers.
func NewServer(config Config) Server {

	mux := bone.New()

	var apiServer *apiServer
	if !config.ReSTServer.Disabled {
		apiServer = newAPIServer(config, mux)
	}

	var pushServer *pushServer
	if !config.WebSocketServer.Disabled {
		pushServer = newPushServer(config, mux)
	}

	srv := &server{
		apiServer:   apiServer,
		pushServer:  pushServer,
		multiplexer: mux,
		stop:        make(chan bool),
		processors:  make(map[string]Processor),
	}

	if !config.WebSocketServer.Disabled {
		pushServer.processorFinder = srv.ProcessorForIdentity
	}

	if !config.ReSTServer.Disabled {
		apiServer.processorFinder = srv.ProcessorForIdentity
		apiServer.pusher = srv.Push
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

// handleExit handle the interupt signal an will try
// to cleanly stop all current routines.
func (b *server) handleExit() {

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	<-c

	b.Stop()
	log.WithFields(log.Fields{
		"package": "bahamut",
	}).Info("Bye!")
}

func (b *server) Start() {

	if b.apiServer != nil {
		go b.apiServer.start()
	}

	if b.pushServer != nil {
		go b.pushServer.start()
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

	b.stop <- true
}
