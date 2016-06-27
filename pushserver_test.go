// Author: Antoine Mercadal
// See LICENSE file for full LICENSE
// Copyright 2016 Aporeto.

package bahamut

import (
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"golang.org/x/net/websocket"

	"github.com/aporeto-inc/elemental"
	"github.com/go-zoo/bone"

	. "github.com/smartystreets/goconvey/convey"
)

func TestPushServer_newPushServer(t *testing.T) {

	Convey("Given I create a new EventServer", t, func() {

		srv := newPushServer("fake", bone.New(), nil)

		Convey("Then address should be set", func() {
			So(srv.address, ShouldEqual, "fake")
		})

		Convey("Then sessions should be initialized", func() {
			So(len(srv.sessions), ShouldEqual, 0)
		})

		Convey("Then register channel should be initialized", func() {
			var m chan *pushSession
			So(srv.register, ShouldHaveSameTypeAs, m)
		})

		Convey("Then unregister channel should be initialized", func() {
			var m chan *pushSession
			So(srv.unregister, ShouldHaveSameTypeAs, m)
		})

		Convey("Then stop channel should be initialized", func() {
			var m chan bool
			So(srv.close, ShouldHaveSameTypeAs, m)
		})

		Convey("Then events channel should be initialized", func() {
			var m chan *elemental.Event
			So(srv.events, ShouldHaveSameTypeAs, m)
		})
	})
}

//
func TestSession_registerSession(t *testing.T) {

	ts := httptest.NewServer(websocket.Handler(func(ws *websocket.Conn) {
		var d []byte
		websocket.Message.Receive(ws, &d)
		ws.Write(d)
	}))
	defer ts.Close()

	Convey("Given I have an EventServer and no registered session", t, func() {

		ws, _ := websocket.Dial("ws"+ts.URL[4:], "", ts.URL)
		defer ws.Close()

		srv := newPushServer("fake", bone.New(), nil)
		session := newSession(ws, srv)

		go srv.start()
		defer srv.stop()

		Convey("When I register a session", func() {

			srv.registerSession(session)
			time.Sleep(300 * time.Millisecond)

			srv.registerSession(session)
			time.Sleep(300 * time.Millisecond)

			Convey("Then there should be 1 registered session", func() {
				So(len(srv.sessions), ShouldEqual, 1)
			})
		})
	})

	Convey("Given I have an EventServer and a registered session", t, func() {

		ws, _ := websocket.Dial("ws"+ts.URL[4:], "", ts.URL)
		defer ws.Close()

		srv := newPushServer("fake", bone.New(), nil)
		session := newSession(ws, srv)

		go srv.start()
		defer srv.stop()

		srv.registerSession(session)
		time.Sleep(300 * time.Millisecond)

		Convey("When I unregister a registered session", func() {

			srv.unregisterSession(session)
			time.Sleep(300 * time.Millisecond)

			srv.unregisterSession(session)
			time.Sleep(300 * time.Millisecond)

			Convey("Then there should be 0 registered session", func() {
				So(len(srv.sessions), ShouldEqual, 0)
			})
		})
	})
}

func TestSession_startStop(t *testing.T) {

	ts := httptest.NewServer(websocket.Handler(func(ws *websocket.Conn) {
		var d []byte
		websocket.Message.Receive(ws, &d)
		ws.Write(d)
	}))
	defer ts.Close()

	Convey("Given I have a started EventServer with a session", t, func() {

		ws, _ := websocket.Dial("ws"+ts.URL[4:], "", ts.URL)
		defer ws.Close()

		srv := newPushServer("fake", bone.New(), nil)
		session := newSession(ws, srv)

		go session.listen()

		var wg sync.WaitGroup

		startFunct := func() {
			srv.start()
			wg.Done()
		}

		wg.Add(1)
		go startFunct()

		Convey("When I stop it", func() {

			srv.registerSession(session)
			time.Sleep(300 * time.Millisecond)

			srv.stop()
			wg.Wait()

			Convey("Then the len of sessions should be 0", func() {
				So(len(srv.sessions), ShouldEqual, 0)
			})
		})
	})
}

func TestSession_PushEvents(t *testing.T) {

	Convey("Given I create a new EventServer with no kafka info", t, func() {

		srv := newPushServer("fake", bone.New(), nil)

		Convey("When I push an event", func() {

			inEvent := elemental.NewEvent(elemental.EventCreate, NewList())
			go func() { srv.pushEvents(inEvent) }()

			var outEvent *elemental.Event
			select {
			case outEvent = <-srv.events:
				break
			case <-time.After(300 * time.Millisecond):
				break
			}

			Convey("Then the event should be sent throught the local channel", func() {
				So(outEvent, ShouldEqual, inEvent)
			})
		})
	})
}

func TestSession_LocalEvents(t *testing.T) {

	ts := httptest.NewServer(websocket.Handler(func(ws *websocket.Conn) {
		var d []byte
		websocket.Message.Receive(ws, &d)
		websocket.Message.Send(ws, d)
	}))
	defer ts.Close()

	Convey("Given I have a started EventServer a session", t, func() {

		ws1, _ := websocket.Dial("ws"+ts.URL[4:], "", ts.URL)
		defer ws1.Close()

		srv := newPushServer("fake", bone.New(), nil)
		session1 := newSession(ws1, srv)

		go srv.start()
		srv.registerSession(session1)

		Convey("When push an event", func() {

			srv.pushEvents(elemental.NewEvent(elemental.EventCreate, NewList()))

			var evt string
			select {
			case evt = <-session1.events:
				break
			case <-time.After(3 * time.Millisecond):
				break
			}

			Convey("Then output event should be correct", func() {
				So(evt, ShouldNotBeEmpty)
			})
		})

		Convey("When push an event with an UnmarshalableList", func() {

			srv.pushEvents(elemental.NewEvent(elemental.EventCreate, NewUnmarshalableList()))

			var evt string
			select {
			case evt = <-session1.events:
				break
			case <-time.After(3 * time.Millisecond):
				break
			}

			Convey("Then output event should be correct", func() {
				So(evt, ShouldBeEmpty)
			})
		})
	})
}
