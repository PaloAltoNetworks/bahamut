package mtlsauthorizer

import (
	"github.com/aporeto-inc/bahamut"
	"github.com/aporeto-inc/elemental"
)

type simpleMTLSAuthorizer struct {
	mandatoryOrganizations       []string
	mandatoryOrganizationalUnits []string
	mandatoryCNs                 []string
	ignoredIdentitied            []elemental.Identity
}

// NewSimpleMTLSAuthorizer returns a new Authorizer that ensures the client certificate contains at least
// one O and/or OUs and/or CNs present in the given list (pass nil to allow all).
// The Authorizer will not enforce this for identities given by ignoredIdentitied.
func NewSimpleMTLSAuthorizer(o, ous, cns []string, ignoredIdentitied []elemental.Identity) bahamut.Authorizer {

	return &simpleMTLSAuthorizer{
		mandatoryOrganizations:       o,
		mandatoryOrganizationalUnits: ous,
		mandatoryCNs:                 cns,
		ignoredIdentitied:            ignoredIdentitied,
	}
}

func (a *simpleMTLSAuthorizer) IsAuthorized(ctx *bahamut.Context) (bool, error) {

	for _, i := range a.ignoredIdentitied {
		if ctx.Request.Identity.IsEqual(i) {
			return true, nil
		}
	}

	err := verifyPeerCertificates(
		ctx.Request.TLSConnectionState.PeerCertificates,
		a.mandatoryOrganizations,
		a.mandatoryOrganizationalUnits,
		a.mandatoryCNs,
	)

	return err == nil, err
}
