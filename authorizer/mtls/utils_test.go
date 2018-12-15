package mtls

import (
	"crypto/x509"
	"encoding/pem"
	"io/ioutil"
	"reflect"
	"testing"
)

func Test_makeClaims(t *testing.T) {

	cdata, _ := ioutil.ReadFile("./fixtures/claim-test-cert.pem")
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
