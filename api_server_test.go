// Author: Antoine Mercadal
// See LICENSE file for full LICENSE
// Copyright 2016 Aporeto.

package bahamut

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-zoo/bone"
	. "github.com/smartystreets/goconvey/convey"
)

func TestServer_corsHandler(t *testing.T) {

	Convey("Given I call the corsHandler", t, func() {

		w := httptest.NewRecorder()
		corsHandler(w, nil)

		Convey("Then the response should be correct", func() {
			So(w.Code, ShouldEqual, http.StatusOK)
		})
	})
}

func TestServer_notFoundHandler(t *testing.T) {

	Convey("Given I call the notFoundHandler", t, func() {

		w := httptest.NewRecorder()
		notFoundHandler(w, nil)

		Convey("Then the response should be correct", func() {
			So(w.Code, ShouldEqual, http.StatusNotFound)
		})
	})
}

func TestServer_Initialization(t *testing.T) {

	Convey("Given I create a new api server", t, func() {

		cfg := MakeAPIServerConfig("address:80", "", "", "", []*Route{}, nil, "", "/h")
		c := newAPIServer(cfg, bone.New())

		Convey("Then it should be correctly initialized", func() {
			So(len(c.multiplexer.Routes), ShouldEqual, 0)
			So(c.config, ShouldResemble, cfg)
		})
	})
}

func TestServer_isTLSEnabled(t *testing.T) {

	Convey("Given I create a new api server without any tls info", t, func() {

		cfg := MakeAPIServerConfig("address:80", "", "", "", []*Route{}, nil, "", "/h")
		c := newAPIServer(cfg, bone.New())

		Convey("Then TLS should not be active", func() {
			So(c.isTLSEnabled(), ShouldBeFalse)
		})
	})

	Convey("Given I create a new api server without all tls info", t, func() {

		cfg := MakeAPIServerConfig("address:80", "a", "b", "c", []*Route{}, nil, "", "/h")
		c := newAPIServer(cfg, bone.New())

		Convey("Then TLS should be active", func() {
			So(c.isTLSEnabled(), ShouldBeTrue)
		})
	})
}

func TestServer_createSecureHTTPServer(t *testing.T) {

	Convey("Given I create a new api server without all valid tls info", t, func() {

		cfg := MakeAPIServerConfig("address:80", "fixtures/ca.pem", "fixtures/cert.pem", "fixtures/key.pem", []*Route{}, nil, "", "/h")
		c := newAPIServer(cfg, bone.New())

		Convey("When I make a secure server", func() {
			srv, err := c.createSecureHTTPServer(cfg.ListenAddress)

			Convey("Then error should be nil", func() {
				So(err, ShouldBeNil)
			})

			Convey("Then the server should be correctly initialized", func() {
				So(srv, ShouldNotBeNil)
			})
		})
	})

	Convey("Given I create a new api server without invalid ca info", t, func() {

		cfg := MakeAPIServerConfig("address:80", "fixtures/nope.pem", "fixtures/cert.pem", "fixtures/key.pem", []*Route{}, nil, "", "/h")
		c := newAPIServer(cfg, bone.New())

		Convey("When I make a secure server", func() {
			srv, err := c.createSecureHTTPServer(cfg.ListenAddress)

			Convey("Then error should not be nil", func() {
				So(err, ShouldNotBeNil)
			})

			Convey("Then the server should be nil", func() {
				So(srv, ShouldBeNil)
			})
		})
	})
}

func TestServer_createUnsecureHTTPServer(t *testing.T) {

	Convey("Given I create a new api server without any tls info", t, func() {

		cfg := MakeAPIServerConfig("address:80", "", "", "", []*Route{}, nil, "", "/h")
		c := newAPIServer(cfg, bone.New())

		Convey("When I make an unsecure server", func() {
			srv, err := c.createUnsecureHTTPServer(cfg.ListenAddress)

			Convey("Then error should be nil", func() {
				So(err, ShouldBeNil)
			})

			Convey("Then the server should be correctly initialized", func() {
				So(srv, ShouldNotBeNil)
			})
		})
	})
}

func TestServer_RouteInstallation(t *testing.T) {

	Convey("Given I create a new api server with routes", t, func() {

		h := func(w http.ResponseWriter, req *http.Request) {}

		var routes []*Route
		routes = append(routes, NewRoute("/lists", http.MethodPost, h))
		routes = append(routes, NewRoute("/lists", http.MethodGet, h))
		routes = append(routes, NewRoute("/lists", http.MethodDelete, h))
		routes = append(routes, NewRoute("/lists", http.MethodPatch, h))
		routes = append(routes, NewRoute("/lists", http.MethodHead, h))
		routes = append(routes, NewRoute("/lists", http.MethodPut, h))

		cfg := MakeAPIServerConfig("address:80", "", "", "", routes, nil, "", "/h")
		cfg.EnableProfiling = true
		c := newAPIServer(cfg, bone.New())

		Convey("When I install the routes", func() {

			c.installRoutes()

			Convey("Then the bone Multiplexer should have correct number of handlers", func() {
				So(len(c.multiplexer.Routes[http.MethodPost]), ShouldEqual, 5)
				So(len(c.multiplexer.Routes[http.MethodGet]), ShouldEqual, 6)
				So(len(c.multiplexer.Routes[http.MethodDelete]), ShouldEqual, 5)
				So(len(c.multiplexer.Routes[http.MethodPatch]), ShouldEqual, 5)
				So(len(c.multiplexer.Routes[http.MethodHead]), ShouldEqual, 5)
				So(len(c.multiplexer.Routes[http.MethodPut]), ShouldEqual, 5)
				So(len(c.multiplexer.Routes[http.MethodOptions]), ShouldEqual, 5)
			})
		})
	})
}

func TestServer_Start(t *testing.T) {

	Convey("Given I create an api without tls server", t, func() {

		Convey("When I start the server", func() {

			h := func(w http.ResponseWriter, req *http.Request) { w.Write([]byte("hello")) }
			routes := []*Route{NewRoute("/hello", http.MethodGet, h)}

			cfg := MakeAPIServerConfig("127.0.0.1:3123", "", "", "", routes, func(w http.ResponseWriter, r *http.Request) {}, "", "/h")
			c := newAPIServer(cfg, bone.New())

			go c.start()
			time.Sleep(1 * time.Second)

			resp, err := http.Get("http://127.0.0.1:3123")

			Convey("Then the status code should be 200", func() {
				So(err, ShouldBeNil)
				So(resp.StatusCode, ShouldEqual, 200)
			})
		})
	})
}
