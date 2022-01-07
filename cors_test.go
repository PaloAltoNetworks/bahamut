package bahamut

import (
	"net/http"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestNewDefaultCORSAccessControlPolicy(t *testing.T) {

	Convey("Calling NewDefaultCORSAccessControlPolicy should work", t, func() {
		c := NewDefaultCORSController("origin.com", []string{"additionalorigin.com"})
		ac := c.PolicyForRequest(nil)

		So(ac.AllowOrigin, ShouldEqual, "origin.com")
		So(ac.additionalOrigins, ShouldResemble, map[string]struct{}{"additionalorigin.com": {}})
		So(ac.AllowCredentials, ShouldBeTrue)
		So(ac.MaxAge, ShouldEqual, 1500)
		So(ac.AllowHeaders, ShouldResemble, []string{
			"Authorization",
			"Accept",
			"Content-Type",
			"Cache-Control",
			"Cookie",
			"If-Modified-Since",
			"X-Requested-With",
			"X-Count-Total",
			"X-Namespace",
			"X-External-Tracking-Type",
			"X-External-Tracking-ID",
			"X-TLS-Client-Certificate",
			"Accept-Encoding",
			"X-Fields",
			"X-Read-Consistency",
			"X-Write-Consistency",
			"Idempotency-Key",
		})
		So(ac.AllowMethods, ShouldResemble, []string{
			"GET",
			"POST",
			"PUT",
			"DELETE",
			"PATCH",
			"HEAD",
			"OPTIONS",
		})
		So(ac.ExposeHeaders, ShouldResemble, []string{
			"X-Requested-With",
			"X-Count-Total",
			"X-Namespace",
			"X-Messages",
			"X-Fields",
			"X-Next",
		})
	})
}

func TestCORSInject(t *testing.T) {

	Convey("Calling inject with no http.Heade should work", t, func() {
		a := NewDefaultCORSController("origin", nil)
		ac := a.PolicyForRequest(nil)
		So(func() { ac.Inject(nil, "", false) }, ShouldNotPanic)
	})

	Convey("Calling inject without passing request origin should work", t, func() {
		a := NewDefaultCORSController("origin", nil)
		ac := a.PolicyForRequest(nil)
		h := http.Header{}
		So(func() { ac.Inject(h, "", false) }, ShouldNotPanic)
		So(h.Get("Access-Control-Allow-Headers"), ShouldBeEmpty)
		So(h.Get("Access-Control-Allow-Methods"), ShouldBeEmpty)
		So(h.Get("Access-Control-Max-Age"), ShouldBeEmpty)
		So(h.Get("Access-Control-Allow-Origin"), ShouldEqual, "origin")
		So(h.Get("Access-Control-Expose-Headers"), ShouldNotBeEmpty)
		So(h.Get("Access-Control-Allow-Credentials"), ShouldEqual, "true")
	})

	Convey("Calling inject with request prefligh should work", t, func() {
		a := NewDefaultCORSController("origin", nil)
		ac := a.PolicyForRequest(nil)
		h := http.Header{}
		So(func() { ac.Inject(h, "", true) }, ShouldNotPanic)
		So(h.Get("Access-Control-Allow-Headers"), ShouldNotBeEmpty)
		So(h.Get("Access-Control-Allow-Methods"), ShouldNotBeEmpty)
		So(h.Get("Access-Control-Max-Age"), ShouldEqual, "1500")
		So(h.Get("Access-Control-Allow-Origin"), ShouldEqual, "origin")
		So(h.Get("Access-Control-Expose-Headers"), ShouldNotBeEmpty)
		So(h.Get("Access-Control-Allow-Credentials"), ShouldEqual, "true")
	})

	Convey("Calling inject with matching origin should work", t, func() {
		a := NewDefaultCORSController("origin", nil)
		ac := a.PolicyForRequest(nil)
		h := http.Header{}
		So(func() { ac.Inject(h, "origin", false) }, ShouldNotPanic)
		So(h.Get("Access-Control-Allow-Headers"), ShouldBeEmpty)
		So(h.Get("Access-Control-Allow-Methods"), ShouldBeEmpty)
		So(h.Get("Access-Control-Max-Age"), ShouldBeEmpty)
		So(h.Get("Access-Control-Allow-Origin"), ShouldEqual, "origin")
		So(h.Get("Access-Control-Expose-Headers"), ShouldNotBeEmpty)
		So(h.Get("Access-Control-Allow-Credentials"), ShouldEqual, "true")
	})

	Convey("Calling inject with non matching origin should work", t, func() {
		a := NewDefaultCORSController("origin", nil)
		ac := a.PolicyForRequest(nil)
		h := http.Header{}
		So(func() { ac.Inject(h, "notorigin", false) }, ShouldNotPanic)
		So(h.Get("Access-Control-Allow-Headers"), ShouldBeEmpty)
		So(h.Get("Access-Control-Allow-Methods"), ShouldBeEmpty)
		So(h.Get("Access-Control-Max-Age"), ShouldBeEmpty)
		So(h.Get("Access-Control-Allow-Origin"), ShouldEqual, "origin")
		So(h.Get("Access-Control-Expose-Headers"), ShouldNotBeEmpty)
		So(h.Get("Access-Control-Allow-Credentials"), ShouldEqual, "true")
	})

	Convey("Calling inject with matching additional origin should work", t, func() {
		a := NewDefaultCORSController("origin", []string{"additional.com"})
		ac := a.PolicyForRequest(nil)
		h := http.Header{}
		So(func() { ac.Inject(h, "additional.com", false) }, ShouldNotPanic)
		So(h.Get("Access-Control-Allow-Headers"), ShouldBeEmpty)
		So(h.Get("Access-Control-Allow-Methods"), ShouldBeEmpty)
		So(h.Get("Access-Control-Max-Age"), ShouldBeEmpty)
		So(h.Get("Access-Control-Allow-Origin"), ShouldEqual, "additional.com")
		So(h.Get("Access-Control-Expose-Headers"), ShouldNotBeEmpty)
		So(h.Get("Access-Control-Allow-Credentials"), ShouldEqual, "true")
	})

	Convey("Calling inject with * configured", t, func() {
		a := NewDefaultCORSController("*", []string{"additional.com"})
		ac := a.PolicyForRequest(nil)
		h := http.Header{}
		So(func() { ac.Inject(h, "additional.com", false) }, ShouldNotPanic)
		So(h.Get("Access-Control-Allow-Headers"), ShouldBeEmpty)
		So(h.Get("Access-Control-Allow-Methods"), ShouldBeEmpty)
		So(h.Get("Access-Control-Max-Age"), ShouldBeEmpty)
		So(h.Get("Access-Control-Allow-Origin"), ShouldEqual, "*")
		So(h.Get("Access-Control-Expose-Headers"), ShouldNotBeEmpty)
		So(h.Get("Access-Control-Allow-Credentials"), ShouldEqual, "")
	})

	Convey("Calling inject with mirroring configured and passed origin", t, func() {
		a := NewDefaultCORSController(CORSOriginMirror, nil)
		ac := a.PolicyForRequest(nil)
		h := http.Header{}
		So(func() { ac.Inject(h, "hello.com", false) }, ShouldNotPanic)
		So(h.Get("Access-Control-Allow-Headers"), ShouldBeEmpty)
		So(h.Get("Access-Control-Allow-Methods"), ShouldBeEmpty)
		So(h.Get("Access-Control-Max-Age"), ShouldBeEmpty)
		So(h.Get("Access-Control-Allow-Origin"), ShouldEqual, "hello.com")
		So(h.Get("Access-Control-Expose-Headers"), ShouldNotBeEmpty)
		So(h.Get("Access-Control-Allow-Credentials"), ShouldEqual, "true")
	})

	Convey("Calling inject with mirroring configured and no passed origin", t, func() {
		a := NewDefaultCORSController(CORSOriginMirror, nil)
		ac := a.PolicyForRequest(nil)
		h := http.Header{}
		So(func() { ac.Inject(h, "", false) }, ShouldNotPanic)
		So(h.Get("Access-Control-Allow-Headers"), ShouldBeEmpty)
		So(h.Get("Access-Control-Allow-Methods"), ShouldBeEmpty)
		So(h.Get("Access-Control-Max-Age"), ShouldBeEmpty)
		So(h.Get("Access-Control-Allow-Origin"), ShouldEqual, "")
		So(h.Get("Access-Control-Expose-Headers"), ShouldNotBeEmpty)
		So(h.Get("Access-Control-Allow-Credentials"), ShouldEqual, "")
	})
}
