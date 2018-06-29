package simple

import (
	"go.aporeto.io/bahamut"
)

// CustomAuthRequestFunc is the type of functions that can be used to
// decide custom authentication operations for requests. It returns a bahamut.AuthAction.
type CustomAuthRequestFunc func(bahamut.Context) (bahamut.AuthAction, error)

// CustomAuthSessionFunc is the type of functions that can be used to
// decide custom authentication operations sessions. It returns a bahamut.AuthAction.
type CustomAuthSessionFunc func(bahamut.Session) (bahamut.AuthAction, error)

// A Authenticator is a bahamut.Authenticator compliant structure to authentify
// requests using a given functions.
type Authenticator struct {
	customAuthRequestFunc CustomAuthRequestFunc
	customAuthSessionFunc CustomAuthSessionFunc
}

// NewAuthenticator returns a new *Authenticator.
func NewAuthenticator(customAuthRequestFunc CustomAuthRequestFunc, customAuthSessionFunc CustomAuthSessionFunc) *Authenticator {

	return &Authenticator{
		customAuthSessionFunc: customAuthSessionFunc,
		customAuthRequestFunc: customAuthRequestFunc,
	}
}

// AuthenticateSession authenticates the given session.
// It will return true if the authentication is a success, false in case of failure
// and an eventual error in case of error.
func (a *Authenticator) AuthenticateSession(session bahamut.Session) (bahamut.AuthAction, error) {

	if a.customAuthSessionFunc == nil {
		return bahamut.AuthActionContinue, nil
	}

	action, err := a.customAuthSessionFunc(session)
	if err != nil {
		return bahamut.AuthActionKO, err
	}

	return action, nil
}

// AuthenticateRequest authenticates the request from the given bahamut.Context.
// It will return true if the authentication is a success, false in case of failure
// and an eventual error in case of error.
func (a *Authenticator) AuthenticateRequest(ctx bahamut.Context) (bahamut.AuthAction, error) {

	if a.customAuthRequestFunc == nil {
		return bahamut.AuthActionContinue, nil
	}

	action, err := a.customAuthRequestFunc(ctx)
	if err != nil {
		return bahamut.AuthActionKO, err
	}

	return action, nil
}
