package mtls

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/aporeto-inc/bahamut"
	"github.com/aporeto-inc/elemental"

	. "github.com/smartystreets/goconvey/convey"
)

type claimsHolder struct {
	claims []string
}

func (h *claimsHolder) SetClaims(c []string)            { h.claims = c }
func (h *claimsHolder) GetClaims() []string             { return h.claims }
func (h *claimsHolder) GetClaimsMap() map[string]string { return nil }

func TestBahamut_MTLSAuthorizer(t *testing.T) {

	Convey("Given I have a some certificates", t, func() {
		caChainAData, _ := ioutil.ReadFile("./fixtures/ca-chain-a.pem")
		certPoolA := x509.NewCertPool()
		certPoolA.AppendCertsFromPEM(caChainAData)

		userCertAData, _ := ioutil.ReadFile("./fixtures/user-a-cert.pem")
		userCertABlock, _ := pem.Decode(userCertAData)
		userCertA, _ := x509.ParseCertificate(userCertABlock.Bytes)

		serverCertAData, _ := ioutil.ReadFile("./fixtures/server-a-cert.pem")
		serverCertABlock, _ := pem.Decode(serverCertAData)
		serverCertA, _ := x509.ParseCertificate(serverCertABlock.Bytes)

		userCertBData, _ := ioutil.ReadFile("./fixtures/user-b-cert.pem")
		userCertBlock, _ := pem.Decode(userCertBData)
		userCertB, _ := x509.ParseCertificate(userCertBlock.Bytes)

		userCertExtData, _ := ioutil.ReadFile("./fixtures/user-ext-cert.pem")
		userCertExtBlock, _ := pem.Decode(userCertExtData)
		userCertExt, _ := x509.ParseCertificate(userCertExtBlock.Bytes)

		Convey("When I try check auth with no certificate provided", func() {

			ctx := &bahamut.Context{
				Request: &elemental.Request{},
			}
			opts := x509.VerifyOptions{
				Roots:     certPoolA,
				KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
			}

			auth := NewMTLSAuthorizer(opts, bahamut.AuthActionOK, bahamut.AuthActionKO, nil, nil)

			action, err := auth.IsAuthorized(ctx)

			Convey("Then err should be nil", func() {
				So(err, ShouldBeNil)
			})

			Convey("Then action should be bahamut.AuthActionContinue", func() {
				So(action, ShouldEqual, bahamut.AuthActionContinue)
			})
		})

		Convey("When I try check auth for user-a using chain-a", func() {

			ctx := &bahamut.Context{
				Request: &elemental.Request{
					TLSConnectionState: &tls.ConnectionState{
						PeerCertificates: []*x509.Certificate{
							userCertA,
						},
					},
				},
			}
			opts := x509.VerifyOptions{
				Roots:     certPoolA,
				KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
			}

			auth := NewMTLSAuthorizer(opts, bahamut.AuthActionOK, bahamut.AuthActionKO, nil, nil)

			action, err := auth.IsAuthorized(ctx)

			Convey("Then err should be nil", func() {
				So(err, ShouldBeNil)
			})

			Convey("Then action should be bahamut.AuthActionOK", func() {
				So(action, ShouldEqual, bahamut.AuthActionOK)
			})
		})

		Convey("When I try check auth for user-a using chain-a as valid inline header", func() {

			header := http.Header{}
			header.Set(tlsHeaderKey, string(userCertAData))
			ctx := &bahamut.Context{
				Request: &elemental.Request{
					Headers: header,
				},
			}
			opts := x509.VerifyOptions{
				Roots:     certPoolA,
				KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
			}

			auth := NewMTLSAuthorizer(opts, bahamut.AuthActionOK, bahamut.AuthActionKO, nil, nil)

			action, err := auth.IsAuthorized(ctx)

			Convey("Then err should be nil", func() {
				So(err, ShouldBeNil)
			})

			Convey("Then action should be bahamut.AuthActionOK", func() {
				So(action, ShouldEqual, bahamut.AuthActionOK)
			})
		})

		Convey("When I try check auth for user-a using chain-a as invalid inline header", func() {

			header := http.Header{}
			header.Set(tlsHeaderKey, "not-good")
			ctx := &bahamut.Context{
				Request: &elemental.Request{
					Headers: header,
				},
			}
			opts := x509.VerifyOptions{
				Roots:     certPoolA,
				KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
			}

			auth := NewMTLSAuthorizer(opts, bahamut.AuthActionOK, bahamut.AuthActionKO, nil, nil)

			action, err := auth.IsAuthorized(ctx)

			Convey("Then err should be nil", func() {
				So(err, ShouldBeNil)
			})

			Convey("Then action should be bahamut.AuthActionOK", func() {
				So(action, ShouldEqual, bahamut.AuthActionKO)
			})
		})

		Convey("When I try check auth for user-a using chain-a whith a verifier function that is ok", func() {

			ctx := &bahamut.Context{
				Request: &elemental.Request{
					TLSConnectionState: &tls.ConnectionState{
						PeerCertificates: []*x509.Certificate{
							userCertA,
						},
					},
				},
			}
			opts := x509.VerifyOptions{
				Roots:     certPoolA,
				KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
			}

			auth := NewMTLSAuthorizer(opts, bahamut.AuthActionOK, bahamut.AuthActionKO, nil, func(cert *x509.Certificate) bool {
				return true
			})

			action, err := auth.IsAuthorized(ctx)

			Convey("Then err should be nil", func() {
				So(err, ShouldBeNil)
			})

			Convey("Then action should be bahamut.AuthActionOK", func() {
				So(action, ShouldEqual, bahamut.AuthActionOK)
			})
		})

		Convey("When I try check auth for user-a using chain-a whith a verifier function that is not ok", func() {

			ctx := &bahamut.Context{
				Request: &elemental.Request{
					TLSConnectionState: &tls.ConnectionState{
						PeerCertificates: []*x509.Certificate{
							userCertA,
						},
					},
				},
			}
			opts := x509.VerifyOptions{
				Roots:     certPoolA,
				KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
			}

			auth := NewMTLSAuthorizer(opts, bahamut.AuthActionOK, bahamut.AuthActionKO, nil, func(cert *x509.Certificate) bool {
				return false
			})

			action, err := auth.IsAuthorized(ctx)

			Convey("Then err should be nil", func() {
				So(err, ShouldBeNil)
			})

			Convey("Then action should be bahamut.AuthActionKO", func() {
				So(action, ShouldEqual, bahamut.AuthActionKO)
			})
		})

		Convey("When I try check auth for server-a using chain-a", func() {

			ctx := &bahamut.Context{
				Request: &elemental.Request{
					TLSConnectionState: &tls.ConnectionState{
						PeerCertificates: []*x509.Certificate{
							serverCertA,
						},
					},
				},
			}
			opts := x509.VerifyOptions{
				Roots:     certPoolA,
				KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
			}

			auth := NewMTLSAuthorizer(opts, bahamut.AuthActionOK, bahamut.AuthActionKO, nil, nil)

			action, err := auth.IsAuthorized(ctx)

			Convey("Then err should be nil", func() {
				So(err, ShouldBeNil)
			})

			Convey("Then action should be bahamut.AuthActionKO", func() {
				So(action, ShouldEqual, bahamut.AuthActionKO)
			})
		})

		Convey("When I try check auth for user-ext using chain-a", func() {

			ctx := &bahamut.Context{
				Request: &elemental.Request{
					TLSConnectionState: &tls.ConnectionState{
						PeerCertificates: []*x509.Certificate{
							userCertExt,
						},
					},
				},
			}
			opts := x509.VerifyOptions{
				Roots:     certPoolA,
				KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
			}

			auth := NewMTLSAuthorizer(opts, bahamut.AuthActionOK, bahamut.AuthActionKO, nil, nil)

			action, err := auth.IsAuthorized(ctx)

			Convey("Then err should be nil", func() {
				So(err, ShouldBeNil)
			})

			Convey("Then action should be bahamut.AuthActionKO", func() {
				So(action, ShouldEqual, bahamut.AuthActionKO)
			})
		})

		Convey("When I try check auth for user-ext using chain-a with a verifier func that is ok", func() {

			ctx := &bahamut.Context{
				Request: &elemental.Request{
					TLSConnectionState: &tls.ConnectionState{
						PeerCertificates: []*x509.Certificate{
							userCertExt,
						},
					},
				},
			}
			opts := x509.VerifyOptions{
				Roots:     certPoolA,
				KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
			}

			auth := NewMTLSAuthorizer(opts, bahamut.AuthActionOK, bahamut.AuthActionKO, nil, func(cert *x509.Certificate) bool {
				return true
			})

			action, err := auth.IsAuthorized(ctx)

			Convey("Then err should be nil", func() {
				So(err, ShouldBeNil)
			})

			Convey("Then action should be bahamut.AuthActionKO", func() {
				So(action, ShouldEqual, bahamut.AuthActionKO)
			})
		})

		Convey("When I try check auth for user-b using chain-a", func() {

			ctx := &bahamut.Context{
				Request: &elemental.Request{
					TLSConnectionState: &tls.ConnectionState{
						PeerCertificates: []*x509.Certificate{
							userCertB,
						},
					},
				},
			}
			opts := x509.VerifyOptions{
				Roots:     certPoolA,
				KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
			}

			auth := NewMTLSAuthorizer(opts, bahamut.AuthActionOK, bahamut.AuthActionKO, nil, nil)

			action, err := auth.IsAuthorized(ctx)

			Convey("Then err should be nil", func() {
				So(err, ShouldBeNil)
			})

			Convey("Then action should be bahamut.AuthActionKO", func() {
				So(action, ShouldEqual, bahamut.AuthActionKO)
			})
		})

		Convey("When I try check auth for server-a using chain-a but using an ignored identity", func() {

			identity := elemental.MakeIdentity("thing", "things")
			ctx := &bahamut.Context{
				Request: &elemental.Request{
					Identity: identity,
					TLSConnectionState: &tls.ConnectionState{
						PeerCertificates: []*x509.Certificate{
							serverCertA,
						},
					},
				},
			}
			opts := x509.VerifyOptions{
				Roots:     certPoolA,
				KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
			}

			auth := NewMTLSAuthorizer(opts, bahamut.AuthActionOK, bahamut.AuthActionKO, []elemental.Identity{identity}, nil)

			action, err := auth.IsAuthorized(ctx)

			Convey("Then err should be nil", func() {
				So(err, ShouldBeNil)
			})

			Convey("Then action should be bahamut.AuthActionContinue", func() {
				So(action, ShouldEqual, bahamut.AuthActionContinue)
			})
		})

	})
}

func TestBahamut_NewMTLSRequestAuthenticator(t *testing.T) {

	Convey("Given I have a some certificates", t, func() {
		caChainAData, _ := ioutil.ReadFile("./fixtures/ca-chain-a.pem")
		certPoolA := x509.NewCertPool()
		certPoolA.AppendCertsFromPEM(caChainAData)

		userCertAData, _ := ioutil.ReadFile("./fixtures/user-a-cert.pem")
		userCertABlock, _ := pem.Decode(userCertAData)
		userCertA, _ := x509.ParseCertificate(userCertABlock.Bytes)

		serverCertAData, _ := ioutil.ReadFile("./fixtures/server-a-cert.pem")
		serverCertABlock, _ := pem.Decode(serverCertAData)
		serverCertA, _ := x509.ParseCertificate(serverCertABlock.Bytes)

		userCertBData, _ := ioutil.ReadFile("./fixtures/user-b-cert.pem")
		userCertBlock, _ := pem.Decode(userCertBData)
		userCertB, _ := x509.ParseCertificate(userCertBlock.Bytes)

		userCertExtData, _ := ioutil.ReadFile("./fixtures/user-ext-cert.pem")
		userCertExtBlock, _ := pem.Decode(userCertExtData)
		userCertExt, _ := x509.ParseCertificate(userCertExtBlock.Bytes)

		Convey("When I try check auth with no certificate provided", func() {

			ctx := &bahamut.Context{Request: &elemental.Request{}}
			opts := x509.VerifyOptions{
				Roots:     certPoolA,
				KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
			}

			auth := NewMTLSRequestAuthenticator(opts, bahamut.AuthActionOK, bahamut.AuthActionKO, nil)

			action, err := auth.AuthenticateRequest(ctx)

			Convey("Then err should be nil", func() {
				So(err, ShouldBeNil)
			})

			Convey("Then action should be bahamut.AuthActionContinue", func() {
				So(action, ShouldEqual, bahamut.AuthActionContinue)
			})
		})

		Convey("When I try check auth for user-a using chain-a", func() {

			ctx := &bahamut.Context{
				Request: &elemental.Request{
					TLSConnectionState: &tls.ConnectionState{
						PeerCertificates: []*x509.Certificate{
							userCertA,
						},
					},
				},
			}

			opts := x509.VerifyOptions{
				Roots:     certPoolA,
				KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
			}

			auth := NewMTLSRequestAuthenticator(opts, bahamut.AuthActionOK, bahamut.AuthActionKO, nil)

			action, err := auth.AuthenticateRequest(ctx)

			Convey("Then err should be nil", func() {
				So(err, ShouldBeNil)
			})

			Convey("Then action should be bahamut.AuthActionOK", func() {
				So(action, ShouldEqual, bahamut.AuthActionOK)
			})

			Convey("Then claims should be correctly populated", func() {
				So(ctx.GetClaims(), ShouldResemble, []string{"@auth:realm=certificate", "@auth:mode=internal", "@auth:serialnumber=23486181163925715704694891313232533542", "@auth:commonname=user-a"})
			})
		})

		Convey("When I try check auth for user-a using chain-a with a verifier func that is ok", func() {

			ctx := &bahamut.Context{
				Request: &elemental.Request{
					TLSConnectionState: &tls.ConnectionState{
						PeerCertificates: []*x509.Certificate{
							userCertA,
						},
					},
				},
			}

			opts := x509.VerifyOptions{
				Roots:     certPoolA,
				KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
			}

			auth := NewMTLSRequestAuthenticator(opts, bahamut.AuthActionOK, bahamut.AuthActionKO, func(cert *x509.Certificate) bool {
				return true
			})

			action, err := auth.AuthenticateRequest(ctx)

			Convey("Then err should be nil", func() {
				So(err, ShouldBeNil)
			})

			Convey("Then action should be bahamut.AuthActionOK", func() {
				So(action, ShouldEqual, bahamut.AuthActionOK)
			})

			Convey("Then claims should be correctly populated", func() {
				So(ctx.GetClaims(), ShouldResemble, []string{"@auth:realm=certificate", "@auth:mode=internal", "@auth:serialnumber=23486181163925715704694891313232533542", "@auth:commonname=user-a"})
			})
		})

		Convey("When I try check auth for user-a using chain-a with a verifier func that is not ok", func() {

			ctx := &bahamut.Context{
				Request: &elemental.Request{
					TLSConnectionState: &tls.ConnectionState{
						PeerCertificates: []*x509.Certificate{
							userCertA,
						},
					},
				},
			}

			opts := x509.VerifyOptions{
				Roots:     certPoolA,
				KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
			}

			auth := NewMTLSRequestAuthenticator(opts, bahamut.AuthActionOK, bahamut.AuthActionKO, func(cert *x509.Certificate) bool {
				return false
			})

			action, err := auth.AuthenticateRequest(ctx)

			Convey("Then err should be nil", func() {
				So(err, ShouldBeNil)
			})

			Convey("Then action should be bahamut.AuthActionOK", func() {
				So(action, ShouldEqual, bahamut.AuthActionKO)
			})

			Convey("Then claims should be correctly populated", func() {
				So(ctx.GetClaims(), ShouldBeNil)
			})
		})

		Convey("When I try check auth for server-a using chain-a", func() {

			ctx := &bahamut.Context{
				Request: &elemental.Request{
					TLSConnectionState: &tls.ConnectionState{
						PeerCertificates: []*x509.Certificate{
							serverCertA,
						},
					},
				},
			}

			opts := x509.VerifyOptions{
				Roots:     certPoolA,
				KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
			}

			auth := NewMTLSRequestAuthenticator(opts, bahamut.AuthActionOK, bahamut.AuthActionKO, nil)

			action, err := auth.AuthenticateRequest(ctx)

			Convey("Then err should be nil", func() {
				So(err, ShouldBeNil)
			})

			Convey("Then action should be bahamut.AuthActionKO", func() {
				So(action, ShouldEqual, bahamut.AuthActionKO)
			})
		})

		Convey("When I try check auth for server-a using chain-a with a verifier func that is ok", func() {

			ctx := &bahamut.Context{
				Request: &elemental.Request{
					TLSConnectionState: &tls.ConnectionState{
						PeerCertificates: []*x509.Certificate{
							serverCertA,
						},
					},
				},
			}

			opts := x509.VerifyOptions{
				Roots:     certPoolA,
				KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
			}

			auth := NewMTLSRequestAuthenticator(opts, bahamut.AuthActionOK, bahamut.AuthActionKO, func(cert *x509.Certificate) bool {
				return true
			})

			action, err := auth.AuthenticateRequest(ctx)

			Convey("Then err should be nil", func() {
				So(err, ShouldBeNil)
			})

			Convey("Then action should be bahamut.AuthActionKO", func() {
				So(action, ShouldEqual, bahamut.AuthActionKO)
			})
		})

		Convey("When I try check auth for user-ext using chain-a", func() {

			ctx := &bahamut.Context{
				Request: &elemental.Request{
					TLSConnectionState: &tls.ConnectionState{
						PeerCertificates: []*x509.Certificate{
							userCertExt,
						},
					},
				},
			}

			opts := x509.VerifyOptions{
				Roots:     certPoolA,
				KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
			}

			auth := NewMTLSRequestAuthenticator(opts, bahamut.AuthActionOK, bahamut.AuthActionKO, nil)

			action, err := auth.AuthenticateRequest(ctx)

			Convey("Then err should be nil", func() {
				So(err, ShouldBeNil)
			})

			Convey("Then action should be bahamut.AuthActionKO", func() {
				So(action, ShouldEqual, bahamut.AuthActionKO)
			})
		})

		Convey("When I try check auth for user-b using chain-a", func() {

			ctx := &bahamut.Context{
				Request: &elemental.Request{
					TLSConnectionState: &tls.ConnectionState{
						PeerCertificates: []*x509.Certificate{
							userCertB,
						},
					},
				},
			}

			opts := x509.VerifyOptions{
				Roots:     certPoolA,
				KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
			}

			auth := NewMTLSRequestAuthenticator(opts, bahamut.AuthActionOK, bahamut.AuthActionKO, nil)

			action, err := auth.AuthenticateRequest(ctx)

			Convey("Then err should be nil", func() {
				So(err, ShouldBeNil)
			})

			Convey("Then action should be bahamut.AuthActionKO", func() {
				So(action, ShouldEqual, bahamut.AuthActionKO)
			})
		})
	})
}
