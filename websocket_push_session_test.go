package bahamut

import (
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"testing"
	"time"

	"github.com/aporeto-inc/elemental/test/model"

	"github.com/aporeto-inc/elemental"

	. "github.com/smartystreets/goconvey/convey"
)

type mockWebsocket struct {
	readErr   error
	writeErr  error
	closeErr  error
	nextRead  interface{}
	lastWrite interface{}
}

func (s *mockWebsocket) ReadJSON(data interface{}) error {

	if s.readErr != nil {
		return s.readErr
	}

	reflect.ValueOf(data).Elem().Set(reflect.ValueOf(s.nextRead))

	return nil
}

func (s *mockWebsocket) WriteJSON(data interface{}) error {
	s.lastWrite = data
	return s.writeErr
}

func (s *mockWebsocket) Close() error {
	return s.closeErr
}

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

		conn := &mockWebsocket{}
		conn.nextRead = filter

		var unregisterCalled int
		unregister := func(ws internalWSSession) {
			unregisterCalled++
		}

		s := newWSPushSession(req, cfg, unregister)
		s.conn = conn

		Convey("When I call read then write a filter", func() {

			var stopped bool
			go func() {
				s.read()
				stopped = true
			}()

			var f *elemental.PushFilter
			select {
			case f = <-s.filters:
			case <-time.After(30 * time.Millisecond):
			}

			Convey("Then read should not be stopped", func() {
				So(stopped, ShouldBeFalse)
			})

			Convey("Then f should not be nil", func() {
				So(f, ShouldNotBeNil)
				So(f.Identities, ShouldResemble, map[string][]elemental.EventType{"list": {elemental.EventType("create")}})
			})
		})

		Convey("When I call read then write an error", func() {

			conn.readErr = errors.New("nooo")

			var stopped bool
			go func() {
				s.read()
				stopped = true
			}()

			<-time.After(30 * time.Millisecond)

			Convey("Then read should be stopped", func() {
				So(stopped, ShouldBeTrue)
			})

			Convey("Then the session should have called unregister once", func() {
				So(unregisterCalled, ShouldEqual, 1)
			})
		})
	})
}

func TestWSPushSession_write(t *testing.T) {

	Convey("Given I have a push session and starts the write routine", t, func() {

		req, _ := http.NewRequest("GET", "bla", nil)
		cfg := Config{}

		conn := &mockWebsocket{}

		var unregisterCalled int
		unregister := func(ws internalWSSession) {
			unregisterCalled++
		}

		s := newWSPushSession(req, cfg, unregister)
		s.conn = conn

		var stopped bool
		go func() {
			s.write()
			stopped = true
		}()

		Convey("When I receive an event", func() {

			evt := elemental.NewEvent(elemental.EventCreate, testmodel.NewList())
			s.events <- evt

			time.Sleep(30 * time.Millisecond)

			Convey("Then the event should have been written in the websocket conn", func() {
				So(conn.lastWrite, ShouldEqual, evt)
			})

			Convey("Then stopped should be false", func() {
				So(stopped, ShouldBeFalse)
			})

			Convey("When I install a filter then write an filtered event ", func() {

				conn.lastWrite = nil

				f := elemental.NewPushFilter()
				f.FilterIdentity("not-list")
				s.setCurrentFilter(f)

				s.events <- elemental.NewEvent(elemental.EventCreate, testmodel.NewList())

				time.Sleep(30 * time.Millisecond)

				Convey("Then no event should have been written in the websocket conn", func() {
					So(conn.lastWrite, ShouldBeNil)
				})

				Convey("Then stopped should be false", func() {
					So(stopped, ShouldBeFalse)
				})

				Convey("When I writing json causes an error", func() {

					conn.lastWrite = nil
					s.setCurrentFilter(nil)
					conn.writeErr = errors.New("nnnnnnooooooo")

					s.events <- elemental.NewEvent(elemental.EventCreate, testmodel.NewList())

					time.Sleep(30 * time.Millisecond)

					Convey("Then stopped should be true", func() {
						So(stopped, ShouldBeTrue)
					})

					Convey("Then the session should have called unregister once", func() {
						So(unregisterCalled, ShouldEqual, 1)
					})
				})

				Convey("When I close the session", func() {

					conn.lastWrite = nil

					s.stop()

					time.Sleep(30 * time.Millisecond)

					Convey("Then stopped should be true", func() {
						So(stopped, ShouldBeTrue)
					})

					Convey("Then the session should have called unregister once", func() {
						So(unregisterCalled, ShouldEqual, 1)
					})
				})
			})
		})
	})
}
