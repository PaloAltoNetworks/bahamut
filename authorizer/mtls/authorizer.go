// Copyright 2019 Aporeto Inc.
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//     http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package mtls

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"strings"

	"go.aporeto.io/bahamut"
	"go.aporeto.io/elemental"
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
// use this function securely, the service using an mtls authenticator preferring header must be behind a proxy
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
// use this function securely, the service using an mtls authenticator preferring header must be behind a proxy
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
// verification on the certificates, like checking against a certificate
// revocation list. Note that CRL checking is not done by
// Go when using x509.VerifyOptions. If you need need advanced CRL check
// you need to implement it in a VerifierFunc.
type VerifierFunc func(*x509.Certificate) bool

// DeciderFunc is the type of function to pass to decide
// what bahamut.Action to return after the MTLS check is done.
// It will be given the mtls result action, and the bahamut.Context or bahamut.Session
// according to the kind of authorization. If bahamut.Context is given, bahamut.Session will
// be nil and vice versa.
type DeciderFunc func(bahamut.AuthAction, bahamut.Context, bahamut.Session) bahamut.AuthAction

type mtlsVerifier struct {
	verifyOptions        x509.VerifyOptions
	ignoredIdentities    []elemental.Identity
	deciderFunc          DeciderFunc
	verifier             VerifierFunc
	certificateCheckMode CertificateCheckMode
}

func newMTLSVerifier(
	verifyOptions x509.VerifyOptions,
	deciderFunc DeciderFunc,
	ignoredIdentities []elemental.Identity,
	verifier VerifierFunc,
	certificateCheckMode CertificateCheckMode,
) *mtlsVerifier {

	return &mtlsVerifier{
		verifyOptions:        verifyOptions,
		ignoredIdentities:    ignoredIdentities,
		deciderFunc:          deciderFunc,
		verifier:             verifier,
		certificateCheckMode: certificateCheckMode,
	}
}

// NewMTLSAuthorizer returns a new Authorizer that ensures the client certificate
// can be verified using the given x509.VerifyOptions.
// The Authorizer will not enforce this for identities given by ignoredIdentities.
//
// deciderFunc is the DeciderFunc to used return the actual action you want the Authorizer
// to return.
func NewMTLSAuthorizer(
	verifyOptions x509.VerifyOptions,
	deciderFunc DeciderFunc,
	ignoredIdentities []elemental.Identity,
	certVerifier VerifierFunc,
	certificateCheckMode CertificateCheckMode,
) bahamut.Authorizer {

	return newMTLSVerifier(verifyOptions, deciderFunc, ignoredIdentities, certVerifier, certificateCheckMode)
}

// NewMTLSRequestAuthenticator returns a new Authenticator that ensures the client certificate
// can be verified using the given x509.VerifyOptions.
// The Authenticator will not enforce this for identities given by ignoredIdentities.
//
// deciderFunc is the DeciderFunc to used return the actual action you want the RequestAuthenticator
// to return.
func NewMTLSRequestAuthenticator(
	verifyOptions x509.VerifyOptions,
	deciderFunc DeciderFunc,
	certVerifier VerifierFunc,
	certificateCheckMode CertificateCheckMode,
) bahamut.RequestAuthenticator {

	return newMTLSVerifier(verifyOptions, deciderFunc, nil, certVerifier, certificateCheckMode)
}

// NewMTLSSessionAuthenticator returns a new Authenticator that ensures the client certificate are
// can be verified using the given x509.VerifyOptions.
// The Authenticator will not enforce this for identities given by ignoredIdentities.
//
// deciderFunc is the DeciderFunc to used return the actual action you want the SessionAuthenticator
// to return.
func NewMTLSSessionAuthenticator(
	verifyOptions x509.VerifyOptions,
	deciderFunc DeciderFunc,
	certVerifier VerifierFunc,
	certificateCheckMode CertificateCheckMode,
) bahamut.SessionAuthenticator {

	return newMTLSVerifier(verifyOptions, deciderFunc, nil, certVerifier, certificateCheckMode)
}

func (a *mtlsVerifier) IsAuthorized(ctx bahamut.Context) (bahamut.AuthAction, error) {

	req := ctx.Request()

	for _, i := range a.ignoredIdentities {
		if req.Identity.IsEqual(i) {
			return bahamut.AuthActionContinue, nil
		}
	}

	if req.TLSConnectionState == nil && req.Headers.Get(tlsHeaderKey) == "" {
		return bahamut.AuthActionContinue, nil
	}

	var certs []*x509.Certificate
	var err error

	switch a.certificateCheckMode {
	case CertificateCheckModeTLSStateOnly:
		certs, err = CertificatesFromTLSState(req.TLSConnectionState)
	case CertificateCheckModeTLSStateThenHeader:
		certs, err = CertificatesFromTLSStateThenHeader(req.TLSConnectionState, req.Headers.Get(tlsHeaderKey))
	case CertificateCheckModeHeaderThenTLSState:
		certs, err = CertificatesFromHeaderThenTLSState(req.TLSConnectionState, req.Headers.Get(tlsHeaderKey))
	case CertificateCheckModeHeaderOnly:
		certs, err = CertificatesFromHeader(req.Headers.Get(tlsHeaderKey))
	}

	if err != nil {
		return a.deciderFunc(bahamut.AuthActionKO, ctx, nil), nil
	}

	// If we can verify, we return the success auth action.
	for _, cert := range certs {
		if _, err := cert.Verify(a.verifyOptions); err == nil {
			if a.verifier == nil || a.verifier(cert) {
				return a.deciderFunc(bahamut.AuthActionOK, ctx, nil), nil
			}
		}
	}

	// If we can't verify, we return the failure auth action.
	return a.deciderFunc(bahamut.AuthActionKO, ctx, nil), nil
}

func (a *mtlsVerifier) AuthenticateRequest(ctx bahamut.Context) (bahamut.AuthAction, error) {

	ac, err := a.checkAction(ctx.Request().TLSConnectionState, ctx.Request().Headers.Get(tlsHeaderKey), ctx.SetClaims)
	if err != nil {
		return bahamut.AuthActionKO, err
	}

	return a.deciderFunc(ac, ctx, nil), nil
}

func (a *mtlsVerifier) AuthenticateSession(session bahamut.Session) (bahamut.AuthAction, error) {

	ac, err := a.checkAction(session.TLSConnectionState(), "", session.SetClaims)
	if err != nil {
		return bahamut.AuthActionKO, err
	}

	return a.deciderFunc(ac, nil, session), nil
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
		return bahamut.AuthActionKO, nil
	}

	// If we can verify, we return the success auth action
	for _, cert := range certs {
		if _, err := cert.Verify(a.verifyOptions); err == nil {
			if a.verifier == nil || a.verifier(cert) {
				claimSetter(makeClaims(cert))
				return bahamut.AuthActionOK, nil
			}
		}
	}

	// If we can't verify, we return the failure auth action.
	return bahamut.AuthActionKO, nil
}

func decodeCertHeader(header string) ([]*x509.Certificate, error) {

	if len(header) < 54 {
		return nil, errors.New("invalid certificate in header")
	}
	// TODO: support multiple of them.
	header = fmt.Sprintf("-----BEGIN CERTIFICATE-----\n%s\n-----END CERTIFICATE-----", strings.Replace(header[28:len(header)-26], " ", "\n", -1))

	var certs []*x509.Certificate
	var pemBlock *pem.Block
	rest := []byte(header)

	for {
		pemBlock, rest = pem.Decode(rest)
		if pemBlock == nil {
			return nil, fmt.Errorf("no valid cert in '%s'", header)
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
