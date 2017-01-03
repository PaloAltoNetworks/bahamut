// Author: Antoine Mercadal
// See LICENSE file for full LICENSE
// Copyright 2016 Aporeto.

package bahamut

import (
	"fmt"

	"github.com/aporeto-inc/elemental"
)

// Page holds pagination information.
type Page struct {
	Current int
	Size    int
	Next    int
	Prev    int
	First   int
	Last    int
}

func (p *Page) String() string {

	return fmt.Sprintf("<page current:%d size:%d>",
		p.Current,
		p.Size,
	)
}

// newPage returns a new *Page.
func newPage() *Page {

	return &Page{}
}

// IndexRange returns the index range of data that needs to be retrieved according to current Page's values.
func (p *Page) IndexRange() (start, end int) {

	start = p.Size * (p.Current - 1)
	end = start + p.Size

	return start, end
}

// FromValues populates the Page from an url.Values.
func (p *Page) fromElementalRequest(request *elemental.Request) {

	p.Current = request.Page
	p.Size = request.PageSize
}

// compute computes the various fields of the Page, like the neighbor page links, etc.
func (p *Page) compute(totalCount int) {

	p.First = 1

	if p.Current > 1 {
		p.Prev = p.Current - 1
	}

	if p.Current*p.Size < totalCount {
		p.Next = p.Current + 1
	}

	if p.Size > 0 {
		last := totalCount / p.Size
		modulo := totalCount % p.Size

		if last == 0 {
			last = 1
		}

		if modulo != 0 {
			last = last + 1
		}

		p.Last = last
	} else {
		p.Last = p.Current
	}
}
