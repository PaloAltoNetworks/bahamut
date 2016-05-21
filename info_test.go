// Author: Antoine Mercadal
// See LICENSE file for full LICENSE
// Copyright 2016 Aporeto.

package bahamut

import (
	"crypto/tls"
	"net/http"
	"net/url"
	"testing"

	"github.com/aporeto-inc/elemental"
	. "github.com/smartystreets/goconvey/convey"
)

func TestInfo_BaseRawURL(t *testing.T) {

	Convey("Given create a new Info", t, func() {

		elemental.RegisterIdentity(elemental.MakeIdentity("parent", "parents"))

		i := NewInfo()

		Convey("When I read from an invalid http request", func() {

			req := &http.Request{
				Host: "test.com",
			}

			Convey("Then it should panic", func() {
				So(func() { i.FromRequest(req) }, ShouldPanic)
			})
		})

		Convey("When I read from a valid http request", func() {

			u, _ := url.Parse("http://test.com/path")
			req := &http.Request{
				Host: "test.com",
				URL:  u,
			}

			i.FromRequest(req)

			Convey("Then BaseRawURL should be correct", func() {
				So(i.BaseRawURL, ShouldEqual, "http://test.com/path")
			})
		})

		Convey("When I read from a valid https request", func() {

			u, _ := url.Parse("http://test.com/path")
			req := &http.Request{
				Host: "test.com",
				TLS:  &tls.ConnectionState{},
				URL:  u,
			}

			i.FromRequest(req)

			Convey("Then BaseRawURL should be correct", func() {
				So(i.BaseRawURL, ShouldEqual, "https://test.com/path")
			})

		})
	})
}

func TestInfo_Parameters(t *testing.T) {

	Convey("Given create a new Info", t, func() {

		i := NewInfo()

		Convey("When I read a request with parameters", func() {

			url, _ := url.Parse("http://link.com/path?param1=1&param2=2")
			req := &http.Request{
				Host:   "link.com",
				URL:    url,
				Method: http.MethodGet,
			}

			i.FromRequest(req)

			Convey("Then parameters 'param1' should be correct", func() {
				So(i.Parameters.Get("param1"), ShouldEqual, "1")
			})

			Convey("Then parameters 'param2' should be correct", func() {
				So(i.Parameters.Get("param2"), ShouldEqual, "2")
			})
		})
	})
}

func TestInfo_Headers(t *testing.T) {

	Convey("Given create a new Info", t, func() {

		i := NewInfo()

		Convey("When I read a request with headers", func() {

			url, _ := url.Parse("http://link.com/path?param1=1&param2=2")
			req := &http.Request{
				Host:   "link.com",
				URL:    url,
				Method: http.MethodGet,
				Header: make(http.Header),
			}

			req.Header.Add("X-Hello", "hello")
			req.Header.Add("X-World", "world")

			i.FromRequest(req)

			Convey("Then the value of Header for X-Hello should be hello", func() {
				So(i.Headers.Get("x-hello"), ShouldEqual, "hello")
			})

			Convey("Then the value of Header for X-World should be world", func() {
				So(i.Headers.Get("x-world"), ShouldEqual, "world")
			})
		})
	})
}

func TestInfo_Components(t *testing.T) {

	parentIdentity := elemental.MakeIdentity("parent", "parents")
	childIdentity := elemental.MakeIdentity("child", "children")
	elemental.RegisterIdentity(parentIdentity)
	elemental.RegisterIdentity(childIdentity)

	Convey("Given create a new Info and an identity", t, func() {

		i := NewInfo()

		Convey("When I read for a root object", func() {

			u, _ := url.Parse("http://test.com/parents")
			req := &http.Request{
				Host: "test.com",
				URL:  u,
			}

			i.FromRequest(req)

			Convey("The the Parent Identity should be correct", func() {
				So(i.ParentIdentity.Name, ShouldEqual, parentIdentity.Name)
			})

			Convey("The the Children Identity should be empty", func() {
				So(i.ChildrenIdentity.IsEmpty(), ShouldBeTrue)
			})
		})

		Convey("When I read for a particular object", func() {

			u, _ := url.Parse("http://test.com/parents/1")
			req := &http.Request{
				Host: "test.com",
				URL:  u,
			}

			i.FromRequest(req)

			Convey("The the Parent Identity should be correct", func() {
				So(i.ParentIdentity.Name, ShouldEqual, parentIdentity.Name)
			})

			Convey("The the Children Identity should be empty", func() {
				So(i.ChildrenIdentity.IsEmpty(), ShouldBeTrue)
			})
		})

		Convey("When I read for a children of an object", func() {

			u, _ := url.Parse("http://test.com/parents/1/children")
			req := &http.Request{
				Host: "test.com",
				URL:  u,
			}

			i.FromRequest(req)

			Convey("The the Parent Identity should be correct", func() {
				So(i.ParentIdentity.Name, ShouldEqual, parentIdentity.Name)
			})

			Convey("The the Children Identity should be correct", func() {
				So(i.ChildrenIdentity.Name, ShouldEqual, childIdentity.Name)
			})
		})
	})
}
