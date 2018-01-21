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

	limited, err := a.config.RateLimiting.RateLimiter.RateLimit(req)

	if err != nil {
		writeHTTPError(w, fakeElementalRequest(req), elemental.NewError("Internal Server Error", err.Error(), "bahamut", http.StatusInternalServerError))
		return
	}

	if limited {
		writeHTTPError(w, fakeElementalRequest(req), elemental.NewError("Rate Limit", "You have exceeded your rate limit", "bahamut", http.StatusTooManyRequests))
		return
	}

	a.multiplexer.ServeHTTP(w, req.WithContext(a.mainContext))
}

func (a *restServer) handleRetrieve(w http.ResponseWriter, req *http.Request) {

	request, err := elemental.NewRequestFromHTTPRequest(req)
	if err != nil {
		writeHTTPError(w, fakeElementalRequest(req), elemental.NewError("Bad Request", err.Error(), "bahamut", http.StatusBadRequest))
		return
	}

	request.StartTracing()
	defer request.FinishTracing()

	if !elemental.IsRetrieveAllowed(a.config.Model.RelationshipsRegistry[request.Version], request.Identity) {
		writeHTTPError(w, request, elemental.NewError("Not allowed", "Retrieve operation not allowed on "+request.Identity.Name, "bahamut", http.StatusMethodNotAllowed))
		return
	}

	ctx := NewContextWithRequest(request)

	runHTTPDispatcher(
		ctx,
		w,
		func() error {
			return dispatchRetrieveOperation(
				ctx,
				a.processorFinder,
				a.config.Model.IdentifiablesFactory,
				a.config.Security.RequestAuthenticators,
				a.config.Security.Authorizers,
				a.pusher,
				a.config.Security.Auditer,
			)
		},
		!a.config.ReSTServer.PanicRecoveryDisabled,
	)
}

func (a *restServer) handleUpdate(w http.ResponseWriter, req *http.Request) {

	request, err := elemental.NewRequestFromHTTPRequest(req)
	if err != nil {
		writeHTTPError(w, fakeElementalRequest(req), elemental.NewError("Bad Request", err.Error(), "bahamut", http.StatusBadRequest))
		return
	}

	request.StartTracing()
	defer request.FinishTracing()

	if !elemental.IsUpdateAllowed(a.config.Model.RelationshipsRegistry[request.Version], request.Identity) {
		writeHTTPError(w, request, elemental.NewError("Not allowed", "Update opration not allowed on "+request.Identity.Name, "bahamut", http.StatusMethodNotAllowed))
		return
	}

	ctx := NewContextWithRequest(request)

	runHTTPDispatcher(
		ctx,
		w,
		func() error {
			return dispatchUpdateOperation(
				ctx,
				a.processorFinder,
				a.config.Model.IdentifiablesFactory,
				a.config.Security.RequestAuthenticators,
				a.config.Security.Authorizers,
				a.pusher,
				a.config.Security.Auditer,
				a.config.Model.ReadOnly,
				a.config.Model.ReadOnlyExcludedIdentities,
			)
		},
		!a.config.ReSTServer.PanicRecoveryDisabled,
	)

}

func (a *restServer) handleDelete(w http.ResponseWriter, req *http.Request) {

	request, err := elemental.NewRequestFromHTTPRequest(req)
	if err != nil {
		writeHTTPError(w, fakeElementalRequest(req), elemental.NewError("Bad Request", err.Error(), "bahamut", http.StatusBadRequest))
		return
	}

	request.StartTracing()
	defer request.FinishTracing()

	if !elemental.IsDeleteAllowed(a.config.Model.RelationshipsRegistry[request.Version], request.Identity) {
		writeHTTPError(w, request, elemental.NewError("Not allowed", "Delete operation not allowed on "+request.Identity.Name, "bahamut", http.StatusMethodNotAllowed))
		return
	}

	ctx := NewContextWithRequest(request)

	runHTTPDispatcher(
		ctx,
		w,
		func() error {
			return dispatchDeleteOperation(
				ctx,
				a.processorFinder,
				a.config.Model.IdentifiablesFactory,
				a.config.Security.RequestAuthenticators,
				a.config.Security.Authorizers,
				a.pusher,
				a.config.Security.Auditer,
				a.config.Model.ReadOnly,
				a.config.Model.ReadOnlyExcludedIdentities,
			)
		},
		!a.config.ReSTServer.PanicRecoveryDisabled,
	)
}

func (a *restServer) handleRetrieveMany(w http.ResponseWriter, req *http.Request) {

	request, err := elemental.NewRequestFromHTTPRequest(req)
	if err != nil {
		writeHTTPError(w, fakeElementalRequest(req), elemental.NewError("Bad Request", err.Error(), "bahamut", http.StatusBadRequest))
		return
	}

	request.StartTracing()
	defer request.FinishTracing()

	if request.ParentIdentity.IsEmpty() {
		request.ParentIdentity = elemental.RootIdentity
	}

	if !elemental.IsRetrieveManyAllowed(a.config.Model.RelationshipsRegistry[request.Version], request.Identity, request.ParentIdentity) {
		writeHTTPError(w, request, elemental.NewError("Not allowed", "RetrieveMany operation not allowed on "+request.Identity.Category, "bahamut", http.StatusMethodNotAllowed))
		return
	}

	ctx := NewContextWithRequest(request)

	runHTTPDispatcher(
		ctx,
		w,
		func() error {
			return dispatchRetrieveManyOperation(
				ctx,
				a.processorFinder,
				a.config.Model.IdentifiablesFactory,
				a.config.Security.RequestAuthenticators,
				a.config.Security.Authorizers,
				a.pusher,
				a.config.Security.Auditer,
			)
		},
		!a.config.ReSTServer.PanicRecoveryDisabled,
	)
}

func (a *restServer) handleCreate(w http.ResponseWriter, req *http.Request) {

	request, err := elemental.NewRequestFromHTTPRequest(req)
	if err != nil {
		writeHTTPError(w, fakeElementalRequest(req), elemental.NewError("Bad Request", err.Error(), "bahamut", http.StatusBadRequest))
		return
	}

	request.StartTracing()
	defer request.FinishTracing()

	if request.ParentIdentity.IsEmpty() {
		request.ParentIdentity = elemental.RootIdentity
	}

	if !elemental.IsCreateAllowed(a.config.Model.RelationshipsRegistry[request.Version], request.Identity, request.ParentIdentity) {
		writeHTTPError(w, request, elemental.NewError("Not allowed", "Create operation not allowed on "+request.Identity.Name, "bahamut", http.StatusMethodNotAllowed))
		return
	}

	ctx := NewContextWithRequest(request)

	runHTTPDispatcher(
		ctx,
		w,
		func() error {
			return dispatchCreateOperation(
				ctx,
				a.processorFinder,
				a.config.Model.IdentifiablesFactory,
				a.config.Security.RequestAuthenticators,
				a.config.Security.Authorizers,
				a.pusher,
				a.config.Security.Auditer,
				a.config.Model.ReadOnly,
				a.config.Model.ReadOnlyExcludedIdentities,
			)
		},
		!a.config.ReSTServer.PanicRecoveryDisabled,
	)
}

func (a *restServer) handleInfo(w http.ResponseWriter, req *http.Request) {

	request, err := elemental.NewRequestFromHTTPRequest(req)
	if err != nil {
		writeHTTPError(w, fakeElementalRequest(req), elemental.NewError("Bad Request", err.Error(), "bahamut", http.StatusBadRequest))
		return
	}

	request.StartTracing()
	defer request.FinishTracing()

	if request.ParentIdentity.IsEmpty() {
		request.ParentIdentity = elemental.RootIdentity
	}

	if !elemental.IsInfoAllowed(a.config.Model.RelationshipsRegistry[request.Version], request.Identity, request.ParentIdentity) {
		writeHTTPError(w, request, elemental.NewError("Not allowed", "Info operation not allowed on "+request.Identity.Category, "bahamut", http.StatusMethodNotAllowed))
		return
	}

	ctx := NewContextWithRequest(request)

	runHTTPDispatcher(
		ctx,
		w,
		func() error {
			return dispatchInfoOperation(
				ctx,
				a.processorFinder,
				a.config.Model.IdentifiablesFactory,
				a.config.Security.RequestAuthenticators,
				a.config.Security.Authorizers,
				a.config.Security.Auditer,
			)
		},
		!a.config.ReSTServer.PanicRecoveryDisabled,
	)
}

func (a *restServer) handlePatch(w http.ResponseWriter, req *http.Request) {

	request, err := elemental.NewRequestFromHTTPRequest(req)
	if err != nil {
		writeHTTPError(w, fakeElementalRequest(req), elemental.NewError("Bad Request", err.Error(), "bahamut", http.StatusBadRequest))
		return
	}

	request.StartTracing()
	defer request.FinishTracing()

	if request.ParentIdentity.IsEmpty() {
		request.ParentIdentity = elemental.RootIdentity
	}

	if !elemental.IsPatchAllowed(a.config.Model.RelationshipsRegistry[request.Version], request.Identity, request.ParentIdentity) {
		writeHTTPError(w, request, elemental.NewError("Not allowed", "Patch operation not allowed on "+request.Identity.Name, "bahamut", http.StatusMethodNotAllowed))
		return
	}

	ctx := NewContextWithRequest(request)

	runHTTPDispatcher(
		ctx,
		w,
		func() error {
			return dispatchPatchOperation(
				ctx,
				a.processorFinder,
				a.config.Model.IdentifiablesFactory,
				a.config.Security.RequestAuthenticators,
				a.config.Security.Authorizers,
				a.pusher,
				a.config.Security.Auditer,
				a.config.Model.ReadOnly,
				a.config.Model.ReadOnlyExcludedIdentities,
			)
		},
		!a.config.ReSTServer.PanicRecoveryDisabled,
	)
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
	a.multiplexer.Get("/:category/:id", http.HandlerFunc(a.handleRetrieve))
	a.multiplexer.Put("/:category/:id", http.HandlerFunc(a.handleUpdate))
	a.multiplexer.Patch("/:category/:id", http.HandlerFunc(a.handlePatch))
	a.multiplexer.Delete("/:category/:id", http.HandlerFunc(a.handleDelete))
	a.multiplexer.Get("/:category", http.HandlerFunc(a.handleRetrieveMany))
	a.multiplexer.Get("/:parentcategory/:id/:category", http.HandlerFunc(a.handleRetrieveMany))
	a.multiplexer.Post("/:category", http.HandlerFunc(a.handleCreate))
	a.multiplexer.Post("/:parentcategory/:id/:category", http.HandlerFunc(a.handleCreate))
	a.multiplexer.Head("/:category", http.HandlerFunc(a.handleInfo))
	a.multiplexer.Head("/:parentcategory/:id/:category", http.HandlerFunc(a.handleInfo))

	// versioned routes
	a.multiplexer.Get("/v/:version/:category/:id", http.HandlerFunc(a.handleRetrieve))
	a.multiplexer.Put("/v/:version/:category/:id", http.HandlerFunc(a.handleUpdate))
	a.multiplexer.Patch("/v/:version/:category/:id", http.HandlerFunc(a.handlePatch))
	a.multiplexer.Delete("/v/:version/:category/:id", http.HandlerFunc(a.handleDelete))
	a.multiplexer.Get("/v/:version/:category", http.HandlerFunc(a.handleRetrieveMany))
	a.multiplexer.Get("/v/:version/:parentcategory/:id/:category", http.HandlerFunc(a.handleRetrieveMany))
	a.multiplexer.Post("/v/:version/:category", http.HandlerFunc(a.handleCreate))
	a.multiplexer.Post("/v/:version/:parentcategory/:id/:category", http.HandlerFunc(a.handleCreate))
	a.multiplexer.Head("/v/:version/:category", http.HandlerFunc(a.handleInfo))
	a.multiplexer.Head("/v/:version/:parentcategory/:id/:category", http.HandlerFunc(a.handleInfo))

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

	// If we have a RateLimiter configured, we use our own main handler.
	if a.config.RateLimiting.RateLimiter != nil {
		a.server.Handler = a
	} else {
		a.server.Handler = a.multiplexer
	}

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
