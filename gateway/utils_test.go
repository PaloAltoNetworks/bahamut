package gateway

import (
	"crypto/tls"
	"net/http"
	"reflect"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
	"go.aporeto.io/bahamut"
	"go.aporeto.io/tg/tglib"
)

func Test_injectGeneralHeader(t *testing.T) {
	type args struct {
		h http.Header
	}
	tests := []struct {
		args args
		want http.Header
		name string
	}{
		{
			name: "simple",
			args: args{
				http.Header{},
			},
			want: http.Header{
				"Strict-Transport-Security": []string{"max-age=31536000; includeSubDomains; preload"},
				"X-Frame-Options":           []string{"DENY"},
				"X-Content-Type-Options":    []string{"nosniff"},
				"X-Xss-Protection":          []string{"1; mode=block"},
				"Cache-Control":             []string{"private, no-transform"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := injectGeneralHeader(tt.args.h); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("injectGeneralHeader() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_injectCORSHeader(t *testing.T) {
	type args struct {
		h                    http.Header
		corsOrigin           string
		origin               string
		method               string
		additionalCorsOrigin []string
		corsAllowCredentials bool
	}
	tests := []struct {
		want http.Header
		name string
		args args
	}{
		{
			name: "default",
			args: args{
				h:                    http.Header{},
				corsOrigin:           bahamut.CORSOriginMirror,
				additionalCorsOrigin: nil,
				corsAllowCredentials: true,
				origin:               "chien",
				method:               http.MethodGet,
			},
			want: http.Header{
				"Access-Control-Allow-Origin":      {"chien"},
				"Access-Control-Expose-Headers":    {"X-Requested-With, X-Count-Total, X-Namespace, X-Messages, X-Fields, X-Next"},
				"Access-Control-Allow-Credentials": {"true"},
			},
		},
		{
			name: "default OPTIONS",
			args: args{
				h:                    http.Header{},
				corsOrigin:           bahamut.CORSOriginMirror,
				additionalCorsOrigin: nil,
				corsAllowCredentials: true,
				origin:               "chien",
				method:               http.MethodOptions,
			},
			want: http.Header{
				"Access-Control-Allow-Origin":      {"chien"},
				"Access-Control-Expose-Headers":    {"X-Requested-With, X-Count-Total, X-Namespace, X-Messages, X-Fields, X-Next"},
				"Access-Control-Allow-Methods":     {"GET, POST, PUT, DELETE, PATCH, HEAD, OPTIONS"},
				"Access-Control-Allow-Headers":     {"Authorization, Accept, Content-Type, Cache-Control, Cookie, If-Modified-Since, X-Requested-With, X-Count-Total, X-Namespace, X-External-Tracking-Type, X-External-Tracking-ID, X-TLS-Client-Certificate, Accept-Encoding, X-Fields, X-Read-Consistency, X-Write-Consistency, Idempotency-Key"},
				"Access-Control-Allow-Credentials": {"true"},
				"Access-Control-Max-Age":           {"1500"},
			},
		},
		{
			name: "dev",
			args: args{
				h:                    http.Header{},
				corsOrigin:           "dog",
				additionalCorsOrigin: nil,
				corsAllowCredentials: true,
				origin:               "chien",
				method:               http.MethodGet,
			},
			want: http.Header{
				"Access-Control-Allow-Origin":      {"dog"},
				"Access-Control-Expose-Headers":    {"X-Requested-With, X-Count-Total, X-Namespace, X-Messages, X-Fields, X-Next"},
				"Access-Control-Allow-Credentials": {"true"},
			},
		},
		{
			name: "dev OPTIONS",
			args: args{
				h:                    http.Header{},
				corsOrigin:           "dog",
				additionalCorsOrigin: nil,
				corsAllowCredentials: true,
				origin:               "chien",
				method:               http.MethodOptions,
			},
			want: http.Header{
				"Access-Control-Allow-Origin":      {"dog"},
				"Access-Control-Expose-Headers":    {"X-Requested-With, X-Count-Total, X-Namespace, X-Messages, X-Fields, X-Next"},
				"Access-Control-Allow-Methods":     {"GET, POST, PUT, DELETE, PATCH, HEAD, OPTIONS"},
				"Access-Control-Allow-Headers":     {"Authorization, Accept, Content-Type, Cache-Control, Cookie, If-Modified-Since, X-Requested-With, X-Count-Total, X-Namespace, X-External-Tracking-Type, X-External-Tracking-ID, X-TLS-Client-Certificate, Accept-Encoding, X-Fields, X-Read-Consistency, X-Write-Consistency, Idempotency-Key"},
				"Access-Control-Allow-Credentials": {"true"},
				"Access-Control-Max-Age":           {"1500"},
			},
		},
		{
			name: "additional origin",
			args: args{
				h:                    http.Header{},
				corsOrigin:           "dog",
				additionalCorsOrigin: []string{"chien"},
				corsAllowCredentials: true,
				origin:               "chien",
				method:               http.MethodGet,
			},
			want: http.Header{
				"Access-Control-Allow-Origin":      {"chien"},
				"Access-Control-Expose-Headers":    {"X-Requested-With, X-Count-Total, X-Namespace, X-Messages, X-Fields, X-Next"},
				"Access-Control-Allow-Credentials": {"true"},
			},
		},
		{
			name: "additional origin OPTIONS",
			args: args{
				h:                    http.Header{},
				corsOrigin:           "dog",
				additionalCorsOrigin: []string{"chien"},
				corsAllowCredentials: true,
				origin:               "chien",
				method:               http.MethodOptions,
			},
			want: http.Header{
				"Access-Control-Allow-Origin":      {"chien"},
				"Access-Control-Expose-Headers":    {"X-Requested-With, X-Count-Total, X-Namespace, X-Messages, X-Fields, X-Next"},
				"Access-Control-Allow-Methods":     {"GET, POST, PUT, DELETE, PATCH, HEAD, OPTIONS"},
				"Access-Control-Allow-Headers":     {"Authorization, Accept, Content-Type, Cache-Control, Cookie, If-Modified-Since, X-Requested-With, X-Count-Total, X-Namespace, X-External-Tracking-Type, X-External-Tracking-ID, X-TLS-Client-Certificate, Accept-Encoding, X-Fields, X-Read-Consistency, X-Write-Consistency, Idempotency-Key"},
				"Access-Control-Allow-Credentials": {"true"},
				"Access-Control-Max-Age":           {"1500"},
			},
		},
		{
			name: "default empty origin",
			args: args{
				h:                    http.Header{},
				corsOrigin:           bahamut.CORSOriginMirror,
				additionalCorsOrigin: nil,
				corsAllowCredentials: true,
				origin:               "",
				method:               http.MethodGet,
			},
			want: http.Header{
				"Access-Control-Expose-Headers": {"X-Requested-With, X-Count-Total, X-Namespace, X-Messages, X-Fields, X-Next"},
			},
		},
		{
			name: "default empty OPTIONS",
			args: args{
				h:                    http.Header{},
				corsOrigin:           bahamut.CORSOriginMirror,
				additionalCorsOrigin: nil,
				corsAllowCredentials: true,
				origin:               "",
				method:               http.MethodOptions,
			},
			want: http.Header{
				"Access-Control-Expose-Headers": {"X-Requested-With, X-Count-Total, X-Namespace, X-Messages, X-Fields, X-Next"},
				"Access-Control-Allow-Methods":  {"GET, POST, PUT, DELETE, PATCH, HEAD, OPTIONS"},
				"Access-Control-Allow-Headers":  {"Authorization, Accept, Content-Type, Cache-Control, Cookie, If-Modified-Since, X-Requested-With, X-Count-Total, X-Namespace, X-External-Tracking-Type, X-External-Tracking-ID, X-TLS-Client-Certificate, Accept-Encoding, X-Fields, X-Read-Consistency, X-Write-Consistency, Idempotency-Key"},
				"Access-Control-Max-Age":        {"1500"},
			},
		},

		{
			name: "star with credential set to true (should be ignored)",
			args: args{
				h:                    http.Header{},
				corsOrigin:           "*",
				additionalCorsOrigin: nil,
				corsAllowCredentials: true,
				origin:               "",
				method:               http.MethodOptions,
			},
			want: http.Header{
				"Access-Control-Allow-Origin":   {"*"},
				"Access-Control-Expose-Headers": {"X-Requested-With, X-Count-Total, X-Namespace, X-Messages, X-Fields, X-Next"},
				"Access-Control-Allow-Methods":  {"GET, POST, PUT, DELETE, PATCH, HEAD, OPTIONS"},
				"Access-Control-Allow-Headers":  {"Authorization, Accept, Content-Type, Cache-Control, Cookie, If-Modified-Since, X-Requested-With, X-Count-Total, X-Namespace, X-External-Tracking-Type, X-External-Tracking-ID, X-TLS-Client-Certificate, Accept-Encoding, X-Fields, X-Read-Consistency, X-Write-Consistency, Idempotency-Key"},
				"Access-Control-Max-Age":        {"1500"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := injectCORSHeader(tt.args.h, tt.args.corsOrigin, tt.args.additionalCorsOrigin, tt.args.corsAllowCredentials, tt.args.origin, tt.args.method); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("injectCORSHeader() = \ngot:  %v\nwant: %v", got, tt.want)
			}
		})
	}
}

type fakeAddr struct {
	net string
	str string
}

func (a fakeAddr) Network() string {
	return a.net
}

func (a fakeAddr) String() string {
	return a.str
}

func Test_makeProxyProtocolSourceChecker(t *testing.T) {

	Convey("Given I create a source checker func with a valid network", t, func() {

		sc, err := makeProxyProtocolSourceChecker("10.0.0/8")
		So(sc, ShouldBeNil)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldEqual, "unable to parse CIDR: invalid CIDR address: 10.0.0/8")
	})

	Convey("Given I create a source checker func with a valid network", t, func() {

		sc, err := makeProxyProtocolSourceChecker("10.0.0.0/8")
		So(err, ShouldBeNil)

		Convey("When call it on a authorized net.Addr", func() {
			ok, err := sc(fakeAddr{str: "10.1.1.1:123"})
			Convey("Then it is working", func() {
				So(ok, ShouldBeTrue)
				So(err, ShouldBeNil)
			})
		})

		Convey("When call it on a unauthorized net.Addr", func() {
			ok, err := sc(fakeAddr{str: "11.1.1.1:123"})
			Convey("Then it is working", func() {
				So(ok, ShouldBeFalse)
				So(err, ShouldBeNil)
			})
		})

		Convey("When call it on a invalid net.Addr", func() {
			ok, err := sc(fakeAddr{str: "11.1.1."})
			Convey("Then it is working", func() {
				So(ok, ShouldBeFalse)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "unable to parse net.Addr: address 11.1.1.: missing port in address")
			})
		})
	})
}

func TestMakeGoodByeServer(t *testing.T) {

	Convey("Given I call makeGoodbyeServer", t, func() {

		cert, key, err := tglib.ReadCertificatePEM("../fixtures/certs/server-cert.pem", "../fixtures/certs/server-key.pem", "")
		if err != nil {
			panic(err)
		}
		tlscert, err := tglib.ToTLSCertificate(cert, key)
		if err != nil {
			panic(err)
		}

		tlscfg := &tls.Config{
			Certificates: []tls.Certificate{tlscert},
		}
		srv := makeGoodbyeServer(":58344", tlscfg)

		Convey("Then srv should be correct", func() {
			So(srv.TLSConfig, ShouldEqual, tlscfg)
			So(srv.Addr, ShouldEqual, ":58344")
		})

		go func() {
			if err := srv.ListenAndServeTLS("", ""); err != nil {
				panic(err)
			}
		}()

		time.Sleep(time.Second)
		cl := http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			},
		}

		req, err := http.NewRequest(http.MethodDelete, "https://127.0.0.1:58344/dog", nil)
		if err != nil {
			panic(err)
		}
		resp, err := cl.Do(req)
		if err != nil {
			panic(err)
		}

		So(resp.StatusCode, ShouldEqual, http.StatusServiceUnavailable)
	})
}
