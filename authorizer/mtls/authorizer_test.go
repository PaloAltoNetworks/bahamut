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
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"io/ioutil"
	"net/http"
	"reflect"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"go.aporeto.io/bahamut"
	"go.aporeto.io/elemental"
)

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

			ctx := bahamut.NewContext(context.TODO(), elemental.NewRequest())

			opts := x509.VerifyOptions{
				Roots:     certPoolA,
				KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
			}

			auth := NewMTLSAuthorizer(opts, func(a bahamut.AuthAction, c bahamut.Context, s bahamut.Session) bahamut.AuthAction { return a }, nil, nil, CertificateCheckModeTLSStateOnly)

			action, err := auth.IsAuthorized(ctx)

			Convey("Then err should be nil", func() {
				So(err, ShouldBeNil)
			})

			Convey("Then action should be bahamut.AuthActionContinue", func() {
				So(action, ShouldEqual, bahamut.AuthActionContinue)
			})
		})

		Convey("When I try check auth for user-a using chain-a", func() {

			ctx := bahamut.NewContext(context.TODO(), &elemental.Request{
				TLSConnectionState: &tls.ConnectionState{
					PeerCertificates: []*x509.Certificate{
						userCertA,
					},
				},
			})

			opts := x509.VerifyOptions{
				Roots:     certPoolA,
				KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
			}

			auth := NewMTLSAuthorizer(opts, func(a bahamut.AuthAction, c bahamut.Context, s bahamut.Session) bahamut.AuthAction { return a }, nil, nil, CertificateCheckModeTLSStateOnly)

			action, err := auth.IsAuthorized(ctx)

			Convey("Then err should be nil", func() {
				So(err, ShouldBeNil)
			})

			Convey("Then action should be bahamut.AuthActionOK", func() {
				So(action, ShouldEqual, bahamut.AuthActionOK)
			})
		})

		Convey("When I try check auth for user-a using chain-a but only checking for header", func() {

			ctx := bahamut.NewContext(context.TODO(), &elemental.Request{
				TLSConnectionState: &tls.ConnectionState{
					PeerCertificates: []*x509.Certificate{
						userCertA,
					},
				},
			})

			opts := x509.VerifyOptions{
				Roots:     certPoolA,
				KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
			}

			auth := NewMTLSAuthorizer(opts, func(a bahamut.AuthAction, c bahamut.Context, s bahamut.Session) bahamut.AuthAction { return a }, nil, nil, CertificateCheckModeHeaderOnly)

			action, err := auth.IsAuthorized(ctx)

			Convey("Then err should be nil", func() {
				So(err, ShouldBeNil)
			})

			Convey("Then action should be bahamut.AuthActionKO", func() {
				So(action, ShouldEqual, bahamut.AuthActionKO)
			})
		})

		Convey("When I try check auth for user-a using chain-a as valid inline header", func() {

			header := http.Header{}
			header.Set(tlsHeaderKey, string(userCertAData))
			ctx := bahamut.NewContext(context.TODO(), &elemental.Request{
				Headers: header,
			})
			opts := x509.VerifyOptions{
				Roots:     certPoolA,
				KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
			}

			auth := NewMTLSAuthorizer(opts, func(a bahamut.AuthAction, c bahamut.Context, s bahamut.Session) bahamut.AuthAction { return a }, nil, nil, CertificateCheckModeHeaderThenTLSState)

			action, err := auth.IsAuthorized(ctx)

			Convey("Then err should be nil", func() {
				So(err, ShouldBeNil)
			})

			Convey("Then action should be bahamut.AuthActionOK", func() {
				So(action, ShouldEqual, bahamut.AuthActionOK)
			})
		})

		Convey("When I try check auth for user-a using chain-a as valid inline header while forbidding checking in header", func() {

			header := http.Header{}
			header.Set(tlsHeaderKey, string(userCertAData))

			ctx := bahamut.NewContext(context.TODO(), &elemental.Request{
				Headers: header,
			})

			opts := x509.VerifyOptions{
				Roots:     certPoolA,
				KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
			}

			auth := NewMTLSAuthorizer(opts, func(a bahamut.AuthAction, c bahamut.Context, s bahamut.Session) bahamut.AuthAction { return a }, nil, nil, CertificateCheckModeTLSStateOnly)

			action, err := auth.IsAuthorized(ctx)

			Convey("Then err should be nil", func() {
				So(err, ShouldBeNil)
			})

			Convey("Then action should be bahamut.AuthActionKO", func() {
				So(action, ShouldEqual, bahamut.AuthActionKO)
			})
		})

		Convey("When I try check auth with valid inline tls header for user-a but tls state presenting user-b while preferring header", func() {

			header := http.Header{}
			header.Set(tlsHeaderKey, string(userCertAData))
			ctx := bahamut.NewContext(context.TODO(), &elemental.Request{
				Headers: header,
				TLSConnectionState: &tls.ConnectionState{
					PeerCertificates: []*x509.Certificate{
						userCertB,
					},
				},
			})

			opts := x509.VerifyOptions{
				Roots:     certPoolA,
				KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
			}

			auth := NewMTLSAuthorizer(opts, func(a bahamut.AuthAction, c bahamut.Context, s bahamut.Session) bahamut.AuthAction { return a }, nil, nil, CertificateCheckModeHeaderThenTLSState)

			action, err := auth.IsAuthorized(ctx)

			Convey("Then err should be nil", func() {
				So(err, ShouldBeNil)
			})

			Convey("Then action should be bahamut.AuthActionOK", func() {
				So(action, ShouldEqual, bahamut.AuthActionOK)
			})
		})

		Convey("When I try check auth with valid inline tls header for user-b but tls state presenting user-a while preferring header", func() {

			header := http.Header{}
			header.Set(tlsHeaderKey, string(userCertBData))
			ctx := bahamut.NewContext(context.TODO(), &elemental.Request{
				Headers: header,
				TLSConnectionState: &tls.ConnectionState{
					PeerCertificates: []*x509.Certificate{
						userCertA,
					},
				},
			})
			opts := x509.VerifyOptions{
				Roots:     certPoolA,
				KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
			}

			auth := NewMTLSAuthorizer(opts, func(a bahamut.AuthAction, c bahamut.Context, s bahamut.Session) bahamut.AuthAction { return a }, nil, nil, CertificateCheckModeHeaderThenTLSState)

			action, err := auth.IsAuthorized(ctx)

			Convey("Then err should be nil", func() {
				So(err, ShouldBeNil)
			})

			Convey("Then action should be bahamut.AuthActionKO", func() {
				So(action, ShouldEqual, bahamut.AuthActionKO)
			})
		})

		Convey("When I try check auth with valid inline tls header for user-a but tls state presenting user-b while preferring state", func() {

			header := http.Header{}
			header.Set(tlsHeaderKey, string(userCertAData))
			ctx := bahamut.NewContext(context.TODO(), &elemental.Request{
				Headers: header,
				TLSConnectionState: &tls.ConnectionState{
					PeerCertificates: []*x509.Certificate{
						userCertB,
					},
				},
			})
			opts := x509.VerifyOptions{
				Roots:     certPoolA,
				KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
			}

			auth := NewMTLSAuthorizer(opts, func(a bahamut.AuthAction, c bahamut.Context, s bahamut.Session) bahamut.AuthAction { return a }, nil, nil, CertificateCheckModeTLSStateThenHeader)

			action, err := auth.IsAuthorized(ctx)

			Convey("Then err should be nil", func() {
				So(err, ShouldBeNil)
			})

			Convey("Then action should be bahamut.AuthActionKO", func() {
				So(action, ShouldEqual, bahamut.AuthActionKO)
			})
		})

		Convey("When I try check auth with valid inline tls header for user-b but tls state presenting user-a while preferring tls state", func() {

			header := http.Header{}
			header.Set(tlsHeaderKey, string(userCertBData))
			ctx := bahamut.NewContext(context.TODO(), &elemental.Request{
				Headers: header,
				TLSConnectionState: &tls.ConnectionState{
					PeerCertificates: []*x509.Certificate{
						userCertA,
					},
				},
			},
			)
			opts := x509.VerifyOptions{
				Roots:     certPoolA,
				KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
			}

			auth := NewMTLSAuthorizer(opts, func(a bahamut.AuthAction, c bahamut.Context, s bahamut.Session) bahamut.AuthAction { return a }, nil, nil, CertificateCheckModeTLSStateOnly)

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
			ctx := bahamut.NewContext(context.TODO(), &elemental.Request{
				Headers: header,
			})
			opts := x509.VerifyOptions{
				Roots:     certPoolA,
				KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
			}

			auth := NewMTLSAuthorizer(opts, func(a bahamut.AuthAction, c bahamut.Context, s bahamut.Session) bahamut.AuthAction { return a }, nil, nil, CertificateCheckModeTLSStateOnly)

			action, err := auth.IsAuthorized(ctx)

			Convey("Then err should be nil", func() {
				So(err, ShouldBeNil)
			})

			Convey("Then action should be bahamut.AuthActionOK", func() {
				So(action, ShouldEqual, bahamut.AuthActionKO)
			})
		})

		Convey("When I try check auth for user-a using chain-a with a verifier function that is ok", func() {

			ctx := bahamut.NewContext(context.TODO(), &elemental.Request{
				TLSConnectionState: &tls.ConnectionState{
					PeerCertificates: []*x509.Certificate{
						userCertA,
					},
				},
			})

			opts := x509.VerifyOptions{
				Roots:     certPoolA,
				KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
			}

			auth := NewMTLSAuthorizer(opts, func(a bahamut.AuthAction, c bahamut.Context, s bahamut.Session) bahamut.AuthAction { return a }, nil, func(cert *x509.Certificate) bool {
				return true
			}, CertificateCheckModeTLSStateOnly)

			action, err := auth.IsAuthorized(ctx)

			Convey("Then err should be nil", func() {
				So(err, ShouldBeNil)
			})

			Convey("Then action should be bahamut.AuthActionOK", func() {
				So(action, ShouldEqual, bahamut.AuthActionOK)
			})
		})

		Convey("When I try check auth for user-a using chain-a with a verifier function that is not ok", func() {

			ctx := bahamut.NewContext(context.TODO(), &elemental.Request{
				TLSConnectionState: &tls.ConnectionState{
					PeerCertificates: []*x509.Certificate{
						userCertA,
					},
				},
			})
			opts := x509.VerifyOptions{
				Roots:     certPoolA,
				KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
			}

			auth := NewMTLSAuthorizer(opts, func(a bahamut.AuthAction, c bahamut.Context, s bahamut.Session) bahamut.AuthAction { return a }, nil, func(cert *x509.Certificate) bool {
				return false
			}, CertificateCheckModeTLSStateOnly)

			action, err := auth.IsAuthorized(ctx)

			Convey("Then err should be nil", func() {
				So(err, ShouldBeNil)
			})

			Convey("Then action should be bahamut.AuthActionKO", func() {
				So(action, ShouldEqual, bahamut.AuthActionKO)
			})
		})

		Convey("When I try check auth for server-a using chain-a", func() {

			ctx := bahamut.NewContext(context.TODO(), &elemental.Request{
				TLSConnectionState: &tls.ConnectionState{
					PeerCertificates: []*x509.Certificate{
						serverCertA,
					},
				},
			})
			opts := x509.VerifyOptions{
				Roots:     certPoolA,
				KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
			}

			auth := NewMTLSAuthorizer(opts, func(a bahamut.AuthAction, c bahamut.Context, s bahamut.Session) bahamut.AuthAction { return a }, nil, nil, CertificateCheckModeTLSStateOnly)

			action, err := auth.IsAuthorized(ctx)

			Convey("Then err should be nil", func() {
				So(err, ShouldBeNil)
			})

			Convey("Then action should be bahamut.AuthActionKO", func() {
				So(action, ShouldEqual, bahamut.AuthActionKO)
			})
		})

		Convey("When I try check auth for user-ext using chain-a", func() {

			ctx := bahamut.NewContext(context.TODO(), &elemental.Request{
				TLSConnectionState: &tls.ConnectionState{
					PeerCertificates: []*x509.Certificate{
						userCertExt,
					},
				},
			})
			opts := x509.VerifyOptions{
				Roots:     certPoolA,
				KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
			}

			auth := NewMTLSAuthorizer(opts, func(a bahamut.AuthAction, c bahamut.Context, s bahamut.Session) bahamut.AuthAction { return a }, nil, nil, CertificateCheckModeTLSStateOnly)

			action, err := auth.IsAuthorized(ctx)

			Convey("Then err should be nil", func() {
				So(err, ShouldBeNil)
			})

			Convey("Then action should be bahamut.AuthActionKO", func() {
				So(action, ShouldEqual, bahamut.AuthActionKO)
			})
		})

		Convey("When I try check auth for user-ext using chain-a with a verifier func that is ok", func() {

			ctx := bahamut.NewContext(context.TODO(), &elemental.Request{
				TLSConnectionState: &tls.ConnectionState{
					PeerCertificates: []*x509.Certificate{
						userCertExt,
					},
				},
			})
			opts := x509.VerifyOptions{
				Roots:     certPoolA,
				KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
			}

			auth := NewMTLSAuthorizer(opts, func(a bahamut.AuthAction, c bahamut.Context, s bahamut.Session) bahamut.AuthAction { return a }, nil, func(cert *x509.Certificate) bool {
				return true
			}, CertificateCheckModeTLSStateOnly)

			action, err := auth.IsAuthorized(ctx)

			Convey("Then err should be nil", func() {
				So(err, ShouldBeNil)
			})

			Convey("Then action should be bahamut.AuthActionKO", func() {
				So(action, ShouldEqual, bahamut.AuthActionKO)
			})
		})

		Convey("When I try check auth for user-b using chain-a", func() {

			ctx := bahamut.NewContext(context.TODO(), &elemental.Request{
				TLSConnectionState: &tls.ConnectionState{
					PeerCertificates: []*x509.Certificate{
						userCertB,
					},
				},
			})
			opts := x509.VerifyOptions{
				Roots:     certPoolA,
				KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
			}

			auth := NewMTLSAuthorizer(opts, func(a bahamut.AuthAction, c bahamut.Context, s bahamut.Session) bahamut.AuthAction { return a }, nil, nil, CertificateCheckModeTLSStateOnly)

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
			ctx := bahamut.NewContext(context.TODO(), &elemental.Request{
				Identity: identity,
				TLSConnectionState: &tls.ConnectionState{
					PeerCertificates: []*x509.Certificate{
						serverCertA,
					},
				},
			})
			opts := x509.VerifyOptions{
				Roots:     certPoolA,
				KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
			}

			auth := NewMTLSAuthorizer(opts, func(a bahamut.AuthAction, c bahamut.Context, s bahamut.Session) bahamut.AuthAction { return a }, []elemental.Identity{identity}, nil, CertificateCheckModeTLSStateOnly)

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

			ctx := bahamut.NewContext(context.TODO(), &elemental.Request{})

			opts := x509.VerifyOptions{
				Roots:     certPoolA,
				KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
			}

			auth := NewMTLSRequestAuthenticator(opts, func(a bahamut.AuthAction, c bahamut.Context, s bahamut.Session) bahamut.AuthAction { return a }, nil, CertificateCheckModeTLSStateOnly)

			action, err := auth.AuthenticateRequest(ctx)

			Convey("Then err should be nil", func() {
				So(err, ShouldBeNil)
			})

			Convey("Then action should be bahamut.AuthActionContinue", func() {
				So(action, ShouldEqual, bahamut.AuthActionContinue)
			})
		})

		Convey("When I try check auth for user-a using chain-a", func() {

			ctx := bahamut.NewContext(context.TODO(), &elemental.Request{
				TLSConnectionState: &tls.ConnectionState{
					PeerCertificates: []*x509.Certificate{
						userCertA,
					},
				},
			})

			opts := x509.VerifyOptions{
				Roots:     certPoolA,
				KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
			}

			auth := NewMTLSRequestAuthenticator(opts, func(a bahamut.AuthAction, c bahamut.Context, s bahamut.Session) bahamut.AuthAction { return a }, nil, CertificateCheckModeTLSStateOnly)

			action, err := auth.AuthenticateRequest(ctx)

			Convey("Then err should be nil", func() {
				So(err, ShouldBeNil)
			})

			Convey("Then action should be bahamut.AuthActionOK", func() {
				So(action, ShouldEqual, bahamut.AuthActionOK)
			})

			Convey("Then claims should be correctly populated", func() {
				So(ctx.Claims(), ShouldResemble, []string{"@auth:realm=certificate", "@auth:mode=internal", "@auth:serialnumber=23486181163925715704694891313232533542", "@auth:commonname=user-a"})
			})
		})

		Convey("When I try check auth for user-a using chain-a but only checking header", func() {

			ctx := bahamut.NewContext(context.TODO(), &elemental.Request{
				TLSConnectionState: &tls.ConnectionState{
					PeerCertificates: []*x509.Certificate{
						userCertA,
					},
				},
			})

			opts := x509.VerifyOptions{
				Roots:     certPoolA,
				KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
			}

			auth := NewMTLSRequestAuthenticator(opts, func(a bahamut.AuthAction, c bahamut.Context, s bahamut.Session) bahamut.AuthAction { return a }, nil, CertificateCheckModeHeaderOnly)

			action, err := auth.AuthenticateRequest(ctx)

			Convey("Then err should be nil", func() {
				So(err, ShouldBeNil)
			})

			Convey("Then action should be bahamut.AuthActionKO", func() {
				So(action, ShouldEqual, bahamut.AuthActionKO)
			})
		})

		Convey("When I try check auth with valid inline tls header for user-a but tls state presenting user-b while preferring header", func() {

			header := http.Header{}
			header.Set(tlsHeaderKey, string(userCertAData))
			ctx := bahamut.NewContext(context.TODO(), &elemental.Request{
				Headers: header,
				TLSConnectionState: &tls.ConnectionState{
					PeerCertificates: []*x509.Certificate{
						userCertB,
					},
				},
			})
			opts := x509.VerifyOptions{
				Roots:     certPoolA,
				KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
			}

			auth := NewMTLSRequestAuthenticator(opts, func(a bahamut.AuthAction, c bahamut.Context, s bahamut.Session) bahamut.AuthAction { return a }, nil, CertificateCheckModeHeaderThenTLSState)

			action, err := auth.AuthenticateRequest(ctx)

			Convey("Then err should be nil", func() {
				So(err, ShouldBeNil)
			})

			Convey("Then action should be bahamut.AuthActionOK", func() {
				So(action, ShouldEqual, bahamut.AuthActionOK)
			})
		})

		Convey("When I try check auth with valid inline tls header for user-b but tls state presenting user-a while preferring header", func() {

			header := http.Header{}
			header.Set(tlsHeaderKey, string(userCertBData))
			ctx := bahamut.NewContext(context.TODO(), &elemental.Request{
				Headers: header,
				TLSConnectionState: &tls.ConnectionState{
					PeerCertificates: []*x509.Certificate{
						userCertA,
					},
				},
			})
			opts := x509.VerifyOptions{
				Roots:     certPoolA,
				KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
			}

			auth := NewMTLSRequestAuthenticator(opts, func(a bahamut.AuthAction, c bahamut.Context, s bahamut.Session) bahamut.AuthAction { return a }, nil, CertificateCheckModeHeaderThenTLSState)

			action, err := auth.AuthenticateRequest(ctx)

			Convey("Then err should be nil", func() {
				So(err, ShouldBeNil)
			})

			Convey("Then action should be bahamut.AuthActionKO", func() {
				So(action, ShouldEqual, bahamut.AuthActionKO)
			})
		})

		Convey("When I try check auth with valid inline tls header for user-a but tls state presenting user-b while preferring state", func() {

			header := http.Header{}
			header.Set(tlsHeaderKey, string(userCertAData))
			ctx := bahamut.NewContext(context.TODO(), &elemental.Request{
				Headers: header,
				TLSConnectionState: &tls.ConnectionState{
					PeerCertificates: []*x509.Certificate{
						userCertB,
					},
				},
			})
			opts := x509.VerifyOptions{
				Roots:     certPoolA,
				KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
			}

			auth := NewMTLSRequestAuthenticator(opts, func(a bahamut.AuthAction, c bahamut.Context, s bahamut.Session) bahamut.AuthAction { return a }, nil, CertificateCheckModeTLSStateThenHeader)

			action, err := auth.AuthenticateRequest(ctx)

			Convey("Then err should be nil", func() {
				So(err, ShouldBeNil)
			})

			Convey("Then action should be bahamut.AuthActionKO", func() {
				So(action, ShouldEqual, bahamut.AuthActionKO)
			})
		})

		Convey("When I try check auth with valid inline tls header for user-b but tls state presenting user-a while preferring tls state", func() {

			header := http.Header{}
			header.Set(tlsHeaderKey, string(userCertBData))
			ctx := bahamut.NewContext(context.TODO(), &elemental.Request{
				Headers: header,
				TLSConnectionState: &tls.ConnectionState{
					PeerCertificates: []*x509.Certificate{
						userCertA,
					},
				},
			})
			opts := x509.VerifyOptions{
				Roots:     certPoolA,
				KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
			}

			auth := NewMTLSRequestAuthenticator(opts, func(a bahamut.AuthAction, c bahamut.Context, s bahamut.Session) bahamut.AuthAction { return a }, nil, CertificateCheckModeTLSStateOnly)

			action, err := auth.AuthenticateRequest(ctx)

			Convey("Then err should be nil", func() {
				So(err, ShouldBeNil)
			})

			Convey("Then action should be bahamut.AuthActionOK", func() {
				So(action, ShouldEqual, bahamut.AuthActionOK)
			})
		})

		Convey("When I try check auth for user-a using chain-a with a verifier func that is ok", func() {

			ctx := bahamut.NewContext(context.TODO(), &elemental.Request{
				TLSConnectionState: &tls.ConnectionState{
					PeerCertificates: []*x509.Certificate{
						userCertA,
					},
				},
			})

			opts := x509.VerifyOptions{
				Roots:     certPoolA,
				KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
			}

			auth := NewMTLSRequestAuthenticator(opts, func(a bahamut.AuthAction, c bahamut.Context, s bahamut.Session) bahamut.AuthAction { return a }, func(cert *x509.Certificate) bool {
				return true
			}, CertificateCheckModeTLSStateOnly)

			action, err := auth.AuthenticateRequest(ctx)

			Convey("Then err should be nil", func() {
				So(err, ShouldBeNil)
			})

			Convey("Then action should be bahamut.AuthActionOK", func() {
				So(action, ShouldEqual, bahamut.AuthActionOK)
			})

			Convey("Then claims should be correctly populated", func() {
				So(ctx.Claims(), ShouldResemble, []string{"@auth:realm=certificate", "@auth:mode=internal", "@auth:serialnumber=23486181163925715704694891313232533542", "@auth:commonname=user-a"})
			})
		})

		Convey("When I try check auth for user-a using chain-a with a verifier func that is not ok", func() {

			ctx := bahamut.NewContext(context.TODO(), &elemental.Request{
				TLSConnectionState: &tls.ConnectionState{
					PeerCertificates: []*x509.Certificate{
						userCertA,
					},
				},
			})

			opts := x509.VerifyOptions{
				Roots:     certPoolA,
				KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
			}

			auth := NewMTLSRequestAuthenticator(opts, func(a bahamut.AuthAction, c bahamut.Context, s bahamut.Session) bahamut.AuthAction { return a }, func(cert *x509.Certificate) bool {
				return false
			}, CertificateCheckModeTLSStateOnly)

			action, err := auth.AuthenticateRequest(ctx)

			Convey("Then err should be nil", func() {
				So(err, ShouldBeNil)
			})

			Convey("Then action should be bahamut.AuthActionOK", func() {
				So(action, ShouldEqual, bahamut.AuthActionKO)
			})

			Convey("Then claims should be correctly populated", func() {
				So(ctx.Claims(), ShouldBeNil)
			})
		})

		Convey("When I try check auth for server-a using chain-a", func() {

			ctx := bahamut.NewContext(context.TODO(), &elemental.Request{
				TLSConnectionState: &tls.ConnectionState{
					PeerCertificates: []*x509.Certificate{
						serverCertA,
					},
				},
			})

			opts := x509.VerifyOptions{
				Roots:     certPoolA,
				KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
			}

			auth := NewMTLSRequestAuthenticator(opts, func(a bahamut.AuthAction, c bahamut.Context, s bahamut.Session) bahamut.AuthAction { return a }, nil, CertificateCheckModeTLSStateOnly)

			action, err := auth.AuthenticateRequest(ctx)

			Convey("Then err should be nil", func() {
				So(err, ShouldBeNil)
			})

			Convey("Then action should be bahamut.AuthActionKO", func() {
				So(action, ShouldEqual, bahamut.AuthActionKO)
			})
		})

		Convey("When I try check auth for server-a using chain-a with a verifier func that is ok", func() {

			ctx := bahamut.NewContext(context.TODO(), &elemental.Request{
				TLSConnectionState: &tls.ConnectionState{
					PeerCertificates: []*x509.Certificate{
						serverCertA,
					},
				},
			})
			opts := x509.VerifyOptions{
				Roots:     certPoolA,
				KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
			}

			auth := NewMTLSRequestAuthenticator(opts, func(a bahamut.AuthAction, c bahamut.Context, s bahamut.Session) bahamut.AuthAction { return a }, func(cert *x509.Certificate) bool {
				return true
			}, CertificateCheckModeTLSStateOnly)

			action, err := auth.AuthenticateRequest(ctx)

			Convey("Then err should be nil", func() {
				So(err, ShouldBeNil)
			})

			Convey("Then action should be bahamut.AuthActionKO", func() {
				So(action, ShouldEqual, bahamut.AuthActionKO)
			})
		})

		Convey("When I try check auth for user-ext using chain-a", func() {

			ctx := bahamut.NewContext(context.TODO(), &elemental.Request{
				TLSConnectionState: &tls.ConnectionState{
					PeerCertificates: []*x509.Certificate{
						userCertExt,
					},
				},
			})

			opts := x509.VerifyOptions{
				Roots:     certPoolA,
				KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
			}

			auth := NewMTLSRequestAuthenticator(opts, func(a bahamut.AuthAction, c bahamut.Context, s bahamut.Session) bahamut.AuthAction { return a }, nil, CertificateCheckModeTLSStateOnly)

			action, err := auth.AuthenticateRequest(ctx)

			Convey("Then err should be nil", func() {
				So(err, ShouldBeNil)
			})

			Convey("Then action should be bahamut.AuthActionKO", func() {
				So(action, ShouldEqual, bahamut.AuthActionKO)
			})
		})

		Convey("When I try check auth for user-b using chain-a", func() {

			ctx := bahamut.NewContext(context.TODO(), &elemental.Request{
				TLSConnectionState: &tls.ConnectionState{
					PeerCertificates: []*x509.Certificate{
						userCertB,
					},
				},
			})

			opts := x509.VerifyOptions{
				Roots:     certPoolA,
				KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
			}

			auth := NewMTLSRequestAuthenticator(opts, func(a bahamut.AuthAction, c bahamut.Context, s bahamut.Session) bahamut.AuthAction { return a }, nil, CertificateCheckModeTLSStateOnly)

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

type mockSession struct {
	state  *tls.ConnectionState
	claims []string
}

func (s *mockSession) Identifier() string                       { return "" }
func (s *mockSession) Parameter(string) string                  { return "" }
func (s *mockSession) Header(string) string                     { return "" }
func (s *mockSession) SetClaims(c []string)                     { s.claims = c }
func (s *mockSession) Claims() []string                         { return s.claims }
func (s *mockSession) ClaimsMap() map[string]string             { return nil }
func (s *mockSession) Token() string                            { return "" }
func (s *mockSession) TLSConnectionState() *tls.ConnectionState { return s.state }
func (s *mockSession) Metadata() interface{}                    { return nil }
func (s *mockSession) SetMetadata(interface{})                  {}
func (s *mockSession) Context() context.Context                 { return context.Background() }
func (s *mockSession) ClientIP() string                         { return "" }

func TestBahamut_NewMTLSSessionAuthenticator(t *testing.T) {

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

			s := &mockSession{}

			opts := x509.VerifyOptions{
				Roots:     certPoolA,
				KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
			}

			auth := NewMTLSSessionAuthenticator(opts, func(a bahamut.AuthAction, c bahamut.Context, s bahamut.Session) bahamut.AuthAction { return a }, nil, CertificateCheckModeTLSStateOnly)

			action, err := auth.AuthenticateSession(s)

			Convey("Then err should be nil", func() {
				So(err, ShouldBeNil)
			})

			Convey("Then action should be bahamut.AuthActionContinue", func() {
				So(action, ShouldEqual, bahamut.AuthActionContinue)
			})
		})

		Convey("When I try check auth for user-a using chain-a", func() {

			s := &mockSession{
				state: &tls.ConnectionState{
					PeerCertificates: []*x509.Certificate{
						userCertA,
					},
				},
			}

			opts := x509.VerifyOptions{
				Roots:     certPoolA,
				KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
			}

			auth := NewMTLSSessionAuthenticator(opts, func(a bahamut.AuthAction, c bahamut.Context, s bahamut.Session) bahamut.AuthAction { return a }, nil, CertificateCheckModeTLSStateOnly)

			action, err := auth.AuthenticateSession(s)

			Convey("Then err should be nil", func() {
				So(err, ShouldBeNil)
			})

			Convey("Then action should be bahamut.AuthActionOK", func() {
				So(action, ShouldEqual, bahamut.AuthActionOK)
			})

			Convey("Then claims should be correctly populated", func() {
				So(s.Claims(), ShouldResemble, []string{"@auth:realm=certificate", "@auth:mode=internal", "@auth:serialnumber=23486181163925715704694891313232533542", "@auth:commonname=user-a"})
			})
		})

		Convey("When I try check auth for user-a using chain-a but only checking header", func() {

			s := &mockSession{
				state: &tls.ConnectionState{
					PeerCertificates: []*x509.Certificate{
						userCertA,
					},
				},
			}

			opts := x509.VerifyOptions{
				Roots:     certPoolA,
				KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
			}

			auth := NewMTLSSessionAuthenticator(opts, func(a bahamut.AuthAction, c bahamut.Context, s bahamut.Session) bahamut.AuthAction { return a }, nil, CertificateCheckModeHeaderOnly)

			action, err := auth.AuthenticateSession(s)

			Convey("Then err should be nil", func() {
				So(err, ShouldBeNil)
			})

			Convey("Then action should be bahamut.AuthActionKO", func() {
				So(action, ShouldEqual, bahamut.AuthActionKO)
			})
		})

		Convey("When I try check auth for user-a using chain-a with a verifier func that is ok", func() {

			s := &mockSession{
				state: &tls.ConnectionState{
					PeerCertificates: []*x509.Certificate{
						userCertA,
					},
				},
			}

			opts := x509.VerifyOptions{
				Roots:     certPoolA,
				KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
			}

			auth := NewMTLSSessionAuthenticator(opts, func(a bahamut.AuthAction, c bahamut.Context, s bahamut.Session) bahamut.AuthAction { return a }, func(cert *x509.Certificate) bool {
				return true
			}, CertificateCheckModeTLSStateOnly)

			action, err := auth.AuthenticateSession(s)

			Convey("Then err should be nil", func() {
				So(err, ShouldBeNil)
			})

			Convey("Then action should be bahamut.AuthActionOK", func() {
				So(action, ShouldEqual, bahamut.AuthActionOK)
			})

			Convey("Then claims should be correctly populated", func() {
				So(s.Claims(), ShouldResemble, []string{"@auth:realm=certificate", "@auth:mode=internal", "@auth:serialnumber=23486181163925715704694891313232533542", "@auth:commonname=user-a"})
			})
		})

		Convey("When I try check auth for user-a using chain-a with a verifier func that is not ok", func() {

			s := &mockSession{
				state: &tls.ConnectionState{
					PeerCertificates: []*x509.Certificate{
						userCertA,
					},
				},
			}

			opts := x509.VerifyOptions{
				Roots:     certPoolA,
				KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
			}

			auth := NewMTLSSessionAuthenticator(opts, func(a bahamut.AuthAction, c bahamut.Context, s bahamut.Session) bahamut.AuthAction { return a }, func(cert *x509.Certificate) bool {
				return false
			}, CertificateCheckModeTLSStateOnly)

			action, err := auth.AuthenticateSession(s)

			Convey("Then err should be nil", func() {
				So(err, ShouldBeNil)
			})

			Convey("Then action should be bahamut.AuthActionOK", func() {
				So(action, ShouldEqual, bahamut.AuthActionKO)
			})

			Convey("Then claims should be correctly populated", func() {
				So(s.Claims(), ShouldBeNil)
			})
		})

		Convey("When I try check auth for server-a using chain-a", func() {

			s := &mockSession{
				state: &tls.ConnectionState{
					PeerCertificates: []*x509.Certificate{
						serverCertA,
					},
				},
			}

			opts := x509.VerifyOptions{
				Roots:     certPoolA,
				KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
			}

			auth := NewMTLSSessionAuthenticator(opts, func(a bahamut.AuthAction, c bahamut.Context, s bahamut.Session) bahamut.AuthAction { return a }, nil, CertificateCheckModeTLSStateOnly)

			action, err := auth.AuthenticateSession(s)

			Convey("Then err should be nil", func() {
				So(err, ShouldBeNil)
			})

			Convey("Then action should be bahamut.AuthActionKO", func() {
				So(action, ShouldEqual, bahamut.AuthActionKO)
			})
		})

		Convey("When I try check auth for server-a using chain-a with a verifier func that is ok", func() {

			s := &mockSession{
				state: &tls.ConnectionState{
					PeerCertificates: []*x509.Certificate{
						serverCertA,
					},
				},
			}

			opts := x509.VerifyOptions{
				Roots:     certPoolA,
				KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
			}

			auth := NewMTLSSessionAuthenticator(opts, func(a bahamut.AuthAction, c bahamut.Context, s bahamut.Session) bahamut.AuthAction { return a }, func(cert *x509.Certificate) bool {
				return true
			}, CertificateCheckModeTLSStateOnly)

			action, err := auth.AuthenticateSession(s)

			Convey("Then err should be nil", func() {
				So(err, ShouldBeNil)
			})

			Convey("Then action should be bahamut.AuthActionKO", func() {
				So(action, ShouldEqual, bahamut.AuthActionKO)
			})
		})

		Convey("When I try check auth for user-ext using chain-a", func() {

			s := &mockSession{
				state: &tls.ConnectionState{
					PeerCertificates: []*x509.Certificate{
						userCertExt,
					},
				},
			}

			opts := x509.VerifyOptions{
				Roots:     certPoolA,
				KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
			}

			auth := NewMTLSSessionAuthenticator(opts, func(a bahamut.AuthAction, c bahamut.Context, s bahamut.Session) bahamut.AuthAction { return a }, nil, CertificateCheckModeTLSStateOnly)

			action, err := auth.AuthenticateSession(s)

			Convey("Then err should be nil", func() {
				So(err, ShouldBeNil)
			})

			Convey("Then action should be bahamut.AuthActionKO", func() {
				So(action, ShouldEqual, bahamut.AuthActionKO)
			})
		})

		Convey("When I try check auth for user-b using chain-a", func() {

			s := &mockSession{
				state: &tls.ConnectionState{
					PeerCertificates: []*x509.Certificate{
						userCertB,
					},
				},
			}

			opts := x509.VerifyOptions{
				Roots:     certPoolA,
				KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
			}

			auth := NewMTLSSessionAuthenticator(opts, func(a bahamut.AuthAction, c bahamut.Context, s bahamut.Session) bahamut.AuthAction { return a }, nil, CertificateCheckModeTLSStateOnly)

			action, err := auth.AuthenticateSession(s)

			Convey("Then err should be nil", func() {
				So(err, ShouldBeNil)
			})

			Convey("Then action should be bahamut.AuthActionKO", func() {
				So(action, ShouldEqual, bahamut.AuthActionKO)
			})
		})
	})
}

func TestCertificatesFromHeaderThenTLSState(t *testing.T) {

	cdata, _ := ioutil.ReadFile("./fixtures/user-a-cert.pem")
	cblock, _ := pem.Decode(cdata)
	cert, _ := x509.ParseCertificate(cblock.Bytes)

	cdata2, _ := ioutil.ReadFile("./fixtures/user-b-cert.pem")
	cblock2, _ := pem.Decode(cdata2)
	cert2, _ := x509.ParseCertificate(cblock2.Bytes)

	type args struct {
		state      *tls.ConnectionState
		headerData string
	}
	tests := []struct {
		name      string
		args      args
		wantCerts []*x509.Certificate
		wantErr   bool
	}{
		{
			"with header data",
			args{
				nil,
				`-----BEGIN CERTIFICATE----- MIIBVTCB/aADAgECAhARq0YiIWt0OVVp+RqLmtgmMAoGCCqGSM49BAMCMBMxETAP BgNVBAMTCHNpZ25lci1hMB4XDTE3MTAxMzIzMTUzMVoXDTI3MDgyMjIzMTUzMVow ETEPMA0GA1UEAxMGdXNlci1hMFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAE6Un9 07SdKaQS+DYeLSQXEHe9TqXZFKMUvxoT7DNFTVAMKD4znqc07A0NnUyya05pRcAW Yup+wvlTyEBFA7FTP6M1MDMwDgYDVR0PAQH/BAQDAgeAMBMGA1UdJQQMMAoGCCsG AQUFBwMCMAwGA1UdEwEB/wQCMAAwCgYIKoZIzj0EAwIDRwAwRAIgEOun3or4nuub 1i2QgNkOOSfxAbEG/stM2nEjTemXtpECIH3KX72mnKbd8eLSYFIsbAz6B55GBeF8 Tuzw3YBRYF5F -----END CERTIFICATE-----`,
			},
			[]*x509.Certificate{cert},
			false,
		},
		{
			"with invalid header data",
			args{
				nil,
				`---`,
			},
			nil,
			true,
		},
		{
			"with emtpty tls state",
			args{
				&tls.ConnectionState{},
				"",
			},
			nil,
			true,
		},
		{
			"with correct tls state",
			args{
				&tls.ConnectionState{
					PeerCertificates: []*x509.Certificate{cert, cert},
				},
				"",
			},
			[]*x509.Certificate{cert, cert},
			false,
		},
		{
			"with both tls state and header",
			args{
				&tls.ConnectionState{
					PeerCertificates: []*x509.Certificate{cert, cert},
				},
				`-----BEGIN CERTIFICATE----- MIIBVzCB/qADAgECAhEAxEXp+z1wWArT0+U85V5BhjAKBggqhkjOPQQDAjATMREw DwYDVQQDEwhzaWduZXItYjAeFw0xNzEwMTMyMzE1NDFaFw0yNzA4MjIyMzE1NDFa MBExDzANBgNVBAMTBnVzZXItYjBZMBMGByqGSM49AgEGCCqGSM49AwEHA0IABDQX w/q6ZMJiN47vM6FRuodSBmMZJUfWGy3GRtabnqf63xIXLdI7C2C6ad8fDhfmAxwy 5Wr8JHcYvcyR7+jWEiyjNTAzMA4GA1UdDwEB/wQEAwIHgDATBgNVHSUEDDAKBggr BgEFBQcDAjAMBgNVHRMBAf8EAjAAMAoGCCqGSM49BAMCA0gAMEUCIQDtf3/E/SdO r0tQSIEAZEAXsoprmc1G3GpZLxyr56fYMQIgFTZ3y3kBi27ec8Kq25RZuVOm8fW5 es4EmhhC8ezi2Jo= -----END CERTIFICATE-----`,
			},
			[]*x509.Certificate{cert2},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotCerts, err := CertificatesFromHeaderThenTLSState(tt.args.state, tt.args.headerData)
			if (err != nil) != tt.wantErr {
				t.Errorf("CertificatesFromHeaderThenTLSState() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotCerts, tt.wantCerts) {
				t.Errorf("CertificatesFromHeaderThenTLSState() = %v, want %v", gotCerts, tt.wantCerts)
			}
		})
	}
}

func TestCertificatesFromTLSStateThenHeader(t *testing.T) {

	cdata, _ := ioutil.ReadFile("./fixtures/user-a-cert.pem")
	cblock, _ := pem.Decode(cdata)
	cert, _ := x509.ParseCertificate(cblock.Bytes)

	type args struct {
		state      *tls.ConnectionState
		headerData string
	}
	tests := []struct {
		name      string
		args      args
		wantCerts []*x509.Certificate
		wantErr   bool
	}{
		{
			"with header data",
			args{
				nil,
				`-----BEGIN CERTIFICATE----- MIIBVTCB/aADAgECAhARq0YiIWt0OVVp+RqLmtgmMAoGCCqGSM49BAMCMBMxETAP BgNVBAMTCHNpZ25lci1hMB4XDTE3MTAxMzIzMTUzMVoXDTI3MDgyMjIzMTUzMVow ETEPMA0GA1UEAxMGdXNlci1hMFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAE6Un9 07SdKaQS+DYeLSQXEHe9TqXZFKMUvxoT7DNFTVAMKD4znqc07A0NnUyya05pRcAW Yup+wvlTyEBFA7FTP6M1MDMwDgYDVR0PAQH/BAQDAgeAMBMGA1UdJQQMMAoGCCsG AQUFBwMCMAwGA1UdEwEB/wQCMAAwCgYIKoZIzj0EAwIDRwAwRAIgEOun3or4nuub 1i2QgNkOOSfxAbEG/stM2nEjTemXtpECIH3KX72mnKbd8eLSYFIsbAz6B55GBeF8 Tuzw3YBRYF5F -----END CERTIFICATE-----`,
			},
			[]*x509.Certificate{cert},
			false,
		},
		{
			"with invalid header data",
			args{
				nil,
				`---`,
			},
			nil,
			true,
		},
		{
			"with emtpty tls state",
			args{
				&tls.ConnectionState{},
				"",
			},
			nil,
			true,
		},
		{
			"with correct tls state",
			args{
				&tls.ConnectionState{
					PeerCertificates: []*x509.Certificate{cert, cert},
				},
				"",
			},
			[]*x509.Certificate{cert, cert},
			false,
		},
		{
			"with both tls state and header",
			args{
				&tls.ConnectionState{
					PeerCertificates: []*x509.Certificate{cert, cert},
				},
				`-----BEGIN CERTIFICATE----- MIIBVzCB/qADAgECAhEAxEXp+z1wWArT0+U85V5BhjAKBggqhkjOPQQDAjATMREw DwYDVQQDEwhzaWduZXItYjAeFw0xNzEwMTMyMzE1NDFaFw0yNzA4MjIyMzE1NDFa MBExDzANBgNVBAMTBnVzZXItYjBZMBMGByqGSM49AgEGCCqGSM49AwEHA0IABDQX w/q6ZMJiN47vM6FRuodSBmMZJUfWGy3GRtabnqf63xIXLdI7C2C6ad8fDhfmAxwy 5Wr8JHcYvcyR7+jWEiyjNTAzMA4GA1UdDwEB/wQEAwIHgDATBgNVHSUEDDAKBggr BgEFBQcDAjAMBgNVHRMBAf8EAjAAMAoGCCqGSM49BAMCA0gAMEUCIQDtf3/E/SdO r0tQSIEAZEAXsoprmc1G3GpZLxyr56fYMQIgFTZ3y3kBi27ec8Kq25RZuVOm8fW5 es4EmhhC8ezi2Jo= -----END CERTIFICATE-----`,
			},
			[]*x509.Certificate{cert, cert},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotCerts, err := CertificatesFromTLSStateThenHeader(tt.args.state, tt.args.headerData)
			if (err != nil) != tt.wantErr {
				t.Errorf("CertificatesFromTLSStateThenHeader() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotCerts, tt.wantCerts) {
				t.Errorf("CertificatesFromTLSStateThenHeader() = %v, want %v", gotCerts, tt.wantCerts)
			}
		})
	}
}

func TestCertificatesFromTLSState(t *testing.T) {

	cdata, _ := ioutil.ReadFile("./fixtures/user-a-cert.pem")
	cblock, _ := pem.Decode(cdata)
	cert, _ := x509.ParseCertificate(cblock.Bytes)

	type args struct {
		state *tls.ConnectionState
	}
	tests := []struct {
		name      string
		args      args
		wantCerts []*x509.Certificate
		wantErr   bool
	}{
		{
			"with correct tls state",
			args{
				&tls.ConnectionState{
					PeerCertificates: []*x509.Certificate{cert, cert},
				},
			},
			[]*x509.Certificate{cert, cert},
			false,
		},
		{
			"with nil tls state",
			args{
				nil,
			},
			nil,
			true,
		},
		{
			"with emtpty tls state",
			args{
				&tls.ConnectionState{},
			},
			nil,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotCerts, err := CertificatesFromTLSState(tt.args.state)
			if (err != nil) != tt.wantErr {
				t.Errorf("CertificatesFromTLSState() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotCerts, tt.wantCerts) {
				t.Errorf("CertificatesFromTLSState() = %v, want %v", gotCerts, tt.wantCerts)
			}
		})
	}
}

func TestCertificatesFromHeader(t *testing.T) {

	cdata, _ := ioutil.ReadFile("./fixtures/user-a-cert.pem")
	cblock, _ := pem.Decode(cdata)
	cert, _ := x509.ParseCertificate(cblock.Bytes)

	type args struct {
		headerData string
	}
	tests := []struct {
		name      string
		args      args
		wantCerts []*x509.Certificate
		wantErr   bool
	}{
		{
			"with header data",
			args{
				`-----BEGIN CERTIFICATE----- MIIBVTCB/aADAgECAhARq0YiIWt0OVVp+RqLmtgmMAoGCCqGSM49BAMCMBMxETAP BgNVBAMTCHNpZ25lci1hMB4XDTE3MTAxMzIzMTUzMVoXDTI3MDgyMjIzMTUzMVow ETEPMA0GA1UEAxMGdXNlci1hMFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAE6Un9 07SdKaQS+DYeLSQXEHe9TqXZFKMUvxoT7DNFTVAMKD4znqc07A0NnUyya05pRcAW Yup+wvlTyEBFA7FTP6M1MDMwDgYDVR0PAQH/BAQDAgeAMBMGA1UdJQQMMAoGCCsG AQUFBwMCMAwGA1UdEwEB/wQCMAAwCgYIKoZIzj0EAwIDRwAwRAIgEOun3or4nuub 1i2QgNkOOSfxAbEG/stM2nEjTemXtpECIH3KX72mnKbd8eLSYFIsbAz6B55GBeF8 Tuzw3YBRYF5F -----END CERTIFICATE-----`,
			},
			[]*x509.Certificate{cert},
			false,
		},
		{
			"with empty header data",
			args{
				``,
			},
			nil,
			true,
		},
		{
			"with invalid header data",
			args{
				`---`,
			},
			nil,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotCerts, err := CertificatesFromHeader(tt.args.headerData)
			if (err != nil) != tt.wantErr {
				t.Errorf("CertificatesFromHeader() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotCerts, tt.wantCerts) {
				t.Errorf("CertificatesFromHeader() = %v, want %v", gotCerts, tt.wantCerts)
			}
		})
	}
}

func Test_decodeCertHeader(t *testing.T) {

	cdata, _ := ioutil.ReadFile("./fixtures/user-a-cert.pem")
	cblock, _ := pem.Decode(cdata)
	cert, _ := x509.ParseCertificate(cblock.Bytes)

	type args struct {
		header string
	}
	tests := []struct {
		name    string
		args    args
		want    []*x509.Certificate
		wantErr bool
	}{
		{
			"valid",
			args{
				`-----BEGIN CERTIFICATE----- MIIBVTCB/aADAgECAhARq0YiIWt0OVVp+RqLmtgmMAoGCCqGSM49BAMCMBMxETAP BgNVBAMTCHNpZ25lci1hMB4XDTE3MTAxMzIzMTUzMVoXDTI3MDgyMjIzMTUzMVow ETEPMA0GA1UEAxMGdXNlci1hMFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAE6Un9 07SdKaQS+DYeLSQXEHe9TqXZFKMUvxoT7DNFTVAMKD4znqc07A0NnUyya05pRcAW Yup+wvlTyEBFA7FTP6M1MDMwDgYDVR0PAQH/BAQDAgeAMBMGA1UdJQQMMAoGCCsG AQUFBwMCMAwGA1UdEwEB/wQCMAAwCgYIKoZIzj0EAwIDRwAwRAIgEOun3or4nuub 1i2QgNkOOSfxAbEG/stM2nEjTemXtpECIH3KX72mnKbd8eLSYFIsbAz6B55GBeF8 Tuzw3YBRYF5F -----END CERTIFICATE-----`,
			},
			[]*x509.Certificate{cert},
			false,
		},
		{
			"too small",
			args{
				`-----BEGIN CERTIFICATE----- -`,
			},
			nil,
			true,
		},
		{
			"empty pem",
			args{
				`-----sdkjfskdjfhdsljkghdksjhgkdfsjhgkjdfhgdfshgkldfsjhgkljdfhglkjhdsfklghjdfsklgh----`,
			},
			nil,
			true,
		},
		{
			"invalid pem",
			args{
				`-----BEGIN CERTIFICATE----- NOPE -----END CERTIFICATE-----`,
			},
			nil,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := decodeCertHeader(tt.args.header)
			if (err != nil) != tt.wantErr {
				t.Errorf("decodeCertHeader() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("decodeCertHeader() = %v, want %v", got, tt.want)
			}
		})
	}
}
