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
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/gorilla/websocket"

	"github.com/opentracing/opentracing-go"
	. "github.com/smartystreets/goconvey/convey"
	"go.aporeto.io/elemental"
	testmodel "go.aporeto.io/elemental/test/model"
	"go.aporeto.io/wsc"
)

func TestWSPushSession_newPushSession(t *testing.T) {

	Convey("Given call newWSPushSession", t, func() {

		u, _ := url.Parse("http://toto.com?a=b")
		conf := config{}
		req := &http.Request{
			Header:     http.Header{"Authorization": {"a"}},
			URL:        u,
			TLS:        &tls.ConnectionState{},
			RemoteAddr: "1.2.3.4",
		}
		unregister := func(i *wsPushSession) {}
		s := newWSPushSession(req, conf, unregister, elemental.EncodingTypeMSGPACK, elemental.EncodingTypeMSGPACK)

		Convey("Then it should be correctly initialized", func() {
			So(s.dataCh, ShouldHaveSameTypeAs, make(chan []byte))
			So(s.Claims(), ShouldResemble, []string{})
			So(s.claimsMap, ShouldResemble, map[string]string{})
			So(s.cfg, ShouldResemble, conf)
			So(s.headers, ShouldResemble, http.Header{"Authorization": {"a"}})
			So(s.Header("Authorization"), ShouldEqual, "a")
			So(s.id, ShouldNotBeEmpty)
			So(s.parameters, ShouldResemble, url.Values{"a": {"b"}})
			So(s.Parameter("a"), ShouldEqual, "b")
			So(s.closeCh, ShouldHaveSameTypeAs, make(chan struct{}))
			So(s.unregister, ShouldEqual, unregister)
			So(s.Context(), ShouldNotBeNil)
			So(s.cancel, ShouldNotBeNil)
			So(s.TLSConnectionState(), ShouldEqual, req.TLS)
			So(s.ClientIP(), ShouldEqual, req.RemoteAddr)
		})
	})
}

func TestWSPushSession_DirectPush(t *testing.T) {

	Convey("Given I have a session and an event", t, func() {

		req, _ := http.NewRequest("GET", "bla", nil)
		cfg := config{}
		s := newWSPushSession(req, cfg, nil, elemental.EncodingTypeMSGPACK, elemental.EncodingTypeMSGPACK)

		evt := elemental.NewEvent(elemental.EventCreate, testmodel.NewList())

		msgpack, _, err := prepareEventData(evt)
		if err != nil {
			panic(err)
		}

		Convey("When I call directPush", func() {

			go s.DirectPush(evt, evt)
			data1 := <-s.dataCh
			data2 := <-s.dataCh

			Convey("Then data1 should be correct", func() {
				So(string(data1), ShouldEqual, string(msgpack))
			})
			Convey("Then data2 should be correct", func() {
				So(string(data2), ShouldEqual, string(msgpack))
			})
		})

		Convey("When I call directPush but event is filtered", func() {

			f := elemental.NewPushConfig()
			f.FilterIdentity("not-list")

			s.setCurrentPushConfig(f)
			go s.DirectPush(evt)

			var data []byte
			select {
			case data = <-s.dataCh:
			case <-time.After(1 * time.Second):
			}

			Convey("Then data should be correct", func() {
				So(data, ShouldBeNil)
			})
		})

		Convey("When I call directPush but event is before session", func() {

			s.startTime = time.Now().Add(1 * time.Second)
			go s.DirectPush(evt)

			var data []byte
			select {
			case data = <-s.dataCh:
			case <-time.After(1 * time.Second):
			}

			Convey("Then data should be correct", func() {
				So(data, ShouldBeNil)
			})
		})

		Convey("When I call directPush with a bad event", func() {

			evt.Encoding = elemental.EncodingTypeJSON
			evt.RawData = []byte("{brodken")

			go s.DirectPush(evt)

			var data []byte
			select {
			case data = <-s.dataCh:
			case <-time.After(1 * time.Second):
			}

			Convey("Then data should be correct", func() {
				So(data, ShouldBeNil)
			})
		})
	})
}

func TestWSPushSession_send(t *testing.T) {

	Convey("Given I have a session and an event", t, func() {

		req, _ := http.NewRequest("GET", "bla", nil)
		cfg := config{}
		s := newWSPushSession(req, cfg, nil, elemental.EncodingTypeMSGPACK, elemental.EncodingTypeMSGPACK)

		Convey("When I call directPush and pull from the event channel", func() {

			s.send([]byte("hello"))
			data := <-s.dataCh

			Convey("Then data should be correct", func() {
				So(string(data), ShouldEqual, "hello")
			})
		})

		Convey("When I call directPush and overflow it", func() {

			for i := 0; i < 2000; i++ {
				s.send([]byte("hello"))
			}

			var total int
			for i := 0; i < 2000; i++ {
				select {
				case <-s.dataCh:
					total++
				default:
				}

			}

			Convey("Then we should get 64 data", func() {
				So(total, ShouldEqual, 64)
			})
		})
	})
}

func TestWSPushSession_String(t *testing.T) {

	Convey("Given I have a session", t, func() {

		req, _ := http.NewRequest("GET", "bla", nil)
		cfg := config{}
		s := newWSPushSession(req, cfg, nil, elemental.EncodingTypeMSGPACK, elemental.EncodingTypeMSGPACK)

		Convey("When I call String", func() {

			str := s.String()

			Convey("Then the string representation should be correct", func() {
				So(str, ShouldEqual, fmt.Sprintf("<pushsession id:%s>", s.Identifier()))
			})
		})
	})
}

func TestWSPushSession_Filtering(t *testing.T) {

	Convey("Given I call setCurrentPushConfig", t, func() {

		req, _ := http.NewRequest("GET", "bla", nil)
		cfg := config{}
		s := newWSPushSession(req, cfg, nil, elemental.EncodingTypeMSGPACK, elemental.EncodingTypeMSGPACK)

		pc := elemental.NewPushConfig()
		pc.SetParameter("hello", "world")

		s.setCurrentPushConfig(pc)

		Convey("Then the filter should be installed", func() {
			So(s.currentPushConfig(), ShouldNotEqual, pc)
			So(s.currentPushConfig(), ShouldResemble, pc)
		})

		Convey("Then the parameters have benn installed", func() {
			So(s.Parameter("hello"), ShouldEqual, "world")
		})

		Convey("When I reset the filter to nil", func() {

			s.setCurrentPushConfig(nil)

			Convey("Then the filter should be uninstalled", func() {
				So(s.currentPushConfig(), ShouldBeNil)
			})
		})
	})
}

func TestWSPushSession_accessors(t *testing.T) {

	Convey("Given create a push session", t, func() {

		u, _ := url.Parse("http://toto.com?a=b&token=token")
		conf := config{}
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

		s := newWSPushSession(req, conf, unregister, elemental.EncodingTypeMSGPACK, elemental.EncodingTypeMSGPACK)

		Convey("When I call Identifier()", func() {

			id := s.Identifier()

			Convey("Then id should be correct", func() {
				So(id, ShouldNotBeEmpty)
			})
		})

		Convey("When I call SetClaims()", func() {

			s.SetClaims([]string{"a=a", "b=b"})

			Convey("Then Claims() should return the correct claims ", func() {
				So(s.Claims(), ShouldResemble, []string{"a=a", "b=b"})
			})

			Convey("Then ClaimsMap() should return the correct claims ", func() {
				m := s.ClaimsMap()
				So(len(m), ShouldEqual, 2)
				So(m["a"], ShouldEqual, "a")
				So(m["b"], ShouldEqual, "b")
			})
		})

		Convey("When I call Token()", func() {

			token := s.Token()

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

			Convey("Then Metadata() should return the correct metadata ", func() {
				So(s.Metadata(), ShouldResemble, "hi")
			})
		})

		Convey("When I call Context()", func() {

			c := s.Context()

			Convey("Then Context() should return the correct context ", func() {
				So(opentracing.SpanFromContext(c), ShouldResemble, span)
			})
		})

		Convey("When I call Parameter()", func() {

			p := s.Parameter("a")

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

		Convey("When I call Cookie()", func() {

			s.cookies = []*http.Cookie{
				&http.Cookie{
					Name:  "toto",
					Value: "hey",
				},
			}

			Convey("Then Cookie on existing key should be correct", func() {
				v, err := s.Cookie("toto")
				So(v.Value, ShouldEqual, "hey")
				So(err, ShouldBeNil)
			})

			Convey("Then Cookie on existing non existing should be correct", func() {
				v, err := s.Cookie("titi")
				So(v, ShouldBeNil)
				So(err, ShouldEqual, http.ErrNoCookie)
			})
		})
	})
}

func TestWSPushSession_listen(t *testing.T) {

	Convey("Given I have a push session", t, FailureHalts, func() {

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		unregistered := make(chan bool, 10)

		s := newWSPushSession(
			(&http.Request{URL: &url.URL{}}).WithContext(ctx),
			config{},
			func(i *wsPushSession) {
				unregistered <- true
			},
			elemental.EncodingTypeMSGPACK,
			elemental.EncodingTypeMSGPACK,
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
				r, _ := elemental.Encode(elemental.EncodingTypeMSGPACK, testEvent)
				So(data, ShouldResemble, r)
			})
		})

		Convey("When the client sends a push config with a filter that cannot be parsed", func() {

			go s.listen()

			testIdentity := "identity-one"
			pc := elemental.NewPushConfig()
			pc.FilterIdentity(testIdentity)
			pc.IdentityFilters = map[string]string{
				testIdentity: "this-will-not-parse",
			}

			rawPushConfig, err := elemental.Encode(elemental.EncodingTypeMSGPACK, pc)
			Convey("should be able to encode a semantically invalid push config", func() {
				So(err, ShouldBeNil)
			})

			conn.NextRead(rawPushConfig)

			var closeErr error
			select {
			case closeErr = <-conn.Done():
			case <-time.After(500 * time.Millisecond):
			}

			Convey("Then an error should be received from the channel returned from the connection's Done() method", func() {
				So(closeErr, ShouldNotBeNil)

				Convey("error copy should include the close code indicating why the socket was closed", func() {
					So(closeErr.Error(), ShouldContainSubstring, fmt.Sprintf("%d", websocket.CloseUnsupportedData))
				})

				Convey("the current push config should be nil", func() {
					So(s.currentPushConfig(), ShouldBeNil)
				})
			})

			Convey("The unregister func should have been called", func() {
				var ok bool
				select {
				case ok = <-unregistered:
				case <-time.After(500 * time.Millisecond):
				}

				So(ok, ShouldBeTrue)
			})
		})

		Convey("When the client sends a push config with a filter on an identity that is NOT declared", func() {

			go s.listen()

			testIdentity := "identity-one"
			identityFilter := elemental.NewFilterComposer().
				WithKey("environment").
				Equals("production").
				Done()

			pc := elemental.NewPushConfig()
			pc.FilterIdentity(testIdentity)
			pc.IdentityFilters = map[string]string{
				// notice how the identity filter below is on an identity that is not declared in the PushConfig's 'Identities'
				// attribute - it only contains "identity-one"
				"undeclared-identity": identityFilter.String(),
			}

			rawPushConfig, err := elemental.Encode(elemental.EncodingTypeMSGPACK, pc)
			Convey("should be able to encode the push config", func() {
				So(err, ShouldBeNil)
			})

			conn.NextRead(rawPushConfig)

			var closeErr error
			select {
			case closeErr = <-conn.Done():
			case <-time.After(500 * time.Millisecond):
			}

			Convey("Then an error should be received from the channel returned from the connection's Done() method", func() {
				So(closeErr, ShouldNotBeNil)

				Convey("error copy should include the close code indicating why the socket was closed", func() {
					So(closeErr.Error(), ShouldContainSubstring, fmt.Sprintf("%d", websocket.CloseUnsupportedData))
				})

				Convey("the current push config should be nil", func() {
					So(s.currentPushConfig(), ShouldBeNil)
				})
			})

			Convey("The unregister func should have been called", func() {
				var ok bool
				select {
				case ok = <-unregistered:
				case <-time.After(500 * time.Millisecond):
				}

				So(ok, ShouldBeTrue)
			})
		})

		Convey("When the client sends a push config with a valid identity filter", func() {

			go s.listen()

			testIdentity := "identity-one"
			identityFilter := elemental.NewFilterComposer().
				WithKey("environment").
				Equals("production").
				Done()

			pc := elemental.NewPushConfig()
			pc.FilterIdentity(testIdentity)
			pc.IdentityFilters = map[string]string{
				testIdentity: identityFilter.String(),
			}

			rawPushConfig, err := elemental.Encode(elemental.EncodingTypeMSGPACK, pc)
			Convey("should be able to encode the push config used in the test", func() {
				So(err, ShouldBeNil)
			})

			conn.NextRead(rawPushConfig)

			var closeErr error
			select {
			case closeErr = <-conn.Done():
			case <-time.After(500 * time.Millisecond):
			}

			Convey("Then no error should be returned from the channel returned by calling the connection's Done() method", func() {
				So(closeErr, ShouldBeNil)

				Convey("the current push config should not be nil", func() {
					pc := s.currentPushConfig()
					So(pc, ShouldNotBeNil)
					So(pc.String(), ShouldEqual, `<pushconfig identities:map[identity-one:[]] identityfilters:map[identity-one:environment == "production"]>`)

					Convey("the push config should contain the parsed identity filter", func() {
						filter, found := pc.FilterForIdentity(testIdentity)
						So(found, ShouldBeTrue)
						So(filter.String(), ShouldEqual, identityFilter.String())
					})
				})
			})
		})

		Convey("When I simulate an incoming event that is manually filtered out", func() {

			go s.listen()

			f := elemental.NewPushConfig()
			f.FilterIdentity("not-list")
			s.setCurrentPushConfig(f)

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
				r, _ := elemental.Encode(elemental.EncodingTypeMSGPACK, testEvent)
				So(data, ShouldResemble, r)
			})
		})

		Convey("When I send a valid filter in the websocket", func() {

			go s.listen()

			s.encodingRead = elemental.EncodingTypeJSON
			s.encodingWrite = elemental.EncodingTypeJSON

			conn.NextRead([]byte(`{"identities":{"not-list": null}}`))
			<-time.After(300 * time.Millisecond)

			Convey("Then the filter should be correctly set", func() {
				So(s.currentPushConfig().String(), ShouldEqual, `<pushconfig identities:map[not-list:[]] identityfilters:map[]>`)
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
				So(s.currentPushConfig(), ShouldBeNil)
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
