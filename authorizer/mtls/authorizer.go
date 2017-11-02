package mtls

import (
	"crypto/tls"
	"crypto/x509"

	"github.com/aporeto-inc/bahamut"
	"github.com/aporeto-inc/elemental"
)

// VerifierFunc is the type of function you can pass to do custom
// verification on the certificates, like checking for the DN.
// NOTE: Not implemented yet.
type VerifierFunc func(*x509.Certificate) (bahamut.AuthAction, error)

type mtlsVerifier struct {
	verifyOptions     x509.VerifyOptions
	ignoredIdentitied []elemental.Identity
	authActionSuccess bahamut.AuthAction
	authActionFailure bahamut.AuthAction
	// verifier          VerifierFunc
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
//
// authActionSuccess is the bahamut.AuthAction to return if the verification succeeds.
// This lets you a chance to return either bahamut.AuthActionOK to definitely validate
// the call, or to return bahamut.AuthActionContinue to continue the authorizer chain.
//
// authActionFailure is the bahamut.AuthAction to return if the verification fails.
// This lets you a chance to return either bahamut.AuthActionKO to definitely fail
// the call, or to return bahamut.AuthActionContinue to continue the authorizer chain.
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
//
// authActionSuccess is the bahamut.AuthAction to return if the verification succeeds.
// This lets you a chance to return either bahamut.AuthActionOK to definitely validate
// the call, or to return bahamut.AuthActionContinue to continue the authorizer chain.
//
// authActionFailure is the bahamut.AuthAction to return if the verification fails.
// This lets you a chance to return either bahamut.AuthActionKO to definitely fail
// the call, or to return bahamut.AuthActionContinue to continue the authorizer chain.
func NewMTLSRequestAuthenticator(
	verifyOptions x509.VerifyOptions,
	authActionSuccess bahamut.AuthAction,
	authActionFailure bahamut.AuthAction,
) bahamut.RequestAuthenticator {

	return newMTLSVerifier(verifyOptions, authActionSuccess, authActionFailure, nil)
}

// NewMTLSSessionAuthenticator returns a new Authenticator that ensures the client certificate are
// can be verified using the given x509.VerifyOptions.
// The Authenticator will not enforce this for identities given by ignoredIdentitied.
//
// authActionSuccess is the bahamut.AuthAction to return if the verification succeeds.
// This lets you a chance to return either bahamut.AuthActionOK to definitely validate
// the call, or to return bahamut.AuthActionContinue to continue the authorizer chain.
//
// authActionFailure is the bahamut.AuthAction to return if the verification fails.
// This lets you a chance to return either bahamut.AuthActionKO to definitely fail
// the call, or to return bahamut.AuthActionContinue to continue the authorizer chain.
func NewMTLSSessionAuthenticator(
	verifyOptions x509.VerifyOptions,
	authActionSuccess bahamut.AuthAction,
	authActionFailure bahamut.AuthAction,
) bahamut.SessionAuthenticator {

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

	// If we can't verify, we return the failure auth action.
	return a.authActionFailure, nil
}

func (a *mtlsVerifier) AuthenticateRequest(ctx *bahamut.Context) (bahamut.AuthAction, error) {

	return a.checkAction(ctx.Request.TLSConnectionState, ctx.SetClaims)
}

func (a *mtlsVerifier) AuthenticateSession(session bahamut.Session) (bahamut.AuthAction, error) {

	return a.checkAction(session.TLSConnectionState(), session.SetClaims)
}

func (a *mtlsVerifier) checkAction(tlsState *tls.ConnectionState, claimSetter func([]string)) (bahamut.AuthAction, error) {

	if tlsState == nil {
		return bahamut.AuthActionContinue, nil
	}

	// If we can verify, we return the success auth action
	for _, cert := range tlsState.PeerCertificates {
		if _, err := cert.Verify(a.verifyOptions); err == nil {
			claimSetter(makeClaims(cert))
			return a.authActionSuccess, nil
		}
	}

	// If we can't verify, we return the failure auth action.
	return a.authActionFailure, nil
}
