package mtls

import (
	"crypto/tls"
	"testing"

	"github.com/aporeto-inc/bahamut"
	"github.com/aporeto-inc/elemental"
	. "github.com/smartystreets/goconvey/convey"
)

func TestBahamut_MTLSAuthorizer(t *testing.T) {

	Convey("Given I have a certificate and a context", t, func() {

		_, _, clientcerts := loadFixtureCertificates()
		testIdentity := elemental.MakeIdentity("test", "test")
		ctx := &bahamut.Context{
			Request: &elemental.Request{
				TLSConnectionState: &tls.ConnectionState{
					PeerCertificates: clientcerts,
				},
				Identity: testIdentity,
			},
		}

		Convey("When I create an empty authorizer", func() {

			a := NewSimpleMTLSAuthorizer(nil, nil, nil, nil)
			ok, err := a.IsAuthorized(ctx)

			Convey("Then ok should be true", func() {
				So(ok, ShouldBeTrue)
			})

			Convey("Then err should be nil", func() {
				So(err, ShouldBeNil)
			})
		})

		Convey("When I create an non empty authorizer with met expectations", func() {

			a := NewSimpleMTLSAuthorizer([]string{"aporeto.com"}, []string{"SuperAdmin"}, nil, nil)
			ok, err := a.IsAuthorized(ctx)

			Convey("Then ok should be true", func() {
				So(ok, ShouldBeTrue)
			})

			Convey("Then err should be nil", func() {
				So(err, ShouldBeNil)
			})
		})

		Convey("When I create an non empty authorizer with met expectations and ignored identity", func() {

			a := NewSimpleMTLSAuthorizer([]string{"aporeto.com"}, []string{"SuperAdmin"}, nil, []elemental.Identity{testIdentity})
			ok, err := a.IsAuthorized(ctx)

			Convey("Then ok should be true", func() {
				So(ok, ShouldBeTrue)
			})

			Convey("Then err should be nil", func() {
				So(err, ShouldBeNil)
			})
		})

		Convey("When I create an non empty authorizer with unmet expectations", func() {

			a := NewSimpleMTLSAuthorizer([]string{"aporeto.com"}, []string{"nope"}, nil, nil)
			ok, err := a.IsAuthorized(ctx)

			Convey("Then ok should be false", func() {
				So(ok, ShouldBeFalse)
			})

			Convey("Then err should not be nil", func() {
				So(err, ShouldNotBeNil)
			})
		})

		Convey("When I create an non empty authorizer with unmet expectations but I ignore the identity", func() {

			a := NewSimpleMTLSAuthorizer([]string{"aporeto.com"}, []string{"nope"}, nil, []elemental.Identity{testIdentity})
			ok, err := a.IsAuthorized(ctx)

			Convey("Then ok should be true", func() {
				So(ok, ShouldBeTrue)
			})

			Convey("Then err should be nil", func() {
				So(err, ShouldBeNil)
			})
		})
	})
}
