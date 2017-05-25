// Author: Antoine Mercadal
// See LICENSE file for full LICENSE
// Copyright 2016 Aporeto.

package bahamut

import (
	"crypto/tls"
	"net"
	"net/http"
	"os"

	"github.com/aporeto-inc/elemental"
	"github.com/go-zoo/bone"
	"go.uber.org/zap"
	"golang.org/x/crypto/acme/autocert"
)

// an apiServer is the structure serving the api routes.
type apiServer struct {
	config          Config
	multiplexer     *bone.Mux
	server          *http.Server
	processorFinder processorFinder
	pusher          eventPusher
}

// newAPIServer returns a new apiServer.
func newAPIServer(config Config, multiplexer *bone.Mux) *apiServer {

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
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
		},
	}

	if !a.config.TLS.EnableLetsEncrypt {

		// If letsencrypt is not enabled we simply set the given list of
		// certificate in the TLS option.
		tlsConfig.Certificates = a.config.TLS.ServerCertificates

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
func (a *apiServer) createUnsecureHTTPServer(address string) (*http.Server, error) {

	return &http.Server{
		Addr: address,
	}, nil
}

// ServeHTTP is the http handler that will be used if an only if a.config.RateLimiting.RateLimiter
// is configured. Otherwise, the main http handler will be directly the multiplexer.
func (a *apiServer) ServeHTTP(w http.ResponseWriter, req *http.Request) {

	limited, err := a.config.RateLimiting.RateLimiter.RateLimit(req)

	if err != nil {
		writeHTTPError(w, req.Header.Get("Origin"), elemental.NewError("Internal Server Error", err.Error(), "bahamut", http.StatusInternalServerError))
		return
	}

	if limited {
		writeHTTPError(w, req.Header.Get("Origin"), elemental.NewError("Rate Limit", "You have exceeded your rate limit", "bahamut", http.StatusTooManyRequests))
		return
	}

	a.multiplexer.ServeHTTP(w, req)
}

func (a *apiServer) handleRetrieve(w http.ResponseWriter, req *http.Request) {

	identity := elemental.IdentityFromCategory(bone.GetValue(req, "category"))

	if !elemental.IsRetrieveAllowed(a.config.Model.RelationshipsRegistry, identity) {
		writeHTTPError(w, req.Header.Get("Origin"), elemental.NewError("Not allowed", "Retrieve operation not allowed on "+identity.Name, "bahamut", http.StatusMethodNotAllowed))
		return
	}

	request, err := elemental.NewRequestFromHTTPRequest(req)
	if err != nil {
		writeHTTPError(w, req.Header.Get("Origin"), elemental.NewError("Bad Request", err.Error(), "bahamut", http.StatusBadRequest))
		return
	}

	ctx, err := dispatchRetrieveOperation(
		request,
		a.processorFinder,
		a.config.Model.IdentifiablesFactory,
		a.config.Security.RequestAuthenticator,
		a.config.Security.Authorizer,
		a.pusher,
		a.config.Security.Auditer,
	)

	if err != nil {
		writeHTTPError(w, w.Header().Get("Origin"), err)
		return
	}

	writeHTTPResponse(w, ctx)
}

func (a *apiServer) handleUpdate(w http.ResponseWriter, req *http.Request) {

	identity := elemental.IdentityFromCategory(bone.GetValue(req, "category"))

	if !elemental.IsUpdateAllowed(a.config.Model.RelationshipsRegistry, identity) {
		writeHTTPError(w, req.Header.Get("Origin"), elemental.NewError("Not allowed", "Update opration not allowed on "+identity.Name, "bahamut", http.StatusMethodNotAllowed))
		return
	}

	request, err := elemental.NewRequestFromHTTPRequest(req)
	if err != nil {
		writeHTTPError(w, req.Header.Get("Origin"), elemental.NewError("Bad Request", err.Error(), "bahamut", http.StatusBadRequest))
		return
	}

	ctx, err := dispatchUpdateOperation(
		request,
		a.processorFinder,
		a.config.Model.IdentifiablesFactory,
		a.config.Security.RequestAuthenticator,
		a.config.Security.Authorizer,
		a.pusher,
		a.config.Security.Auditer,
	)

	if err != nil {
		writeHTTPError(w, w.Header().Get("Origin"), err)
		return
	}

	writeHTTPResponse(w, ctx)
}

func (a *apiServer) handleDelete(w http.ResponseWriter, req *http.Request) {

	identity := elemental.IdentityFromCategory(bone.GetValue(req, "category"))

	if !elemental.IsDeleteAllowed(a.config.Model.RelationshipsRegistry, identity) {
		writeHTTPError(w, req.Header.Get("Origin"), elemental.NewError("Not allowed", "Delete operation not allowed on "+identity.Name, "bahamut", http.StatusMethodNotAllowed))
		return
	}

	request, err := elemental.NewRequestFromHTTPRequest(req)
	if err != nil {
		writeHTTPError(w, req.Header.Get("Origin"), elemental.NewError("Bad Request", err.Error(), "bahamut", http.StatusBadRequest))
		return
	}

	ctx, err := dispatchDeleteOperation(
		request,
		a.processorFinder,
		a.config.Model.IdentifiablesFactory,
		a.config.Security.RequestAuthenticator,
		a.config.Security.Authorizer,
		a.pusher,
		a.config.Security.Auditer,
	)

	if err != nil {
		writeHTTPError(w, w.Header().Get("Origin"), err)
		return
	}

	writeHTTPResponse(w, ctx)
}

func (a *apiServer) handleRetrieveMany(w http.ResponseWriter, req *http.Request) {

	identity := elemental.IdentityFromCategory(bone.GetValue(req, "category"))
	parentIdentity := elemental.IdentityFromCategory(bone.GetValue(req, "parentcategory"))

	if parentIdentity.IsEmpty() {
		parentIdentity = elemental.RootIdentity
	}

	if !elemental.IsRetrieveManyAllowed(a.config.Model.RelationshipsRegistry, identity, parentIdentity) {
		writeHTTPError(w, req.Header.Get("Origin"), elemental.NewError("Not allowed", "RetrieveMany operation not allowed on "+identity.Category, "bahamut", http.StatusMethodNotAllowed))
		return
	}

	request, err := elemental.NewRequestFromHTTPRequest(req)
	if err != nil {
		writeHTTPError(w, req.Header.Get("Origin"), elemental.NewError("Bad Request", err.Error(), "bahamut", http.StatusBadRequest))
		return
	}

	ctx, err := dispatchRetrieveManyOperation(
		request,
		a.processorFinder,
		a.config.Model.IdentifiablesFactory,
		a.config.Security.RequestAuthenticator,
		a.config.Security.Authorizer,
		a.pusher,
		a.config.Security.Auditer,
	)

	if err != nil {
		writeHTTPError(w, w.Header().Get("Origin"), err)
		return
	}

	writeHTTPResponse(w, ctx)
}

func (a *apiServer) handleCreate(w http.ResponseWriter, req *http.Request) {

	identity := elemental.IdentityFromCategory(bone.GetValue(req, "category"))
	parentIdentity := elemental.IdentityFromCategory(bone.GetValue(req, "parentcategory"))

	if parentIdentity.IsEmpty() {
		parentIdentity = elemental.RootIdentity
	}

	if !elemental.IsCreateAllowed(a.config.Model.RelationshipsRegistry, identity, parentIdentity) {
		writeHTTPError(w, req.Header.Get("Origin"), elemental.NewError("Not allowed", "Create operation not allowed on "+identity.Name, "bahamut", http.StatusMethodNotAllowed))
		return
	}

	request, err := elemental.NewRequestFromHTTPRequest(req)
	if err != nil {
		writeHTTPError(w, req.Header.Get("Origin"), elemental.NewError("Bad Request", err.Error(), "bahamut", http.StatusBadRequest))
		return
	}

	ctx, err := dispatchCreateOperation(
		request,
		a.processorFinder,
		a.config.Model.IdentifiablesFactory,
		a.config.Security.RequestAuthenticator,
		a.config.Security.Authorizer,
		a.pusher,
		a.config.Security.Auditer,
	)

	if err != nil {
		writeHTTPError(w, w.Header().Get("Origin"), err)
		return
	}

	writeHTTPResponse(w, ctx)
}

func (a *apiServer) handleInfo(w http.ResponseWriter, req *http.Request) {

	identity := elemental.IdentityFromCategory(bone.GetValue(req, "category"))
	parentIdentity := elemental.IdentityFromCategory(bone.GetValue(req, "parentcategory"))

	if parentIdentity.IsEmpty() {
		parentIdentity = elemental.RootIdentity
	}

	if !elemental.IsInfoAllowed(a.config.Model.RelationshipsRegistry, identity, parentIdentity) {
		writeHTTPError(w, req.Header.Get("Origin"), elemental.NewError("Not allowed", "Info operation not allowed on "+identity.Category, "bahamut", http.StatusMethodNotAllowed))
		return
	}

	request, err := elemental.NewRequestFromHTTPRequest(req)
	if err != nil {
		writeHTTPError(w, req.Header.Get("Origin"), elemental.NewError("Bad Request", err.Error(), "bahamut", http.StatusBadRequest))
		return
	}

	ctx, err := dispatchInfoOperation(
		request,
		a.processorFinder,
		a.config.Model.IdentifiablesFactory,
		a.config.Security.RequestAuthenticator,
		a.config.Security.Authorizer,
		a.config.Security.Auditer,
	)

	if err != nil {
		writeHTTPError(w, w.Header().Get("Origin"), err)
		return
	}

	writeHTTPResponse(w, ctx)
}

func (a *apiServer) handlePatch(w http.ResponseWriter, req *http.Request) {

	identity := elemental.IdentityFromCategory(bone.GetValue(req, "category"))
	parentIdentity := elemental.IdentityFromCategory(bone.GetValue(req, "parentcategory"))

	if parentIdentity.IsEmpty() {
		parentIdentity = elemental.RootIdentity
	}

	if !elemental.IsPatchAllowed(a.config.Model.RelationshipsRegistry, identity, parentIdentity) {
		writeHTTPError(w, req.Header.Get("Origin"), elemental.NewError("Not allowed", "Patch operation not allowed on "+identity.Name, "bahamut", http.StatusMethodNotAllowed))
		return
	}

	request, err := elemental.NewRequestFromHTTPRequest(req)
	if err != nil {
		writeHTTPError(w, req.Header.Get("Origin"), elemental.NewError("Bad Request", err.Error(), "bahamut", http.StatusBadRequest))
		return
	}

	ctx, err := dispatchPatchOperation(
		request,
		a.processorFinder,
		a.config.Model.IdentifiablesFactory,
		a.config.Security.RequestAuthenticator,
		a.config.Security.Authorizer,
		a.pusher,
		a.config.Security.Auditer,
	)

	if err != nil {
		writeHTTPError(w, w.Header().Get("Origin"), err)
		return
	}

	writeHTTPResponse(w, ctx)
}

// installRoutes installs all the routes declared in the APIServerConfig.
func (a *apiServer) installRoutes() {

	a.multiplexer.Options("*", http.HandlerFunc(corsHandler))
	a.multiplexer.Get("/", http.HandlerFunc(corsHandler))
	a.multiplexer.NotFound(http.HandlerFunc(notFoundHandler))

	// non versioned routes
	a.multiplexer.Get("/:category/:id", http.HandlerFunc(a.handleRetrieve))
	a.multiplexer.Put("/:category/:id", http.HandlerFunc(a.handleUpdate))
	a.multiplexer.Delete("/:category/:id", http.HandlerFunc(a.handleDelete))
	a.multiplexer.Get("/:category", http.HandlerFunc(a.handleRetrieveMany))
	a.multiplexer.Get("/:parentcategory/:id/:category", http.HandlerFunc(a.handleRetrieveMany))
	a.multiplexer.Post("/:category", http.HandlerFunc(a.handleCreate))
	a.multiplexer.Post("/:parentcategory/:id/:category", http.HandlerFunc(a.handleCreate))
	a.multiplexer.Head("/:category", http.HandlerFunc(a.handleInfo))
	a.multiplexer.Head("/:parentcategory/:id/:category", http.HandlerFunc(a.handleInfo))
	a.multiplexer.Patch("/:category", http.HandlerFunc(a.handlePatch))
	a.multiplexer.Patch("/:parentcategory/:id/:category", http.HandlerFunc(a.handlePatch))

	// versioned routes
	a.multiplexer.Get("/v/:version/:category/:id", http.HandlerFunc(a.handleRetrieve))
	a.multiplexer.Put("/v/:version/:category/:id", http.HandlerFunc(a.handleUpdate))
	a.multiplexer.Delete("/v/:version/:category/:id", http.HandlerFunc(a.handleDelete))
	a.multiplexer.Get("/v/:version/:category", http.HandlerFunc(a.handleRetrieveMany))
	a.multiplexer.Get("/v/:version/:parentcategory/:id/:category", http.HandlerFunc(a.handleRetrieveMany))
	a.multiplexer.Post("/v/:version/:category", http.HandlerFunc(a.handleCreate))
	a.multiplexer.Post("/v/:version/:parentcategory/:id/:category", http.HandlerFunc(a.handleCreate))
	a.multiplexer.Head("/v/:version/:category", http.HandlerFunc(a.handleInfo))
	a.multiplexer.Head("/v/:version/:parentcategory/:id/:category", http.HandlerFunc(a.handleInfo))
	a.multiplexer.Patch("/v/:version/:category", http.HandlerFunc(a.handlePatch))
	a.multiplexer.Patch("/v/:version/:parentcategory/:id/:category", http.HandlerFunc(a.handlePatch))
}

// start starts the apiServer.
func (a *apiServer) start() {

	a.installRoutes()

	zap.L().Info("Starting api server", zap.String("address", a.config.ReSTServer.ListenAddress))

	var err error
	if a.config.TLS.ServerCertificates != nil {
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

	if a.config.TLS.ServerCertificates != nil {
		err = a.server.ListenAndServeTLS("", "")
	} else {
		err = a.server.ListenAndServe()
	}

	if err != nil {
		zap.L().Fatal("Unable to start api server", zap.Error(err))
	}
}

// stop stops the apiServer.
func (a *apiServer) stop() {

	// a.server.Shutdown() // Uncomment with Go 1.8
	// a.server = nil
}
