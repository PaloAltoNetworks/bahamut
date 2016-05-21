// Author: Antoine Mercadal
// See LICENSE file for full LICENSE
// Copyright 2016 Aporeto.

package bahamut

import (
	"testing"

	"github.com/aporeto-inc/elemental"
	. "github.com/smartystreets/goconvey/convey"
)

type FakeProcessor struct {
}

func TestBahamut_NewBahamut(t *testing.T) {

	Convey("Given I create a new Bahamut with no server", t, func() {

		b := NewBahamut("fake", []*Route{}, false, false, false)

		Convey("Then apiServer should be nil", func() {
			So(b.apiServer, ShouldBeNil)
		})

		Convey("Then pushServer should be nil", func() {
			So(b.pushServer, ShouldBeNil)
		})

		Convey("Then number of routes should be 0", func() {
			So(len(b.multiplexer.Routes), ShouldEqual, 0)
		})

		Convey("Then pushing an event should panic", func() {
			So(func() { b.Push(elemental.NewEvent(elemental.EventCreate, NewList())) }, ShouldPanic)
		})
	})

	Convey("Given I create a new Bahamut with all servers", t, func() {

		b := NewBahamut("fake", []*Route{}, true, true, true)

		Convey("Then apiServer should not be nil", func() {
			So(b.apiServer, ShouldNotBeNil)
		})

		Convey("Then pushServer should be nil", func() {
			So(b.pushServer, ShouldNotBeNil)
		})

		Convey("Then number of routes should be 0", func() {
			So(len(b.multiplexer.Routes), ShouldEqual, 8)
		})
	})
}

func TestBahamut_DefaultBahamut(t *testing.T) {

	Convey("Given I create a Bahamut", t, func() {

		b := NewBahamut("fake", []*Route{}, false, false, false)

		Convey("Then the defaultBahamut should be set", func() {
			So(DefaultBahamut(), ShouldEqual, b)
		})
	})
}

func TestBahamut_ProcessorRegistration(t *testing.T) {

	Convey("Given I create a Bahamut, aProcessor and an Identity", t, func() {

		p := &FakeProcessor{}
		ident := elemental.MakeIdentity("identity", "random")
		b := NewBahamut("fake", []*Route{}, false, false, false)

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
		})

		Convey("When I register it twie", func() {

			b.RegisterProcessor(p, ident)
			err := b.RegisterProcessor(p, ident)

			Convey("Then err should not be nil", func() {
				So(err, ShouldNotBeNil)
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

func TestBahamut_Authenticator(t *testing.T) {

	Convey("Given I create a new Bahamut", t, func() {

		b := NewBahamut("fake", []*Route{}, false, false, false)
		auth := &Auth{}

		Convey("When I access an Authenticator while there is none", func() {

			a, err := b.Authenticator()

			Convey("Then the authenticator should be set", func() {
				So(a, ShouldBeNil)
				So(err, ShouldNotBeNil)
			})
		})

		Convey("When I set an Authenticator", func() {

			b.SetAuthenticator(auth)
			a, err := b.Authenticator()

			Convey("Then the authenticator should be set", func() {
				So(a, ShouldNotBeNil)
				So(err, ShouldBeNil)
			})
		})
	})
}
