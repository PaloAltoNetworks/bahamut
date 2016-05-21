// Author: Antoine Mercadal
// See LICENSE file for full LICENSE
// Copyright 2016 Aporeto.

package bahamut

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/aporeto-inc/cid/materia/elemental"
	"github.com/go-zoo/bone"
)

// Info represents general information about the initial request.
type Info struct {
	Parameters       url.Values
	BaseRawURL       string
	ParentIdentifier string
	ParentIdentity   elemental.Identity
	ChildrenIdentity elemental.Identity
	Headers          http.Header
}

// NewInfo creates a new *Info.
func NewInfo() *Info {

	return &Info{}
}

// FromRequest populates the Info from an http.Request.
func (i *Info) FromRequest(req *http.Request) {

	if req.URL == nil {
		panic("request must have an url")
	}

	i.Parameters = req.URL.Query()

	i.Headers = req.Header

	var scheme string
	if req.TLS != nil {
		scheme = "https"
	} else {
		scheme = "http"
	}

	i.BaseRawURL = scheme + "://" + req.Host + req.URL.Path

	i.ParentIdentifier = bone.GetValue(req, "id")

	components := strings.Split(req.URL.Path, "/")

	if l := len(components); l == 2 || l == 3 {
		i.ParentIdentity = elemental.IdentityFromCategory(components[1])
	}

	if l := len(components); l == 4 {
		i.ParentIdentity = elemental.IdentityFromCategory(components[1])
		i.ChildrenIdentity = elemental.IdentityFromCategory(components[3])
	}
}
