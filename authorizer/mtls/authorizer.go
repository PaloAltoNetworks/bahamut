package mtls

import (
	"crypto/x509"

	"github.com/aporeto-inc/bahamut"
	"github.com/aporeto-inc/elemental"
)

type mtlsAuthenticator struct {
	ignoredIdentitied []elemental.Identity
	defaultAction     bahamut.AuthAction
}

type simpleMTLSAuthorizer struct {
	mandatoryOrganizations       []string
	mandatoryOrganizationalUnits []string
	mandatoryCNs                 []string

	mtlsAuthenticator
}

// NewSimpleMTLSAuthorizer returns a new Authorizer that ensures the client certificate contains at least
// one O and/or OUs and/or CNs present in the given list (pass nil to allow all).
// The Authorizer will not enforce this for identities given by ignoredIdentitied.
func NewSimpleMTLSAuthorizer(o, ous, cns []string, ignoredIdentitied []elemental.Identity, defaultAction bahamut.AuthAction) bahamut.Authorizer {

	return &simpleMTLSAuthorizer{
		mandatoryOrganizations:       o,
		mandatoryOrganizationalUnits: ous,
		mandatoryCNs:                 cns,
		mtlsAuthenticator: mtlsAuthenticator{
			ignoredIdentitied: ignoredIdentitied,
			defaultAction:     defaultAction,
		},
	}
}

func (a *simpleMTLSAuthorizer) IsAuthorized(ctx *bahamut.Context) (bahamut.AuthAction, error) {

	for _, i := range a.ignoredIdentitied {
		if ctx.Request.Identity.IsEqual(i) {
			return bahamut.AuthActionContinue, nil
		}
	}

	err := verifyPeerCertificates(
		ctx.Request.TLSConnectionState.PeerCertificates,
		a.mandatoryOrganizations,
		a.mandatoryOrganizationalUnits,
		a.mandatoryCNs,
	)
	if err != nil {
		return bahamut.AuthActionKO, err
	}

	return a.defaultAction, err
}

type verifierMTLSAuthorizer struct {
	verifyOptions x509.VerifyOptions

	mtlsAuthenticator
}

// NewMTLSVerifierAuthorizer returns a new Authorizer that ensures the client certificate are
// can be verified using the given x509.VerifyOptions.
// The Authorizer will not enforce this for identities given by ignoredIdentitied.
func NewMTLSVerifierAuthorizer(verifyOptions x509.VerifyOptions, ignoredIdentitied []elemental.Identity, defaultAction bahamut.AuthAction) bahamut.Authorizer {

	return &verifierMTLSAuthorizer{
		verifyOptions: verifyOptions,
		mtlsAuthenticator: mtlsAuthenticator{
			ignoredIdentitied: ignoredIdentitied,
			defaultAction:     defaultAction,
		},
	}
}

func (a *verifierMTLSAuthorizer) IsAuthorized(ctx *bahamut.Context) (bahamut.AuthAction, error) {

	for _, i := range a.ignoredIdentitied {
		if ctx.Request.Identity.IsEqual(i) {
			return bahamut.AuthActionContinue, nil
		}
	}

	for _, cert := range ctx.Request.TLSConnectionState.PeerCertificates {
		if _, err := cert.Verify(a.verifyOptions); err != nil {
			return bahamut.AuthActionKO, nil
		}
	}

	return a.defaultAction, nil
}
