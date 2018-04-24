package bahamut

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/aporeto-inc/addedeffect/wsc"
	"github.com/aporeto-inc/elemental"
	"github.com/aporeto-inc/elemental/test/model"

	opentracing "github.com/opentracing/opentracing-go"
	. "github.com/smartystreets/goconvey/convey"
)

func TestWSPushSession_newPushSession(t *testing.T) {

	Convey("Given call newWSPushSession", t, func() {

		u, _ := url.Parse("http://toto.com?a=b")
		conf := Config{}
		req := &http.Request{
			Header:     http.Header{"h1": {"a"}},
			URL:        u,
			TLS:        &tls.ConnectionState{},
			RemoteAddr: "1.2.3.4",
		}
		unregister := func(i *wsPushSession) {}
		s := newWSPushSession(req, conf, unregister)

		Convey("Then it should be correctly initialized", func() {
			So(s.events, ShouldHaveSameTypeAs, make(chan *elemental.Event))
			So(s.filters, ShouldHaveSameTypeAs, make(chan *elemental.PushFilter))
			So(s.currentFilterLock, ShouldNotBeNil)
			So(s.claims, ShouldResemble, []string{})
			So(s.claimsMap, ShouldResemble, map[string]string{})
			So(s.config, ShouldResemble, conf)
			So(s.headers, ShouldEqual, req.Header)
			So(s.id, ShouldNotBeEmpty)
			So(s.parameters, ShouldResemble, url.Values{"a": {"b"}})
			So(s.closeCh, ShouldHaveSameTypeAs, make(chan struct{}))
			So(s.unregister, ShouldEqual, unregister)
			So(s.ctx, ShouldNotBeNil)
			So(s.cancel, ShouldNotBeNil)
			So(s.tlsConnectionState, ShouldEqual, req.TLS)
			So(s.remoteAddr, ShouldEqual, req.RemoteAddr)
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

func TestWSPushSession_accessors(t *testing.T) {

	Convey("Given create a push session", t, func() {

		u, _ := url.Parse("http://toto.com?a=b&token=token")
		conf := Config{}
		req := &http.Request{
			Header:     http.Header{"h1": {"a"}},
			URL:        u,
			TLS:        &tls.ConnectionState{},
			RemoteAddr: "1.2.3.4",
		}
		span := opentracing.StartSpan("test")
		ctx := opentracing.ContextWithSpan(context.Background(), span)
		req = req.WithContext(ctx)
		unregister := func(i *wsPushSession) {}

		s := newWSPushSession(req, conf, unregister)

		Convey("When I call Identifier()", func() {

			id := s.Identifier()

			Convey("Then id should be correct", func() {
				So(id, ShouldNotBeEmpty)
			})
		})

		Convey("When I call SetClaims()", func() {

			s.SetClaims([]string{"a=a", "b=b"})

			Convey("Then GetClaims() should return the correct claims ", func() {
				So(s.GetClaims(), ShouldResemble, []string{"a=a", "b=b"})
			})

			Convey("Then GetClaimsMap() should return the correct claims ", func() {
				m := s.GetClaimsMap()
				So(len(m), ShouldEqual, 2)
				So(m["a"], ShouldEqual, "a")
				So(m["b"], ShouldEqual, "b")
			})
		})

		Convey("When I call GetToken()", func() {

			token := s.GetToken()

			Convey("Then token should be correct", func() {
				So(token, ShouldEqual, "token")
			})
		})

		Convey("When I call TLSConnectionState()", func() {

			s := s.TLSConnectionState()

			Convey("Then TLSConnectionState should be correct", func() {
				So(s, ShouldEqual, req.TLS)
			})
		})

		Convey("When I call SetMetadata()", func() {

			s.SetMetadata("hi")

			Convey("Then GetMetadata() should return the correct metadata ", func() {
				So(s.GetMetadata(), ShouldResemble, "hi")
			})
		})

		Convey("When I call GetContext()", func() {

			c := s.GetContext()

			Convey("Then GetContext() should return the correct context ", func() {
				So(opentracing.SpanFromContext(c), ShouldResemble, span)
			})
		})

		Convey("When I call GetParameter()", func() {

			p := s.GetParameter("a")

			Convey("Then parameter should be correct", func() {
				So(p, ShouldEqual, "b")
			})
		})

		Convey("When I call setRemoteAddress()", func() {

			s.setRemoteAddress("a.b.c.d")

			Convey("Then address should be correct", func() {
				So(s.remoteAddr, ShouldEqual, "a.b.c.d")
			})
		})

		Convey("When I call setTLSConnectionState()", func() {

			tcs := &tls.ConnectionState{}
			s.setTLSConnectionState(tcs)

			Convey("Then address should be correct", func() {
				So(s.tlsConnectionState, ShouldEqual, tcs)
			})
		})

		Convey("When I call setSocket()", func() {

			ws := wsc.NewMockWebsocket(context.TODO())
			s.setConn(ws)

			Convey("Then ws should be correct", func() {
				So(s.conn, ShouldEqual, ws)
			})
		})
	})
}

func TestWSPushSession_listen(t *testing.T) {

	Convey("Given I have a push session", t, func() {

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		unregistered := make(chan bool, 10)

		s := newWSPushSession(
			(&http.Request{URL: &url.URL{}}).WithContext(ctx),
			Config{},
			func(i *wsPushSession) {
				unregistered <- true
			},
		)

		conn := wsc.NewMockWebsocket(ctx)
		s.setConn(conn)

		testEvent := elemental.NewEvent(elemental.EventUpdate, testmodel.NewList())

		Convey("When I simulate an incoming event that is not filtered out", func() {

			go s.listen()
			s.DirectPush(testEvent)

			var data []byte
			select {
			case data = <-conn.LastWrite():
			case <-ctx.Done():
				panic("test: did not receive data in time")
			}

			Convey("Then the websocket should send the event", func() {
				So(string(data), ShouldStartWith, `{"entity":{"creationOnly":"","date":"0001-01-01T00:00:00Z","description":"","name":"","readOnly":"","slice":null,"ID":"","parentID":"","parentType":""},"identity":"list","type":"update","timestamp":"`)
			})
		})

		Convey("When I simulate an incoming event that is manually filtered out", func() {

			go s.listen()

			f := elemental.NewPushFilter()
			f.FilterIdentity("not-list")
			s.setCurrentFilter(f)

			s.DirectPush(testEvent)

			var data []byte
			select {
			case data = <-conn.LastWrite():
			case <-time.After(800 * time.Millisecond):
			}

			Convey("Then the websocket should not send the event", func() {
				So(data, ShouldBeNil)
			})
		})

		Convey("When I simulate an incoming event that is older than the session", func() {

			go s.listen()

			testEvent.Timestamp = time.Now().Add(-1 * time.Hour)
			s.DirectPush(testEvent)

			var data []byte
			select {
			case data = <-conn.LastWrite():
			case <-time.After(800 * time.Millisecond):
			}

			Convey("Then the websocket should not send the event", func() {
				So(data, ShouldBeNil)
			})
		})

		Convey("When I simulate an incoming event with broken json", func() {

			go s.listen()
			s.DirectPush(testEvent)

			var data []byte
			select {
			case data = <-conn.LastWrite():
			case <-ctx.Done():
				panic("test: did not receive data in time")
			}

			Convey("Then the websocket should send the event", func() {
				So(string(data), ShouldStartWith, `{"entity":{"creationOnly":"","date":"0001-01-01T00:00:00Z","description":"","name":"","readOnly":"","slice":null,"ID":"","parentID":"","parentType":""},"identity":"list","type":"update","timestamp":"`)
			})
		})

		Convey("When I send a valid filter in the websocket", func() {

			go s.listen()

			conn.NextRead([]byte(`{"identities":{"not-list": null}}`))
			<-time.After(300 * time.Millisecond)

			Convey("Then the filter should be correctly set", func() {
				So(s.currentFilter().String(), ShouldEqual, `<pushfilter identities:map[not-list:[]]>`)
			})
		})

		Convey("When I send an invalid filter in the websocket", func() {

			go s.listen()

			conn.NextRead([]byte(`{"identities":{"not`))

			var doneErr error
			select {
			case doneErr = <-conn.Done():
			case <-ctx.Done():
				panic("test: did not receive message in time")
			}

			Convey("Then the filter should be nil", func() {
				So(s.currentFilter(), ShouldBeNil)
				So(doneErr, ShouldNotBeNil)
				So(doneErr.Error(), ShouldEqual, "1003")
			})

			Convey("Then the session should be unregistered", func() {
				var u bool
				select {
				case u = <-unregistered:
				case <-ctx.Done():
					panic("test: did not receive response in time")
				}

				So(u, ShouldBeTrue)
			})
		})

		Convey("When the client closes the websocket", func() {

			go s.listen()

			conn.NextDone(errors.New("bye"))

			Convey("Then the session should be unregistered", func() {
				var u bool
				select {
				case u = <-unregistered:
				case <-ctx.Done():
					panic("test: did not receive response in time")
				}

				So(u, ShouldBeTrue)
			})
		})

		Convey("When the server closes the websocket", func() {

			go s.listen()

			cancel()

			Convey("Then the session should be unregistered", func() {
				var u bool
				select {
				case u = <-unregistered:
				case <-time.After(1 * time.Second):
					panic("test: did not receive response in time")
				}

				So(u, ShouldBeTrue)
			})
		})
	})
}
