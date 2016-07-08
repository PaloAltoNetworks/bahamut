// Author: Antoine Mercadal
// See LICENSE file for full LICENSE
// Copyright 2016 Aporeto.

package bahamut

import (
	"net/http"

	log "github.com/Sirupsen/logrus"
	"github.com/aporeto-inc/elemental"
	"github.com/go-zoo/bone"
)

type apiServer struct {
	address     string
	multiplexer *bone.Mux
}

func newAPIServer(address string, multiplexer *bone.Mux, routes []*Route) *apiServer {

	srv := &apiServer{
		address:     address,
		multiplexer: multiplexer,
	}

	for _, route := range routes {

		if route.Method == http.MethodHead {
			srv.multiplexer.Head(route.Pattern, http.HandlerFunc(route.Handler))
		} else if route.Method == http.MethodGet {
			srv.multiplexer.Get(route.Pattern, http.HandlerFunc(route.Handler))
		} else if route.Method == http.MethodPost {
			srv.multiplexer.Post(route.Pattern, http.HandlerFunc(route.Handler))
		} else if route.Method == http.MethodPut {
			srv.multiplexer.Put(route.Pattern, http.HandlerFunc(route.Handler))
		} else if route.Method == http.MethodDelete {
			srv.multiplexer.Delete(route.Pattern, http.HandlerFunc(route.Handler))
		} else if route.Method == http.MethodPatch {
			srv.multiplexer.Patch(route.Pattern, http.HandlerFunc(route.Handler))
		}

		log.WithFields(log.Fields{
			"pattern": route.Pattern,
			"method":  route.Method,
		}).Debug("route installed")
	}

	srv.multiplexer.Options("*", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		setCommonHeader(w)
		w.WriteHeader(http.StatusOK)
	}))

	srv.multiplexer.Get("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		setCommonHeader(w)
		w.WriteHeader(http.StatusOK)
	}))

	srv.multiplexer.NotFound(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		WriteHTTPError(w, http.StatusNotFound, elemental.NewError("Not Found", "Unable to find the requested resource", "http", http.StatusNotFound))
	}))

	log.WithFields(log.Fields{
		"routes": len(routes) + 3,
	}).Info("all routes installed")

	return srv
}
