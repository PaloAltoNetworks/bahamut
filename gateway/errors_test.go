package gateway

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mailgun/multibuf"
	// nolint:revive // Allow dot imports for readability in tests
	. "github.com/smartystreets/goconvey/convey"
	"github.com/vulcand/oxy/ratelimit"
	"github.com/vulcand/oxy/v2/connlimit"
	"go.aporeto.io/elemental"
)

type fakeNetErr struct {
	err       error
	timeout   bool
	temporary bool
}

func (e *fakeNetErr) Error() string   { return e.err.Error() }
func (e *fakeNetErr) Timeout() bool   { return e.timeout }
func (e *fakeNetErr) Temporary() bool { return e.temporary }

func TestErrorHandler(t *testing.T) {

	Convey("Given I create a new error handler", t, func() {

		corsOriginInjectorFunc := func(w http.ResponseWriter, r *http.Request) http.Header {
			w.Header().Set("Access-Control-Allow-Origin", "foo")
			return w.Header()
		}

		eh := &errorHandler{corsOriginInjector: corsOriginInjectorFunc}
		req := httptest.NewRequest(http.MethodGet, "/target", nil)
		w := httptest.NewRecorder()

		Convey("When I call ServeHTTP with no error, nothing should happen", func() {
			eh.ServeHTTP(w, req, nil)
		})

		Convey("When I call ServeHTTP with an io.EOF error", func() {
			eh.ServeHTTP(w, req, io.EOF)
			data, _ := io.ReadAll(w.Body)
			So(string(data), ShouldEqual, `[{"code":502,"description":"The requested service is not available. Please try again in a moment.","subject":"gateway","title":"Bad Gateway"}]`)
			So(w.Header(), ShouldResemble, http.Header{
				"Access-Control-Allow-Origin": {"foo"},
				"Content-Type":                {"application/json; charset=UTF-8"},
			})
		})

		Convey("When I call ServeHTTP with a context.Canceled error", func() {
			eh.ServeHTTP(w, req, context.Canceled)
			data, _ := io.ReadAll(w.Body)
			So(string(data), ShouldEqual, `[{"code":499,"description":"The client closed the connection before it could complete.","subject":"gateway","title":"Client Closed Connection"}]`)
		})

		Convey("When I call ServeHTTP with a random error", func() {
			eh.ServeHTTP(w, req, fmt.Errorf("random error"))
			data, _ := io.ReadAll(w.Body)
			So(string(data), ShouldEqual, `[{"code":500,"description":"random error","subject":"gateway","title":"Internal Server Error"}]`)
		})

		Convey("When I call ServeHTTP with net.Error that is not a timeout", func() {
			ne := &fakeNetErr{
				err: &ratelimit.MaxRateError{},
			}
			eh.ServeHTTP(w, req, ne)
			data, _ := io.ReadAll(w.Body)
			So(string(data), ShouldEqual, `[{"code":502,"description":"The requested service is not available. Please try again in a moment.","subject":"gateway","title":"Bad Gateway"}]`)
			So(w.Header(), ShouldResemble, http.Header{
				"Access-Control-Allow-Origin": {"foo"},
				"Content-Type":                {"application/json; charset=UTF-8"},
			})
		})

		Convey("When I call ServeHTTP with net.Error that is a timeout", func() {
			ne := &fakeNetErr{
				err:     &ratelimit.MaxRateError{},
				timeout: true,
			}
			eh.ServeHTTP(w, req, ne)
			data, _ := io.ReadAll(w.Body)
			So(string(data), ShouldEqual, `[{"code":504,"description":"The requested service took too long to respond. Please try again in a moment.","subject":"gateway","title":"Gateway Timeout"}]`)
			So(w.Header(), ShouldResemble, http.Header{
				"Access-Control-Allow-Origin": {"foo"},
				"Content-Type":                {"application/json; charset=UTF-8"},
			})
		})

		Convey("When I call ServeHTTP with a errRateLimit", func() {
			eh.ServeHTTP(w, req, errTooManyRequest)
			data, _ := io.ReadAll(w.Body)
			_ = data
			So(string(data), ShouldEqual, `[{"code":429,"description":"Please retry in a moment.","subject":"gateway","title":"Too Many Requests"}]`)
			So(w.Header(), ShouldResemble, http.Header{
				"Access-Control-Allow-Origin": {"foo"},
				"Content-Type":                {"application/json; charset=UTF-8"},
			})
		})

		Convey("When I call ServeHTTP with a connlimit.MaxConnError", func() {
			eh.ServeHTTP(w, req, &connlimit.MaxConnError{})
			data, _ := io.ReadAll(w.Body)
			So(string(data), ShouldEqual, `[{"code":429,"description":"Please retry in a moment.","subject":"gateway","title":"Too Many Connections"}]`)
			So(w.Header(), ShouldResemble, http.Header{
				"Access-Control-Allow-Origin": {"foo"},
				"Content-Type":                {"application/json; charset=UTF-8"},
			})
		})

		Convey("When I call ServeHTTP with a multibuf.MaxSizeReachedErrortf", func() {
			eh.ServeHTTP(w, req, &multibuf.MaxSizeReachedError{})
			data, _ := io.ReadAll(w.Body)
			So(string(data), ShouldEqual, `[{"code":413,"description":"Payload size exceeds the maximum allowed size (0 bytes)","subject":"gateway","title":"Entity Too Large"}]`)
		})

		Convey("When I call ServeHTTP with a error.errorString too large returned by MaxBytesReader", func() {
			eh.ServeHTTP(w, req, errors.New("http: request body too large"))
			data, _ := io.ReadAll(w.Body)
			So(string(data), ShouldEqual, `[{"code":413,"description":"http: request body too large","subject":"gateway","title":"Entity Too Large"}]`)
		})

		Convey("When I call ServeHTTP with a x509.UnknownAuthorityError", func() {
			eh.ServeHTTP(w, req, x509.UnknownAuthorityError{})
			data, _ := io.ReadAll(w.Body)
			So(string(data), ShouldEqual, `[{"code":495,"description":"x509: certificate signed by unknown authority","subject":"gateway","title":"TLS Error"}]`)
		})

		Convey("When I call ServeHTTP with a x509.HostnameError", func() {
			eh.ServeHTTP(w, req, x509.HostnameError{
				Host:        "toto",
				Certificate: &x509.Certificate{},
			})
			data, _ := io.ReadAll(w.Body)
			So(string(data), ShouldEqual, `[{"code":495,"description":"x509: certificate is not valid for any names, but wanted to match toto","subject":"gateway","title":"TLS Error"}]`)
		})

		Convey("When I call ServeHTTP with a x509.CertificateInvalidError", func() {
			eh.ServeHTTP(w, req, x509.CertificateInvalidError{})
			data, _ := io.ReadAll(w.Body)
			So(string(data), ShouldEqual, `[{"code":495,"description":"x509: certificate is not authorized to sign other certificates","subject":"gateway","title":"TLS Error"}]`)
		})

		Convey("When I call ServeHTTP with a x509.ConstraintViolationError", func() {
			eh.ServeHTTP(w, req, x509.ConstraintViolationError{})
			data, _ := io.ReadAll(w.Body)
			So(string(data), ShouldEqual, `[{"code":495,"description":"x509: invalid signature: parent certificate cannot sign this kind of certificate","subject":"gateway","title":"TLS Error"}]`)
		})

		Convey("When I call ServeHTTP with a tls.CertificateVerificationError", func() {
			eh.ServeHTTP(w, req, &tls.CertificateVerificationError{})
			data, _ := io.ReadAll(w.Body)
			So(string(data), ShouldStartWith, `[{"code":495,"description":"tls: failed to verify certificate:`)
		})
	})
}

func TestWriteError(t *testing.T) {

	Convey("Given call writeError on request with no encoding and everything is fine", t, func() {

		req := httptest.NewRequest(http.MethodGet, "/target", nil)
		w := httptest.NewRecorder()

		writeError(w, req, elemental.NewError("error", "that's an error", "gateway", http.StatusNotFound))

		data, _ := io.ReadAll(w.Body)
		So(string(data), ShouldEqual, `[{"code":404,"description":"that's an error","subject":"gateway","title":"error"}]`)
	})

	Convey("Given call writeError on request with a bad encoding and everything is fine", t, func() {

		req := httptest.NewRequest(http.MethodGet, "/target", nil)
		req.Header.Add("Accept", "dog")
		w := httptest.NewRecorder()

		writeError(w, req, elemental.NewError("error", "that's an error", "gateway", http.StatusNotFound))

		data, _ := io.ReadAll(w.Body)
		So(string(data), ShouldEqual, `[{"code":404,"description":"that's an error","subject":"gateway","title":"error"}]`)
	})

	Convey("Given call writeError on request with msgpack and everything is fine", t, func() {

		req := httptest.NewRequest(http.MethodGet, "/target", nil)
		req.Header.Add("Accept", "application/msgpack")
		w := httptest.NewRecorder()

		writeError(w, req, elemental.NewError("error", "that's an error", "gateway", http.StatusNotFound))

		data, _ := io.ReadAll(w.Body)
		So(data, ShouldResemble, []byte{145, 134, 164, 99, 111, 100, 101, 209, 1, 148, 164, 100, 97, 116, 97, 192, 171, 100, 101, 115, 99, 114, 105, 112, 116, 105, 111, 110, 175, 116, 104, 97, 116, 39, 115, 32, 97, 110, 32, 101, 114, 114, 111, 114, 167, 115, 117, 98, 106, 101, 99, 116, 167, 103, 97, 116, 101, 119, 97, 121, 165, 116, 105, 116, 108, 101, 165, 101, 114, 114, 111, 114, 165, 116, 114, 97, 99, 101, 160})
	})
}
