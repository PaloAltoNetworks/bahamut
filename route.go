// Author: Antoine Mercadal
// See LICENSE file for full LICENSE
// Copyright 2016 Aporeto.

package bahamut

import "net/http"

// A Route represents a Bahamut Route.
type Route struct {
	Method  string
	Handler http.HandlerFunc
	Pattern string
}

// NewRoute returns a new Route.
func NewRoute(pattern, method string, handler http.HandlerFunc) *Route {

	return &Route{
		Method:  method,
		Pattern: pattern,
		Handler: handler,
	}
}
