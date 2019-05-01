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

import "go.aporeto.io/bahamut"

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
func (a *Authorizer) IsAuthorized(ctx bahamut.Context) (bahamut.AuthAction, error) {

	if a.customAuthFunc == nil {
		return bahamut.AuthActionContinue, nil
	}

	action, err := a.customAuthFunc(ctx)
	if err != nil {
		return bahamut.AuthActionKO, err
	}

	return action, nil
}
