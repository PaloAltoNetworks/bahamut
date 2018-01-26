package bahamut

// import (
// 	. "github.com/smartystreets/goconvey/convey"
// )

// func TestWebsocketSession_newWSSession(t *testing.T) {

// 	Convey("Given create a new wsSession", t, func() {

// 		u, _ := url.Parse("http://toto.com?a=b")
// 		conf := Config{}
// 		req := &http.Request{
// 			Header:     http.Header{"h1": {"a"}},
// 			URL:        u,
// 			TLS:        &tls.ConnectionState{},
// 			RemoteAddr: "1.2.3.4",
// 		}
// 		unregister := func(i internalWSSession) {}
// 		span := opentracing.StartSpan("test")

// 		s := newWSSession(req, conf, unregister, span)

// 		Convey("Then s should be correctly initialized", func() {
// 			So(s.claims, ShouldResemble, []string{})
// 			So(s.claimsMap, ShouldResemble, map[string]string{})
// 			So(s.config, ShouldResemble, conf)
// 			So(s.headers, ShouldEqual, req.Header)
// 			So(s.id, ShouldNotBeEmpty)
// 			So(s.parameters, ShouldResemble, url.Values{"a": {"b"}})
// 			So(s.closeCh, ShouldHaveSameTypeAs, make(chan struct{}))
// 			So(s.unregister, ShouldEqual, unregister)
// 			So(s.context, ShouldNotBeNil)
// 			So(s.cancel, ShouldNotBeNil)
// 			So(s.tlsConnectionState, ShouldEqual, req.TLS)
// 			So(s.remoteAddr, ShouldEqual, req.RemoteAddr)
// 		})
// 	})
// }

// func TestWebsocketSession_accessors(t *testing.T) {

// 	Convey("Given create a new wsSession", t, func() {

// 		u, _ := url.Parse("http://toto.com?a=b&token=token")
// 		conf := Config{}
// 		req := &http.Request{
// 			Header:     http.Header{"h1": {"a"}},
// 			URL:        u,
// 			TLS:        &tls.ConnectionState{},
// 			RemoteAddr: "1.2.3.4",
// 		}
// 		unregister := func(i internalWSSession) {}
// 		span := opentracing.StartSpan("test")

// 		s := newWSSession(req, conf, unregister, span)

// 		Convey("When I call Identifier()", func() {

// 			id := s.Identifier()

// 			Convey("Then id should be correct", func() {
// 				So(id, ShouldNotBeEmpty)
// 			})
// 		})

// 		Convey("When I call SetClaims()", func() {

// 			s.SetClaims([]string{"a=a", "b=b"})

// 			Convey("Then GetClaims() should return the correct claims ", func() {
// 				So(s.GetClaims(), ShouldResemble, []string{"a=a", "b=b"})
// 			})

// 			Convey("Then GetClaimsMap() should return the correct claims ", func() {
// 				m := s.GetClaimsMap()
// 				So(len(m), ShouldEqual, 2)
// 				So(m["a"], ShouldEqual, "a")
// 				So(m["b"], ShouldEqual, "b")
// 			})
// 		})

// 		Convey("When I call GetToken()", func() {

// 			token := s.GetToken()

// 			Convey("Then token should be correct", func() {
// 				So(token, ShouldEqual, "token")
// 			})
// 		})

// 		Convey("When I call TLSConnectionState()", func() {

// 			s := s.TLSConnectionState()

// 			Convey("Then TLSConnectionState should be correct", func() {
// 				So(s, ShouldEqual, req.TLS)
// 			})
// 		})

// 		Convey("When I call SetMetadata()", func() {

// 			s.SetMetadata("hi")

// 			Convey("Then GetMetadata() should return the correct metadata ", func() {
// 				So(s.GetMetadata(), ShouldResemble, "hi")
// 			})
// 		})

// 		Convey("When I call GetParameter()", func() {

// 			p := s.GetParameter("a")

// 			Convey("Then parameter should be correct", func() {
// 				So(p, ShouldEqual, "b")
// 			})
// 		})

// 		Convey("When I call Span()", func() {

// 			s := s.Span()

// 			Convey("Then span should be correct", func() {
// 				So(s, ShouldResemble, span)
// 			})
// 		})

// 		Convey("When I call NewChildSpan()", func() {

// 			s := s.NewChildSpan("new")

// 			Convey("Then span should be correct", func() {
// 				So(s, ShouldNotBeEmpty)
// 				So(s, ShouldHaveSameTypeAs, span)
// 			})
// 		})

// 		Convey("When I call setRemoteAddress()", func() {

// 			s.setRemoteAddress("a.b.c.d")

// 			Convey("Then address should be correct", func() {
// 				So(s.remoteAddr, ShouldEqual, "a.b.c.d")
// 			})
// 		})

// 		Convey("When I call setTLSConnectionState()", func() {

// 			tcs := &tls.ConnectionState{}
// 			s.setTLSConnectionState(tcs)

// 			Convey("Then address should be correct", func() {
// 				So(s.tlsConnectionState, ShouldEqual, tcs)
// 			})
// 		})

// 		Convey("When I call setSocket()", func() {

// 			ws := &websocket.Conn{}
// 			s.setConn(ws)

// 			Convey("Then ws should be correct", func() {
// 				So(s.conn, ShouldEqual, ws)
// 			})
// 		})
// 	})
// }

// func TestWebsocketSession_close(t *testing.T) {

// 	Convey("Given create a new wsSession", t, func() {

// 		u, _ := url.Parse("http://toto.com?a=b&token=token")
// 		conf := Config{}
// 		req := &http.Request{
// 			Header:     http.Header{"h1": {"a"}},
// 			URL:        u,
// 			TLS:        &tls.ConnectionState{},
// 			RemoteAddr: "1.2.3.4",
// 		}
// 		var unregisterCalled int
// 		unregister := func(i internalWSSession) { unregisterCalled++ }
// 		span := opentracing.StartSpan("test")

// 		s := newWSSession(req, conf, unregister, span)

// 		Convey("When I call stop", func() {
// 			s.stop()

// 			Convey("Then the session should be closed", func() {
// 				So(s.closeCh, ShouldBeNil)
// 				So(unregisterCalled, ShouldEqual, 1)
// 			})

// 			Convey("When I stop it again", func() {

// 				Convey("Then it should not panic", func() {
// 					So(func() { s.stop() }, ShouldNotPanic)
// 					So(unregisterCalled, ShouldEqual, 1)
// 				})
// 			})
// 		})

// 	})
// }

// func TestWebsocketSession_listen(t *testing.T) {

// 	Convey("Given create a new wsSession", t, func() {

// 		u, _ := url.Parse("http://toto.com?a=b&token=token")
// 		conf := Config{}
// 		req := &http.Request{
// 			Header:     http.Header{"h1": {"a"}},
// 			URL:        u,
// 			TLS:        &tls.ConnectionState{},
// 			RemoteAddr: "1.2.3.4",
// 		}
// 		unregister := func(i internalWSSession) {}
// 		span := opentracing.StartSpan("test")

// 		s := newWSSession(req, conf, unregister, span)

// 		Convey("When I call listen", func() {

// 			s.listen()

// 			Convey("Then nothing should happen", func() {
// 				So(true, ShouldBeTrue)
// 			})
// 		})

// 	})
// }
