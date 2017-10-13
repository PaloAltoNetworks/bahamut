// Package barret provides a bahamut.Authorizer used to check if
// a certificate that has been used to issue the token is revoked.
package barret

import (
	"time"

	"github.com/aporeto-inc/addedeffect/cache"
	"github.com/aporeto-inc/addedeffect/certification"
	"github.com/aporeto-inc/bahamut"
	"github.com/aporeto-inc/manipulate"
)

const (
	authClaimRealmKey            = "@auth:realm"
	authClaimCertSerialNumberKey = "@auth:serialnumber"
	authRealmCertificateKey      = "certificate"
)

// A Authorizer is a simple authenticator that will verify
// if the given Certificate realm issued Midgard Token has been revoked.
type barretAuthorizer struct {
	manipulator   manipulate.Manipulator
	authCache     cache.Cacher
	cacheDuration time.Duration
}

// NewBarretAuthorizer returns a new bahamut.Authorizer backed by Barret.
func NewBarretAuthorizer(m manipulate.Manipulator, cacheDuration time.Duration) bahamut.Authorizer {

	return &barretAuthorizer{
		manipulator:   m,
		authCache:     cache.NewMemoryCache(),
		cacheDuration: cacheDuration,
	}
}

// IsAuthorized is the main method that returns whether the API call is authorized or not.
func (a *barretAuthorizer) IsAuthorized(ctx *bahamut.Context) (bahamut.AuthAction, error) {

	// If it is not a token issued from a certificate, we do nothing.
	cm := ctx.GetClaimsMap()
	if cm[authClaimRealmKey] != authRealmCertificateKey {
		return bahamut.AuthActionContinue, nil
	}

	token := ctx.Request.Password
	if ok := a.authCache.Get(token); ok != nil {
		return ok.(bahamut.AuthAction), nil
	}

	if err := certification.CheckRevocation(a.manipulator, cm[authClaimCertSerialNumberKey], nil); err != nil {
		a.authCache.SetWithExpiration(token, bahamut.AuthActionKO, a.cacheDuration)
		return bahamut.AuthActionKO, err
	}

	a.authCache.SetWithExpiration(token, bahamut.AuthActionContinue, a.cacheDuration)
	return bahamut.AuthActionContinue, nil
}
