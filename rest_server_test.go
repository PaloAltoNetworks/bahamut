// Author: Antoine Mercadal
// See LICENSE file for full LICENSE
// Copyright 2016 Aporeto.

package bahamut

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/go-zoo/bone"
	. "github.com/smartystreets/goconvey/convey"
)

func loadFixtureCertificates() (*x509.CertPool, *x509.CertPool, []tls.Certificate) {

	systemCAPool, _ := x509.SystemCertPool()

	clientCACertData, _ := ioutil.ReadFile("fixtures/certs/ca-cert.pem")
	clientCAPool := x509.NewCertPool()
	clientCAPool.AppendCertsFromPEM(clientCACertData)

	serverCert, _ := tls.LoadX509KeyPair("fixtures/certs/server-cert.pem", "fixtures/certs/server-key.pem")
	return systemCAPool, clientCAPool, []tls.Certificate{serverCert}
}

func TestServer_Initialization(t *testing.T) {

	Convey("Given I create a new api server", t, func() {

		cfg := Config{}
		cfg.ReSTServer.ListenAddress = "address:80"

		c := newRestServer(cfg, bone.New(), nil, nil)

		Convey("Then it should be correctly initialized", func() {
			So(len(c.multiplexer.Routes), ShouldEqual, 0)
			So(c.config, ShouldResemble, cfg)
		})
	})
}

func TestServer_createSecureHTTPServer(t *testing.T) {

	Convey("Given I create a new api server without all valid tls info", t, func() {

		syscapool, clientcapool, servercerts := loadFixtureCertificates()

		cfg := Config{}
		cfg.ReSTServer.ListenAddress = "address:80"
		cfg.TLS.RootCAPool = syscapool
		cfg.TLS.ClientCAPool = clientcapool
		cfg.TLS.ServerCertificates = servercerts
		cfg.TLS.AuthType = tls.RequireAndVerifyClientCert

		c := newRestServer(cfg, bone.New(), nil, nil)

		Convey("When I make a secure server", func() {
			srv, err := c.createSecureHTTPServer(cfg.ReSTServer.ListenAddress)

			Convey("Then error should be nil", func() {
				So(err, ShouldBeNil)
			})

			Convey("Then the server should be correctly initialized", func() {
				So(srv, ShouldNotBeNil)
			})
		})
	})
}

func TestServer_createUnsecureHTTPServer(t *testing.T) {

	Convey("Given I create a new api server without any tls info", t, func() {

		cfg := Config{}
		cfg.ReSTServer.ListenAddress = "address:80"

		c := newRestServer(cfg, bone.New(), nil, nil)

		Convey("When I make an unsecure server", func() {
			srv, err := c.createUnsecureHTTPServer(cfg.ReSTServer.ListenAddress)

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

	// Convey("Given I create a new api server with routes", t, func() {
	//
	// 	h := func(w http.ResponseWriter, req *http.Request) {}
	//
	// 	var routes []*Route
	// 	routes = append(routes, NewRoute("/lists", http.MethodPost, h))
	// 	routes = append(routes, NewRoute("/lists", http.MethodGet, h))
	// 	routes = append(routes, NewRoute("/lists", http.MethodDelete, h))
	// 	routes = append(routes, NewRoute("/lists", http.MethodPatch, h))
	// 	routes = append(routes, NewRoute("/lists", http.MethodHead, h))
	// 	routes = append(routes, NewRoute("/lists", http.MethodPut, h))
	//
	// 	cfg := APIServerConfig{
	// 		ListenAddress:          "address:80",
	// 		ProfilingListenAddress: "address:3434",
	// 		Routes:                 routes,
	// 		EnableProfiling:        true,
	// 	}
	//
	// 	c := newAPIServer(cfg, bone.New())
	//
	// 	Convey("When I install the routes", func() {
	//
	// 		c.installRoutes()
	//
	// 		Convey("Then the bone Multiplexer should have correct number of handlers", func() {
	// 			So(len(c.multiplexer.Routes[http.MethodPost]), ShouldEqual, 1)
	// 			So(len(c.multiplexer.Routes[http.MethodGet]), ShouldEqual, 2)
	// 			So(len(c.multiplexer.Routes[http.MethodDelete]), ShouldEqual, 1)
	// 			So(len(c.multiplexer.Routes[http.MethodPatch]), ShouldEqual, 1)
	// 			So(len(c.multiplexer.Routes[http.MethodHead]), ShouldEqual, 1)
	// 			So(len(c.multiplexer.Routes[http.MethodPut]), ShouldEqual, 1)
	// 			So(len(c.multiplexer.Routes[http.MethodOptions]), ShouldEqual, 1)
	// 		})
	// 	})
	// })
}

func TestServer_Start(t *testing.T) {

	// yeah, well, until Go provides a way to stop an http server...
	rand.Seed(time.Now().UTC().UnixNano())

	Convey("Given I create an api without tls server", t, func() {

		Convey("When I start the server", func() {

			port1 := strconv.Itoa(rand.Intn(10000) + 20000)

			cfg := Config{}
			cfg.ReSTServer.ListenAddress = "127.0.0.1:" + port1

			c := newRestServer(cfg, bone.New(), nil, nil)

			go c.start(context.TODO())
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

			// h := func(w http.ResponseWriter, req *http.Request) { w.Write([]byte("hello")) }

			syscapool, clientcapool, servercerts := loadFixtureCertificates()

			cfg := Config{}
			cfg.ReSTServer.ListenAddress = "127.0.0.1:" + port1
			cfg.TLS.RootCAPool = syscapool
			cfg.TLS.ClientCAPool = clientcapool
			cfg.TLS.ServerCertificates = servercerts
			cfg.TLS.AuthType = tls.RequireAndVerifyClientCert

			c := newRestServer(cfg, bone.New(), nil, nil)

			go c.start(context.TODO())
			time.Sleep(1 * time.Second)

			cert, _ := tls.LoadX509KeyPair("fixtures/certs/client-cert.pem", "fixtures/certs/client-key.pem")
			cacert, _ := ioutil.ReadFile("fixtures/certs/ca-cert.pem")
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
