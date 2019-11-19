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
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/go-zoo/bone"
	. "github.com/smartystreets/goconvey/convey"
	"go.aporeto.io/elemental"
	testmodel "go.aporeto.io/elemental/test/model"
	"go.aporeto.io/wsc"
)

type mockPubSubServer struct {
	publications []*Publication
	PublishErr   error
}

func (p *mockPubSubServer) Connect() Waiter   { return nil }
func (p *mockPubSubServer) Disconnect() error { return nil }

func (p *mockPubSubServer) Publish(publication *Publication, opts ...PubSubOptPublish) error {
	p.publications = append(p.publications, publication)
	return p.PublishErr
}

func (p *mockPubSubServer) Subscribe(pubs chan *Publication, errors chan error, topic string, opts ...PubSubOptSubscribe) func() {
	return nil
}

type mockSessionAuthenticator struct {
	action AuthAction
	err    error
}

func (a *mockSessionAuthenticator) AuthenticateSession(Session) (AuthAction, error) {
	return a.action, a.err
}

type mockSessionHandler struct {
	onPushSessionInitCalled  int
	onPushSessionInitOK      bool
	onPushSessionInitErr     error
	onPushSessionStartCalled int
	onPushSessionStopCalled  int
	shouldPublishCalled      int
	shouldPublishOK          bool
	shouldPublishErr         error
	shouldDispatchCalled     int
	shouldDispatchOK         bool
	shouldDispatchErr        error
	relatedIdentitiesCalled  int
	relatedIdentities        []string
	summarizeEvent           interface{}
	summarizeEventErr        error
	summarizeEventCalled     int

	sync.Mutex
}

func (h *mockSessionHandler) OnPushSessionInit(PushSession) (bool, error) {
	h.Lock()
	defer h.Unlock()

	h.onPushSessionInitCalled++
	return h.onPushSessionInitOK, h.onPushSessionInitErr
}

func (h *mockSessionHandler) OnPushSessionStart(PushSession) {
	h.Lock()
	defer h.Unlock()

	h.onPushSessionStartCalled++
}

func (h *mockSessionHandler) OnPushSessionStop(PushSession) {
	h.Lock()
	defer h.Unlock()

	h.onPushSessionStopCalled++
}

func (h *mockSessionHandler) ShouldPublish(*elemental.Event) (bool, error) {
	h.Lock()
	defer h.Unlock()

	h.shouldPublishCalled++
	return h.shouldPublishOK, h.shouldPublishErr
}

func (h *mockSessionHandler) ShouldDispatch(PushSession, *elemental.Event, interface{}) (bool, error) {
	h.Lock()
	defer h.Unlock()

	h.shouldDispatchCalled++
	return h.shouldDispatchOK, h.shouldDispatchErr
}

func (h *mockSessionHandler) RelatedEventIdentities(i string) []string {

	h.Lock()
	defer h.Unlock()

	h.relatedIdentitiesCalled++
	return h.relatedIdentities
}

func (h *mockSessionHandler) SummarizeEvent(evt *elemental.Event) (interface{}, error) {

	h.Lock()
	defer h.Unlock()

	h.summarizeEventCalled++
	return h.summarizeEvent, h.summarizeEventErr
}

func TestWebsocketServer_newWebsocketServer(t *testing.T) {

	Convey("Given I have a processor finder", t, func() {

		pf := func(identity elemental.Identity) (Processor, error) {
			return struct{}{}, nil
		}

		Convey("When I create a new websocket server with push", func() {

			mux := bone.New()
			cfg := config{}
			cfg.pushServer.enabled = true
			cfg.pushServer.publishEnabled = true
			cfg.pushServer.dispatchEnabled = true

			wss := newPushServer(cfg, mux, pf)

			Convey("Then the websocket sever should be correctly initialized", func() {
				So(wss.sessions, ShouldResemble, map[string]*wsPushSession{})
				So(wss.multiplexer, ShouldEqual, mux)
				So(wss.cfg, ShouldResemble, cfg)
				So(wss.processorFinder, ShouldEqual, pf)
			})

			Convey("Then the handlers should be installed in the mux", func() {
				So(len(mux.Routes), ShouldEqual, 1)
				So(len(mux.Routes["GET"]), ShouldEqual, 1)
				So(mux.Routes["GET"][0].Path, ShouldEqual, "/events")
			})
		})

		Convey("When I create a new websocket server with everything disabled", func() {

			mux := bone.New()
			cfg := config{}

			_ = newPushServer(cfg, mux, pf)

			Convey("Then the handlers should be installed in the mux", func() {
				So(len(mux.Routes), ShouldEqual, 0)
			})
		})
	})
}

func TestWebsockerServer_SessionRegistration(t *testing.T) {

	Convey("Given I have a websocket server", t, func() {

		pf := func(identity elemental.Identity) (Processor, error) {
			return struct{}{}, nil
		}

		req, _ := http.NewRequest("GET", "bla", nil)
		mux := bone.New()
		cfg := config{}
		h := &mockSessionHandler{}
		cfg.pushServer.dispatchHandler = h

		wss := newPushServer(cfg, mux, pf)

		Convey("When I register a valid push session", func() {

			s := newWSPushSession(req, cfg, nil, elemental.EncodingTypeJSON, elemental.EncodingTypeJSON)
			wss.registerSession(s)

			Convey("Then the session should correctly registered", func() {
				So(len(wss.sessions), ShouldEqual, 1)
				So(wss.sessions[s.Identifier()], ShouldEqual, s)
			})

			Convey("Then handler.onPushSessionStart should have been called", func() {
				So(h.onPushSessionStartCalled, ShouldEqual, 1)
			})

			Convey("When I unregister it", func() {

				wss.unregisterSession(s)

				Convey("Then the session should correctly unregistered", func() {
					So(len(wss.sessions), ShouldEqual, 0)
				})

				Convey("Then handler.onPushSessionStop should have been called", func() {
					So(h.onPushSessionStopCalled, ShouldEqual, 1)
				})
			})
		})

		Convey("When I register a valid session with no id", func() {

			s := &wsPushSession{}

			Convey("Then it should panic", func() {
				So(func() { wss.registerSession(s) }, ShouldPanicWith, "cannot register websocket session. empty identifier")
			})
		})

		Convey("When I unregister a valid session with no id", func() {

			s := &wsPushSession{}

			Convey("Then it should panic", func() {
				So(func() { wss.unregisterSession(s) }, ShouldPanicWith, "cannot unregister websocket session. empty identifier")
			})
		})
	})
}

func TestWebsocketServer_authSession(t *testing.T) {

	Convey("Given I have a websocket server", t, func() {

		pf := func(identity elemental.Identity) (Processor, error) {
			return struct{}{}, nil
		}

		req, _ := http.NewRequest("GET", "bla", nil)
		mux := bone.New()

		Convey("When I call authSession on when there is no authenticator configured", func() {

			cfg := config{}

			wss := newPushServer(cfg, mux, pf)

			s := newWSPushSession(req, cfg, nil, elemental.EncodingTypeJSON, elemental.EncodingTypeJSON)
			err := wss.authSession(s)

			Convey("Then err should be nil", func() {
				So(err, ShouldBeNil)
			})
		})

		Convey("When I call authSession with a configured authenticator that is ok", func() {

			a := &mockSessionAuthenticator{}
			a.action = AuthActionOK

			cfg := config{}
			cfg.security.sessionAuthenticators = []SessionAuthenticator{a}

			wss := newPushServer(cfg, mux, pf)

			s := newWSPushSession(req, cfg, nil, elemental.EncodingTypeJSON, elemental.EncodingTypeJSON)
			err := wss.authSession(s)

			Convey("Then err should be nil", func() {
				So(err, ShouldBeNil)
			})
		})

		Convey("When I call authSession with a configured authenticator that is not ok", func() {

			a := &mockSessionAuthenticator{}
			a.action = AuthActionKO

			cfg := config{}
			cfg.security.sessionAuthenticators = []SessionAuthenticator{a}

			wss := newPushServer(cfg, mux, pf)

			s := newWSPushSession(req, cfg, nil, elemental.EncodingTypeJSON, elemental.EncodingTypeJSON)
			err := wss.authSession(s)

			Convey("Then err should not be nil", func() {
				So(err.Error(), ShouldEqual, "error 401 (bahamut): Unauthorized: You are not authorized to start a session")
			})
		})

		Convey("When I call authSession with a configured authenticator that returns an error", func() {

			a := &mockSessionAuthenticator{}
			a.action = AuthActionOK // we wan't to check that error takes precedence
			a.err = errors.New("nope")

			cfg := config{}
			cfg.security.sessionAuthenticators = []SessionAuthenticator{a}

			wss := newPushServer(cfg, mux, pf)

			s := newWSPushSession(req, cfg, nil, elemental.EncodingTypeJSON, elemental.EncodingTypeJSON)
			err := wss.authSession(s)

			Convey("Then err should not be nil", func() {
				So(err.Error(), ShouldEqual, "error 401 (bahamut): Unauthorized: nope")
			})
		})
	})
}

func TestWebsocketServer_initPushSession(t *testing.T) {

	Convey("Given I have a websocket server", t, func() {

		pf := func(identity elemental.Identity) (Processor, error) {
			return struct{}{}, nil
		}

		req, _ := http.NewRequest("GET", "bla", nil)
		mux := bone.New()

		Convey("When I call initSession on when there is no session handler configured", func() {

			cfg := config{}

			wss := newPushServer(cfg, mux, pf)

			s := newWSPushSession(req, cfg, nil, elemental.EncodingTypeJSON, elemental.EncodingTypeJSON)
			err := wss.initPushSession(s)

			Convey("Then err should be nil", func() {
				So(err, ShouldBeNil)
			})
		})

		Convey("When I call initSession on when there is a session handler that is ok", func() {

			h := &mockSessionHandler{}
			h.onPushSessionInitOK = true

			cfg := config{}
			cfg.pushServer.dispatchHandler = h

			wss := newPushServer(cfg, mux, pf)

			s := newWSPushSession(req, cfg, nil, elemental.EncodingTypeJSON, elemental.EncodingTypeJSON)
			err := wss.initPushSession(s)

			Convey("Then err should be nil", func() {
				So(err, ShouldBeNil)
			})
		})

		Convey("When I call initSession on when there is a session handler that is not ok", func() {

			h := &mockSessionHandler{}
			h.onPushSessionInitOK = false

			cfg := config{}
			cfg.pushServer.dispatchHandler = h

			wss := newPushServer(cfg, mux, pf)

			s := newWSPushSession(req, cfg, nil, elemental.EncodingTypeJSON, elemental.EncodingTypeJSON)
			err := wss.initPushSession(s)

			Convey("Then err should not be nil", func() {
				So(err.Error(), ShouldEqual, "error 403 (bahamut): Forbidden: You are not authorized to initiate a push session")
			})
		})

		Convey("When I call initSession on when there is a session handler that returns an error", func() {

			h := &mockSessionHandler{}
			h.onPushSessionInitOK = true // we wan't to check that error takes precedence
			h.onPushSessionInitErr = errors.New("nope")

			cfg := config{}
			cfg.pushServer.dispatchHandler = h

			wss := newPushServer(cfg, mux, pf)

			s := newWSPushSession(req, cfg, nil, elemental.EncodingTypeJSON, elemental.EncodingTypeJSON)
			err := wss.initPushSession(s)

			Convey("Then err should not be nil", func() {
				So(err.Error(), ShouldEqual, "error 403 (bahamut): Forbidden: nope")
			})
		})
	})
}

func TestWebsocketServer_pushEvents(t *testing.T) {

	Convey("Given I have a websocket server", t, func() {

		pf := func(identity elemental.Identity) (Processor, error) {
			return struct{}{}, nil
		}

		mux := bone.New()

		Convey("When I call pushEvents when no service is configured", func() {

			cfg := config{}

			wss := newPushServer(cfg, mux, pf)
			wss.pushEvents(nil)

			Convey("Then nothing special should happen", func() {})
		})

		Convey("When I call pushEvents with a service is configured but no sessions handler", func() {

			srv := &mockPubSubServer{}

			cfg := config{}
			cfg.pushServer.service = srv
			cfg.pushServer.enabled = true
			cfg.pushServer.publishEnabled = true
			cfg.pushServer.dispatchEnabled = true

			wss := newPushServer(cfg, mux, pf)
			evtin := elemental.NewEvent(elemental.EventCreate, testmodel.NewList())
			wss.pushEvents(evtin)

			Convey("Then I should find one publication", func() {
				evtout := elemental.NewEvent(elemental.EventCreate, testmodel.NewList())
				evtout.Timestamp = evtin.Timestamp
				r, _ := elemental.Encode(elemental.EncodingTypeMSGPACK, evtout)
				So(len(srv.publications), ShouldEqual, 1)
				So(string(srv.publications[0].Data), ShouldResemble, string(r))
			})
		})

		Convey("When I call pushEvents with a service is configured and sessions handler that is ok to push", func() {

			srv := &mockPubSubServer{}
			h := &mockSessionHandler{}
			h.shouldPublishOK = true

			cfg := config{}
			cfg.pushServer.service = srv
			cfg.pushServer.enabled = true
			cfg.pushServer.publishEnabled = true
			cfg.pushServer.dispatchEnabled = true
			cfg.pushServer.publishHandler = h

			wss := newPushServer(cfg, mux, pf)
			evtin := elemental.NewEvent(elemental.EventCreate, testmodel.NewList())
			wss.pushEvents(evtin)

			Convey("Then I should find one publication", func() {
				evtout := elemental.NewEvent(elemental.EventCreate, testmodel.NewList())
				evtout.Timestamp = evtin.Timestamp
				r, _ := elemental.Encode(elemental.EncodingTypeMSGPACK, evtout)
				So(len(srv.publications), ShouldEqual, 1)
				So(string(srv.publications[0].Data), ShouldResemble, string(r))
			})
		})

		Convey("When I call pushEvents with a service is configured and sessions handler that is not ok to push", func() {

			srv := &mockPubSubServer{}
			h := &mockSessionHandler{}
			h.shouldPublishOK = false

			cfg := config{}
			cfg.pushServer.service = srv
			cfg.pushServer.enabled = true
			cfg.pushServer.publishEnabled = true
			cfg.pushServer.dispatchEnabled = true
			cfg.pushServer.publishHandler = h

			wss := newPushServer(cfg, mux, pf)
			wss.pushEvents(elemental.NewEvent(elemental.EventCreate, testmodel.NewList()))

			Convey("Then I should find one publication", func() {
				So(len(srv.publications), ShouldEqual, 0)
			})
		})

		Convey("When I call pushEvents with a service is configured and sessions handler that returns an error", func() {

			srv := &mockPubSubServer{}
			h := &mockSessionHandler{}
			h.shouldPublishOK = true // we want to be sure error takes precedence
			h.shouldPublishErr = errors.New("nop")

			cfg := config{}
			cfg.pushServer.service = srv
			cfg.pushServer.enabled = true
			cfg.pushServer.publishEnabled = true
			cfg.pushServer.dispatchEnabled = true
			cfg.pushServer.publishHandler = h

			wss := newPushServer(cfg, mux, pf)
			wss.pushEvents(elemental.NewEvent(elemental.EventCreate, testmodel.NewList()))

			Convey("Then I should find one publication", func() {
				So(len(srv.publications), ShouldEqual, 0)
			})
		})
	})
}

func TestWebsocketServer_start(t *testing.T) {

	pf := func(identity elemental.Identity) (Processor, error) {
		return struct{}{}, nil
	}

	makePubsub := func(ctx context.Context, idprefix string) (*pushServer, *mockSessionHandler, *wsPushSession, *wsPushSession) {

		pubsub := NewLocalPubSubClient()
		if !pubsub.Connect().Wait(3 * time.Second) {
			panic("could not connect to local pubsub")
		}

		pushHandler := &mockSessionHandler{}

		mux := bone.New()
		cfg := config{}
		cfg.pushServer.service = pubsub
		cfg.pushServer.enabled = true
		cfg.pushServer.publishEnabled = true
		cfg.pushServer.dispatchEnabled = true
		cfg.pushServer.dispatchHandler = pushHandler

		wss := newPushServer(cfg, mux, pf)

		go wss.start(ctx)

		s1 := newWSPushSession(
			(&http.Request{URL: &url.URL{}}).WithContext(ctx),
			config{},
			wss.unregisterSession,
			elemental.EncodingTypeMSGPACK,
			elemental.EncodingTypeMSGPACK,
		)
		conn1 := wsc.NewMockWebsocket(ctx)
		s1.setConn(conn1)
		s1.id = idprefix + "s1"

		go s1.listen()

		s2 := newWSPushSession(
			(&http.Request{URL: &url.URL{}}).WithContext(ctx),
			config{},
			wss.unregisterSession,
			elemental.EncodingTypeMSGPACK,
			elemental.EncodingTypeMSGPACK,
		)
		conn2 := wsc.NewMockWebsocket(ctx)
		s2.setConn(conn2)
		s2.id = "s2"

		go s2.listen()

		wss.registerSession(s1)
		wss.registerSession(s2)

		return wss, pushHandler, s1, s2
	}

	Convey("Given I push an event that is filtered out by one session", t, func() {

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		pushServer, pushHandler, s1, s2 := makePubsub(ctx, "")
		conn1 := s1.conn.(wsc.MockWebsocket)
		conn2 := s2.conn.(wsc.MockWebsocket)

		filter := elemental.NewPushConfig()
		filter.FilterIdentity("something-else")
		s1.setCurrentPushConfig(filter)

		filter = elemental.NewPushConfig()
		filter.FilterIdentity("list")
		s2.setCurrentPushConfig(filter)

		pushHandler.shouldDispatchOK = true

		evt := elemental.NewEvent(elemental.EventCreate, testmodel.NewList())
		evt.Timestamp = time.Now().Add(time.Second)
		pub := NewPublication("")
		if err := pub.Encode(evt); err != nil {
			panic(err)
		}

		pushServer.publications <- pub

		var msg1 []byte
		var msg2 []byte
		var l sync.Mutex
		finished := make(chan struct{})
		go func() {
			defer close(finished)

			for {
				select {
				case data := <-conn1.LastWrite():
					l.Lock()
					msg1 = data
					l.Unlock()
				case data := <-conn2.LastWrite():
					l.Lock()
					msg2 = data
					l.Unlock()
				case <-time.After(1 * time.Second):
					return
				}
			}
		}()
		<-finished

		So(msg1, ShouldBeNil)
		So(msg2, ShouldNotBeNil)
	})

	Convey("Given I push an event that is filtered out but related by one session", t, func() {

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		pushServer, pushHandler, s1, s2 := makePubsub(ctx, "")
		conn1 := s1.conn.(wsc.MockWebsocket)
		conn2 := s2.conn.(wsc.MockWebsocket)

		pushHandler.relatedIdentities = []string{"something-else"}

		filter := elemental.NewPushConfig()
		filter.FilterIdentity("something-else")
		s1.setCurrentPushConfig(filter)

		pushHandler.shouldDispatchOK = true

		evt := elemental.NewEvent(elemental.EventCreate, testmodel.NewList())
		evt.Timestamp = time.Now().Add(time.Second)
		pub := NewPublication("")
		if err := pub.Encode(evt); err != nil {
			panic(err)
		}

		pushServer.publications <- pub

		var msg1 []byte
		var msg2 []byte
		var l sync.Mutex
		finished := make(chan struct{})
		go func() {
			defer close(finished)

			for {
				select {
				case data := <-conn1.LastWrite():
					l.Lock()
					msg1 = data
					l.Unlock()
				case data := <-conn2.LastWrite():
					l.Lock()
					msg2 = data
					l.Unlock()
				case <-time.After(1 * time.Second):
					return
				}
			}
		}()
		<-finished

		So(msg1, ShouldNotBeNil)
		So(msg2, ShouldNotBeNil)
	})

	Convey("Given I push an event and the handler is ok then both sessions should receive the event", t, func() {

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		pushServer, pushHandler, s1, s2 := makePubsub(ctx, "toto")
		conn1 := s1.conn.(wsc.MockWebsocket)
		conn2 := s2.conn.(wsc.MockWebsocket)

		pushHandler.shouldDispatchOK = true

		evt := elemental.NewEvent(elemental.EventCreate, testmodel.NewList())
		pub := NewPublication("")
		if err := pub.Encode(evt); err != nil {
			panic(err)
		}

		pushServer.publications <- pub

		var msg1 []byte
		var msg2 []byte
		var l sync.Mutex
		finished := make(chan struct{})
		go func() {

			for {

				if len(msg1) > 0 && len(msg2) > 0 {
					close(finished)
					return
				}

				select {
				case data := <-conn1.LastWrite():
					fmt.Println("ZOBN")
					l.Lock()
					msg1 = data
					l.Unlock()
				case data := <-conn2.LastWrite():
					fmt.Println("P{ROUT")
					l.Lock()
					msg2 = data
					l.Unlock()

				case <-ctx.Done():
					panic("test: no response in time")
				}
			}
		}()
		<-finished

		d1, _ := elemental.Encode(elemental.EncodingTypeMSGPACK, evt)

		l.Lock()
		So(msg1, ShouldResemble, d1)
		So(msg2, ShouldResemble, d1)
		l.Unlock()
	})

	Convey("Given I push an event and the handler is not ok then no session should receive the event", t, func() {

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		pushServer, pushHandler, s1, s2 := makePubsub(ctx, "")
		conn1 := s1.conn.(wsc.MockWebsocket)
		conn2 := s2.conn.(wsc.MockWebsocket)

		pushHandler.shouldDispatchOK = false

		evt := elemental.NewEvent(elemental.EventCreate, testmodel.NewList())
		pub := NewPublication("")
		if err := pub.Encode(evt); err != nil {
			panic(err)
		}

		pushServer.publications <- pub

		var msg1 []byte
		var msg2 []byte
		var l sync.Mutex
		finished := make(chan struct{})
		go func() {
			defer close(finished)

			for {
				select {
				case data := <-conn1.LastWrite():
					l.Lock()
					msg1 = data
					l.Unlock()
				case data := <-conn2.LastWrite():
					l.Lock()
					msg2 = data
					l.Unlock()
				case <-time.After(1 * time.Second):
					return
				}
			}
		}()
		<-finished

		l.Lock()
		So(msg1, ShouldBeNil)
		So(msg2, ShouldBeNil)
		l.Unlock()
	})

	Convey("Given I push an event that is that older than session connection time", t, func() {

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		pushServer, pushHandler, s1, s2 := makePubsub(ctx, "")
		conn1 := s1.conn.(wsc.MockWebsocket)
		conn2 := s2.conn.(wsc.MockWebsocket)

		pushHandler.shouldDispatchOK = true

		evt := elemental.NewEvent(elemental.EventCreate, testmodel.NewList())
		evt.Timestamp = time.Now().Add(-time.Hour)

		pub := NewPublication("")
		if err := pub.Encode(evt); err != nil {
			panic(err)
		}

		pushServer.publications <- pub

		var msg1 []byte
		var msg2 []byte
		var l sync.Mutex
		finished := make(chan struct{})
		go func() {
			defer close(finished)

			for {
				select {
				case data := <-conn1.LastWrite():
					l.Lock()
					msg1 = data
					l.Unlock()
				case data := <-conn2.LastWrite():
					l.Lock()
					msg2 = data
					l.Unlock()
				case <-time.After(1 * time.Second):
					return
				}
			}
		}()
		<-finished

		l.Lock()
		So(msg1, ShouldBeNil)
		So(msg2, ShouldBeNil)
		l.Unlock()
	})

	Convey("Given I push an event and the handler returns an error then then both sessions should receive no event", t, func() {

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		pushServer, pushHandler, s1, s2 := makePubsub(ctx, "")
		conn1 := s1.conn.(wsc.MockWebsocket)
		conn2 := s2.conn.(wsc.MockWebsocket)

		pushHandler.shouldDispatchOK = true
		pushHandler.shouldDispatchErr = errors.New("nope")

		evt := elemental.NewEvent(elemental.EventCreate, testmodel.NewList())
		pub := NewPublication("")
		if err := pub.Encode(evt); err != nil {
			panic(err)
		}

		pushServer.publications <- pub

		var msg1 []byte
		var msg2 []byte
		var l sync.Mutex
		finished := make(chan struct{})
		go func() {
			defer close(finished)

			for {
				select {
				case data := <-conn1.LastWrite():
					l.Lock()
					msg1 = data
					l.Unlock()
				case data := <-conn2.LastWrite():
					l.Lock()
					msg2 = data
					l.Unlock()
				case <-time.After(1 * time.Second):
					return
				}
			}
		}()
		<-finished

		l.Lock()
		So(msg1, ShouldBeNil)
		So(msg2, ShouldBeNil)
		l.Unlock()
	})

	Convey("Given I dispatcher returns an error then both sessions should receive no event", t, func() {

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		pushServer, pushHandler, s1, s2 := makePubsub(ctx, "")
		conn1 := s1.conn.(wsc.MockWebsocket)
		conn2 := s2.conn.(wsc.MockWebsocket)

		pushHandler.shouldDispatchOK = true
		pushHandler.shouldDispatchErr = fmt.Errorf("boom")

		evt := elemental.NewEvent(elemental.EventCreate, testmodel.NewList())
		pub := NewPublication("")
		if err := pub.Encode(evt); err != nil {
			panic(err)
		}

		pushServer.publications <- pub

		var msg1 []byte
		var msg2 []byte
		var l sync.Mutex
		finished := make(chan struct{})
		go func() {
			defer close(finished)

			for {
				select {
				case data := <-conn1.LastWrite():
					l.Lock()
					msg1 = data
					l.Unlock()
				case data := <-conn2.LastWrite():
					l.Lock()
					msg2 = data
					l.Unlock()
				case <-time.After(1 * time.Second):
					return
				}
			}
		}()
		<-finished

		l.Lock()
		So(msg1, ShouldBeNil)
		So(msg2, ShouldBeNil)
		l.Unlock()
	})

	Convey("Given I I receive a bad event in the publication", t, func() {

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		pushServer, pushHandler, s1, s2 := makePubsub(ctx, "")
		conn1 := s1.conn.(wsc.MockWebsocket)
		conn2 := s2.conn.(wsc.MockWebsocket)

		pushHandler.shouldDispatchOK = true

		evt := elemental.NewEvent(elemental.EventCreate, testmodel.NewList())
		pub := NewPublication("")
		evt.RawData = []byte("{broken")
		if err := pub.Encode(evt); err != nil {
			panic(err)
		}

		pushServer.publications <- pub

		var msg1 []byte
		var msg2 []byte
		var l sync.Mutex
		finished := make(chan struct{})
		go func() {
			defer close(finished)

			for {
				select {
				case data := <-conn1.LastWrite():
					l.Lock()
					msg1 = data
					l.Unlock()
				case data := <-conn2.LastWrite():
					l.Lock()
					msg2 = data
					l.Unlock()
				case <-time.After(1 * time.Second):
					return
				}
			}
		}()
		<-finished

		l.Lock()
		So(msg1, ShouldBeNil)
		So(msg2, ShouldBeNil)
		l.Unlock()
	})

	Convey("Given I I receive a bad  publication", t, func() {

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		pushServer, pushHandler, s1, s2 := makePubsub(ctx, "")
		conn1 := s1.conn.(wsc.MockWebsocket)
		conn2 := s2.conn.(wsc.MockWebsocket)

		pushHandler.shouldDispatchOK = true

		pub := NewPublication("")
		pub.Data = []byte("{broken")

		pushServer.publications <- pub

		var msg1 []byte
		var msg2 []byte
		var l sync.Mutex
		finished := make(chan struct{})
		go func() {
			defer close(finished)

			for {
				select {
				case data := <-conn1.LastWrite():
					l.Lock()
					msg1 = data
					l.Unlock()
				case data := <-conn2.LastWrite():
					l.Lock()
					msg2 = data
					l.Unlock()
				case <-time.After(1 * time.Second):
					return
				}
			}
		}()
		<-finished

		l.Lock()
		So(msg1, ShouldBeNil)
		So(msg2, ShouldBeNil)
		l.Unlock()
	})

	Convey("Given I start a websocket server with no push dispatching", t, func() {

		mux := bone.New()
		cfg := config{}

		wss := newPushServer(cfg, mux, pf)

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		out := make(chan bool)
		go func() {
			wss.start(ctx)
			out <- true
		}()

		So(
			func() {
				select {
				case <-out:
					panic("test: unexpected response")
				case <-time.After(3 * time.Second):
				}
			},
			ShouldNotPanic,
		)

		cancel()

		var exited bool
		select {
		case exited = <-out:
		case <-time.After(3 * time.Second):
			panic("test: no respons in time")
		}

		So(exited, ShouldBeTrue)
	})
}

func TestWebsocketServer_handleRequest(t *testing.T) {

	Convey("Given I have a webserver", t, func() {

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		pf := func(identity elemental.Identity) (Processor, error) {
			return struct{}{}, nil
		}

		pushHandler := &mockSessionHandler{}
		authenticator := &mockSessionAuthenticator{}

		mux := bone.New()
		cfg := config{}
		cfg.pushServer.dispatchHandler = pushHandler
		cfg.pushServer.enabled = true
		cfg.pushServer.publishEnabled = true
		cfg.pushServer.dispatchEnabled = true
		cfg.security.sessionAuthenticators = []SessionAuthenticator{authenticator}

		wss := newPushServer(cfg, mux, pf)
		wss.mainContext = ctx

		ts := httptest.NewServer(http.HandlerFunc(wss.handleRequest))
		defer ts.Close()

		Convey("When I connect to the server with no issue", func() {

			authenticator.action = AuthActionOK

			pushHandler.Lock()
			pushHandler.onPushSessionInitOK = true
			pushHandler.Unlock()

			ws, resp, err := wsc.Connect(ctx, strings.Replace(ts.URL, "http://", "ws://", 1), wsc.Config{
				Headers: http.Header{"X-Forwarded-For": []string{"12.12.12.12"}},
			})
			Convey("Then err should should be nil", func() {
				So(err, ShouldBeNil)
			})
			defer ws.Close(0) // nolint

			Convey("Then resp should should be correct", func() {
				So(resp.Status, ShouldEqual, "101 Switching Protocols")
			})
		})

		Convey("When I connect to the server but I am not authenticated", func() {

			authenticator.action = AuthActionKO

			pushHandler.Lock()
			pushHandler.onPushSessionInitOK = true
			pushHandler.Unlock()

			ws, resp, err := wsc.Connect(ctx, strings.Replace(ts.URL, "http://", "ws://", 1), wsc.Config{
				Headers: http.Header{"X-Real-IP": []string{"12.12.12.12"}},
			})

			Convey("Then ws should be nil", func() {
				So(ws, ShouldBeNil)
			})

			Convey("Then err should should not be nil", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "websocket: bad handshake")
			})

			Convey("Then resp should should be correct", func() {
				So(resp.Status, ShouldEqual, "401 Unauthorized")
			})
		})

		Convey("When I connect to the server but I am not authorized", func() {

			authenticator.action = AuthActionOK
			pushHandler.Lock()
			pushHandler.onPushSessionInitOK = false
			pushHandler.Unlock()

			ws, resp, err := wsc.Connect(ctx, strings.Replace(ts.URL, "http://", "ws://", 1), wsc.Config{})

			Convey("Then ws should be nil", func() {
				So(ws, ShouldBeNil)
			})

			Convey("Then err should should not be nil", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "websocket: bad handshake")
			})

			Convey("Then resp should should be correct", func() {
				So(resp.Status, ShouldEqual, "403 Forbidden")
			})
		})
	})
}

func Test_prepareEventData(t *testing.T) {

	pristineEvent := elemental.NewEvent(
		elemental.EventUpdate,
		&testmodel.List{ID: "ID1", Name: "Hello"},
	)

	// keep a copy
	event := pristineEvent.Duplicate()

	// prepare some known conversions
	msgpackEventData, err := elemental.Encode(elemental.EncodingTypeMSGPACK, event)
	if err != nil {
		panic(err)
	}

	err = event.Convert(elemental.EncodingTypeJSON)
	if err != nil {
		panic(err)
	}

	jsonEventData, err := elemental.Encode(elemental.EncodingTypeJSON, event)
	if err != nil {
		panic(err)
	}

	type args struct {
		event *elemental.Event
	}

	tests := []struct {
		name        string
		args        args
		wantMSGPACK string
		wantJSON    string
		checkErr    func(error) (bool, string)
	}{
		{
			"msgpack event",
			args{
				pristineEvent,
			},
			string(msgpackEventData),
			string(jsonEventData),
			func(err error) (bool, string) {
				return false, ""
			},
		},
		{
			"unencodable msgpack event",
			args{
				func() *elemental.Event {
					dupe := pristineEvent.Duplicate()
					dupe.RawData = []byte("broken")
					return dupe
				}(),
			},
			"",
			"",
			func(err error) (bool, string) {
				return true, "unable to convert original msgpack encoding to json: unable to decode application/msgpack: msgpack decode error [pos 1]: cannot read container length: unrecognized descriptor byte: hex: 62, decimal: 98"
			},
		},

		{
			"json event",
			args{
				event,
			},
			string(msgpackEventData),
			string(jsonEventData),
			func(err error) (bool, string) {
				return false, ""
			},
		},

		{
			"unencodable json event",
			args{
				func() *elemental.Event {
					dupe := event.Duplicate()
					dupe.JSONData = []byte("broken")
					return dupe
				}(),
			},
			"",
			"",
			func(err error) (bool, string) {
				return true, "unable to convert original json encoding to msgpack: unable to decode application/json: json decode error [pos 1]: read map - expect char '{' but got char 'b'"
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotMSGPACK, gotJSON, err := prepareEventData(tt.args.event)

			wantErr, wantErrStr := tt.checkErr(err)

			if wantErr && err == nil {
				t.Errorf("prepareEventData() error = %v want %s", err, wantErrStr)
			}

			if wantErr && err.Error() != wantErrStr {
				t.Errorf("prepareEventData() error = %v want %s", err, wantErrStr)
			}

			if string(gotMSGPACK) != tt.wantMSGPACK {
				t.Errorf("prepareEventData() gotMsgpack = %v, want %v", string(gotMSGPACK), tt.wantMSGPACK)
			}
			if string(gotJSON) != tt.wantJSON {
				t.Errorf("prepareEventData() gotJson = %v, want %v", string(gotJSON), tt.wantJSON)
			}
		})
	}
}
