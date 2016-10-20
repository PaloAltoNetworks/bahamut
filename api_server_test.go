// Author: Antoine Mercadal
// See LICENSE file for full LICENSE
// Copyright 2016 Aporeto.

package bahamut

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/go-zoo/bone"
	. "github.com/smartystreets/goconvey/convey"
)

func TestServer_corsHandler(t *testing.T) {

	Convey("Given I call the corsHandler", t, func() {

		h := http.Header{}
		h.Add("Origin", "toto")

		w := httptest.NewRecorder()
		corsHandler(w, &http.Request{Header: h})

		Convey("Then the response should be correct", func() {
			So(w.Code, ShouldEqual, http.StatusOK)
		})
	})
}

func TestServer_notFoundHandler(t *testing.T) {

	Convey("Given I call the notFoundHandler", t, func() {

		h := http.Header{}
		h.Add("Origin", "toto")

		w := httptest.NewRecorder()
		notFoundHandler(w, &http.Request{Header: h})

		Convey("Then the response should be correct", func() {
			So(w.Code, ShouldEqual, http.StatusNotFound)
		})
	})
}

func TestServer_Initialization(t *testing.T) {

	Convey("Given I create a new api server", t, func() {

		cfg := APIServerConfig{
			ListenAddress: "address:80",
			Routes:        []*Route{},
		}
		c := newAPIServer(cfg, bone.New())

		Convey("Then it should be correctly initialized", func() {
			So(len(c.multiplexer.Routes), ShouldEqual, 0)
			So(c.config, ShouldResemble, cfg)
		})
	})
}

func TestServer_isTLSEnabled(t *testing.T) {

	Convey("Given I create a new api server without any tls info", t, func() {

		cfg := APIServerConfig{
			ListenAddress: "address:80",
			Routes:        []*Route{},
		}

		c := newAPIServer(cfg, bone.New())

		Convey("Then TLS should not be active", func() {
			So(c.isTLSEnabled(), ShouldBeFalse)
		})
	})

	Convey("Given I create a new api server without all tls info", t, func() {

		cfg := APIServerConfig{
			ListenAddress:      "address:80",
			Routes:             []*Route{},
			TLSCAPath:          "a",
			TLSCertificatePath: "b",
			TLSKeyPath:         "c",
		}

		c := newAPIServer(cfg, bone.New())

		Convey("Then TLS should be active", func() {
			So(c.isTLSEnabled(), ShouldBeTrue)
		})
	})
}

func TestServer_createSecureHTTPServer(t *testing.T) {

	Convey("Given I create a new api server without all valid tls info", t, func() {

		cfg := APIServerConfig{
			ListenAddress:      "address:80",
			Routes:             []*Route{},
			TLSCAPath:          "fixtures/ca.pem",
			TLSCertificatePath: "fixtures/server-cert.pem",
			TLSKeyPath:         "fixtures/server-key.pem",
			TLSAuthType:        tls.RequireAndVerifyClientCert,
		}

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

		cfg := APIServerConfig{
			ListenAddress:      "address:80",
			Routes:             []*Route{},
			TLSCAPath:          "fixtures/nope.pem",
			TLSCertificatePath: "fixtures/server-cert.pem",
			TLSKeyPath:         "fixtures/server-key.pem",
			TLSAuthType:        tls.RequireAndVerifyClientCert,
		}

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

		cfg := APIServerConfig{
			ListenAddress: "address:80",
			Routes:        []*Route{},
		}
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

		cfg := APIServerConfig{
			ListenAddress:          "address:80",
			ProfilingListenAddress: "address:3434",
			Routes:                 routes,
			EnableProfiling:        true,
		}

		c := newAPIServer(cfg, bone.New())

		Convey("When I install the routes", func() {

			c.installRoutes()

			Convey("Then the bone Multiplexer should have correct number of handlers", func() {
				So(len(c.multiplexer.Routes[http.MethodPost]), ShouldEqual, 1)
				So(len(c.multiplexer.Routes[http.MethodGet]), ShouldEqual, 2)
				So(len(c.multiplexer.Routes[http.MethodDelete]), ShouldEqual, 1)
				So(len(c.multiplexer.Routes[http.MethodPatch]), ShouldEqual, 1)
				So(len(c.multiplexer.Routes[http.MethodHead]), ShouldEqual, 1)
				So(len(c.multiplexer.Routes[http.MethodPut]), ShouldEqual, 1)
				So(len(c.multiplexer.Routes[http.MethodOptions]), ShouldEqual, 1)
			})
		})
	})
}

func TestServer_Start(t *testing.T) {

	// yeah, well, until Go provides a way to stop an http server...
	rand.Seed(time.Now().UTC().UnixNano())

	Convey("Given I create an api without tls server", t, func() {

		Convey("When I start the server", func() {

			port1 := strconv.Itoa(rand.Intn(10000) + 20000)
			port2 := strconv.Itoa(rand.Intn(10000) + 30000)

			cfg := APIServerConfig{
				ListenAddress:          "127.0.0.1:" + port1,
				Routes:                 []*Route{},
				EnableProfiling:        true,
				ProfilingListenAddress: "127.0.0.1:" + port2,
			}

			c := newAPIServer(cfg, bone.New())

			go c.start()
			time.Sleep(1 * time.Second)

			resp, err := http.Get("http://127.0.0.1:" + port1)

			Convey("Then the status code should be OK", func() {
				So(err, ShouldBeNil)
				So(resp.StatusCode, ShouldEqual, 200)
			})
		})
	})

	Convey("Given I create an api with tls server", t, func() {

		Convey("When I start the server", func() {

			port1 := strconv.Itoa(rand.Intn(10000) + 40000)
			port2 := strconv.Itoa(rand.Intn(10000) + 50000)

			h := func(w http.ResponseWriter, req *http.Request) { w.Write([]byte("hello")) }

			cfg := APIServerConfig{
				ListenAddress:          "127.0.0.1:" + port1,
				TLSCAPath:              "fixtures/ca.pem",
				TLSCertificatePath:     "fixtures/server-cert.pem",
				TLSKeyPath:             "fixtures/server-key.pem",
				Routes:                 []*Route{NewRoute("/hello", http.MethodGet, h)},
				EnableProfiling:        true,
				ProfilingListenAddress: "127.0.0.1:" + port2,
			}

			c := newAPIServer(cfg, bone.New())

			go c.start()
			time.Sleep(1 * time.Second)

			cert, _ := tls.LoadX509KeyPair("fixtures/client-cert.pem", "fixtures/client-key.pem")
			cacert, _ := ioutil.ReadFile("fixtures/ca.pem")
			pool := x509.NewCertPool()
			pool.AppendCertsFromPEM(cacert)
			tlsConfig := &tls.Config{
				Certificates: []tls.Certificate{cert},
				RootCAs:      pool,
			}
			tlsConfig.BuildNameToCertificate()
			transport := &http.Transport{TLSClientConfig: tlsConfig}
			client := &http.Client{Transport: transport}

			resp, err := client.Get("https://localhost:" + port1)

			Convey("Then the status code should be 200", func() {
				So(err, ShouldBeNil)
				So(resp.StatusCode, ShouldEqual, 200)
			})
		})
	})
}
