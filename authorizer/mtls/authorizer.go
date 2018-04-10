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

// CertificateCheckMode represents the mode to use to
// check the certificate.
type CertificateCheckMode int

// Various value for CertificateCheckMode.
const (
	CertificateCheckModeTLSStateOnly CertificateCheckMode = iota
	CertificateCheckModeTLSStateThenHeader
	CertificateCheckModeHeaderThenTLSState
	CertificateCheckModeHeaderOnly
)

// CertificatesFromHeaderThenTLSState retrieves the certificates in either from the header `X-TLS-Client-Certificate`
// or from the tls connection state in that order.
//
// Note: Using this function on a service directly available on the internet is extremely dangerous as it assumes
// the given certificate has already been validated by a third party and is just used as informative data. To
// use this function securely, the service using an mtls authenticator prefering header must be behind a proxy
// that does mtls authentication first.
func CertificatesFromHeaderThenTLSState(state *tls.ConnectionState, headerData string) (certs []*x509.Certificate, err error) {

	if headerData != "" {
		return decodeCertHeader(headerData)
	}

	if state != nil && len(state.PeerCertificates) > 0 {
		return state.PeerCertificates, nil
	}

	return nil, errors.New("no valid certificate found in header or tls state")
}

// CertificatesFromTLSStateThenHeader retrieves the certificates in either from the tls connection state or
// from the header `X-TLS-Client-Certificate` in that order.
//
// Note: Using this function on a service directly available on the internet is extremely dangerous as it assumes
// the given certificate has already been validated by a third party and is just used as informative data. To
// use this function securely, the service using an mtls authenticator prefering header must be behind a proxy
// that does mtls authentication first.
func CertificatesFromTLSStateThenHeader(state *tls.ConnectionState, headerData string) (certs []*x509.Certificate, err error) {

	if state != nil && len(state.PeerCertificates) > 0 {
		return state.PeerCertificates, nil
	}

	if headerData != "" {
		return decodeCertHeader(headerData)
	}

	return nil, errors.New("no valid certificate found in tls state or header")
}

// CertificatesFromTLSState retrieves the certificates from the tls connection state.
func CertificatesFromTLSState(state *tls.ConnectionState) (certs []*x509.Certificate, err error) {

	if state == nil || len(state.PeerCertificates) == 0 {
		return nil, errors.New("no valid certificate found in tls state or header")
	}

	return state.PeerCertificates, nil
}

// CertificatesFromHeader retrieves the certificates from the http header `X-TLS-Client-Certificate`.
func CertificatesFromHeader(headerData string) (certs []*x509.Certificate, err error) {

	if headerData == "" {
		return nil, errors.New("no valid certificate found in header")
	}

	return decodeCertHeader(headerData)
}

// VerifierFunc is the type of function you can pass to do custom
// verification on the certificates, like checking for the DN.
type VerifierFunc func(*x509.Certificate) bool

type mtlsVerifier struct {
	verifyOptions        x509.VerifyOptions
	ignoredIdentities    []elemental.Identity
	authActionSuccess    bahamut.AuthAction
	authActionFailure    bahamut.AuthAction
	verifier             VerifierFunc
	certificateCheckMode CertificateCheckMode
}

func newMTLSVerifier(
	verifyOptions x509.VerifyOptions,
	authActionSuccess bahamut.AuthAction,
	authActionFailure bahamut.AuthAction,
	ignoredIdentities []elemental.Identity,
	verifier VerifierFunc,
	certificateCheckMode CertificateCheckMode,
) *mtlsVerifier {

	return &mtlsVerifier{
		verifyOptions:        verifyOptions,
		ignoredIdentities:    ignoredIdentities,
		authActionSuccess:    authActionSuccess,
		authActionFailure:    authActionFailure,
		verifier:             verifier,
		certificateCheckMode: certificateCheckMode,
	}
}

// NewMTLSAuthorizer returns a new Authorizer that ensures the client certificate
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
	certificateCheckMode CertificateCheckMode,
) bahamut.Authorizer {

	return newMTLSVerifier(verifyOptions, authActionSuccess, authActionFailure, ignoredIdentities, certVerifier, certificateCheckMode)
}

// NewMTLSRequestAuthenticator returns a new Authenticator that ensures the client certificate
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
	certificateCheckMode CertificateCheckMode,
) bahamut.RequestAuthenticator {

	return newMTLSVerifier(verifyOptions, authActionSuccess, authActionFailure, nil, certVerifier, certificateCheckMode)
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
	certificateCheckMode CertificateCheckMode,
) bahamut.SessionAuthenticator {

	return newMTLSVerifier(verifyOptions, authActionSuccess, authActionFailure, nil, certVerifier, certificateCheckMode)
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

	var certs []*x509.Certificate
	var err error

	switch a.certificateCheckMode {
	case CertificateCheckModeTLSStateOnly:
		certs, err = CertificatesFromTLSState(ctx.Request.TLSConnectionState)
	case CertificateCheckModeTLSStateThenHeader:
		certs, err = CertificatesFromTLSStateThenHeader(ctx.Request.TLSConnectionState, ctx.Request.Headers.Get(tlsHeaderKey))
	case CertificateCheckModeHeaderThenTLSState:
		certs, err = CertificatesFromHeaderThenTLSState(ctx.Request.TLSConnectionState, ctx.Request.Headers.Get(tlsHeaderKey))
	case CertificateCheckModeHeaderOnly:
		certs, err = CertificatesFromHeader(ctx.Request.Headers.Get(tlsHeaderKey))
	}

	if err != nil {
		return a.authActionFailure, nil
	}

	// If we can verify, we return the success auth action.
	for _, cert := range certs {
		if _, err := cert.Verify(a.verifyOptions); err == nil {

			if paliateGo110VerificationBug(a.verifyOptions, cert) {

				if a.verifier == nil || a.verifier(cert) {
					return a.authActionSuccess, nil
				}
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

	var certs []*x509.Certificate
	var err error

	switch a.certificateCheckMode {
	case CertificateCheckModeTLSStateOnly:
		certs, err = CertificatesFromTLSState(tlsState)
	case CertificateCheckModeTLSStateThenHeader:
		certs, err = CertificatesFromTLSStateThenHeader(tlsState, headerCert)
	case CertificateCheckModeHeaderThenTLSState:
		certs, err = CertificatesFromHeaderThenTLSState(tlsState, headerCert)
	case CertificateCheckModeHeaderOnly:
		certs, err = CertificatesFromHeader(headerCert)
	}

	if err != nil {
		return a.authActionFailure, nil
	}

	// If we can verify, we return the success auth action
	for _, cert := range certs {
		if _, err := cert.Verify(a.verifyOptions); err == nil {
			if paliateGo110VerificationBug(a.verifyOptions, cert) {

				if a.verifier == nil || a.verifier(cert) {
					claimSetter(makeClaims(cert))
					return a.authActionSuccess, nil
				}
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
