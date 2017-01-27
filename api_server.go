// Author: Antoine Mercadal
// See LICENSE file for full LICENSE
// Copyright 2016 Aporeto.

package bahamut

import (
	"crypto/tls"
	"net/http"

	"github.com/aporeto-inc/elemental"
	"github.com/go-zoo/bone"
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
		Certificates:             a.config.TLS.ServerCertificates,
		ClientAuth:               a.config.TLS.AuthType,
		ClientCAs:                a.config.TLS.ClientCAPool,
		RootCAs:                  a.config.TLS.RootCAPool,
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
		ReadTimeout:  a.config.ReSTServer.ReadTimeout,
		WriteTimeout: a.config.ReSTServer.WriteTimeout,
		// IdleTimeout:  a.config.ReSTServer.IdleTimeout, // Uncomment with Go 1.8
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

func (a *apiServer) handleRetrieve(w http.ResponseWriter, req *http.Request) {

	identity := elemental.IdentityFromCategory(bone.GetValue(req, "category"))

	if !elemental.IsRetrieveAllowed(a.config.Model.RelationshipsRegistry, identity) {
		writeHTTPError(w, req.Header.Get("Origin"), elemental.NewError("Not allowed", "Method not allowed on "+identity.Name, "bahamut", http.StatusMethodNotAllowed))
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
		a.config.Security.Authenticator,
		a.config.Security.Authorizer,
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
		writeHTTPError(w, req.Header.Get("Origin"), elemental.NewError("Not allowed", "Method not allowed on "+identity.Name, "bahamut", http.StatusMethodNotAllowed))
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
		a.config.Security.Authenticator,
		a.config.Security.Authorizer,
		a.pusher,
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
		writeHTTPError(w, req.Header.Get("Origin"), elemental.NewError("Not allowed", "Method not allowed on "+identity.Name, "bahamut", http.StatusMethodNotAllowed))
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
		a.config.Security.Authenticator,
		a.config.Security.Authorizer,
		a.pusher,
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
		writeHTTPError(w, req.Header.Get("Origin"), elemental.NewError("Not allowed", "Method not allowed on "+identity.Name, "bahamut", http.StatusMethodNotAllowed))
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
		a.config.Security.Authenticator,
		a.config.Security.Authorizer,
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
		writeHTTPError(w, req.Header.Get("Origin"), elemental.NewError("Not allowed", "Method not allowed on "+identity.Name, "bahamut", http.StatusMethodNotAllowed))
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
		a.config.Security.Authenticator,
		a.config.Security.Authorizer,
		a.pusher,
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
		writeHTTPError(w, req.Header.Get("Origin"), elemental.NewError("Not allowed", "Method not allowed on "+identity.Name, "bahamut", http.StatusMethodNotAllowed))
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
		a.config.Security.Authenticator,
		a.config.Security.Authorizer,
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
		writeHTTPError(w, req.Header.Get("Origin"), elemental.NewError("Not allowed", "Method not allowed on "+identity.Name, "bahamut", http.StatusMethodNotAllowed))
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
		a.config.Security.Authenticator,
		a.config.Security.Authorizer,
		a.pusher,
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
}

// start starts the apiServer.
func (a *apiServer) start() {

	a.installRoutes()

	log.WithField("address", a.config.ReSTServer.ListenAddress).Info("Starting api server.")

	var err error
	if a.config.TLS.ServerCertificates != nil {
		a.server, err = a.createSecureHTTPServer(a.config.ReSTServer.ListenAddress)
	} else {
		a.server, err = a.createUnsecureHTTPServer(a.config.ReSTServer.ListenAddress)
	}
	if err != nil {
		log.WithError(err).Fatal("Unable to create api server.")
	}

	a.server.Handler = a.multiplexer
	a.server.SetKeepAlivesEnabled(true)

	if a.config.TLS.ServerCertificates != nil {
		err = a.server.ListenAndServeTLS("", "")
	} else {
		err = a.server.ListenAndServe()
	}

	if err != nil {
		log.WithError(err).Fatal("Unable to start api server.")
	}
}

// stop stops the apiServer.
func (a *apiServer) stop() {

	// a.server.Shutdown() // Uncomment with Go 1.8
	// a.server = nil
}
