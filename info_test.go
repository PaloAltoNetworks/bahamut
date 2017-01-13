// Author: Antoine Mercadal
// See LICENSE file for full LICENSE
// Copyright 2016 Aporeto.

package bahamut

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/aporeto-inc/elemental"
	. "github.com/smartystreets/goconvey/convey"
)

func TestInfo_Parameters(t *testing.T) {

	Convey("Given create a new Info", t, func() {

		i := newInfo()

		Convey("When I read a request with parameters", func() {

			url, _ := url.Parse("http://link.com/path?param1=1&param2=2")
			req := &http.Request{
				Host:   "link.com",
				URL:    url,
				Method: http.MethodGet,
			}

			request, _ := elemental.NewRequestFromHTTPRequest(req)
			i.fromElementalRequest(request)

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

		i := newInfo()

		Convey("When I read a request with parameters", func() {

			url, _ := url.Parse("http://link.com/path?param1=1&param2=2")
			req := &http.Request{
				Host:   "link.com",
				URL:    url,
				Method: http.MethodGet,
				Header: make(http.Header),
			}

			request, _ := elemental.NewRequestFromHTTPRequest(req)
			i.fromElementalRequest(request)

			Convey("Then the value of param1 should be 1", func() {
				So(i.Parameters.Get("param1"), ShouldEqual, "1")
			})

			Convey("Then the value of param2 should be 2", func() {
				So(i.Parameters.Get("param2"), ShouldEqual, "2")
			})
		})
	})
}

func TestInfo_Components(t *testing.T) {

	Convey("Given create a new Info and an identity", t, func() {

		i := newInfo()

		Convey("When I read for a root object", func() {

			u, _ := url.Parse("http://test.com/lists")
			req := &http.Request{
				Host: "test.com",
				URL:  u,
			}

			request, _ := elemental.NewRequestFromHTTPRequest(req)
			i.fromElementalRequest(request)

			Convey("The the Parent Identity should be correct", func() {
				So(i.ChildrenIdentity.Name, ShouldEqual, ListIdentity.Name)
			})

			Convey("The the Children Identity should be empty", func() {
				So(i.ParentIdentity.IsEmpty(), ShouldBeTrue)
			})
		})

		Convey("When I read for a particular object", func() {

			u, _ := url.Parse("http://test.com/lists/1")
			req := &http.Request{
				Host: "test.com",
				URL:  u,
			}
			request, _ := elemental.NewRequestFromHTTPRequest(req)
			i.fromElementalRequest(request)

			Convey("The the Parent Identity should be correct", func() {
				So(i.ChildrenIdentity.Name, ShouldEqual, ListIdentity.Name)
			})

			Convey("The the Children Identity should be empty", func() {
				So(i.ParentIdentity.IsEmpty(), ShouldBeTrue)
			})
		})

		Convey("When I read for a children of an object", func() {

			u, _ := url.Parse("http://test.com/lists/1/tasks")
			req := &http.Request{
				Host: "test.com",
				URL:  u,
			}

			request, _ := elemental.NewRequestFromHTTPRequest(req)
			i.fromElementalRequest(request)

			Convey("The the Parent Identity should be correct", func() {
				So(i.ParentIdentity.Name, ShouldEqual, ListIdentity.Name)
			})

			Convey("The the Children Identity should be correct", func() {
				So(i.ChildrenIdentity.Name, ShouldEqual, TaskIdentity.Name)
			})
		})
	})
}

func TestInfo_String(t *testing.T) {

	Convey("Given I have an Info", t, func() {

		i := &Info{
			Parameters:       url.Values{"hello": []string{"world"}},
			Headers:          http.Header{"header": []string{"h1"}},
			ParentIdentity:   elemental.EmptyIdentity,
			ParentIdentifier: "xxxx",
			ChildrenIdentity: elemental.EmptyIdentity,
		}

		Convey("When I use the String method", func() {

			s := i.String()

			Convey("Then the should string should be correct", func() {
				So(s, ShouldEqual, "<info parameters:map[hello:[world]] headers:map[header:[h1]] parent-identity: <Identity |> parent-id: xxxx children-identity: <Identity |>>")
			})
		})
	})
}

func TestInfo_fromElementalRequest(t *testing.T) {

	Convey("Given I have an Info and a elemental Request", t, func() {

		r := elemental.NewRequest()
		r.Namespace = "ns"
		r.Operation = elemental.OperationCreate
		r.Identity = ListIdentity
		r.ObjectID = "1"
		r.Username = "toto"
		r.Password = "password"

		i := &Info{}

		Convey("When I run fromElementalRequest", func() {

			i.fromElementalRequest(r)

			Convey("Then the parentIdentifier be set", func() {
				So(i.ParentIdentifier, ShouldEqual, "1")
			})

			Convey("Then the parentIdentity be set", func() {
				So(i.ChildrenIdentity, ShouldResemble, ListIdentity)
			})

			Convey("Then the Headers be correct", func() {
				So(i.Headers.Get("X-Namespace"), ShouldEqual, "ns")
				So(i.Headers.Get("Authorization"), ShouldEqual, "toto password")
			})
		})
	})

	Convey("Given I have an Info and a elemental Request with a parent", t, func() {

		r := elemental.NewRequest()
		r.Namespace = "ns"
		r.Operation = elemental.OperationCreate
		r.Identity = ListIdentity
		r.ObjectID = "1"
		r.Username = "toto"
		r.Password = "password"
		r.ParentIdentity = TaskIdentity
		r.ParentID = "2"
		r.Recursive = true

		i := &Info{}

		Convey("When I run fromElementalRequest", func() {

			i.fromElementalRequest(r)

			Convey("Then the parentIdentifier be set", func() {
				So(i.ParentIdentifier, ShouldEqual, "2")
			})

			Convey("Then the parentIdentity be set", func() {
				So(i.ParentIdentity, ShouldResemble, TaskIdentity)
			})

			Convey("Then the childrenIdentity be set", func() {
				So(i.ChildrenIdentity, ShouldResemble, ListIdentity)
			})

			Convey("Then the Headers be correct", func() {
				So(i.Headers.Get("X-Namespace"), ShouldEqual, "ns")
				So(i.Headers.Get("Authorization"), ShouldEqual, "toto password")
				So(i.Headers.Get("X-Request-Recursive"), ShouldEqual, "true")
			})
		})
	})
}
