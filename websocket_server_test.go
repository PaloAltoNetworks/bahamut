package bahamut

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"

	"go.aporeto.io/addedeffect/wsc"
	"go.aporeto.io/elemental"
	"go.aporeto.io/elemental/test/model"
	"github.com/go-zoo/bone"

	. "github.com/smartystreets/goconvey/convey"
)

type mockPubSubServer struct {
	publications []*Publication
	PublishErr   error
}

func (p *mockPubSubServer) Connect() Waiter   { return nil }
func (p *mockPubSubServer) Disconnect() error { return nil }

func (p *mockPubSubServer) Publish(publication *Publication) error {
	p.publications = append(p.publications, publication)
	return p.PublishErr
}

func (p *mockPubSubServer) Subscribe(pubs chan *Publication, errors chan error, topic string, args ...interface{}) func() {
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

func (h *mockSessionHandler) ShouldDispatch(PushSession, *elemental.Event) (bool, error) {
	h.Lock()
	defer h.Unlock()

	h.shouldDispatchCalled++
	return h.shouldDispatchOK, h.shouldDispatchErr
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

			s := newWSPushSession(req, cfg, nil)
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

			s := newWSPushSession(req, cfg, nil)
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

			s := newWSPushSession(req, cfg, nil)
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

			s := newWSPushSession(req, cfg, nil)
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

			s := newWSPushSession(req, cfg, nil)
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

			s := newWSPushSession(req, cfg, nil)
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

			s := newWSPushSession(req, cfg, nil)
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

			s := newWSPushSession(req, cfg, nil)
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

			s := newWSPushSession(req, cfg, nil)
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

			Convey("Then nothing special should happen", func() {
			})
		})

		Convey("When I call pushEvents with a service is configured but no sessions handler", func() {

			srv := &mockPubSubServer{}

			cfg := config{}
			cfg.pushServer.service = srv
			cfg.pushServer.enabled = true
			cfg.pushServer.publishEnabled = true
			cfg.pushServer.dispatchEnabled = true

			wss := newPushServer(cfg, mux, pf)
			wss.pushEvents(elemental.NewEvent(elemental.EventCreate, testmodel.NewList()))

			Convey("Then I should find one publication", func() {
				So(len(srv.publications), ShouldEqual, 1)
				So(string(srv.publications[0].Data), ShouldStartWith, `{"entity":{"ID":"","creationOnly":"","date":"0001-01-01T00:00:00Z","description":"","name":"","parentID":"","parentType":"","readOnly":"","slice":null},"identity":"list","type":"create","timestamp":"`)
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
			wss.pushEvents(elemental.NewEvent(elemental.EventCreate, testmodel.NewList()))

			Convey("Then I should find one publication", func() {
				So(len(srv.publications), ShouldEqual, 1)
				So(string(srv.publications[0].Data), ShouldStartWith, `{"entity":{"ID":"","creationOnly":"","date":"0001-01-01T00:00:00Z","description":"","name":"","parentID":"","parentType":"","readOnly":"","slice":null},"identity":"list","type":"create","timestamp":"`)
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

	Convey("Given I have a websocket server with 2 registered sessions", t, func() {

		pubsub := NewLocalPubSubClient(nil)
		if !pubsub.Connect().Wait(2 * time.Second) {
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

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		go wss.start(ctx)

		s1 := newWSPushSession(
			(&http.Request{URL: &url.URL{}}).WithContext(ctx),
			config{},
			wss.unregisterSession,
		)
		conn1 := wsc.NewMockWebsocket(ctx)
		s1.setConn(conn1)
		s1.id = "s1"
		go s1.listen()

		s2 := newWSPushSession(
			(&http.Request{URL: &url.URL{}}).WithContext(ctx),
			config{},
			wss.unregisterSession,
		)
		conn2 := wsc.NewMockWebsocket(ctx)
		s2.setConn(conn2)
		s2.id = "s2"
		go s2.listen()

		wss.registerSession(s1)
		wss.registerSession(s2)

		Convey("When I push an event and the handler is ok", func() {

			pushHandler.shouldDispatchOK = true

			evt := elemental.NewEvent(elemental.EventCreate, testmodel.NewList())
			pub := NewPublication("")
			pub.Encode(evt) // nolint: errcheck

			pubsub.Publish(pub) // nolint: errcheck

			var msg1 []byte
			select {
			case msg1 = <-conn1.LastWrite():
			case <-ctx.Done():
				panic("test: no response in time")
			}

			var msg2 []byte
			select {
			case msg2 = <-conn2.LastWrite():
			case <-ctx.Done():
				panic("test: no response in time")
			}

			Convey("Then both sessions should receive the event", func() {
				So(string(msg1), ShouldStartWith, `{"entity":{"ID":"","creationOnly":"","date":"0001-01-01T00:00:00Z","description":"","name":"","parentID":"","parentType":"","readOnly":"","slice":null},"identity":"list","type":"create","timestamp":"`)
				So(string(msg2), ShouldStartWith, `{"entity":{"ID":"","creationOnly":"","date":"0001-01-01T00:00:00Z","description":"","name":"","parentID":"","parentType":"","readOnly":"","slice":null},"identity":"list","type":"create","timestamp":"`)
			})
		})

		Convey("When I push an event and the handler is not ok", func() {

			pushHandler.shouldDispatchOK = false

			evt := elemental.NewEvent(elemental.EventCreate, testmodel.NewList())
			pub := NewPublication("")
			pub.Encode(evt) // nolint: errcheck

			pubsub.Publish(pub) // nolint: errcheck

			var msg1 []byte
			select {
			case msg1 = <-conn1.LastWrite():
			case <-time.After(300 * time.Millisecond):
			case <-ctx.Done():
				panic("test: no response in time")
			}

			var msg2 []byte
			select {
			case msg2 = <-conn2.LastWrite():
			case <-time.After(300 * time.Millisecond):
			case <-ctx.Done():
				panic("test: no response in time")
			}

			Convey("Then both sessions should receive the event", func() {
				So(msg1, ShouldBeNil)
				So(msg2, ShouldBeNil)
			})
		})

		Convey("When I push an event and the handler returns an error", func() {

			pushHandler.shouldDispatchOK = true
			pushHandler.shouldDispatchErr = errors.New("nope")

			evt := elemental.NewEvent(elemental.EventCreate, testmodel.NewList())
			pub := NewPublication("")
			pub.Encode(evt) // nolint: errcheck

			pubsub.Publish(pub) // nolint: errcheck

			var msg1 []byte
			select {
			case msg1 = <-conn1.LastWrite():
			case <-time.After(300 * time.Millisecond):
			case <-ctx.Done():
				panic("test: no response in time")
			}

			var msg2 []byte
			select {
			case msg2 = <-conn2.LastWrite():
			case <-time.After(300 * time.Millisecond):
			case <-ctx.Done():
				panic("test: no response in time")
			}

			Convey("Then both sessions should receive the event", func() {
				So(msg1, ShouldBeNil)
				So(msg2, ShouldBeNil)
			})
		})

		Convey("When I push bad event", func() {

			pushHandler.shouldDispatchOK = true
			pushHandler.shouldDispatchErr = errors.New("nope")

			evt := elemental.NewEvent(elemental.EventCreate, testmodel.NewList())
			pub := NewPublication("")
			evt.Entity = []byte(`{ broken`)
			pub.Encode(evt) // nolint: errcheck

			pubsub.Publish(pub) // nolint: errcheck

			var msg1 []byte
			select {
			case msg1 = <-conn1.LastWrite():
			case <-time.After(300 * time.Millisecond):
			case <-ctx.Done():
				panic("test: no response in time")
			}

			var msg2 []byte
			select {
			case msg2 = <-conn2.LastWrite():
			case <-time.After(300 * time.Millisecond):
			case <-ctx.Done():
				panic("test: no response in time")
			}

			Convey("Then both sessions should receive the event", func() {
				So(msg1, ShouldBeNil)
				So(msg2, ShouldBeNil)
			})
		})
	})

	Convey("Given I start a websocket server with no push dispatching", t, func() {

		mux := bone.New()
		cfg := config{}

		wss := newPushServer(cfg, mux, pf)

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		Convey("When I start it", func() {

			out := make(chan bool)
			go func() {
				wss.start(ctx)
				out <- true
			}()

			Convey("Then it be running", func() {
				So(
					func() {
						select {
						case <-out:
							panic("test: unexpected response")
						case <-time.After(1 * time.Second):
						}
					},
					ShouldNotPanic,
				)
			})

			Convey("When I stop it", func() {

				cancel()

				var exited bool
				select {
				case exited = <-out:
				case <-time.After(1 * time.Second):
					panic("test: no respons in time")
				}

				Convey("Then the server should should exit", func() {
					So(exited, ShouldBeTrue)
				})
			})
		})
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

			ws, resp, err := wsc.Connect(ctx, strings.Replace(ts.URL, "http://", "ws://", 1), wsc.Config{})
			defer ws.Close(0) // nolint

			Convey("Then err should should be nil", func() {
				So(err, ShouldBeNil)
			})

			Convey("Then resp should should be correct", func() {
				So(resp.Status, ShouldEqual, "101 Switching Protocols")
			})
		})

		Convey("When I connect to the server but I am not authenticated", func() {

			authenticator.action = AuthActionKO

			pushHandler.Lock()
			pushHandler.onPushSessionInitOK = true
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
