package mtls

import (
	"crypto/x509"

	"github.com/aporeto-inc/bahamut"
	"github.com/aporeto-inc/elemental"
)

type mtlsVerifier struct {
	verifyOptions     x509.VerifyOptions
	ignoredIdentitied []elemental.Identity
	authActionSuccess bahamut.AuthAction
	authActionFailure bahamut.AuthAction
}

func newMTLSVerifier(
	verifyOptions x509.VerifyOptions,
	authActionSuccess bahamut.AuthAction,
	authActionFailure bahamut.AuthAction,
	ignoredIdentitied []elemental.Identity,
) *mtlsVerifier {

	return &mtlsVerifier{
		verifyOptions:     verifyOptions,
		ignoredIdentitied: ignoredIdentitied,
		authActionSuccess: authActionSuccess,
		authActionFailure: authActionFailure,
	}
}

// NewMTLSAuthorizer returns a new Authorizer that ensures the client certificate are
// can be verified using the given x509.VerifyOptions.
// The Authorizer will not enforce this for identities given by ignoredIdentitied.
func NewMTLSAuthorizer(
	verifyOptions x509.VerifyOptions,
	authActionSuccess bahamut.AuthAction,
	authActionFailure bahamut.AuthAction,
	ignoredIdentitied []elemental.Identity,
) bahamut.Authorizer {

	return newMTLSVerifier(verifyOptions, authActionSuccess, authActionFailure, ignoredIdentitied)
}

// NewMTLSRequestAuthenticator returns a new Authenticator that ensures the client certificate are
// can be verified using the given x509.VerifyOptions.
// The Authenticator will not enforce this for identities given by ignoredIdentitied.
func NewMTLSRequestAuthenticator(
	verifyOptions x509.VerifyOptions,
	authActionSuccess bahamut.AuthAction,
	authActionFailure bahamut.AuthAction,
) bahamut.RequestAuthenticator {

	return newMTLSVerifier(verifyOptions, authActionSuccess, authActionFailure, nil)
}

func (a *mtlsVerifier) IsAuthorized(ctx *bahamut.Context) (bahamut.AuthAction, error) {

	for _, i := range a.ignoredIdentitied {
		if ctx.Request.Identity.IsEqual(i) {
			return bahamut.AuthActionContinue, nil
		}
	}

	if ctx.Request.TLSConnectionState == nil {
		return bahamut.AuthActionContinue, nil
	}

	// If we can verify, we return the success auth action.
	for _, cert := range ctx.Request.TLSConnectionState.PeerCertificates {
		if _, err := cert.Verify(a.verifyOptions); err == nil {
			return a.authActionSuccess, nil
		}
	}

	// If we can verify, we return the failure auth action.
	return a.authActionFailure, nil
}

func (a *mtlsVerifier) AuthenticateRequest(req *elemental.Request, claimsHolder elemental.ClaimsHolder) (bahamut.AuthAction, error) {

	if req.TLSConnectionState == nil {
		return bahamut.AuthActionContinue, nil
	}

	// If we can verify, we return the success auth action
	for _, cert := range req.TLSConnectionState.PeerCertificates {
		if _, err := cert.Verify(a.verifyOptions); err == nil {
			return a.authActionSuccess, nil
		}
	}

	// If we can verify, we return the failure auth action.
	return a.authActionFailure, nil
}
