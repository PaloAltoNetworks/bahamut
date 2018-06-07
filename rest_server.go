// Author: Antoine Mercadal
// See LICENSE file for full LICENSE
// Copyright 2016 Aporeto.

package bahamut

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/opentracing/opentracing-go"

	"github.com/aporeto-inc/elemental"
	"github.com/go-zoo/bone"
	"go.uber.org/zap"
	"golang.org/x/crypto/acme/autocert"
)

// an restServer is the structure serving the api routes.
type restServer struct {
	cfg             config
	multiplexer     *bone.Mux
	server          *http.Server
	processorFinder processorFinderFunc
	pusher          eventPusherFunc
}

// newRestServer returns a new apiServer.
func newRestServer(cfg config, multiplexer *bone.Mux, processorFinder processorFinderFunc, pusher eventPusherFunc) *restServer {

	return &restServer{
		cfg:             cfg,
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
		ClientAuth:               a.cfg.tls.authType,
		ClientCAs:                a.cfg.tls.clientCAPool,
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

	if !a.cfg.tls.enableLetsEncrypt {

		// If letsencrypt is not enabled we simply set the given list of
		// certificates or the ServerCertificatesRetrieverFunc in the TLS option.

		if a.cfg.tls.serverCertificatesRetrieverFunc != nil {
			tlsConfig.GetCertificate = a.cfg.tls.serverCertificatesRetrieverFunc
		} else {
			tlsConfig.Certificates = a.cfg.tls.serverCertificates
		}

	} else {

		cachePath := a.cfg.tls.letsEncryptCertificateCacheFolder
		if cachePath == "" {
			cachePath = os.TempDir()
		}

		// Otherwise, we create an autocert manager
		m := autocert.Manager{
			Prompt:     autocert.AcceptTOS,
			HostPolicy: autocert.HostWhitelist(a.cfg.tls.letsEncryptDomainWhiteList...),
			Cache:      autocert.DirCache(cachePath),
		}

		// Then we build a custom GetCertificate function to first use the certificate passed
		// by the config, then eventually try to get a certificate from letsencrypt.
		localCertMap := buildNameAndIPsToCertificate(a.cfg.tls.serverCertificates)
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
		ReadTimeout:  a.cfg.restServer.readTimeout,
		WriteTimeout: a.cfg.restServer.writeTimeout,
		IdleTimeout:  a.cfg.restServer.idleTimeout,
	}

	server.SetKeepAlivesEnabled(!a.cfg.restServer.disableKeepalive)

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

	if a.cfg.rateLimiting.rateLimiter != nil {
		limited, err := a.cfg.rateLimiting.rateLimiter.RateLimit(req)
		if err != nil {
			writeHTTPResponse(
				w,
				makeErrorResponse(
					req.Context(),
					elemental.NewResponse(elemental.NewRequest()),
					elemental.NewError("Internal Server Error", err.Error(), "bahamut", http.StatusInternalServerError),
				),
			)
			return
		}

		if limited {
			writeHTTPResponse(
				w,
				makeErrorResponse(
					req.Context(),
					elemental.NewResponse(elemental.NewRequest()),
					ErrRateLimit,
				),
			)
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

	if a.cfg.restServer.customRootHandlerFunc != nil {
		a.multiplexer.Handle("/", a.cfg.restServer.customRootHandlerFunc)
	} else {
		a.multiplexer.Get("/", http.HandlerFunc(corsHandler))
	}

	if a.cfg.meta.serviceName != "" {
		a.multiplexer.Get("/_meta/name", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			setCommonHeader(w, r.Header.Get("Origin"))
			w.WriteHeader(200)
			w.Write([]byte(a.cfg.meta.serviceName)) // nolint: errcheck
		}))
	}

	if !a.cfg.meta.disableMetaRoute {

		routesInfo := buildVersionedRoutes(a.cfg.model.relationshipsRegistry, a.cfg.model.identifiableFactories[0], a.processorFinder)

		encodedRoutesInfo, err := json.Marshal(routesInfo)
		if err != nil {
			panic(fmt.Sprintf("Unable to build route info: %s", err))
		}

		a.multiplexer.Get("/_meta/routes", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			setCommonHeader(w, r.Header.Get("Origin"))
			w.WriteHeader(200)
			w.Write(encodedRoutesInfo) // nolint: errcheck
		}))
	}

	if a.cfg.meta.version != nil {

		encodedVersionInfo, err := json.MarshalIndent(a.cfg.meta.version, "", "    ")
		if err != nil {
			panic(fmt.Sprintf("Unable to build route info: %s", err))
		}

		a.multiplexer.Get("/_meta/version", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			setCommonHeader(w, r.Header.Get("Origin"))
			w.WriteHeader(200)
			w.Write(encodedVersionInfo) // nolint: errcheck
		}))
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

	a.installRoutes()

	var err error
	if a.cfg.tls.serverCertificates != nil || a.cfg.tls.serverCertificatesRetrieverFunc != nil {
		a.server, err = a.createSecureHTTPServer(a.cfg.restServer.listenAddress)
	} else {
		a.server, err = a.createUnsecureHTTPServer(a.cfg.restServer.listenAddress)
	}
	if err != nil {
		zap.L().Fatal("Unable to create api server", zap.Error(err))
	}

	a.server.Handler = a

	go func() {
		if a.cfg.tls.serverCertificates != nil || a.cfg.tls.serverCertificatesRetrieverFunc != nil {
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

	zap.L().Info("API server started", zap.String("address", a.cfg.restServer.listenAddress))

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

		request, err := elemental.NewRequestFromHTTPRequest(req, a.cfg.model.identifiableFactories[0])
		if err != nil {
			writeHTTPResponse(
				w,
				makeErrorResponse(
					req.Context(),
					elemental.NewResponse(elemental.NewRequest()),
					elemental.NewError("Bad Request", err.Error(), "bahamut", http.StatusBadRequest),
				),
			)
			return
		}

		ctx := traceRequest(req.Context(), request, opentracing.GlobalTracer())
		defer finishTracing(ctx)

		bctx := NewContextWithRequest(request)
		bctx.ctx = ctx

		writeHTTPResponse(w, handler(bctx, a.cfg, a.processorFinder, a.pusher))
	}
}
