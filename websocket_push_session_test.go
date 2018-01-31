package bahamut

import (
	"errors"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/aporeto-inc/elemental/test/model"

	"github.com/aporeto-inc/elemental"

	. "github.com/smartystreets/goconvey/convey"
)

func TestWSPushSession_newPushSession(t *testing.T) {

	Convey("Given call newWSPushSession", t, func() {

		req, _ := http.NewRequest("GET", "bla", nil)
		cfg := Config{}
		s := newWSPushSession(req, cfg, nil)

		Convey("Then it should be correctly initialized", func() {
			So(s.events, ShouldHaveSameTypeAs, make(chan *elemental.Event))
			So(s.filters, ShouldHaveSameTypeAs, make(chan *elemental.PushFilter))
			So(s.currentFilterLock, ShouldNotBeNil)
			So(s, ShouldImplement, (*internalWSSession)(nil))
		})
	})
}

func TestWSPushSession_DirectPush(t *testing.T) {

	Convey("Given I have a session and an event", t, func() {

		req, _ := http.NewRequest("GET", "bla", nil)
		cfg := Config{}
		s := newWSPushSession(req, cfg, nil)

		evt := elemental.NewEvent(elemental.EventCreate, testmodel.NewList())

		Convey("When I call directPush and pull from the event channel", func() {

			go s.DirectPush(evt, evt)
			evt1 := <-s.events
			evt2 := <-s.events

			Convey("Then evt1 should be correct", func() {
				So(evt1, ShouldEqual, evt)
			})
			Convey("Then evt2 should be correct", func() {
				So(evt2, ShouldEqual, evt)
			})
		})
	})
}

func TestWSPushSession_String(t *testing.T) {

	Convey("Given I have a session", t, func() {

		req, _ := http.NewRequest("GET", "bla", nil)
		cfg := Config{}
		s := newWSPushSession(req, cfg, nil)

		Convey("When I call String", func() {

			str := s.String()

			Convey("Then the string representation should be correct", func() {
				So(str, ShouldEqual, fmt.Sprintf("<pushsession id:%s parameters:%v>", s.Identifier(), s.parameters))
			})
		})
	})
}

func TestWSPushSession_Filtering(t *testing.T) {

	Convey("Given I have a session and a filter", t, func() {

		req, _ := http.NewRequest("GET", "bla", nil)
		cfg := Config{}
		s := newWSPushSession(req, cfg, nil)

		f := elemental.NewPushFilter()

		Convey("When I call setCurrentFilter", func() {

			s.setCurrentFilter(f)

			Convey("Then the filter should be installed", func() {
				So(s.currentFilter(), ShouldNotEqual, f)
				So(s.currentFilter(), ShouldResemble, f)
			})

			Convey("When I reset the filter to nil", func() {

				s.setCurrentFilter(nil)

				Convey("Then the filter should be uninstalled", func() {
					So(s.currentFilter(), ShouldBeNil)
				})
			})
		})
	})
}

func TestWSPushSession_read(t *testing.T) {

	Convey("Given I have a push session", t, func() {

		req, _ := http.NewRequest("GET", "bla", nil)
		cfg := Config{}

		filter := elemental.NewPushFilter()
		filter.FilterIdentity("list", elemental.EventCreate)

		conn := newMockWebsocket()

		calledCounter := &counter{}
		unregister := func(ws internalWSSession) {
			calledCounter.Add(1)
		}

		s := newWSPushSession(req, cfg, unregister)
		s.conn = conn

		stopper := newStopper()
		go func() {
			s.read()
			stopper.Stop()
		}()

		Convey("When it receives a new filter", func() {

			conn.setNextRead(filter)

			var f *elemental.PushFilter
			select {
			case f = <-s.filters:
			case <-time.After(30 * time.Millisecond):
				panic("getting filter took too long")
			}

			Convey("Then read should not be stopped", func() {
				So(stopper.isStopped(), ShouldBeFalse)
			})

			Convey("Then f should not be nil", func() {
				So(f, ShouldNotBeNil)
				So(f.Identities, ShouldResemble, map[string][]elemental.EventType{"list": {elemental.EventType("create")}})
			})
		})

		Convey("When I call read then write an error", func() {

			conn.setNextRead(errors.New("nooo"))

			select {
			case <-stopper.Done():
			case <-time.After(300 * time.Millisecond):
				panic("closing session took too long")
			}

			Convey("Then read should be stopped", func() {
				So(stopper.isStopped(), ShouldBeTrue)
			})

			Convey("Then the session should have called unregister once", func() {
				So(calledCounter.Value(), ShouldEqual, 1)
			})
		})
	})
}

func TestWSPushSession_write(t *testing.T) {

	Convey("Given I have a push session and starts the write routine", t, func() {

		req, _ := http.NewRequest("GET", "bla", nil)
		conn := newMockWebsocket()

		calledCounter := &counter{}
		unregister := func(ws internalWSSession) {
			calledCounter.Add(1)
		}

		s := newWSPushSession(req, Config{}, unregister)
		s.conn = conn

		stopper := newStopper()
		go func() {
			s.write()
			stopper.Stop()
		}()

		Convey("When I receive an event that is not filtered out", func() {

			evt := elemental.NewEvent(elemental.EventCreate, testmodel.NewList())
			s.events <- evt

			Convey("Then the event should have been written in the websocket conn", func() {
				So(<-conn.getLastWrite(), ShouldEqual, evt)
			})

			Convey("Then stopped should be false", func() {
				So(stopper.isStopped(), ShouldBeFalse)
			})
		})

		Convey("When I receive an event that is filtered out ", func() {

			f := elemental.NewPushFilter()
			f.FilterIdentity("not-list")
			s.setCurrentFilter(f)

			s.events <- elemental.NewEvent(elemental.EventCreate, testmodel.NewList())

			var d interface{}
			select {
			case d = <-conn.getLastWrite():
			case <-time.After(30 * time.Millisecond):
			}

			Convey("Then no event should have been written in the websocket conn", func() {
				So(d, ShouldBeNil)
			})

			Convey("Then stopped should be false", func() {
				So(stopper.isStopped(), ShouldBeFalse)
			})
		})

		Convey("When I receive an error from the websocket", func() {

			conn.setWriteErr(errors.New("nnnnnnooooooo"))

			s.events <- elemental.NewEvent(elemental.EventCreate, testmodel.NewList())

			select {
			case <-stopper.Done():
			case <-time.After(300 * time.Millisecond):
				panic("closing session took too long")
			}

			Convey("Then stopped should be true", func() {
				So(stopper.isStopped(), ShouldBeTrue)
			})

			Convey("Then the session should have called unregister once", func() {
				So(calledCounter.Value(), ShouldEqual, 1)
			})
		})

		Convey("When I close the session", func() {

			// Not sure why but calling s.stop() causes a race condition with s.write() at line 195...
			close(s.closeCh)

			select {
			case <-stopper.Done():
			case <-time.After(300 * time.Millisecond):
				panic("closing session took too long")
			}

			Convey("Then stopped should be true", func() {
				So(stopper.isStopped(), ShouldBeTrue)
			})

			// Convey("Then the session should have called unregister once", func() {
			// 	So(calledCounter.Value(), ShouldEqual, 1)
			// })
		})
	})
}
