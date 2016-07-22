// Author: Antoine Mercadal
// See LICENSE file for full LICENSE
// Copyright 2016 Aporeto.

package bahamut

import (
	"fmt"
	"os"
	"os/signal"

	log "github.com/Sirupsen/logrus"
	"github.com/aporeto-inc/bahamut/pubsub"
	"github.com/aporeto-inc/elemental"
	"github.com/go-zoo/bone"
)

var defaultBahamut *Bahamut

// DefaultBahamut returns the defaut Bahamut.
// Needless to say I don't like this. but that will be ok for now.
func DefaultBahamut() *Bahamut {
	return defaultBahamut
}

// A Bahamut is an application server.
//
// It serves various configured api routes, routes the calls to the
// correct processors, creates a push notification system using websocket.
//
// This is the main entrypoint of the library.
type Bahamut struct {
	multiplexer *bone.Mux
	processors  map[string]Processor

	apiServer    *apiServer
	pushServer   *pushServer
	pubSubServer pubsub.Server

	authenticator Authenticator
	authorizer    Authorizer

	stop chan bool
}

// NewBahamut returns a new Bahamut.
//
// It will use the given apiConfig and pushConfig to initialize the various servers.
func NewBahamut(apiConfig APIServerConfig, pushConfig PushServerConfig) *Bahamut {

	mux := bone.New()

	var apiServer *apiServer
	if apiConfig.enabled {
		apiServer = newAPIServer(apiConfig, mux)
	}

	var pushServer *pushServer
	var pubsubServer pubsub.Server
	if pushConfig.enabled {
		pubsubServer = pubsub.NewServer(pushConfig.kafkaAddresses)
		pushServer = newPushServer(pushConfig, pubsubServer, mux)
	}

	srv := &Bahamut{
		apiServer:    apiServer,
		pushServer:   pushServer,
		pubSubServer: pubsubServer,

		multiplexer: mux,
		stop:        make(chan bool),
		processors:  make(map[string]Processor),
	}

	defaultBahamut = srv

	return srv
}

// RegisterProcessor registers a new Processor for a particular Identity.
func (b *Bahamut) RegisterProcessor(processor Processor, identity elemental.Identity) error {

	if _, ok := b.processors[identity.Name]; ok {
		return fmt.Errorf("identity %s already has a registered processor", identity)
	}

	b.processors[identity.Name] = processor

	return nil
}

// UnregisterProcessor unregisters a registered Processor for a particular identity.
func (b *Bahamut) UnregisterProcessor(identity elemental.Identity) error {

	if _, ok := b.processors[identity.Name]; !ok {
		return fmt.Errorf("no registered processor for identity %s", identity)
	}

	delete(b.processors, identity.Name)

	return nil
}

// ProcessorForIdentity returns the registered Processor for a particular identity.
func (b *Bahamut) ProcessorForIdentity(identity elemental.Identity) (Processor, error) {

	if _, ok := b.processors[identity.Name]; !ok {
		return nil, fmt.Errorf("no registered processor for identity %s", identity)
	}

	return b.processors[identity.Name], nil
}

// Push pushes the given events to all active sessions.
//
// Depending on the configuration of the pushServer, it may use
// internal local push system, or may use Kafka to publish the events
// to a cluster of Bahamut Servers.
func (b *Bahamut) Push(events ...*elemental.Event) {

	if b.apiServer == nil {
		panic("you cannot push events as it is not enabled.")
	}

	b.pushServer.pushEvents(events...)
}

// SetAuthenticator sets the main Authenticator to use for the Bahamut server.
//
// An authenticator must implement the Authenticator interface.
func (b *Bahamut) SetAuthenticator(authenticator Authenticator) {
	b.authenticator = authenticator
}

// Authenticator returns the current authenticator.
//
// It will return an error if none is set.
func (b *Bahamut) Authenticator() (Authenticator, error) {

	if b.authenticator == nil {
		return nil, fmt.Errorf("no authenticator set")
	}

	return b.authenticator, nil
}

// SetAuthorizer sets the main Authorizer to use for the Bahamut server.
//
// An authorizer must implement the Authorizer interface.
func (b *Bahamut) SetAuthorizer(authorizer Authorizer) {

	b.authorizer = authorizer
}

// Authorizer returns the current authenticator.
//
// It will return an error if none is set.
func (b *Bahamut) Authorizer() (Authorizer, error) {

	if b.authorizer == nil {
		return nil, fmt.Errorf("no authorizer set")
	}

	return b.authorizer, nil
}

// handleExit handle the interupt signal an will try
// to cleanly stop all current routines.
func (b *Bahamut) handleExit() {

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	<-c

	b.Stop()
	log.WithFields(log.Fields{
		"materia": "bahamut",
	}).Info("Bye!")
}

// Start starts the Bahamut server.
func (b *Bahamut) Start() {

	if b.apiServer != nil {
		go b.apiServer.start()
	}

	if b.pubSubServer != nil {
		go b.pubSubServer.Start()
	}

	if b.pushServer != nil {
		go b.pushServer.start()
	}

	go b.handleExit()

	<-b.stop
}

// Stop stops the Bahamut server.
func (b *Bahamut) Stop() {

	if b.apiServer != nil {
		b.apiServer.stop()
	}

	if b.pushServer != nil {
		b.pushServer.stop()
	}

	if b.pubSubServer != nil {
		b.pubSubServer.Stop()
	}

	b.stop <- true
}
