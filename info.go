// Author: Antoine Mercadal
// See LICENSE file for full LICENSE
// Copyright 2016 Aporeto.

package bahamut

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/aporeto-inc/elemental"
	"github.com/go-zoo/bone"
)

// Info contains general information about the initial request.
type Info struct {
	Parameters         url.Values
	BaseRawURL         string
	ParentIdentifier   string
	ParentIdentity     elemental.Identity
	ChildrenIdentity   elemental.Identity
	Headers            http.Header
	TLSConnectionState *tls.ConnectionState
}

// newInfo returns a new Info.
func newInfo() *Info {

	return &Info{
		Headers:    make(http.Header),
		Parameters: make(url.Values),
	}
}

// FromRequest populates the Info from an http.Request.
func (i *Info) fromRequest(req *http.Request) {

	if req.URL == nil {
		panic("request must have an url")
	}

	i.Parameters = req.URL.Query()
	i.Headers = req.Header
	i.TLSConnectionState = req.TLS

	var scheme string
	if i.TLSConnectionState != nil {
		scheme = "https"
	} else {
		scheme = "http"
	}

	i.BaseRawURL = scheme + "://" + req.Host + req.URL.Path

	i.ParentIdentifier = bone.GetValue(req, "id")

	components := strings.Split(req.URL.Path, "/")

	if l := len(components); l == 2 || l == 3 {
		i.ChildrenIdentity = elemental.IdentityFromCategory(components[1])
	}

	if l := len(components); l == 4 {
		i.ParentIdentity = elemental.IdentityFromCategory(components[1])
		i.ChildrenIdentity = elemental.IdentityFromCategory(components[3])
	}
}

// FromRequest populates the Info from an elemental.Request.
func (i *Info) fromElementalRequest(req *elemental.Request) {

	i.Parameters = req.Parameters
	i.Headers = http.Header{
		"X-Namespace":   []string{req.Namespace},
		"Authorization": []string{req.Username + " " + req.Password},
	}

	if !req.ParentIdentity.IsEmpty() {
		i.ParentIdentity = req.ParentIdentity
		i.ParentIdentifier = req.ParentID
		i.ChildrenIdentity = req.Identity
	} else {
		i.ParentIdentity = req.Identity
		i.ParentIdentifier = req.ObjectID
	}
}

func (i *Info) String() string {

	return fmt.Sprintf("<info parameters:%v headers:%v parent-identity: %v parent-id: %s children-identity: %v>",
		i.Parameters,
		i.Headers,
		i.ParentIdentity,
		i.ParentIdentifier,
		i.ChildrenIdentity,
	)
}
