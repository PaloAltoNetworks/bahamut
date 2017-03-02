// Author: Antoine Mercadal
// See LICENSE file for full LICENSE
// Copyright 2016 Aporeto.

package bahamut

import (
	"fmt"
	"testing"

	"github.com/aporeto-inc/elemental"
	. "github.com/smartystreets/goconvey/convey"
)

type FakeProcessor struct {
}

type Auth struct {
	authenticated bool
	authorized    bool
	errored       bool
}

func (a *Auth) AuthenticateRequest(req *elemental.Request, ch elemental.ClaimsHolder) (bool, error) {

	if a.errored {
		return false, fmt.Errorf("this is an %s", "error")
	}

	return a.authenticated, nil
}

func (a *Auth) IsAuthorized(ctx *Context) (bool, error) {

	if a.errored {
		return false, fmt.Errorf("this is an %s", "error")
	}

	return a.authorized, nil
}

type testSessionHandler struct {
	sessionCount int
	shouldCalls  int
	block        bool
}

func (h *testSessionHandler) OnPushSessionStart(session *Session)            { h.sessionCount++ }
func (h *testSessionHandler) OnPushSessionStop(session *Session)             { h.sessionCount-- }
func (h *testSessionHandler) IsAuthenticated(session *Session) (bool, error) { return true, nil }
func (h *testSessionHandler) ShouldPush(session *Session, event *elemental.Event) (bool, error) {
	h.shouldCalls++
	return !h.block, nil
}

func TestBahamut_NewBahamut(t *testing.T) {

	Convey("Given I create a new Bahamut with no server", t, func() {

		cfg := Config{}
		cfg.ReSTServer.Disabled = true
		cfg.WebSocketServer.Disabled = true

		b := NewServer(cfg)

		Convey("Then apiServer should be nil", func() {
			So(b.(*server).apiServer, ShouldBeNil)
		})

		Convey("Then pushServer should be nil", func() {
			So(b.(*server).pushServer, ShouldBeNil)
		})

		Convey("Then number of routes should be 0", func() {
			So(len(b.(*server).multiplexer.Routes), ShouldEqual, 0)
		})

		Convey("Then pushing an event should not panic", func() {
			So(func() { b.Push(elemental.NewEvent(elemental.EventCreate, NewList())) }, ShouldNotPanic)
		})
	})

	Convey("Given I create a new Bahamut with all servers", t, func() {

		cfg := Config{}

		b := NewServer(cfg)

		Convey("Then apiServer should not be nil", func() {
			So(b.(*server).apiServer, ShouldNotBeNil)
		})

		Convey("Then pushServer should be nil", func() {
			So(b.(*server).pushServer, ShouldNotBeNil)
		})

		Convey("Then number of routes should be 0", func() {
			So(len(b.(*server).multiplexer.Routes), ShouldEqual, 7)
		})
	})
}

func TestBahamut_ProcessorRegistration(t *testing.T) {

	Convey("Given I create a Bahamut, aProcessor and an Identity", t, func() {

		p := &FakeProcessor{}
		ident := elemental.MakeIdentity("identity", "random")
		b := NewServer(Config{})

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

			b.RegisterProcessor(p, ident)
			err := b.RegisterProcessor(p, ident)

			Convey("Then err should not be nil", func() {
				So(err, ShouldNotBeNil)
			})

			Convey("Then the number of registered processors should be 1", func() {
				So(b.ProcessorsCount(), ShouldEqual, 1)
			})
		})

		Convey("When I unregister it", func() {

			b.RegisterProcessor(p, ident)
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

			b.UnregisterProcessor(ident)
			err := b.UnregisterProcessor(ident)

			Convey("Then err should not be nil", func() {
				So(err, ShouldNotBeNil)
			})
		})
	})
}
