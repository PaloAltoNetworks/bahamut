// Author: Antoine Mercadal
// See LICENSE file for full LICENSE
// Copyright 2016 Aporeto.

package bahamut

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"go.aporeto.io/elemental"
	"go.aporeto.io/elemental/test/model"
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
			So(len(logs), ShouldEqual, 2)
			So(logs[0].Message, ShouldEqual, "No rest server or push server configured. Use bahamut.OptRestServer() and/or bahamaut.OptPushServer()")
			So(logs[1].Message, ShouldEqual, "No elemental.ModelManager is defined. Use bahamut.OptModel()")
		})
	})

	Convey("Given I create a bahamut with a push server, but no dispatch and publish option", t, func() {

		zc, obs := observer.New(zapcore.WarnLevel)
		zap.ReplaceGlobals(zap.New(zc))

		New(OptRestServer(":123"), OptPushServer(NewLocalPubSubClient(nil), "coucou"), OptModel(map[int]elemental.ModelManager{0: testmodel.Manager()}))

		Convey("Then some warnings should be printed", func() {
			logs := obs.AllUntimed()
			So(len(logs), ShouldEqual, 1)
			So(logs[0].Message, ShouldEqual, "Push server is enabled but neither dispatching or publishing is. Use bahamut.OptPushPublishHandler() and/or bahamut.OptPushDispatchHandler()")
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

		b := NewServer(cfg)

		Convey("Then apiServer should not be nil", func() {
			So(b.(*server).restServer, ShouldNotBeNil)
		})

		Convey("Then pushServer should not be nil", func() {
			So(b.(*server).pushServer, ShouldNotBeNil)
		})

		Convey("Then number of routes should be 1", func() {
			So(len(b.(*server).multiplexer.Routes), ShouldEqual, 1)
			So(b.(*server).multiplexer.Routes["GET"][0].Path, ShouldEqual, "/events")
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
