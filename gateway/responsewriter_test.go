package gateway

import (
	"bufio"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

type testResponseWritterHijacker struct {
	http.ResponseWriter
}

func (rw *testResponseWritterHijacker) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return nil, nil, fmt.Errorf("hello!")
}

func TestResponseWritter(t *testing.T) {

	Convey("Given I have a http.ResponseWriter and a responseWriter", t, func() {

		rw := httptest.NewRecorder()
		w := newResponseWriter(rw)

		Convey("Then WriteHeader should ", func() {
			w.WriteHeader(12)
			So(w.code, ShouldEqual, 12)
		})

		Convey("Then Write should ", func() {
			w.Write([]byte("hello"))
			So(rw.Body.Bytes(), ShouldResemble, []byte("hello"))
		})

		Convey("Then Hijack should work", func() {
			_, _, err := w.Hijack()
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, `the responseWriter doesn't support the Hijacker interface`)
		})

		Convey("Then FLush should work", func() {
			w.Flush()
			So(rw.Flushed, ShouldBeTrue)
		})
	})

	Convey("Given I have a testResponseWritterHijacker and a responseWriter", t, func() {

		rw := &testResponseWritterHijacker{}
		w := newResponseWriter(rw)

		Convey("Then Hijack should work", func() {
			_, _, err := w.Hijack()
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, `hello!`)
		})
	})
}
