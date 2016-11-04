// Author: Antoine Mercadal
// See LICENSE file for full LICENSE
// Copyright 2016 Aporeto.

package bahamut

import (
	"crypto/tls"
	"crypto/x509"
	"net/http"
	"net/http/pprof"

	"github.com/aporeto-inc/elemental"
	"github.com/go-zoo/bone"

	log "github.com/Sirupsen/logrus"
)

func corsHandler(w http.ResponseWriter, r *http.Request) {
	setCommonHeader(w, r.Header.Get("Origin"))
	w.WriteHeader(http.StatusOK)
}

func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	WriteHTTPError(w, r.Header.Get("Origin"), elemental.NewError("Not Found", "Unable to find the requested resource", "bahamut", http.StatusNotFound))
}

// an apiServer is the structure serving the api routes.
type apiServer struct {
	config      APIServerConfig
	multiplexer *bone.Mux
}

// newAPIServer returns a new apiServer.
func newAPIServer(config APIServerConfig, multiplexer *bone.Mux) *apiServer {

	return &apiServer{
		config:      config,
		multiplexer: multiplexer,
	}
}

// createSecureHTTPServer returns the main HTTP Server.
//
// It will return an error if any.
func (a *apiServer) createSecureHTTPServer(address string) (*http.Server, error) {

	CAPool := a.config.TLSServerCAPool

	if CAPool == nil {
		var err error
		CAPool, err = x509.SystemCertPool()
		if err != nil {
			CAPool = x509.NewCertPool()
		}
	}

	tlsConfig := &tls.Config{
		Certificates:           a.config.TLSServerCertificates,
		ClientAuth:             a.config.TLSAuthType,
		ClientCAs:              a.config.TLSClientCAPool,
		RootCAs:                CAPool,
		SessionTicketsDisabled: true,
		MinVersion:             tls.VersionSSL30,
	}

	tlsConfig.BuildNameToCertificate()

	return &http.Server{
		Addr:      address,
		TLSConfig: tlsConfig,
	}, nil
}

// createSecureHTTPServer returns a insecure HTTP Server.
//
// It will return an error if any.
func (a *apiServer) createUnsecureHTTPServer(address string) (*http.Server, error) {

	return &http.Server{
		Addr: address,
	}, nil
}

// installRoutes installs all the routes declared in the APIServerConfig.
func (a *apiServer) installRoutes() {

	for _, route := range a.config.Routes {

		switch route.Method {
		case http.MethodHead:
			a.multiplexer.Head(route.Pattern, route.Handler)
		case http.MethodGet:
			a.multiplexer.Get(route.Pattern, route.Handler)
		case http.MethodPost:
			a.multiplexer.Post(route.Pattern, route.Handler)
		case http.MethodPut:
			a.multiplexer.Put(route.Pattern, route.Handler)
		case http.MethodDelete:
			a.multiplexer.Delete(route.Pattern, route.Handler)
		case http.MethodPatch:
			a.multiplexer.Patch(route.Pattern, route.Handler)
		}

		log.WithFields(log.Fields{
			"pattern": route.Pattern,
			"method":  route.Method,
			"package": "bahamut",
		}).Debug("API route installed.")
	}

	a.multiplexer.Options("*", http.HandlerFunc(corsHandler))
	a.multiplexer.Get("/", http.HandlerFunc(corsHandler))
	a.multiplexer.NotFound(http.HandlerFunc(notFoundHandler))
}

func (a *apiServer) startProfilingServer() {

	log.WithFields(log.Fields{
		"address": a.config.ProfilingListenAddress,
		"package": "bahamut",
	}).Info("Starting profiling server.")

	srv, err := a.createUnsecureHTTPServer(a.config.ProfilingListenAddress)
	if err != nil {
		log.WithFields(log.Fields{
			"error":   err,
			"package": "bahamut",
		}).Fatal("Unable to create profiling server.")
	}

	mux := bone.New()
	mux.Handle("/debug/pprof/", http.HandlerFunc(pprof.Index))
	mux.Handle("/debug/pprof/cmdline", http.HandlerFunc(pprof.Cmdline))
	mux.Handle("/debug/pprof/profile", http.HandlerFunc(pprof.Profile))
	mux.Handle("/debug/pprof/symbol", http.HandlerFunc(pprof.Symbol))
	mux.Handle("/debug/pprof/trace", http.HandlerFunc(pprof.Trace))

	srv.Handler = mux
	if err := srv.ListenAndServe(); err != nil {
		log.WithFields(log.Fields{
			"error":   err,
			"package": "bahamut",
		}).Fatal("Unable to start profiling http server.")
	}
}

// start starts the apiServer.
func (a *apiServer) start() {

	if a.config.EnableProfiling {
		go a.startProfilingServer()
	}

	a.installRoutes()

	log.WithFields(log.Fields{
		"address": a.config.ListenAddress,
		"package": "bahamut",
		"routes":  len(a.config.Routes),
	}).Info("Starting api server.")

	var err error
	var server *http.Server
	if a.config.TLSServerCertificates != nil {
		server, err = a.createSecureHTTPServer(a.config.ListenAddress)
	} else {
		server, err = a.createUnsecureHTTPServer(a.config.ListenAddress)
	}
	if err != nil {
		log.WithFields(log.Fields{
			"error":   err,
			"package": "bahamut",
		}).Fatal("Unable to create api server.")
	}

	server.Handler = a.multiplexer
	server.SetKeepAlivesEnabled(true)

	if a.config.TLSServerCertificates != nil {
		err = server.ListenAndServeTLS("", "")
	} else {
		err = server.ListenAndServe()
	}

	if err != nil {
		log.WithFields(log.Fields{
			"error":   err,
			"package": "bahamut",
		}).Fatal("Unable to start api server.")
	}
}

// stop stops the apiServer.
//
// In reality right now, it does nothing :).
func (a *apiServer) stop() {

}
