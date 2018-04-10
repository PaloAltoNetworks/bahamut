package mtls

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"strings"

	"github.com/aporeto-inc/bahamut"
	"github.com/aporeto-inc/elemental"
)

const tlsHeaderKey = "X-TLS-Client-Certificate"

// CertificatesFromStateOrHeader retrieves the certificates in either from the tls connection
// state or from the header X-TLS-Client-Certificate in that order.
func CertificatesFromStateOrHeader(state *tls.ConnectionState, headerData string) (certs []*x509.Certificate, err error) {

	if state != nil && len(state.PeerCertificates) > 0 {
		return state.PeerCertificates, nil
	}

	if headerData != "" {
		return decodeCertHeader(headerData)
	}

	return nil, errors.New("no valid certificate found in tls state or header")
}

// VerifierFunc is the type of function you can pass to do custom
// verification on the certificates, like checking for the DN.
type VerifierFunc func(*x509.Certificate) bool

type mtlsVerifier struct {
	verifyOptions     x509.VerifyOptions
	ignoredIdentities []elemental.Identity
	authActionSuccess bahamut.AuthAction
	authActionFailure bahamut.AuthAction
	verifier          VerifierFunc
}

func newMTLSVerifier(
	verifyOptions x509.VerifyOptions,
	authActionSuccess bahamut.AuthAction,
	authActionFailure bahamut.AuthAction,
	ignoredIdentities []elemental.Identity,
	verifier VerifierFunc,
) *mtlsVerifier {

	return &mtlsVerifier{
		verifyOptions:     verifyOptions,
		ignoredIdentities: ignoredIdentities,
		authActionSuccess: authActionSuccess,
		authActionFailure: authActionFailure,
		verifier:          verifier,
	}
}

// NewMTLSAuthorizer returns a new Authorizer that ensures the client certificate are
// can be verified using the given x509.VerifyOptions.
// The Authorizer will not enforce this for identities given by ignoredIdentities.
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
	ignoredIdentities []elemental.Identity,
	certVerifier VerifierFunc,
) bahamut.Authorizer {

	return newMTLSVerifier(verifyOptions, authActionSuccess, authActionFailure, ignoredIdentities, certVerifier)
}

// NewMTLSRequestAuthenticator returns a new Authenticator that ensures the client certificate are
// can be verified using the given x509.VerifyOptions.
// The Authenticator will not enforce this for identities given by ignoredIdentities.
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
	certVerifier VerifierFunc,
) bahamut.RequestAuthenticator {

	return newMTLSVerifier(verifyOptions, authActionSuccess, authActionFailure, nil, certVerifier)
}

// NewMTLSSessionAuthenticator returns a new Authenticator that ensures the client certificate are
// can be verified using the given x509.VerifyOptions.
// The Authenticator will not enforce this for identities given by ignoredIdentities.
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
	certVerifier VerifierFunc,
) bahamut.SessionAuthenticator {

	return newMTLSVerifier(verifyOptions, authActionSuccess, authActionFailure, nil, certVerifier)
}

func (a *mtlsVerifier) IsAuthorized(ctx *bahamut.Context) (bahamut.AuthAction, error) {

	for _, i := range a.ignoredIdentities {
		if ctx.Request.Identity.IsEqual(i) {
			return bahamut.AuthActionContinue, nil
		}
	}

	if ctx.Request.TLSConnectionState == nil && ctx.Request.Headers.Get(tlsHeaderKey) == "" {
		return bahamut.AuthActionContinue, nil
	}

	certs, err := CertificatesFromStateOrHeader(ctx.Request.TLSConnectionState, ctx.Request.Headers.Get(tlsHeaderKey))
	if err != nil {
		return a.authActionFailure, nil
	}

	// If we can verify, we return the success auth action.
	for _, cert := range certs {
		if _, err := cert.Verify(a.verifyOptions); err == nil {

			if paliateGo110VerificationBug(a.verifyOptions, cert) {

				if a.verifier != nil {
					if !a.verifier(cert) {
						return a.authActionFailure, nil
					}
				}

				return a.authActionSuccess, nil
			}
		}
	}

	// If we can't verify, we return the failure auth action.
	return a.authActionFailure, nil
}

func (a *mtlsVerifier) AuthenticateRequest(ctx *bahamut.Context) (bahamut.AuthAction, error) {

	return a.checkAction(ctx.Request.TLSConnectionState, ctx.Request.Headers.Get(tlsHeaderKey), ctx.SetClaims)
}

func (a *mtlsVerifier) AuthenticateSession(session bahamut.Session) (bahamut.AuthAction, error) {

	return a.checkAction(session.TLSConnectionState(), "", session.SetClaims)
}

func (a *mtlsVerifier) checkAction(tlsState *tls.ConnectionState, headerCert string, claimSetter func([]string)) (bahamut.AuthAction, error) {

	if tlsState == nil && headerCert == "" {
		return bahamut.AuthActionContinue, nil
	}

	certs, err := CertificatesFromStateOrHeader(tlsState, headerCert)
	if err != nil {
		return a.authActionFailure, nil
	}

	// If we can verify, we return the success auth action
	for _, cert := range certs {
		if _, err := cert.Verify(a.verifyOptions); err == nil {
			if paliateGo110VerificationBug(a.verifyOptions, cert) {

				if a.verifier != nil {
					if !a.verifier(cert) {
						return a.authActionFailure, nil
					}
				}

				claimSetter(makeClaims(cert))
				return a.authActionSuccess, nil
			}
		}
	}

	// If we can't verify, we return the failure auth action.
	return a.authActionFailure, nil
}

func decodeCertHeader(header string) ([]*x509.Certificate, error) {

	if len(header) < 54 {
		return nil, errors.New("Invalid certificate in header")
	}
	// TODO: support multiple of them.
	header = fmt.Sprintf("-----BEGIN CERTIFICATE-----\n%s\n-----END CERTIFICATE-----", strings.Replace(header[28:len(header)-26], " ", "\n", -1))

	var certs []*x509.Certificate
	var pemBlock *pem.Block
	rest := []byte(header)

	for {
		pemBlock, rest = pem.Decode(rest)
		if pemBlock == nil {
			return nil, fmt.Errorf("No valid cert in: %s", header)
		}
		cert, err := x509.ParseCertificate(pemBlock.Bytes)
		if err != nil {
			return nil, err
		}
		certs = append(certs, cert)
		if len(rest) == 0 {
			break
		}
	}

	return certs, nil
}

func paliateGo110VerificationBug(opts x509.VerifyOptions, cert *x509.Certificate) bool {

	for _, neededEKU := range opts.KeyUsages {
		for _, currentEKU := range cert.ExtKeyUsage {
			if neededEKU == currentEKU {
				return true
			}
		}
	}

	return false
}
