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

		cfg := config{}
		cfg.restServer.listenAddress = "address:80"

		c := newRestServer(cfg, bone.New(), nil, nil)

		Convey("Then it should be correctly initialized", func() {
			So(len(c.multiplexer.Routes), ShouldEqual, 0)
			So(c.cfg, ShouldResemble, cfg)
		})
	})
}

func TestServer_createSecureHTTPServer(t *testing.T) {

	Convey("Given I create a new api server without all valid tls info", t, func() {

		_, clientcapool, servercerts := loadFixtureCertificates()

		cfg := config{}
		cfg.restServer.listenAddress = "address:80"
		cfg.tls.clientCAPool = clientcapool
		cfg.tls.serverCertificates = servercerts
		cfg.tls.authType = tls.RequireAndVerifyClientCert

		c := newRestServer(cfg, bone.New(), nil, nil)

		Convey("When I make a secure server", func() {
			srv := c.createSecureHTTPServer(cfg.restServer.listenAddress)

			Convey("Then the server should be correctly initialized", func() {
				So(srv, ShouldNotBeNil)
			})
		})
	})

	Convey("Given I create a new api server without all custom tls cert func", t, func() {

		r := func(*tls.ClientHelloInfo) (*tls.Certificate, error) { return nil, nil }

		cfg := config{}
		cfg.restServer.listenAddress = "address:80"
		cfg.tls.serverCertificatesRetrieverFunc = r
		c := newRestServer(cfg, bone.New(), nil, nil)

		Convey("When I make a secure server", func() {
			srv := c.createSecureHTTPServer(cfg.restServer.listenAddress)

			Convey("Then the server should be correctly initialized", func() {
				So(srv.TLSConfig.GetCertificate, ShouldEqual, r)
			})
		})
	})
}

func TestServer_createUnsecureHTTPServer(t *testing.T) {

	Convey("Given I create a new api server without any tls info", t, func() {

		cfg := config{}
		cfg.restServer.listenAddress = "address:80"

		c := newRestServer(cfg, bone.New(), nil, nil)

		Convey("When I make an unsecure server", func() {
			srv := c.createUnsecureHTTPServer(cfg.restServer.listenAddress)

			Convey("Then the server should be correctly initialized", func() {
				So(srv, ShouldNotBeNil)
			})
		})
	})
}

func TestServer_RouteInstallation(t *testing.T) {

	Convey("Given I create a new api server with routes", t, func() {

		routes := map[int][]RouteInfo{
			1: {
				{
					URL:   "/a",
					Verbs: []string{"GET"},
				},
			},
			2: {
				{
					URL:   "/b",
					Verbs: []string{"POST"},
				},
				{
					URL:   "/c/d",
					Verbs: []string{"POST", "GET"},
				},
			},
		}

		cfg := config{}
		cfg.restServer.customRootHandlerFunc = http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})
		cfg.restServer.listenAddress = "address:80"
		cfg.meta.serviceName = "hello"
		cfg.meta.version = map[string]interface{}{}

		c := newRestServer(cfg, bone.New(), nil, nil)

		Convey("When I install the routes", func() {

			c.installRoutes(routes)

			Convey("Then the bone Multiplexer should have correct number of handlers", func() {
				So(len(c.multiplexer.Routes[http.MethodPost]), ShouldEqual, 5)
				So(len(c.multiplexer.Routes[http.MethodGet]), ShouldEqual, 10)
				So(len(c.multiplexer.Routes[http.MethodDelete]), ShouldEqual, 3)
				So(len(c.multiplexer.Routes[http.MethodPatch]), ShouldEqual, 3)
				So(len(c.multiplexer.Routes[http.MethodHead]), ShouldEqual, 5)
				So(len(c.multiplexer.Routes[http.MethodPut]), ShouldEqual, 3)
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

			cfg := config{}
			cfg.restServer.listenAddress = "127.0.0.1:" + port1

			c := newRestServer(cfg, bone.New(), nil, nil)

			go c.start(context.TODO(), nil)
			time.Sleep(30 * time.Millisecond)

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

			_, clientcapool, servercerts := loadFixtureCertificates()

			cfg := config{}
			cfg.restServer.listenAddress = "127.0.0.1:" + port1
			cfg.tls.clientCAPool = clientcapool
			cfg.tls.serverCertificates = servercerts
			cfg.tls.authType = tls.RequireAndVerifyClientCert

			c := newRestServer(cfg, bone.New(), nil, nil)

			go c.start(context.TODO(), nil)
			time.Sleep(30 * time.Millisecond)

			cert, _ := tls.LoadX509KeyPair("fixtures/certs/client-cert.pem", "fixtures/certs/client-key.pem")
			cacert, _ := ioutil.ReadFile("fixtures/certs/ca-cert.pem")
			pool := x509.NewCertPool()
			pool.AppendCertsFromPEM(cacert)
			tlsConfig := &tls.Config{
				Certificates: []tls.Certificate{cert},
				RootCAs:      pool,
			}

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
