// Author: Antoine Mercadal
// See LICENSE file for full LICENSE
// Copyright 2016 Aporeto.

package bahamut

import (
	"fmt"
	"os"
	"os/signal"

	log "github.com/Sirupsen/logrus"
	"github.com/aporeto-inc/elemental"
	"github.com/go-zoo/bone"
)

var defaultBahamut *Bahamut

// DefaultBahamut returns the defaut Bahamut.
// Needless to say I don't like this. but that will be ok for now.
func DefaultBahamut() *Bahamut {
	return defaultBahamut
}

// Bahamut is crazy
type Bahamut struct {
	apiServer     *apiServer
	pushServer    *pushServer
	multiplexer   *bone.Mux
	stop          chan bool
	processors    map[string]Processor
	authenticator Authenticator
}

// NewBahamut creates a new Bahamut.
func NewBahamut(apiConfig APIServerConfig, pushConfig PushServerConfig) *Bahamut {

	mux := bone.New()

	var apiServer *apiServer
	if apiConfig.enabled {
		apiServer = newAPIServer(apiConfig, mux)
	}

	var pushServer *pushServer
	if pushConfig.enabled {
		pushServer = newPushServer(pushConfig, mux)
	}

	srv := &Bahamut{
		apiServer:   apiServer,
		pushServer:  pushServer,
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
func (b *Bahamut) Push(events ...*elemental.Event) {

	if b.apiServer == nil {
		panic("you cannot push events as it is not enabled.")
	}

	b.pushServer.pushEvents(events...)
}

// SetAuthenticator sets the Authenticator to use for the Bahamut server.
func (b *Bahamut) SetAuthenticator(authenticator Authenticator) {
	b.authenticator = authenticator
}

// Authenticator returns the current authenticator
func (b *Bahamut) Authenticator() (Authenticator, error) {

	if b.authenticator == nil {
		return nil, fmt.Errorf("no authenticator set")
	}

	return b.authenticator, nil
}

// Start starts the Bahamut server.
func (b *Bahamut) Start() {

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		<-c
		log.Info("shutting down...")
		b.Stop()
		log.Info("bye!")
	}()

	if b.apiServer != nil {
		go b.apiServer.start()
	}

	if b.pushServer != nil {
		go b.pushServer.start()
	}

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

	b.stop <- true
}
