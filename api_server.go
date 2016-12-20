// Author: Antoine Mercadal
// See LICENSE file for full LICENSE
// Copyright 2016 Aporeto.

package bahamut

import (
	"crypto/tls"
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
	server      *http.Server
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

	tlsConfig := &tls.Config{
		Certificates:             a.config.TLSServerCertificates,
		ClientAuth:               a.config.TLSAuthType,
		ClientCAs:                a.config.TLSClientCAPool,
		RootCAs:                  a.config.TLSRootCAPool,
		MinVersion:               tls.VersionTLS12,
		SessionTicketsDisabled:   true,
		PreferServerCipherSuites: true,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			// tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305, // Uncomment with Go 1.8
			// tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,   // Uncomment with Go 1.8
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
		},
	}

	tlsConfig.BuildNameToCertificate()

	return &http.Server{
		Addr:         address,
		TLSConfig:    tlsConfig,
		ReadTimeout:  a.config.ReadTimeout,
		WriteTimeout: a.config.WriteTimeout,
		// IdleTimeout:  a.config.IdleTimeout, // Uncomment with Go 1.8
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
	if a.config.TLSServerCertificates != nil {
		a.server, err = a.createSecureHTTPServer(a.config.ListenAddress)
	} else {
		a.server, err = a.createUnsecureHTTPServer(a.config.ListenAddress)
	}
	if err != nil {
		log.WithFields(log.Fields{
			"error":   err,
			"package": "bahamut",
		}).Fatal("Unable to create api server.")
	}

	a.server.Handler = a.multiplexer
	a.server.SetKeepAlivesEnabled(true)

	if a.config.TLSServerCertificates != nil {
		err = a.server.ListenAndServeTLS("", "")
	} else {
		err = a.server.ListenAndServe()
	}

	if err != nil {
		log.WithFields(log.Fields{
			"error":   err,
			"package": "bahamut",
		}).Fatal("Unable to start api server.")
	}
}

// stop stops the apiServer.
func (a *apiServer) stop() {

	// a.server.Shutdown() // Uncomment with Go 1.8
	// a.server = nil
}
