// Author: Antoine Mercadal
// See LICENSE file for full LICENSE
// Copyright 2016 Aporeto.

package bahamut

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/aporeto-inc/elemental"
	"github.com/go-zoo/bone"
	. "github.com/smartystreets/goconvey/convey"
)

func loadFixtureCertificates() (*x509.CertPool, *x509.CertPool, []tls.Certificate) {

	systemCAPool, _ := x509.SystemCertPool()

	clientCACertData, _ := ioutil.ReadFile("fixtures/ca.pem")
	clientCAPool := x509.NewCertPool()
	clientCAPool.AppendCertsFromPEM(clientCACertData)

	serverCert, _ := tls.LoadX509KeyPair("fixtures/server-cert.pem", "fixtures/server-key.pem")
	return systemCAPool, clientCAPool, []tls.Certificate{serverCert}
}

func TestServer_Initialization(t *testing.T) {

	Convey("Given I create a new api server", t, func() {

		cfg := Config{}
		cfg.ReSTServer.ListenAddress = "address:80"

		c := newAPIServer(cfg, bone.New())

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
		cfg.Model.RelationshipsRegistry = Relationships()
		cfg.Model.IdentifiablesFactory = IdentifiableForIdentity

		c := newAPIServer(cfg, bone.New())

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

		c := newAPIServer(cfg, bone.New())

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

func TestServer_Start(t *testing.T) {

	// yeah, well, until Go provides a way to stop an http server...
	rand.Seed(time.Now().UTC().UnixNano())

	Convey("Given I create an api without tls server", t, func() {

		Convey("When I start the server", func() {

			port1 := strconv.Itoa(rand.Intn(10000) + 20000)

			cfg := Config{}
			cfg.ReSTServer.ListenAddress = "127.0.0.1:" + port1

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

			syscapool, clientcapool, servercerts := loadFixtureCertificates()

			cfg := Config{}
			cfg.ReSTServer.ListenAddress = "127.0.0.1:" + port1
			cfg.TLS.RootCAPool = syscapool
			cfg.TLS.ClientCAPool = clientcapool
			cfg.TLS.ServerCertificates = servercerts
			cfg.TLS.AuthType = tls.RequireAndVerifyClientCert

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

func TestServer_handleRetrieve(t *testing.T) {

	// yeah, well, until Go provides a way to stop an http server...
	rand.Seed(time.Now().UTC().UnixNano())

	Convey("Given I create an api server", t, func() {

		Convey("When I start the server and call handleRetrieve", func() {

			port := strconv.Itoa(rand.Intn(10000) + 20000)
			listenAddress := "127.0.0.1:" + port

			cfg := Config{}
			cfg.ReSTServer.ListenAddress = listenAddress
			cfg.Model.RelationshipsRegistry = Relationships()
			cfg.Model.IdentifiablesFactory = IdentifiableForIdentity

			w := httptest.NewRecorder()
			// elReq := elemental.NewRequest()

			processorFinder := func(identity elemental.Identity) (Processor, error) {
				return &FakeCompleteProcessor{}, nil
			}

			c := newAPIServer(cfg, bone.New())
			c.processorFinder = processorFinder
			go c.start()
			time.Sleep(1 * time.Second)

			// Simulate action
			httptest.NewRequest("GET", "http://"+listenAddress+"/lists/x", nil)
			// c.handleRetrieve(w, req)

			resp := w.Result()

			Convey("Then the status code should be OK", func() {
				So(resp.StatusCode, ShouldEqual, 200)
			})
		})
	})

	Convey("Given I create an api server", t, func() {

		Convey("When I start the server and call handleRetrieve with unknown processor", func() {

			port := strconv.Itoa(rand.Intn(10000) + 20000)
			listenAddress := "127.0.0.1:" + port

			cfg := Config{}
			cfg.ReSTServer.ListenAddress = listenAddress
			cfg.Model.RelationshipsRegistry = Relationships()
			cfg.Model.IdentifiablesFactory = IdentifiableForIdentity

			processorFinder := func(identity elemental.Identity) (Processor, error) {
				return &FakeCompleteProcessor{}, nil
			}

			c := newAPIServer(cfg, bone.New())
			c.processorFinder = processorFinder
			go c.start()
			time.Sleep(3 * time.Second)

			// Simulate action
			resp, _ := http.Get("http://" + listenAddress + "/unknown/x")

			body, _ := ioutil.ReadAll(resp.Body)
			fmt.Println(resp.StatusCode)
			fmt.Println(string(body))

			Convey("Then the status code should be a 405, method not allowed", func() {
				So(resp.StatusCode, ShouldEqual, 405)
			})
		})
	})
}
