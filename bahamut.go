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
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-zoo/bone"
	"go.aporeto.io/elemental"
	"go.uber.org/zap"
)

// CustomUmarshaller is the type of function use to create custom unmarshalling.
type CustomUmarshaller func(*elemental.Request) (elemental.Identifiable, error)

// CustomMarshaller is the type of function use to create custom marshalling.
type CustomMarshaller func(*elemental.Response, interface{}, error) ([]byte, error)

// RegisterProcessorOrDie will register the given Processor for the given
// Identity and will exit in case of errors. This is just a helper for
// Server.RegisterProcessor function.
func RegisterProcessorOrDie(server Server, processor Processor, identity elemental.Identity) {

	if server == nil {
		panic("bahamut server must not be nil")
	}

	if err := server.RegisterProcessor(processor, identity); err != nil {
		panic(fmt.Sprintf("cannot register processor: %s", err))
	}
}

// RegisterCustomHandlerOrDie will register a handler for a given
// path. This is just a helper for the Server.Register
// Identity and will exit in case of errors. This is just a helper for
// Server.RegisterCustomRouteHandler function.
func RegisterCustomHandlerOrDie(server Server, handler http.HandlerFunc, path string) {

	if server == nil {
		panic("bahamut server must not be nil")
	}

	if err := server.RegisterCustomRouteHandler(path, handler); err != nil {
		panic(fmt.Sprintf("cannot register processor: %s", err))
	}
}

// InstallSIGINTHandler installs signal handlers for graceful shutdown.
func InstallSIGINTHandler(cancelFunc context.CancelFunc) {

	signalCh := make(chan os.Signal, 1)
	signal.Reset(syscall.SIGINT, syscall.SIGTERM)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-signalCh
		cancelFunc()
		signal.Stop(signalCh)
		close(signalCh)
	}()
}

type server struct {
	multiplexer          *bone.Mux
	processors           map[string]Processor
	customRoutesHandlers map[string]http.HandlerFunc
	cfg                  config
	restServer           *restServer
	pushServer           *pushServer
	healthServer         *healthServer
	profilingServer      *profilingServer
}

// New returns a new bahamut Server configured with
// the given options.
func New(options ...Option) Server {

	c := config{}
	for _, opt := range options {
		opt(&c)
	}

	if !c.restServer.enabled && !c.pushServer.enabled && !c.profilingServer.enabled && !c.healthServer.enabled {
		zap.L().Warn("No server configured. Enable some servers through options")
	}

	if c.pushServer.enabled && (!c.pushServer.dispatchEnabled && !c.pushServer.publishEnabled) {
		zap.L().Warn("Push server is enabled but neither dispatching or publishing is. Use bahamut.OptPushPublishHandler() and/or bahamut.OptPushDispatchHandler()")
	}

	if (c.restServer.enabled || c.pushServer.enabled) && len(c.model.modelManagers) == 0 {
		zap.L().Warn("No elemental.ModelManager is defined. Use bahamut.OptModel()")
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
		multiplexer:          mux,
		processors:           make(map[string]Processor),
		customRoutesHandlers: make(map[string]http.HandlerFunc),
		cfg:                  cfg,
	}

	if cfg.restServer.enabled {
		srv.restServer = newRestServer(cfg, mux, srv.ProcessorForIdentity, srv.CustomHandlers, srv.Push)
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

func (b *server) RegisterCustomRouteHandler(path string, handler http.HandlerFunc) error {

	if !(b.restServer != nil &&
		b.cfg.restServer.apiPrefix != "" &&
		b.cfg.restServer.customRoutePrefix != "" &&
		b.cfg.restServer.apiPrefix != b.cfg.restServer.customRoutePrefix) {
		return fmt.Errorf(
			"API root path '%s' and custom handler path '%s' must not overlap",
			b.cfg.restServer.apiPrefix,
			b.cfg.restServer.customRoutePrefix,
		)
	}

	if _, ok := b.customRoutesHandlers[path]; ok {
		return fmt.Errorf("path %s has a registered handler already", path)
	}

	b.customRoutesHandlers[path] = handler

	return nil
}

func (b *server) UnregisterCustomRouteHandler(path string) error {

	if _, ok := b.customRoutesHandlers[path]; !ok {
		return fmt.Errorf("path %s has no existing handler", path)
	}

	delete(b.customRoutesHandlers, path)

	return nil
}

func (b *server) ProcessorForIdentity(identity elemental.Identity) (Processor, error) {

	if _, ok := b.processors[identity.Name]; !ok {
		return nil, fmt.Errorf("no registered processor for identity %s", identity)
	}

	return b.processors[identity.Name], nil
}

func (b *server) CustomHandlers() map[string]http.HandlerFunc {

	return b.customRoutesHandlers
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

func (b *server) RoutesInfo() map[int][]RouteInfo {

	return buildVersionedRoutes(b.cfg.model.modelManagers, b.ProcessorForIdentity)
}

func (b *server) VersionsInfo() map[string]interface{} {

	return b.cfg.meta.version
}

func (b *server) PushEndpoint() string {

	if !b.cfg.pushServer.enabled || !b.cfg.pushServer.dispatchEnabled {
		return ""
	}

	if b.cfg.pushServer.endpoint == "" {
		return "/events"
	}

	return b.cfg.pushServer.endpoint
}

func (b *server) Run(ctx context.Context) {

	if b.profilingServer != nil {
		go b.profilingServer.start(ctx)
	}

	if b.restServer != nil {
		go b.restServer.start(ctx, b.RoutesInfo())
	}

	if b.pushServer != nil {
		go b.pushServer.start(ctx)
	}

	if b.healthServer != nil {
		go b.healthServer.start(ctx)
	}

	if hook := b.cfg.hooks.postStart; hook != nil {
		if err := hook(b); err != nil {
			zap.L().Fatal("Unable to execute bahamut postStart hook", zap.Error(err))
		}
	}

	<-ctx.Done()

	if hook := b.cfg.hooks.preStop; hook != nil {
		if err := hook(b); err != nil {
			zap.L().Error("Unable to execute bahamut preStop hook", zap.Error(err))
		}
	}

	// Stop the health server first so we become unhealthy.
	if b.healthServer != nil {
		<-b.healthServer.stop().Done()
	}

	// Stop the push server to disconnect everybody.
	if b.pushServer != nil {
		b.pushServer.stop()
	}

	// Stop the restserver and wait for current requests to complete.
	if b.restServer != nil {
		<-b.restServer.stop().Done()
	}

	// Stop the profiling server.
	if b.profilingServer != nil {
		b.profilingServer.stop()
	}
}
