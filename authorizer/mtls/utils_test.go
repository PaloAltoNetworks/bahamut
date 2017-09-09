package mtlsauthorizer

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func loadFixtureCertificates() (*x509.CertPool, *x509.CertPool, []*x509.Certificate) {

	systemCAPool, _ := x509.SystemCertPool()

	clientCACertData, _ := ioutil.ReadFile("../../fixtures/ca.pem")
	clientCAPool := x509.NewCertPool()
	clientCAPool.AppendCertsFromPEM(clientCACertData)

	cert, err := tls.LoadX509KeyPair("../../fixtures/client-cert.pem", "../../fixtures/client-key.pem")
	if err != nil {
		panic(err)
	}

	c, err := x509.ParseCertificate(cert.Certificate[0])
	if err != nil {
		panic(err)
	}

	return systemCAPool, clientCAPool, []*x509.Certificate{c}
}

func TestBahamut_verifyPeerCertificates(t *testing.T) {

	Convey("Given I have a certificate", t, func() {

		_, _, clientcerts := loadFixtureCertificates()

		Convey("When I call verifyPeerCertificates with expected os, ous, and cn", func() {

			err := verifyPeerCertificates(clientcerts, []string{"aporeto.com"}, []string{"SuperAdmin"}, []string{"superadmin"})

			Convey("Then err should be nil", func() {
				So(err, ShouldBeNil)
			})
		})

		Convey("When I call verifyPeerCertificates with expected os, ous", func() {

			err := verifyPeerCertificates(clientcerts, []string{"aporeto.com"}, []string{"SuperAdmin"}, nil)

			Convey("Then err should be nil", func() {
				So(err, ShouldBeNil)
			})
		})

		Convey("When I call verifyPeerCertificates with expected os", func() {

			err := verifyPeerCertificates(clientcerts, []string{"aporeto.com"}, nil, nil)

			Convey("Then err should be nil", func() {
				So(err, ShouldBeNil)
			})
		})

		Convey("When I call verifyPeerCertificates with expected os and cn", func() {

			err := verifyPeerCertificates(clientcerts, []string{"aporeto.com"}, nil, []string{"superadmin"})

			Convey("Then err should be nil", func() {
				So(err, ShouldBeNil)
			})
		})

		Convey("When I call verifyPeerCertificates with expected ous and cn", func() {

			err := verifyPeerCertificates(clientcerts, nil, []string{"SuperAdmin"}, []string{"superadmin"})

			Convey("Then err should be nil", func() {
				So(err, ShouldBeNil)
			})
		})

		Convey("When I call verifyPeerCertificates with additional o, additional ous and cn", func() {

			err := verifyPeerCertificates(clientcerts, []string{"aporeto.com", "biloute.com"}, []string{"SuperAdmin", "toto"}, []string{"superadmin"})

			Convey("Then err should be nil", func() {
				So(err, ShouldBeNil)
			})
		})

		Convey("When I call verifyPeerCertificates with missing o, ous and cn", func() {

			err := verifyPeerCertificates(clientcerts, []string{"biloute.com"}, []string{"SuperAdmin", "toto"}, []string{"superadmin"})

			Convey("Then err should not be nil", func() {
				So(err, ShouldNotBeNil)
			})
		})

		Convey("When I call verifyPeerCertificates with o, missing ous and cn", func() {

			err := verifyPeerCertificates(clientcerts, []string{"aporeto.com", "biloute.com"}, []string{"toto"}, []string{"superadmin"})

			Convey("Then err should not be nil", func() {
				So(err, ShouldNotBeNil)
			})
		})

		Convey("When I call verifyPeerCertificates with o, ous and missing cn", func() {

			err := verifyPeerCertificates(clientcerts, []string{"aporeto.com", "biloute.com"}, []string{"SuperAdmin", "toto"}, []string{"nop"})

			Convey("Then err should not be nil", func() {
				So(err, ShouldNotBeNil)
			})
		})

		Convey("When I call verifyPeerCertificates with no certificate", func() {

			err := verifyPeerCertificates(nil, []string{"aporeto.com", "biloute.com"}, []string{"SuperAdmin", "toto"}, []string{"nop"})

			Convey("Then err should not be nil", func() {
				So(err, ShouldNotBeNil)
			})
		})
	})
}
