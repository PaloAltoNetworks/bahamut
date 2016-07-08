// Author: Antoine Mercadal
// See LICENSE file for full LICENSE
// Copyright 2016 Aporeto.

package bahamut

import (
	"net/http"
	"testing"

	"github.com/go-zoo/bone"
	. "github.com/smartystreets/goconvey/convey"
)

func TestServer_Initialization(t *testing.T) {

	Convey("Given I create a new Cid server", t, func() {

		c := newAPIServer("address:80", bone.New(), []*Route{})

		Convey("Then it should be correctly initialized", func() {
			So(c.address, ShouldEqual, "address:80")
			So(len(c.multiplexer.Routes), ShouldEqual, 2)
		})
	})
}

func TestServer_RouteInstallation(t *testing.T) {

	Convey("Given I create a new Cid Server with routes", t, func() {

		h := func(w http.ResponseWriter, req *http.Request) {}

		var routes []*Route
		routes = append(routes, NewRoute("/lists", http.MethodPost, h))
		routes = append(routes, NewRoute("/lists", http.MethodGet, h))
		routes = append(routes, NewRoute("/lists", http.MethodDelete, h))
		routes = append(routes, NewRoute("/lists", http.MethodPatch, h))
		routes = append(routes, NewRoute("/lists", http.MethodHead, h))
		routes = append(routes, NewRoute("/lists", http.MethodPut, h))

		c := newAPIServer("address:80", bone.New(), routes)

		Convey("Then the bon2 Multiplexer should be correctly populated", func() {

			So(len(c.multiplexer.Routes[http.MethodPost]), ShouldEqual, 1)
			So(len(c.multiplexer.Routes[http.MethodGet]), ShouldEqual, 2)
			So(len(c.multiplexer.Routes[http.MethodDelete]), ShouldEqual, 1)
			So(len(c.multiplexer.Routes[http.MethodPatch]), ShouldEqual, 1)
			So(len(c.multiplexer.Routes[http.MethodHead]), ShouldEqual, 1)
			So(len(c.multiplexer.Routes[http.MethodPut]), ShouldEqual, 1)
			So(len(c.multiplexer.Routes[http.MethodOptions]), ShouldEqual, 1)
		})
	})
}
