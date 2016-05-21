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

		srv := newPushServer("fake", bone.New())

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
			So(srv.stop, ShouldHaveSameTypeAs, m)
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

		srv := newPushServer("fake", bone.New())
		session := newSession(ws, srv)

		go srv.start()
		defer srv.Stop()

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

		srv := newPushServer("fake", bone.New())
		session := newSession(ws, srv)

		go srv.start()
		defer srv.Stop()

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

		srv := newPushServer("fake", bone.New())
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

			srv.Stop()
			wg.Wait()

			Convey("Then the len of sessions should be 0", func() {
				So(len(srv.sessions), ShouldEqual, 0)
			})
		})
	})
}

// func TestSession_Events(t *testing.T) {
//
// 	ts := httptest.NewServer(websocket.Handler(func(ws *websocket.Conn) {
// 		var d []byte
// 		websocket.Message.Receive(ws, &d)
// 		ws.Write(d)
// 		time.Sleep(5000 * time.Millisecond)
// 	}))
// 	defer ts.Close()
//
// 	Convey("Given I have a started EventServer an 3 sessions", t, func() {
//
// 		ws1, _ := websocket.Dial("ws"+ts.URL[4:], "", ts.URL)
// 		ws2, _ := websocket.Dial("ws"+ts.URL[4:], "", ts.URL)
// 		ws3, _ := websocket.Dial("ws"+ts.URL[4:], "", ts.URL)
// 		defer ws1.Close()
// 		defer ws2.Close()
// 		defer ws3.Close()
//
// 		srv := newPushServer("fake", bone.New())
// 		session1 := newSession(ws1, srv)
// 		session2 := newSession(ws2, srv)
// 		session3 := newSession(ws3, srv)
//
// 		go srv.start()
// 		go session1.listen()
// 		go session2.listen()
// 		go session3.listen()
//
// 		srv.registerSession(session1)
// 		srv.registerSession(session2)
// 		srv.registerSession(session3)
//
// 		time.Sleep(300 * time.Millisecond)
//
// 		Convey("When push an event", func() {
//
// 			srv.pushEvents(elemental.NewEvent(elemental.EventCreate, NewList()))
//
// 			var output1, output2, output3 string
// 			websocket.Message.Receive(session1.socket, &output1)
// 			websocket.Message.Receive(session2.socket, &output2)
// 			websocket.Message.Receive(session3.socket, &output3)
//
// 			Convey("Then output1 should be correct", func() {
// 				So(output1, ShouldNotBeEmpty)
// 			})
//
// 			Convey("Then output1 should be the same than output2", func() {
// 				So(output1, ShouldEqual, output2)
// 			})
//
// 			Convey("Then output1 should be the same than output3", func() {
// 				So(output1, ShouldEqual, output3)
// 			})
//
// 		})
// 	})
// }
