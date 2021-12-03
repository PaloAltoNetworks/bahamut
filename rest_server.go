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
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/NYTimes/gziphandler"
	"github.com/go-zoo/bone"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/valyala/tcplisten"
	"go.aporeto.io/elemental"
	"go.uber.org/zap"
)

var (
	// ErrInvalidAPIVersion is returned when the url cannot be parsed
	// to find a valid version.
	ErrInvalidAPIVersion = elemental.NewError("Bad Request", fmt.Sprintf("Invalid api version number"), "bahamut", http.StatusBadRequest)

	// ErrUnknownAPIVersion when there is no elemental.ModelManager configured
	// to handle the requested version of the api.
	ErrUnknownAPIVersion = elemental.NewError("Bad Request", fmt.Sprintf("Unknown api version"), "bahamut", http.StatusBadRequest)
)

// an restServer is the structure serving the api routes.
type restServer struct {
	cfg             config
	multiplexer     *bone.Mux
	server          *http.Server
	processorFinder processorFinderFunc
	pusher          eventPusherFunc
	customHandlers  retrieveHandlersFunc
}

// newRestServer returns a new apiServer.
func newRestServer(cfg config, multiplexer *bone.Mux, processorFinder processorFinderFunc, customHandlers retrieveHandlersFunc, pusher eventPusherFunc) *restServer {

	return &restServer{
		cfg:             cfg,
		multiplexer:     multiplexer,
		processorFinder: processorFinder,
		pusher:          pusher,
		customHandlers:  customHandlers,
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
		SessionTicketsDisabled:   a.cfg.tls.disableSessionTicket,
		PreferServerCipherSuites: true,
		NextProtos:               a.cfg.tls.nextProtos,
		CurvePreferences:         []tls.CurveID{tls.CurveP521, tls.CurveP384, tls.CurveP256},
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

	server := &http.Server{
		Addr:         address,
		TLSConfig:    tlsConfig,
		ReadTimeout:  a.cfg.restServer.readTimeout,
		WriteTimeout: a.cfg.restServer.writeTimeout,
		IdleTimeout:  a.cfg.restServer.idleTimeout,
		ErrorLog:     a.cfg.restServer.httpLogger,
	}

	server.SetKeepAlivesEnabled(!a.cfg.restServer.disableKeepalive)

	return server
}

// createUnsecureHTTPServer returns a insecure HTTP Server.
//
// It will return an error if any.
func (a *restServer) createUnsecureHTTPServer(address string) *http.Server {

	return &http.Server{
		Addr:         address,
		ReadTimeout:  a.cfg.restServer.readTimeout,
		WriteTimeout: a.cfg.restServer.writeTimeout,
		IdleTimeout:  a.cfg.restServer.idleTimeout,
		ErrorLog:     a.cfg.restServer.httpLogger,
	}
}

// installRoutes installs all the routes declared in the APIServerConfig.
func (a *restServer) installRoutes(routesInfo map[int][]RouteInfo) {

	a.multiplexer.NotFound(http.HandlerFunc(makeNotFoundHandler(a.cfg.security.accessControl)))

	if a.cfg.restServer.customRootHandlerFunc != nil {
		a.multiplexer.Handle("/", a.cfg.restServer.customRootHandlerFunc)
	} else {
		a.multiplexer.Get("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
		}))
	}

	if a.cfg.meta.serviceName != "" {
		a.multiplexer.Get("/_meta/name", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			setCommonHeader(w, "text/plain")
			w.WriteHeader(200)
			_, _ = w.Write([]byte(a.cfg.meta.serviceName)) // nolint: errcheck
		}))
	}

	if !a.cfg.meta.disableMetaRoute {

		encodedRoutesInfo, err := json.Marshal(routesInfo)
		if err != nil {
			panic(fmt.Sprintf("Unable to build route info: %s", err))
		}

		a.multiplexer.Get("/_meta/routes", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			setCommonHeader(w, elemental.EncodingTypeJSON)
			w.WriteHeader(200)
			_, _ = w.Write(encodedRoutesInfo) // nolint: errcheck
		}))
	}

	if a.cfg.meta.version != nil {

		encodedVersionInfo, err := json.MarshalIndent(a.cfg.meta.version, "", "    ")
		if err != nil {
			panic(fmt.Sprintf("Unable to build route info: %s", err))
		}

		a.multiplexer.Get("/_meta/version", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			setCommonHeader(w, elemental.EncodingTypeJSON)
			w.WriteHeader(200)
			_, _ = w.Write(encodedVersionInfo) // nolint: errcheck
		}))
	}

	if a.customHandlers != nil {
		for customRoute, f := range a.customHandlers() {
			a.multiplexer.Handle(path.Join(a.cfg.restServer.customRoutePrefix, customRoute), f)
		}
	}

	// non versioned routes
	a.multiplexer.Get(path.Join(a.cfg.restServer.apiPrefix, "/:category/:id"), a.makeHandler(handleRetrieve))
	a.multiplexer.Put(path.Join(a.cfg.restServer.apiPrefix, "/:category/:id"), a.makeHandler(handleUpdate))
	a.multiplexer.Patch(path.Join(a.cfg.restServer.apiPrefix, "/:category/:id"), a.makeHandler(handlePatch))
	a.multiplexer.Delete(path.Join(a.cfg.restServer.apiPrefix, "/:category/:id"), a.makeHandler(handleDelete))
	a.multiplexer.Get(path.Join(a.cfg.restServer.apiPrefix, "/:category"), a.makeHandler(handleRetrieveMany))
	a.multiplexer.Get(path.Join(a.cfg.restServer.apiPrefix, "/:parentcategory/:id/:category"), a.makeHandler(handleRetrieveMany))
	a.multiplexer.Post(path.Join(a.cfg.restServer.apiPrefix, "/:category"), a.makeHandler(handleCreate))
	a.multiplexer.Post(path.Join(a.cfg.restServer.apiPrefix, "/:parentcategory/:id/:category"), a.makeHandler(handleCreate))
	a.multiplexer.Head(path.Join(a.cfg.restServer.apiPrefix, "/:category"), a.makeHandler(handleInfo))
	a.multiplexer.Head(path.Join(a.cfg.restServer.apiPrefix, "/:parentcategory/:id/:category"), a.makeHandler(handleInfo))

	// versioned routes
	a.multiplexer.Get(path.Join(a.cfg.restServer.apiPrefix, "/v/:version/:category/:id"), a.makeHandler(handleRetrieve))
	a.multiplexer.Put(path.Join(a.cfg.restServer.apiPrefix, "/v/:version/:category/:id"), a.makeHandler(handleUpdate))
	a.multiplexer.Patch(path.Join(a.cfg.restServer.apiPrefix, "/v/:version/:category/:id"), a.makeHandler(handlePatch))
	a.multiplexer.Delete(path.Join(a.cfg.restServer.apiPrefix, "/v/:version/:category/:id"), a.makeHandler(handleDelete))
	a.multiplexer.Get(path.Join(a.cfg.restServer.apiPrefix, "/v/:version/:category"), a.makeHandler(handleRetrieveMany))
	a.multiplexer.Get(path.Join(a.cfg.restServer.apiPrefix, "/v/:version/:parentcategory/:id/:category"), a.makeHandler(handleRetrieveMany))
	a.multiplexer.Post(path.Join(a.cfg.restServer.apiPrefix, "/v/:version/:category"), a.makeHandler(handleCreate))
	a.multiplexer.Post(path.Join(a.cfg.restServer.apiPrefix, "/v/:version/:parentcategory/:id/:category"), a.makeHandler(handleCreate))
	a.multiplexer.Head(path.Join(a.cfg.restServer.apiPrefix, "/v/:version/:category"), a.makeHandler(handleInfo))
	a.multiplexer.Head(path.Join(a.cfg.restServer.apiPrefix, "/v/:version/:parentcategory/:id/:category"), a.makeHandler(handleInfo))

	if a.cfg.security.accessControl != nil {
		a.multiplexer.Options("*", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			setCommonHeader(w, elemental.EncodingTypeJSON)
			a.cfg.security.accessControl.Inject(w.Header(), r.Header.Get("origin"), true)
		}))
	}
}

func (a *restServer) start(ctx context.Context, routesInfo map[int][]RouteInfo) {

	a.installRoutes(routesInfo)

	var err error
	if a.cfg.tls.serverCertificates != nil || a.cfg.tls.serverCertificatesRetrieverFunc != nil {
		a.server = a.createSecureHTTPServer(a.cfg.restServer.listenAddress)
	} else {
		a.server = a.createUnsecureHTTPServer(a.cfg.restServer.listenAddress)
	}

	// This is just noise.
	a.server.Handler = a.multiplexer

	if metricManager := a.cfg.healthServer.metricsManager; metricManager != nil {
		a.server.ConnState = func(conn net.Conn, state http.ConnState) {
			switch state {
			case http.StateNew:
				metricManager.RegisterTCPConnection()
			case http.StateClosed, http.StateHijacked:
				metricManager.UnregisterTCPConnection()
			}
		}
	}

	go func() {

		listener := a.cfg.restServer.customListener
		if listener == nil {
			listener, err = (&tcplisten.Config{
				ReusePort:   true,
				DeferAccept: true,
				FastOpen:    true,
			}).NewListener("tcp4", a.server.Addr)
			if err != nil {
				zap.L().Fatal("Unable to dial", zap.Error(err))
			}
		}

		listener = newListener(listener, a.cfg.restServer.maxConnection)

		if a.cfg.tls.serverCertificates != nil || a.cfg.tls.serverCertificatesRetrieverFunc != nil {
			err = a.server.ServeTLS(listener, "", "")
		} else {
			err = a.server.Serve(listener)
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

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)

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

	h := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {

		var measure FinishMeasurementFunc
		if a.cfg.healthServer.metricsManager != nil {
			measure = a.cfg.healthServer.metricsManager.MeasureRequest(req.Method, req.URL.Path)
		}

		// Trim our custom prefix out of the request URI.
		// TODO: The elemental function needs to moved in here
		// and potentially find a cleaner way rather than triming the prefix here.
		if a.cfg.restServer.apiPrefix != "" {
			req.URL.Path = strings.TrimPrefix(req.URL.Path, a.cfg.restServer.apiPrefix)
		}

		// Get API version
		version, err := extractAPIVersion(req.URL.Path)
		if err != nil {
			code := writeHTTPResponse(
				w,
				makeErrorResponse(
					req.Context(),
					elemental.NewResponse(elemental.NewRequest()),
					ErrInvalidAPIVersion,
					nil,
					nil,
				),
				req.Header.Get("origin"),
				a.cfg.security.accessControl,
			)
			if measure != nil {
				measure(code, nil)
			}
			return
		}

		// Check API Version
		manager, ok := a.cfg.model.modelManagers[version]
		if !ok {
			code := writeHTTPResponse(
				w,
				makeErrorResponse(
					req.Context(),
					elemental.NewResponse(elemental.NewRequest()),
					ErrUnknownAPIVersion,
					nil,
					nil,
				),
				req.Header.Get("origin"),
				a.cfg.security.accessControl,
			)
			if measure != nil {
				measure(code, nil)
			}
			return
		}

		request, err := elemental.NewRequestFromHTTPRequest(req, manager)
		if err != nil {
			code := writeHTTPResponse(
				w,
				makeErrorResponse(
					req.Context(),
					elemental.NewResponse(elemental.NewRequest()),
					err,
					nil,
					nil,
				),
				req.Header.Get("origin"),
				a.cfg.security.accessControl,
			)
			if measure != nil {
				measure(code, nil)
			}
			return
		}

		ctx := traceRequest(req.Context(), request, a.cfg.opentracing.tracer, a.cfg.opentracing.excludedIdentities, a.cfg.opentracing.traceCleaner)
		defer finishTracing(ctx)

		// Global rate limiting
		if a.cfg.rateLimiting.rateLimiter != nil {
			if !a.cfg.rateLimiting.rateLimiter.Allow() {
				code := writeHTTPResponse(
					w,
					makeErrorResponse(ctx, elemental.NewResponse(request),
						ErrRateLimit,
						nil,
						nil,
					),
					req.Header.Get("origin"),
					a.cfg.security.accessControl,
				)
				if measure != nil {
					measure(code, opentracing.SpanFromContext(ctx))
				}
				return
			}
		}

		// Per api rate limiting
		if a.cfg.rateLimiting.apiRateLimiters != nil {
			if rlm, ok := a.cfg.rateLimiting.apiRateLimiters[request.Identity]; ok {
				if rlm.condition == nil || rlm.condition(request) {
					if !rlm.limiter.Allow() {
						code := writeHTTPResponse(
							w,
							makeErrorResponse(
								ctx,
								elemental.NewResponse(request),
								ErrRateLimit,
								nil,
								nil,
							),
							req.Header.Get("origin"),
							a.cfg.security.accessControl,
						)
						if measure != nil {
							measure(code, opentracing.SpanFromContext(ctx))
						}
						return
					}
				}
			}
		}

		bctx := newContext(ctx, request)
		resp := handler(bctx, a.cfg, a.processorFinder, a.pusher)
		var code int

		switch {
		case bctx.responseWriter != nil:
			code = bctx.responseWriter(w)
		default:
			code = writeHTTPResponse(
				w,
				resp,
				req.Header.Get("origin"),
				a.cfg.security.accessControl,
			)
		}

		if measure != nil {
			measure(code, opentracing.SpanFromContext(ctx))
		}
	})

	if a.cfg.restServer.disableCompression {
		return h
	}

	return gziphandler.GzipHandler(h).(http.HandlerFunc)
}
