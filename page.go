// Author: Antoine Mercadal
// See LICENSE file for full LICENSE
// Copyright 2016 Aporeto.

package bahamut

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

const (
	// DefaultPageSize is Default page size value
	DefaultPageSize = 100
)

// Page holds pagination information.
type Page struct {
	Current int
	Size    int
	Next    string
	Prev    string
	First   string
	Last    string
}

func (p *Page) String() string {

	return fmt.Sprintf("<page current:%d size:%d>",
		p.Current,
		p.Size,
	)
}

// newPage returns a new *Page.
func newPage() *Page {

	return &Page{
		Current: 1,
		Size:    DefaultPageSize,
	}
}

// IndexRange returns the index range of data that needs to be retrieved according to current Page's values.
func (p *Page) IndexRange() (start, end int) {

	start = p.Size * (p.Current - 1)
	end = start + p.Size

	return start, end
}

// FromValues populates the Page from an url.Values.
func (p *Page) fromValues(query url.Values) {

	var err error
	p.Current, err = strconv.Atoi(query.Get("page"))
	if err != nil {
		p.Current = 1
	}

	p.Size, err = strconv.Atoi(query.Get("per_page"))
	if err != nil {
		p.Size = DefaultPageSize
	}
}

// compute computes the various fields of the Page, like the neighbor page links, etc.
func (p *Page) compute(baseURL string, query url.Values, totalCount int) {

	query.Set("page", strconv.Itoa(1))
	p.First = strings.Join([]string{baseURL, query.Encode()}, "?")

	if p.Current > 1 {
		query.Set("page", strconv.Itoa(p.Current-1))
		p.Prev = strings.Join([]string{baseURL, query.Encode()}, "?")
	}

	if p.Current*p.Size < totalCount {
		query.Set("page", strconv.Itoa(p.Current+1))
		p.Next = strings.Join([]string{baseURL, query.Encode()}, "?")
	}

	last := totalCount / p.Size
	modulo := totalCount % p.Size

	if last == 0 {
		last = 1
	}

	if modulo != 0 {
		last = last + 1
	}

	query.Set("page", strconv.Itoa(last))
	p.Last = strings.Join([]string{baseURL, query.Encode()}, "?")
}
