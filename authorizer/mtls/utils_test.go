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
	"crypto/x509"
	"encoding/pem"
	"os"
	"reflect"
	"testing"
)

func Test_makeClaims(t *testing.T) {

	cdata, _ := os.ReadFile("./fixtures/claim-test-cert.pem")
	cblock, _ := pem.Decode(cdata)
	cert, _ := x509.ParseCertificate(cblock.Bytes)

	type args struct {
		cert *x509.Certificate
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			"simple",
			args{
				cert,
			},
			[]string{
				"@auth:realm=certificate",
				"@auth:mode=internal",
				"@auth:serialnumber=240974276977353940447659278772794983018",
				"@auth:commonname=test",
				"@auth:organization=A",
				"@auth:organizationalunit=B",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := makeClaims(tt.args.cert); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("makeClaims() = %v, want %v", got, tt.want)
			}
		})
	}
}
