// Author: Antoine Mercadal
// See LICENSE file for full LICENSE
// Copyright 2016 Aporeto.

package bahamut

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"

	"github.com/aporeto-inc/elemental"
)

// Info contains general information about the initial request.
type Info struct {
	Parameters         url.Values
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

// FromRequest populates the Info from an elemental.Request.
func (i *Info) fromElementalRequest(req *elemental.Request) {

	i.Parameters = req.Parameters
	i.TLSConnectionState = req.TLSConnectionState

	i.Headers = http.Header{
		"X-Namespace":   []string{req.Namespace},
		"Authorization": []string{req.Username + " " + req.Password},
	}

	if !req.ParentIdentity.IsEmpty() {
		i.ParentIdentity = req.ParentIdentity
		i.ParentIdentifier = req.ParentID
		i.ChildrenIdentity = req.Identity
	} else {
		i.ChildrenIdentity = req.Identity
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
