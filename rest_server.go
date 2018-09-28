// Author: Antoine Mercadal
// See LICENSE file for full LICENSE
// Copyright 2016 Aporeto.

package bahamut

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/NYTimes/gziphandler"
	"github.com/go-zoo/bone"
	"github.com/opentracing/opentracing-go"
	"go.aporeto.io/elemental"
	"go.uber.org/zap"
)

// an restServer is the structure serving the api routes.
type restServer struct {
	cfg             config
	multiplexer     *bone.Mux
	server          *http.Server
	processorFinder processorFinderFunc
	pusher          eventPusherFunc
	reqCounter      *uint64
}

// newRestServer returns a new apiServer.
func newRestServer(cfg config, multiplexer *bone.Mux, processorFinder processorFinderFunc, pusher eventPusherFunc, reqCounter *uint64) *restServer {

	return &restServer{
		cfg:             cfg,
		multiplexer:     multiplexer,
		processorFinder: processorFinder,
		pusher:          pusher,
		reqCounter:      reqCounter,
	}
}

// createSecureHTTPServer returns the main HTTP Server.
//
// It will return an error if any.
func (a *restServer) createSecureHTTPServer(address string) *http.Server {

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

	if a.cfg.tls.serverCertificatesRetrieverFunc != nil {
		tlsConfig.GetCertificate = a.cfg.tls.serverCertificatesRetrieverFunc
	} else {
		tlsConfig.Certificates = a.cfg.tls.serverCertificates
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

	return server
}

// createUnsecureHTTPServer returns a insecure HTTP Server.
//
// It will return an error if any.
func (a *restServer) createUnsecureHTTPServer(address string) *http.Server {

	return &http.Server{
		Addr: address,
	}
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

		routesInfo := buildVersionedRoutes(a.cfg.model.modelManagers, a.processorFinder)

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
		a.server = a.createSecureHTTPServer(a.cfg.restServer.listenAddress)
	} else {
		a.server = a.createUnsecureHTTPServer(a.cfg.restServer.listenAddress)
	}

	a.server.Handler = a.multiplexer

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

func (a *restServer) stop() context.Context {

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)

	go func() {
		defer cancel()
		if err := a.server.Shutdown(ctx); err != nil {
			zap.L().Error("Could not gracefully stop API server", zap.Error(err))
		} else {
			zap.L().Debug("API server stopped")
		}
	}()

	return ctx
}

func (a *restServer) makeHandler(handler handlerFunc) http.HandlerFunc {

	return gziphandler.GzipHandler(
		http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {

			// TODO: find a way to support tracing in case of bad request here.
			request, err := elemental.NewRequestFromHTTPRequest(req, a.cfg.model.modelManagers[0])
			if err != nil {
				writeHTTPResponse(w, makeErrorResponse(req.Context(), elemental.NewResponse(elemental.NewRequest()), err))
				return
			}

			ctx := traceRequest(req.Context(), request, opentracing.GlobalTracer())
			defer finishTracing(ctx)

			if a.cfg.rateLimiting.rateLimiter != nil {
				if err = a.cfg.rateLimiting.rateLimiter.Wait(ctx); err != nil {
					writeHTTPResponse(w, makeErrorResponse(ctx, elemental.NewResponse(elemental.NewRequest()), ErrRateLimit))
					return
				}
			}

			atomic.AddUint64(a.reqCounter, 1)
			setCommonHeader(w, req.Header.Get("Origin"))
			writeHTTPResponse(w, handler(newContext(ctx, request), a.cfg, a.processorFinder, a.pusher))
		}),
	).(http.HandlerFunc)
}
