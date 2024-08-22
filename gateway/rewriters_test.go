package gateway

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"strings"
	"testing"

	// nolint:revive // Allow dot imports for readability in tests
	. "github.com/smartystreets/goconvey/convey"
	"go.aporeto.io/tg/tglib"
)

func Test_requestRewriter(t *testing.T) {

	Convey("Given I have a requestRewriter", t, func() {

		rw := &requestRewriter{}

		customReqriter := func(req *httputil.ProxyRequest, private bool) error {
			switch private {
			case true:
				req.Out.Header.Add("X-Aporeto-Gateway", "private")
			case false:
				req.Out.Header.Add("X-Aporeto-Gateway", "public")
			}

			return nil
		}

		Convey("When I call Rewrite with a custom rewriter handling private mode", func() {

			rw.private = true
			rw.customRewriter = customReqriter

			r, _ := http.NewRequest(http.MethodGet, "http://127.0.0.1", nil)
			r.TLS = &tls.ConnectionState{}
			r2 := r.Clone(context.Background())

			rw.Rewrite(&httputil.ProxyRequest{In: r, Out: r2})

			Convey("Then the response should be correct", func() {
				So(r2.Header.Get("X-Aporeto-Gateway"), ShouldEqual, "private")
			})
		})

		Convey("When I call Rewrite with a custom rewriter handling public mode", func() {

			rw.private = false
			rw.customRewriter = customReqriter

			r, _ := http.NewRequest(http.MethodGet, "http://127.0.0.1", nil)
			r.TLS = &tls.ConnectionState{}
			r2 := r.Clone(context.Background())

			rw.Rewrite(&httputil.ProxyRequest{In: r, Out: r2})

			Convey("Then the response should be correct", func() {
				So(r2.Header.Get("X-Aporeto-Gateway"), ShouldEqual, "public")
			})
		})

		Convey("When I call Rewrite it with custom X-Forwarded-For with no client IP", func() {

			r, _ := http.NewRequest(http.MethodGet, "http://127.0.0.1", nil)
			r.TLS = &tls.ConnectionState{}
			r2 := r.Clone(context.Background())

			r.Header.Set("X-Forwarded-For", "A")

			rw.Rewrite(&httputil.ProxyRequest{In: r, Out: r2})

			Convey("Then the response should be correct", func() {
				So(r2.Header.Get("X-Forwarded-For"), ShouldEqual, "")
			})
		})

		Convey("When I call Rewrite it with custom X-Forwarded-For and with client IP", func() {

			r, _ := http.NewRequest(http.MethodGet, "http://127.0.0.1", nil)
			r.RemoteAddr = "C:7878"
			r.TLS = &tls.ConnectionState{}
			r2 := r.Clone(context.Background())

			r.Header.Set("X-Forwarded-For", "A")

			rw.Rewrite(&httputil.ProxyRequest{In: r, Out: r2})

			Convey("Then the response should be correct", func() {
				So(r2.Header.Get("X-Forwarded-For"), ShouldEqual, "C")
			})
		})

		Convey("When I call Rewrite it with custom X-Forwarded-For and trustForwardHeader is set", func() {

			rw.trustForwardHeader = true

			r, _ := http.NewRequest(http.MethodGet, "http://127.0.0.1", nil)
			r.TLS = &tls.ConnectionState{}
			r2 := r.Clone(context.Background())

			r.Header.Set("X-Forwarded-For", "A")

			rw.Rewrite(&httputil.ProxyRequest{In: r, Out: r2})

			Convey("Then the response should be correct", func() {
				So(r2.Header.Get("X-Forwarded-For"), ShouldEqual, "A")
			})
		})

		Convey("When I call Rewrite it with marked as ws internal", func() {

			r, _ := http.NewRequest(http.MethodGet, "http://127.0.0.1", nil)
			r.TLS = &tls.ConnectionState{}
			r.RemoteAddr = "1.1.1.1:11"
			r2 := r.Clone(context.Background())

			r.Header.Set(internalWSMarkingHeader, "1")

			rw.Rewrite(&httputil.ProxyRequest{In: r, Out: r2})

			Convey("Then the response should be correct", func() {
				So(r2.Header.Get("X-Forwarded-For"), ShouldEqual, "1.1.1.1")
			})
		})

		Convey("When I call Rewrite it with marked as ws internal with bad address", func() {

			r, _ := http.NewRequest(http.MethodGet, "http://127.0.0.1", nil)
			r.TLS = &tls.ConnectionState{}
			r.RemoteAddr = "oh no"
			r2 := r.Clone(context.Background())
			r.Header.Set(internalWSMarkingHeader, "1")

			rw.Rewrite(&httputil.ProxyRequest{In: r, Out: r2})

			Convey("Then the response should be correct", func() {
				So(r2.Header.Get("X-Forwarded-For"), ShouldEqual, "")
			})
		})

		Convey("When I call Rewrite it with a valid TLS client certificate", func() {

			certData := []byte(`-----BEGIN CERTIFICATE-----
MIIBKjCB0qADAgECAhBLliCl1URppVpoHheDFLdKMAoGCCqGSM49BAMCMA8xDTAL
BgNVBAMTBHRvdG8wHhcNMjAwMjI4MTgzMTA2WhcNMzAwMTA2MTgzMTA2WjAPMQ0w
CwYDVQQDEwR0b3RvMFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEy6WbbGdtOM0b
DXryjj9EZiFKphTurtQTeo5whOoTTodNHJwMxAj2pSl+lAYDxokdk6PdUd/jZ5s2
+LXXqIhT0aMQMA4wDAYDVR0TAQH/BAIwADAKBggqhkjOPQQDAgNHADBEAiAxmm8w
Ag5DuK1V5vjCqZeuXWyVrfoL3rbMHdrpsYVqSAIgeY9F6wqMEpqIPjAtbCkSC+DG
f6eiTREm5FRLzNkfhxQ=
-----END CERTIFICATE-----
`,
			)

			cert, err := tglib.ParseCertificate(certData)
			if err != nil {
				panic(err)
			}

			r, _ := http.NewRequest(http.MethodGet, "http://127.0.0.1", nil)
			r.TLS = &tls.ConnectionState{
				PeerCertificates: []*x509.Certificate{cert},
			}
			r2 := r.Clone(context.Background())

			rw.Rewrite(&httputil.ProxyRequest{In: r, Out: r2})

			Convey("Then the response should be correct", func() {
				So(r2.Header.Get("X-TLS-Client-Certificate"), ShouldEqual, strings.ReplaceAll(string(certData), "\n", " "))
			})
		})

		Convey("When I call Rewrite it with a invalid TLS client certificate", func() {

			r, _ := http.NewRequest(http.MethodGet, "http://127.0.0.1", nil)
			r.TLS = &tls.ConnectionState{
				PeerCertificates: []*x509.Certificate{nil},
			}
			r2 := r.Clone(context.Background())

			Convey("Then the response should be correct", func() {
				So(func() { rw.Rewrite(&httputil.ProxyRequest{In: r, Out: r2}) }, ShouldPanicWith, "unable to handle client TLS certificate: nil certificate provided")
			})
		})

		Convey("When I call Rewrite without blocking opentracing", func() {

			r, _ := http.NewRequest(http.MethodGet, "http://127.0.0.1", nil)
			r.TLS = &tls.ConnectionState{}

			r.Header.Set("X-B3-TraceID", "X-B3-TraceID")
			r.Header.Set("X-B3-SpanID", "X-B3-SpanID")
			r.Header.Set("X-B3-ParentSpanID", "X-B3-ParentSpanID")
			r.Header.Set("X-B3-Sampled", "X-B3-Sampled")
			r.Header.Set("Uber-Trace-ID", "Uber-Trace-ID")
			r.Header.Set("Jaeger-Baggage", "Jaeger-Baggage")
			r.Header.Set("TraceParent", "TraceParent")
			r.Header.Set("TraceState", "TraceState")

			r2 := r.Clone(context.Background())

			rw.Rewrite(&httputil.ProxyRequest{In: r, Out: r2})

			Convey("Then the response should be correct", func() {
				So(r2.Header.Get("X-B3-TraceID"), ShouldEqual, "X-B3-TraceID")
				So(r2.Header.Get("X-B3-SpanID"), ShouldEqual, "X-B3-SpanID")
				So(r2.Header.Get("X-B3-ParentSpanID"), ShouldEqual, "X-B3-ParentSpanID")
				So(r2.Header.Get("X-B3-Sampled"), ShouldEqual, "X-B3-Sampled")
				So(r2.Header.Get("Uber-Trace-ID"), ShouldEqual, "Uber-Trace-ID")
				So(r2.Header.Get("Jaeger-Baggage"), ShouldEqual, "Jaeger-Baggage")
				So(r2.Header.Get("TraceParent"), ShouldEqual, "TraceParent")
				So(r2.Header.Get("TraceState"), ShouldEqual, "TraceState")

			})
		})

		Convey("When I call Rewrite blocking opentracing", func() {

			rw.blockOpenTracing = true

			r, _ := http.NewRequest(http.MethodGet, "http://127.0.0.1", nil)
			r.TLS = &tls.ConnectionState{}

			r.Header.Set("X-B3-TraceID", "X-B3-TraceID")
			r.Header.Set("X-B3-SpanID", "X-B3-SpanID")
			r.Header.Set("X-B3-ParentSpanID", "X-B3-ParentSpanID")
			r.Header.Set("X-B3-Sampled", "X-B3-Sampled")
			r.Header.Set("Uber-Trace-ID", "Uber-Trace-ID")
			r.Header.Set("Jaeger-Baggage", "Jaeger-Baggage")
			r.Header.Set("TraceParent", "TraceParent")
			r.Header.Set("TraceState", "TraceState")

			r2 := r.Clone(context.Background())

			rw.Rewrite(&httputil.ProxyRequest{In: r, Out: r2})

			Convey("Then the response should be correct", func() {
				So(r2.Header.Get("X-B3-TraceID"), ShouldEqual, "")
				So(r2.Header.Get("X-B3-SpanID"), ShouldEqual, "")
				So(r2.Header.Get("X-B3-ParentSpanID"), ShouldEqual, "")
				So(r2.Header.Get("X-B3-Sampled"), ShouldEqual, "")
				So(r2.Header.Get("Uber-Trace-ID"), ShouldEqual, "")
				So(r2.Header.Get("Jaeger-Baggage"), ShouldEqual, "")
				So(r2.Header.Get("TraceParent"), ShouldEqual, "")
				So(r2.Header.Get("TraceState"), ShouldEqual, "")

			})
		})

		Convey("When I call Rewrite with a custom rewriter that works", func() {

			rw.customRewriter = func(resp *httputil.ProxyRequest, private bool) error {
				resp.Out.Header.Set("yey", "ouais")
				return nil
			}

			r, _ := http.NewRequest(http.MethodGet, "http://127.0.0.1", nil)
			r.TLS = &tls.ConnectionState{}
			r2 := r.Clone(context.Background())

			rw.Rewrite(&httputil.ProxyRequest{In: r, Out: r2})

			Convey("Then the response should be correct", func() {
				So(r2.Header.Get("yey"), ShouldEqual, "ouais")
			})
		})

		Convey("When I call Rewrite with a custom rewriter that returns an error", func() {

			rw.customRewriter = func(resp *httputil.ProxyRequest, private bool) error {
				resp.Out.Header.Set("yey", "ouais")
				return fmt.Errorf("boom")
			}

			r, _ := http.NewRequest(http.MethodGet, "http://127.0.0.1", nil)
			r.TLS = &tls.ConnectionState{}
			r2 := r.Clone(context.Background())

			Convey("Then the response should be correct", func() {
				So(func() { rw.Rewrite(&httputil.ProxyRequest{In: r, Out: r2}) }, ShouldPanicWith, "unable to rewrite request with custom rewriter: boom")
			})
		})
	})
}

func Test_circuitBreakerHandler(t *testing.T) {

	Convey("Given I have a circuitBreakerHandler", t, func() {
		cb := &circuitBreakerHandler{}

		Convey("When I call ServeHTTP", func() {

			recorder := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodGet, "http://127.0.0.1", nil)
			cb.ServeHTTP(recorder, req)

			Convey("Then the response should be correct", func() {
				data, _ := io.ReadAll(recorder.Body)
				So(string(data), ShouldEqual, `[{"code":503,"description":"The service is busy handling requests. Please retry in a moment","subject":"gateway","title":"Service Unavailable"}]`)
			})
		})
	})
}
