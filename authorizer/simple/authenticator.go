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
