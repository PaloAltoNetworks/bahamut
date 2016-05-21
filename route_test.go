// Author: Antoine Mercadal
// See LICENSE file for full LICENSE
// Copyright 2016 Aporeto.

package bahamut

import (
	"net/http"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestRoute_Initialization(t *testing.T) {

	h := func(w http.ResponseWriter, req *http.Request) {}

	Convey("Given I create a new Route", t, func() {

		r := NewRoute("/object/:id", http.MethodGet, h)

		Convey("Then the Route should be correctly initialized", func() {
			So(r.Handler, ShouldEqual, h)
			So(r.Method, ShouldEqual, http.MethodGet)
			So(r.Pattern, ShouldEqual, "/object/:id")
		})
	})
}
