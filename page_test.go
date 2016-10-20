// Author: Antoine Mercadal
// See LICENSE file for full LICENSE
// Copyright 2016 Aporeto.

package bahamut

import (
	"net/url"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestPage_FromQuery(t *testing.T) {

	Convey("Given I have a Page", t, func() {

		p := Page{}

		Convey("When I pass an empty query", func() {

			q := url.Values{}
			p.fromValues(q)

			Convey("Then the current page should be 1", func() {
				So(p.Current, ShouldEqual, 1)
			})

			Convey("Then the size should be 100", func() {
				So(p.Size, ShouldEqual, 100)
			})
		})

		Convey("When I pass an query with page set to 42 and per_page set to 4242", func() {

			q := url.Values{}
			q.Set("page", "42")
			q.Set("per_page", "4242")
			p.fromValues(q)

			Convey("Then the current page should be 42", func() {
				So(p.Current, ShouldEqual, 42)
			})

			Convey("Then the size page should be 4242", func() {
				So(p.Size, ShouldEqual, 4242)
			})
		})
	})
}

func TestPage_IndexRange(t *testing.T) {

	Convey("Given I have a Page", t, func() {

		p := Page{}

		Convey("When I set the size to 50 and the current to 1", func() {

			p.Size = 50
			p.Current = 1

			start, end := p.IndexRange()

			Convey("Then start should be 0", func() {
				So(start, ShouldEqual, 0)
			})

			Convey("Then end should be 49", func() {
				So(end, ShouldEqual, 50)
			})
		})

		Convey("When I set the size to 1 and the current to 1", func() {

			p.Size = 1
			p.Current = 1

			start, end := p.IndexRange()

			Convey("Then start should be 0", func() {
				So(start, ShouldEqual, 0)
			})

			Convey("Then end should be 0", func() {
				So(end, ShouldEqual, 1)
			})
		})

		Convey("When I set the size to 10 and the current to 2", func() {

			p.Size = 10
			p.Current = 2

			start, end := p.IndexRange()

			Convey("Then start should be 10", func() {
				So(start, ShouldEqual, 10)
			})

			Convey("Then end should be 20", func() {
				So(end, ShouldEqual, 20)
			})
		})
	})
}

func TestPage_Compute(t *testing.T) {

	Convey("Given I have Response", t, func() {

		p := newPage()

		Convey("When I get the first page and there is no element", func() {

			q := url.Values{}
			q.Set("foo", "bar")
			q.Set("page", "1")
			q.Set("per_page", "10")

			p.fromValues(q)
			p.compute("http://link.com/path", q, 0)

			Convey("Then the first page should be correct", func() {
				So(p.First, ShouldEqual, "http://link.com/path?foo=bar&page=1&per_page=10")
			})

			Convey("Then the prev page should be empty", func() {
				So(p.Prev, ShouldEqual, "")
			})

			Convey("Then the next page should be empty", func() {
				So(p.Next, ShouldEqual, "")
			})

			Convey("Then the last page should be correct", func() {
				So(p.Last, ShouldEqual, "http://link.com/path?foo=bar&page=1&per_page=10")
			})
		})

		Convey("When get the first page of a list that has 2", func() {

			q := url.Values{}
			q.Set("foo", "bar")
			q.Set("page", "1")
			q.Set("per_page", "10")

			p.fromValues(q)
			p.compute("http://link.com/path", q, 20)

			Convey("Then the first page should be correct", func() {
				So(p.First, ShouldEqual, "http://link.com/path?foo=bar&page=1&per_page=10")
			})

			Convey("Then the prev page should be empty", func() {
				So(p.Prev, ShouldEqual, "")
			})

			Convey("Then the next page should be empty", func() {
				So(p.Next, ShouldEqual, "http://link.com/path?foo=bar&page=2&per_page=10")
			})

			Convey("Then the last page should be correct", func() {
				So(p.Last, ShouldEqual, "http://link.com/path?foo=bar&page=2&per_page=10")
			})
		})

		Convey("When get the last page of a list that has 2", func() {

			q := url.Values{}
			q.Set("foo", "bar")
			q.Set("page", "2")
			q.Set("per_page", "10")

			p.fromValues(q)
			p.compute("http://link.com/path", q, 20)

			Convey("Then the first page should be correct", func() {
				So(p.First, ShouldEqual, "http://link.com/path?foo=bar&page=1&per_page=10")
			})

			Convey("Then the previous page should be correct", func() {
				So(p.Prev, ShouldEqual, "http://link.com/path?foo=bar&page=1&per_page=10")
			})

			Convey("Then the next page should be empty", func() {
				So(p.Next, ShouldEqual, "")
			})

			Convey("Then the last page should be correct", func() {
				So(p.Last, ShouldEqual, "http://link.com/path?foo=bar&page=2&per_page=10")
			})

		})

		Convey("When get the middle page of a list that has 3", func() {

			q := url.Values{}
			q.Set("foo", "bar")
			q.Set("page", "2")
			q.Set("per_page", "10")

			p.fromValues(q)
			p.compute("http://link.com/path", q, 30)

			Convey("Then the first page should be correct", func() {
				So(p.First, ShouldEqual, "http://link.com/path?foo=bar&page=1&per_page=10")
			})

			Convey("Then the previous page should be correct", func() {
				So(p.Prev, ShouldEqual, "http://link.com/path?foo=bar&page=1&per_page=10")
			})

			Convey("Then the next page should be correct", func() {
				So(p.Next, ShouldEqual, "http://link.com/path?foo=bar&page=3&per_page=10")
			})

			Convey("Then the last page should be correct", func() {
				So(p.Last, ShouldEqual, "http://link.com/path?foo=bar&page=3&per_page=10")
			})
		})

		Convey("When get the middle page of a list that has 4", func() {

			q := url.Values{}
			q.Set("foo", "bar")
			q.Set("page", "2")
			q.Set("per_page", "10")

			p.fromValues(q)
			p.compute("http://link.com/path", q, 40)

			Convey("Then the first page should be correct", func() {
				So(p.First, ShouldEqual, "http://link.com/path?foo=bar&page=1&per_page=10")
			})

			Convey("Then the previous page should be correct", func() {
				So(p.Prev, ShouldEqual, "http://link.com/path?foo=bar&page=1&per_page=10")
			})

			Convey("Then the next page should be correct", func() {
				So(p.Next, ShouldEqual, "http://link.com/path?foo=bar&page=3&per_page=10")
			})

			Convey("Then the last page should be correct", func() {
				So(p.Last, ShouldEqual, "http://link.com/path?foo=bar&page=4&per_page=10")
			})
		})

		Convey("When get a random page  with an odd number", func() {

			q := url.Values{}
			q.Set("foo", "bar")
			q.Set("page", "2")
			q.Set("per_page", "10")

			p.fromValues(q)
			p.compute("http://link.com/path", q, 41)

			Convey("Then the first page should be correct", func() {
				So(p.First, ShouldEqual, "http://link.com/path?foo=bar&page=1&per_page=10")
			})

			Convey("Then the previous page should be correct", func() {
				So(p.Prev, ShouldEqual, "http://link.com/path?foo=bar&page=1&per_page=10")
			})

			Convey("Then the next page should be correct", func() {
				So(p.Next, ShouldEqual, "http://link.com/path?foo=bar&page=3&per_page=10")
			})

			Convey("Then the last page should be correct", func() {
				So(p.Last, ShouldEqual, "http://link.com/path?foo=bar&page=5&per_page=10")
			})
		})
	})
}

func TestPage_String(t *testing.T) {

	Convey("Given I have Page", t, func() {

		p := &Page{
			Current: 1,
			First:   "http://server.com?page=1",
			Last:    "http://server.com?page=1",
			Next:    "http://server.com?page=2",
			Prev:    "http://server.com?page=0",
			Size:    5,
		}

		Convey("When I use the String method", func() {
			s := p.String()

			Convey("Then the string should be correct", func() {
				So(s, ShouldEqual, "<page current:1 size:5>")
			})
		})
	})
}
