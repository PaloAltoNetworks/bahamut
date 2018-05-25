package simple

import "github.com/aporeto-inc/bahamut"

// A Authorizer is a bahamut.Authorizer compliant structure to authorize
// requests using a given functions.
type Authorizer struct {
	customAuthFunc CustomAuthRequestFunc
}

// NewAuthorizer returns a new *Authorizer.
func NewAuthorizer(customAuthFunc CustomAuthRequestFunc) *Authorizer {

	return &Authorizer{
		customAuthFunc: customAuthFunc,
	}
}

// IsAuthorized authorizer the given context.
// It will return true if the authentication is a success, false in case of failure
// and an eventual error in case of error.
func (a *Authorizer) IsAuthorized(ctx *bahamut.Context) (bahamut.AuthAction, error) {

	if a.customAuthFunc == nil {
		return bahamut.AuthActionContinue, nil
	}

	action, err := a.customAuthFunc(ctx)
	if err != nil {
		return bahamut.AuthActionKO, err
	}

	return action, nil
}
