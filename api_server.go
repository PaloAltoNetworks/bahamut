// Author: Antoine Mercadal
// See LICENSE file for full LICENSE
// Copyright 2016 Aporeto.

package bahamut

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"net/http"
	"net/http/pprof"

	log "github.com/Sirupsen/logrus"
	"github.com/aporeto-inc/elemental"
	"github.com/go-zoo/bone"
)

type apiServer struct {
	config      APIServerConfig
	address     string
	multiplexer *bone.Mux
}

func newAPIServer(config APIServerConfig, multiplexer *bone.Mux) *apiServer {

	return &apiServer{
		config:      config,
		multiplexer: multiplexer,
	}
}

func (a *apiServer) isTLSEnabled() bool {

	return a.config.TLSCAPath != "" && a.config.TLSCertificatePath != "" && a.config.TLSKeyPath != ""
}

func (a *apiServer) createSecureHTTPServer() (*http.Server, error) {

	caCert, err := ioutil.ReadFile(a.config.TLSCAPath)
	if err != nil {
		return nil, err
	}

	clientCertPool := x509.NewCertPool()
	clientCertPool.AppendCertsFromPEM(caCert)

	tlsConfig := &tls.Config{
		ClientAuth: tls.RequireAndVerifyClientCert,
		ClientCAs:  clientCertPool,
	}

	tlsConfig.BuildNameToCertificate()

	httpServer := &http.Server{
		Addr:      a.config.ListenAddress,
		TLSConfig: tlsConfig,
	}

	return httpServer, nil
}

func (a *apiServer) createUnsecureHTTPServer() (*http.Server, error) {

	httpServer := &http.Server{
		Addr: a.config.ListenAddress,
	}

	return httpServer, nil
}

func (a *apiServer) installRoutes() {

	for _, route := range a.config.Routes {

		if route.Method == http.MethodHead {
			a.multiplexer.Head(route.Pattern, http.HandlerFunc(route.Handler))
		} else if route.Method == http.MethodGet {
			a.multiplexer.Get(route.Pattern, http.HandlerFunc(route.Handler))
		} else if route.Method == http.MethodPost {
			a.multiplexer.Post(route.Pattern, http.HandlerFunc(route.Handler))
		} else if route.Method == http.MethodPut {
			a.multiplexer.Put(route.Pattern, http.HandlerFunc(route.Handler))
		} else if route.Method == http.MethodDelete {
			a.multiplexer.Delete(route.Pattern, http.HandlerFunc(route.Handler))
		} else if route.Method == http.MethodPatch {
			a.multiplexer.Patch(route.Pattern, http.HandlerFunc(route.Handler))
		}

		log.WithFields(log.Fields{
			"pattern": route.Pattern,
			"method":  route.Method,
		}).Debug("api route installed")
	}

	a.multiplexer.Options("*", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		setCommonHeader(w)
		w.WriteHeader(http.StatusOK)
	}))

	a.multiplexer.Get("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		setCommonHeader(w)
		w.WriteHeader(http.StatusOK)
	}))

	a.multiplexer.NotFound(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		WriteHTTPError(w, http.StatusNotFound, elemental.NewError("Not Found", "Unable to find the requested resource", "http", http.StatusNotFound))
	}))

	if a.config.EnableProfiling {
		a.multiplexer.Handle("/debug/pprof/", http.HandlerFunc(pprof.Index))
		a.multiplexer.Handle("/debug/pprof/cmdline", http.HandlerFunc(pprof.Cmdline))
		a.multiplexer.Handle("/debug/pprof/profile", http.HandlerFunc(pprof.Profile))
		a.multiplexer.Handle("/debug/pprof/symbol", http.HandlerFunc(pprof.Symbol))
		a.multiplexer.Handle("/debug/pprof/trace", http.HandlerFunc(pprof.Trace))

		log.Warn("profiling route installed")
	}

	log.WithFields(log.Fields{
		"routes": len(a.multiplexer.Routes),
	}).Info("all routes installed")
}

func (a *apiServer) start() {

	a.installRoutes()

	var server *http.Server
	var err error

	if a.isTLSEnabled() {

		log.WithFields(log.Fields{
			"endpoint": a.config.ListenAddress,
		}).Info("creating secure http server.")

		server, err = a.createSecureHTTPServer()
		if err != nil {
			log.WithFields(log.Fields{
				"error": err,
			}).Fatal("unable to create secure http server")
		}

		http.Handle("/", a.multiplexer)
		err = server.ListenAndServeTLS(a.config.TLSCertificatePath, a.config.TLSKeyPath)

		if err != nil {
			log.WithFields(log.Fields{
				"error": err,
			}).Fatal("unable to start secure http server")
		}

	} else {
		log.WithFields(log.Fields{
			"endpoint": a.config.ListenAddress,
		}).Warn("creating unsecure http server")

		server, err = a.createUnsecureHTTPServer()
		if err != nil {
			log.WithFields(log.Fields{
				"error": err,
			}).Fatal("unable to create unsecure http server")
		}

		http.Handle("/", a.multiplexer)
		err = server.ListenAndServe()

		if err != nil {
			log.WithFields(log.Fields{
				"error": err,
			}).Fatal("unable to start unsecure http server")
		}
	}
}

func (a *apiServer) stop() {

}
