// Author: Antoine Mercadal
// See LICENSE file for full LICENSE
// Copyright 2016 Aporeto.

package bahamut

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/aporeto-inc/elemental"
	"github.com/go-zoo/bone"
	"go.uber.org/zap"
	"golang.org/x/crypto/acme/autocert"
)

// an restServer is the structure serving the api routes.
type restServer struct {
	config          Config
	multiplexer     *bone.Mux
	server          *http.Server
	processorFinder processorFinderFunc
	pusher          eventPusherFunc
	mainContext     context.Context
}

// newRestServer returns a new apiServer.
func newRestServer(config Config, multiplexer *bone.Mux, processorFinder processorFinderFunc, pusher eventPusherFunc) *restServer {

	return &restServer{
		config:          config,
		multiplexer:     multiplexer,
		processorFinder: processorFinder,
		pusher:          pusher,
	}
}

// createSecureHTTPServer returns the main HTTP Server.
//
// It will return an error if any.
func (a *restServer) createSecureHTTPServer(address string) (*http.Server, error) {

	tlsConfig := &tls.Config{
		ClientAuth:               a.config.TLS.AuthType,
		ClientCAs:                a.config.TLS.ClientCAPool,
		RootCAs:                  a.config.TLS.RootCAPool,
		MinVersion:               tls.VersionTLS12,
		SessionTicketsDisabled:   true,
		PreferServerCipherSuites: true,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
		},
	}

	if !a.config.TLS.EnableLetsEncrypt {

		// If letsencrypt is not enabled we simply set the given list of
		// certificates or the ServerCertificatesRetrieverFunc in the TLS option.

		if a.config.TLS.ServerCertificatesRetrieverFunc != nil {
			tlsConfig.GetCertificate = a.config.TLS.ServerCertificatesRetrieverFunc
		} else {
			tlsConfig.Certificates = a.config.TLS.ServerCertificates
		}

	} else {

		cachePath := a.config.TLS.LetsEncryptCertificateCacheFolder
		if cachePath == "" {
			cachePath = os.TempDir()
		}

		// Otherwise, we create an autocert manager
		m := autocert.Manager{
			Prompt:     autocert.AcceptTOS,
			HostPolicy: autocert.HostWhitelist(a.config.TLS.LetsEncryptDomainWhiteList...),
			Cache:      autocert.DirCache(cachePath),
		}

		// Then we build a custom GetCertificate function to first use the certificate passed
		// by the config, then eventually try to get a certificate from letsencrypt.
		localCertMap := buildNameAndIPsToCertificate(a.config.TLS.ServerCertificates)
		tlsConfig.GetCertificate = func(hello *tls.ClientHelloInfo) (*tls.Certificate, error) {
			if hello.ServerName != "" {
				if c, ok := localCertMap[hello.ServerName]; ok {
					return c, nil
				}
			} else {
				host, _, err := net.SplitHostPort(hello.Conn.LocalAddr().String())
				if err != nil {
					return nil, err
				}
				if c, ok := localCertMap[host]; ok {
					return c, nil
				}
			}
			return m.GetCertificate(hello)
		}
	}

	tlsConfig.BuildNameToCertificate()

	server := &http.Server{
		Addr:         address,
		TLSConfig:    tlsConfig,
		ReadTimeout:  a.config.ReSTServer.ReadTimeout,
		WriteTimeout: a.config.ReSTServer.WriteTimeout,
		IdleTimeout:  a.config.ReSTServer.IdleTimeout,
	}

	server.SetKeepAlivesEnabled(!a.config.ReSTServer.DisableKeepalive)

	return server, nil
}

// createUnsecureHTTPServer returns a insecure HTTP Server.
//
// It will return an error if any.
func (a *restServer) createUnsecureHTTPServer(address string) (*http.Server, error) {

	return &http.Server{
		Addr: address,
	}, nil
}

// ServeHTTP is the http handler that will be used if an only if a.config.RateLimiting.RateLimiter
// is configured. Otherwise, the main http handler will be directly the multiplexer.
func (a *restServer) ServeHTTP(w http.ResponseWriter, req *http.Request) {

	req = req.WithContext(a.mainContext)

	if a.config.RateLimiting.RateLimiter != nil {
		limited, err := a.config.RateLimiting.RateLimiter.RateLimit(req)
		if err != nil {
			writeHTTPResponse(w, makeErrorResponse(elemental.NewResponse(req.Context()), elemental.NewError("Internal Server Error", err.Error(), "bahamut", http.StatusInternalServerError)))
			return
		}

		if limited {
			writeHTTPResponse(w, makeErrorResponse(elemental.NewResponse(req.Context()), ErrRateLimit))
			return
		}
	}

	setCommonHeader(w, req.Header.Get("Origin"))

	a.multiplexer.ServeHTTP(w, req)
}

// installRoutes installs all the routes declared in the APIServerConfig.
func (a *restServer) installRoutes() {

	a.multiplexer.Options("*", http.HandlerFunc(corsHandler))
	a.multiplexer.NotFound(http.HandlerFunc(notFoundHandler))

	if a.config.ReSTServer.CustomRootHandlerFunc != nil {
		a.multiplexer.Handle("/", a.config.ReSTServer.CustomRootHandlerFunc)
	} else {
		a.multiplexer.Get("/", http.HandlerFunc(corsHandler))
	}

	// non versioned routes
	a.multiplexer.Get("/:category/:id", a.makeHandler(handleRetrieve))
	a.multiplexer.Put("/:category/:id", a.makeHandler(handleUpdate))
	a.multiplexer.Patch("/:category/:id", a.makeHandler(handlePatch))
	a.multiplexer.Delete("/:category/:id", a.makeHandler(handleDelete))
	a.multiplexer.Get("/:category", a.makeHandler(handleRetrieveMany))
	a.multiplexer.Get("/:parentcategory/:id/:category", a.makeHandler(handleRetrieveMany))
	a.multiplexer.Post("/:category", a.makeHandler(handleCreate))
	a.multiplexer.Post("/:parentcategory/:id/:category", a.makeHandler(handleCreate))
	a.multiplexer.Head("/:category", a.makeHandler(handleInfo))
	a.multiplexer.Head("/:parentcategory/:id/:category", a.makeHandler(handleInfo))

	// versioned routes
	a.multiplexer.Get("/v/:version/:category/:id", a.makeHandler(handleRetrieve))
	a.multiplexer.Put("/v/:version/:category/:id", a.makeHandler(handleUpdate))
	a.multiplexer.Patch("/v/:version/:category/:id", a.makeHandler(handlePatch))
	a.multiplexer.Delete("/v/:version/:category/:id", a.makeHandler(handleDelete))
	a.multiplexer.Get("/v/:version/:category", a.makeHandler(handleRetrieveMany))
	a.multiplexer.Get("/v/:version/:parentcategory/:id/:category", a.makeHandler(handleRetrieveMany))
	a.multiplexer.Post("/v/:version/:category", a.makeHandler(handleCreate))
	a.multiplexer.Post("/v/:version/:parentcategory/:id/:category", a.makeHandler(handleCreate))
	a.multiplexer.Head("/v/:version/:category", a.makeHandler(handleInfo))
	a.multiplexer.Head("/v/:version/:parentcategory/:id/:category", a.makeHandler(handleInfo))
}

func (a *restServer) start(ctx context.Context) {

	a.mainContext = ctx
	defer func() { a.mainContext = nil }()

	a.installRoutes()

	var err error
	if a.config.TLS.ServerCertificates != nil || a.config.TLS.ServerCertificatesRetrieverFunc != nil {
		a.server, err = a.createSecureHTTPServer(a.config.ReSTServer.ListenAddress)
	} else {
		a.server, err = a.createUnsecureHTTPServer(a.config.ReSTServer.ListenAddress)
	}
	if err != nil {
		zap.L().Fatal("Unable to create api server", zap.Error(err))
	}

	a.server.Handler = a

	go func() {
		if a.config.TLS.ServerCertificates != nil || a.config.TLS.ServerCertificatesRetrieverFunc != nil {
			err = a.server.ListenAndServeTLS("", "")
		} else {
			err = a.server.ListenAndServe()
		}

		if err != nil {
			if err == http.ErrServerClosed {
				return
			}
			zap.L().Fatal("Unable to start api server", zap.Error(err))
		}
	}()

	zap.L().Info("API server started", zap.String("address", a.config.ReSTServer.ListenAddress))

	<-ctx.Done()
}

func (a *restServer) stop() {

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := a.server.Shutdown(ctx); err != nil {
		zap.L().Error("Could not gracefully stop API server", zap.Error(err))
	}

	zap.L().Debug("API server stopped")
}

func (a *restServer) makeHandler(handler handlerFunc) http.HandlerFunc {

	return func(w http.ResponseWriter, req *http.Request) {

		request, err := elemental.NewRequestFromHTTPRequest(req)
		if err != nil {
			writeHTTPResponse(w, makeErrorResponse(elemental.NewResponse(req.Context()), elemental.NewError("Bad Request", err.Error(), "bahamut", http.StatusBadRequest)))
			return
		}

		traceRequest(request)
		defer finishTracing(request)

		writeHTTPResponse(w, handler(a.config, request, a.processorFinder, a.pusher))
	}
}
