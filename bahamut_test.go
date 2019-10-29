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
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"go.aporeto.io/elemental"
	testmodel "go.aporeto.io/elemental/test/model"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func TestBahamut_New(t *testing.T) {

	Convey("Given I create a bahamut with no options", t, func() {

		zc, obs := observer.New(zapcore.WarnLevel)
		zap.ReplaceGlobals(zap.New(zc))

		New()

		Convey("Then some warnings should be printed", func() {
			logs := obs.AllUntimed()
			So(len(logs), ShouldEqual, 1)
			So(logs[0].Message, ShouldEqual, "No server configured. Enable some servers through options")
		})
	})

	Convey("Given I create a bahamut with a push server, but no dispatch and publish option", t, func() {

		zc, obs := observer.New(zapcore.WarnLevel)
		zap.ReplaceGlobals(zap.New(zc))

		New(OptRestServer(":123"), OptPushServer(NewLocalPubSubClient(), "coucou"), OptModel(map[int]elemental.ModelManager{0: testmodel.Manager()}))

		Convey("Then some warnings should be printed", func() {
			logs := obs.AllUntimed()
			So(len(logs), ShouldEqual, 1)
			So(logs[0].Message, ShouldEqual, "Push server is enabled but neither dispatching or publishing is. Use bahamut.OptPushPublishHandler() and/or bahamut.OptPushDispatchHandler()")
		})
	})

	Convey("Given I create a bahamut with a rest server, but no model manager", t, func() {

		zc, obs := observer.New(zapcore.WarnLevel)
		zap.ReplaceGlobals(zap.New(zc))

		New(OptRestServer(":123"))

		Convey("Then some warnings should be printed", func() {
			logs := obs.AllUntimed()
			So(len(logs), ShouldEqual, 1)
			So(logs[0].Message, ShouldEqual, "No elemental.ModelManager is defined. Use bahamut.OptModel()")
		})
	})
}

func TestBahamut_NewBahamut(t *testing.T) {

	Convey("Given I create a new Bahamut with no server", t, func() {

		cfg := config{}

		b := NewServer(cfg)

		Convey("Then apiServer should be nil", func() {
			So(b.(*server).restServer, ShouldBeNil)
		})

		Convey("Then pushServer should be nil", func() {
			So(b.(*server).pushServer, ShouldBeNil)
		})

		Convey("Then number of routes should be 0", func() {
			So(len(b.(*server).multiplexer.Routes), ShouldEqual, 0)
		})

		Convey("Then pushing an event should not panic", func() {
			So(func() { b.Push(elemental.NewEvent(elemental.EventCreate, testmodel.NewList())) }, ShouldNotPanic)
		})
	})

	Convey("Given I create a new Bahamut with all servers", t, func() {

		cfg := config{}
		cfg.pushServer.enabled = true
		cfg.pushServer.dispatchEnabled = true
		cfg.restServer.enabled = true
		cfg.healthServer.enabled = true
		cfg.profilingServer.enabled = true

		b := NewServer(cfg)

		Convey("Then apiServer should not be nil", func() {
			So(b.(*server).restServer, ShouldNotBeNil)
		})

		Convey("Then pushServer should not be nil", func() {
			So(b.(*server).pushServer, ShouldNotBeNil)
		})

		Convey("Then healthServer should not be nil", func() {
			So(b.(*server).healthServer, ShouldNotBeNil)
		})

		Convey("Then profilingServer should not be nil", func() {
			So(b.(*server).profilingServer, ShouldNotBeNil)
		})

		Convey("Then number of routes should be 1", func() {
			So(len(b.(*server).multiplexer.Routes), ShouldEqual, 1)
			So(b.(*server).multiplexer.Routes["GET"][0].Path, ShouldEqual, "/events")
		})
	})
}

func TestBahamut_RouteInfos(t *testing.T) {

	Convey("Given I have a bahamut server loaded with test model", t, func() {

		cfg := config{}
		cfg.model.modelManagers = map[int]elemental.ModelManager{
			0: testmodel.Manager(),
		}

		b := NewServer(cfg)

		Convey("When I call RoutesInfo", func() {

			ri := b.RoutesInfo()

			// full test of content in buildVersionedRoutes test
			Convey("Then ri should not be nil", func() {
				So(ri, ShouldNotBeNil)
			})
		})
	})
}

func TestBahamut_VersionsInfo(t *testing.T) {

	Convey("Given I have a bahamut server", t, func() {

		cfg := config{}
		cfg.meta.version = map[string]interface{}{}

		b := NewServer(cfg)

		Convey("When I call VersionsInfo", func() {

			vi := b.VersionsInfo()

			Convey("Then it should return the version info", func() {
				So(vi, ShouldEqual, cfg.meta.version)
			})
		})
	})
}

func TestBahamut_PushEndpoint(t *testing.T) {

	Convey("Given I have a bahamut server with no push enabled", t, func() {

		cfg := config{}

		b := NewServer(cfg)

		Convey("When I call PushEndpoint", func() {

			pe := b.PushEndpoint()

			Convey("Then it should be correct", func() {
				So(pe, ShouldEqual, "")
			})
		})
	})

	Convey("Given I have a bahamut server with push enabled but no dispatcher", t, func() {

		cfg := config{}
		cfg.pushServer.enabled = true
		cfg.pushServer.dispatchEnabled = false

		b := NewServer(cfg)

		Convey("When I call PushEndpoint", func() {

			pe := b.PushEndpoint()

			Convey("Then it should be correct", func() {
				So(pe, ShouldEqual, "")
			})
		})
	})

	Convey("Given I have a bahamut server with push with no custom endpoint", t, func() {

		cfg := config{}
		cfg.pushServer.enabled = true
		cfg.pushServer.dispatchEnabled = true

		b := NewServer(cfg)

		Convey("When I call PushEndpoint", func() {

			pe := b.PushEndpoint()

			Convey("Then it should be correct", func() {
				So(pe, ShouldEqual, "/events")
			})
		})
	})

	Convey("Given I have a bahamut server with push with custom endpoint", t, func() {

		cfg := config{}
		cfg.pushServer.enabled = true
		cfg.pushServer.dispatchEnabled = true
		cfg.pushServer.endpoint = "/custom"

		b := NewServer(cfg)

		Convey("When I call PushEndpoint", func() {

			pe := b.PushEndpoint()

			Convey("Then it should be correct", func() {
				So(pe, ShouldEqual, "/custom")
			})
		})
	})
}

func TestBahamut_ProcessorRegistration(t *testing.T) {

	Convey("Given I create a Bahamut, aProcessor and an Identity", t, func() {

		p := &mockEmptyProcessor{}
		ident := elemental.MakeIdentity("identity", "random")
		b := NewServer(config{})

		Convey("When I register it for an identity", func() {

			err := b.RegisterProcessor(p, ident)

			Convey("Then err should be nil", func() {
				So(err, ShouldBeNil)
			})

			Convey("Then it should be registered", func() {
				processor, err := b.ProcessorForIdentity(ident)
				So(processor, ShouldEqual, p)
				So(err, ShouldBeNil)
			})

			Convey("Then the number of registered processors should be 1", func() {
				So(b.ProcessorsCount(), ShouldEqual, 1)
			})
		})

		Convey("When I register it twie", func() {

			_ = b.RegisterProcessor(p, ident)
			err := b.RegisterProcessor(p, ident)

			Convey("Then err should not be nil", func() {
				So(err, ShouldNotBeNil)
			})

			Convey("Then the number of registered processors should be 1", func() {
				So(b.ProcessorsCount(), ShouldEqual, 1)
			})
		})

		Convey("When I unregister it", func() {

			_ = b.RegisterProcessor(p, ident)
			err := b.UnregisterProcessor(ident)

			Convey("Then err should be nil", func() {
				So(err, ShouldBeNil)
			})

			Convey("Then it should be unregistered", func() {
				processor, err := b.ProcessorForIdentity(ident)
				So(processor, ShouldBeNil)
				So(err, ShouldNotBeNil)
			})

			Convey("Then the number of registered processors should be 0", func() {
				So(b.ProcessorsCount(), ShouldEqual, 0)
			})
		})

		Convey("When I unregister it twice", func() {

			_ = b.UnregisterProcessor(ident)
			err := b.UnregisterProcessor(ident)

			Convey("Then err should not be nil", func() {
				So(err, ShouldNotBeNil)
			})
		})
	})
}

func TestBahamyt_RegistorProcessorOrDie(t *testing.T) {

	Convey("Given I have no bahamut server ", t, func() {

		Convey("When I call RegisterProcessorOrDie", func() {

			Convey("Then it should panic", func() {
				So(func() { RegisterProcessorOrDie(nil, &mockEmptyProcessor{}, testmodel.ListIdentity) }, ShouldPanicWith, "bahamut server must not be nil")
			})
		})
	})

	Convey("Given I have a bahamut server ", t, func() {

		cfg := config{}
		b := NewServer(cfg)

		Convey("When I call RegisterProcessorOrDie", func() {

			Convey("Then it should panic", func() {
				So(func() { RegisterProcessorOrDie(b, &mockEmptyProcessor{}, testmodel.ListIdentity) }, ShouldNotPanic)
				rp, _ := b.ProcessorForIdentity(testmodel.ListIdentity)
				So(rp, ShouldNotBeNil)
			})
		})
	})

	Convey("Given I have a bahamut server ", t, func() {

		cfg := config{}
		b := NewServer(cfg)

		Convey("When I call RegisterProcessorOrDie on the same identity twice", func() {

			Convey("Then it should panic", func() {
				So(func() { RegisterProcessorOrDie(b, &mockEmptyProcessor{}, testmodel.ListIdentity) }, ShouldNotPanic)
				So(func() { RegisterProcessorOrDie(b, &mockEmptyProcessor{}, testmodel.ListIdentity) }, ShouldPanicWith, "cannot register processor: identity <Identity list|lists> already has a registered processor")
			})
		})
	})
}
