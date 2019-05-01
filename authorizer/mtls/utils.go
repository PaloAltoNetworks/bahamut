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

import "crypto/x509"

func makeClaims(cert *x509.Certificate) []string {

	claims := []string{
		"@auth:realm=certificate",
		"@auth:mode=internal",
		"@auth:serialnumber=" + cert.SerialNumber.String(),
		"@auth:commonname=" + cert.Subject.CommonName,
	}

	for _, o := range cert.Subject.Organization {
		claims = append(claims, "@auth:organization="+o)
	}

	for _, ou := range cert.Subject.OrganizationalUnit {
		claims = append(claims, "@auth:organizationalunit="+ou)
	}

	return claims
}
