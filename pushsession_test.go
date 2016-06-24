// Author: Antoine Mercadal
// See LICENSE file for full LICENSE
// Copyright 2016 Aporeto.

package bahamut

import (
	"testing"

	"github.com/go-zoo/bone"
	. "github.com/smartystreets/goconvey/convey"
	"golang.org/x/net/websocket"
)

func TestSession_newSession(t *testing.T) {

	Convey("Given I create have a new pushServer", t, func() {

		Convey("When I create a new session", func() {

			ws := &websocket.Conn{}
			session := newSession(ws, newPushServer("fake", bone.New(), nil))

			Convey("Then the session id should not be empty", func() {
				So(session.id, ShouldNotBeEmpty)
			})

			Convey("Then the socket should be nil", func() {
				So(session.socket, ShouldEqual, ws)
			})

			Convey("Then the events channel should be a chan of bytes", func() {
				So(session.events, ShouldHaveSameTypeAs, make(chan string))
			})
		})
	})
}

// func TestSession_write(t *testing.T) {
//
// 	ts := httptest.NewServer(websocket.Handler(func(ws *websocket.Conn) {
// 		var d []byte
// 		websocket.Message.Receive(ws, &d)
// 		ws.Write(d)
// 		go websocket.Message.Receive(ws, &d)
// 	}))
// 	defer ts.Close()
//
// 	Convey("Given I have an open socket", t, func() {
//
// 		ws, _ := websocket.Dial("ws"+ts.URL[4:], "", ts.URL)
// 		defer ws.Close()
//
// 		session := newSession(ws, nil)
//
// 		Convey("When I write some data on a working socket", func() {
//
// 			go session.listen()
// 			defer func() { session.close <- true }()
//
// 			session.events <- elemental.NewEvent(elemental.EventCreate, NewList())
//
// 			var output string
// 			websocket.Message.Receive(session.socket, &output)
//
// 			Convey("Then msg should not be empty", func() {
// 				So(output, ShouldNotBeEmpty)
// 			})
// 		})
// 	})
// }
