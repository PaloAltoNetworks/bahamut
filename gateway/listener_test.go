package gateway

import (
	"fmt"
	"net"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

type fakeConn struct {
	closed bool
}

func (c *fakeConn) Read(_ []byte) (n int, err error) { return 0, nil }

func (c *fakeConn) Write(_ []byte) (n int, err error) { return 0, nil }

func (c *fakeConn) Close() error { c.closed = true; return nil }

func (c *fakeConn) LocalAddr() net.Addr { return nil }

func (c *fakeConn) RemoteAddr() net.Addr { return nil }

func (c *fakeConn) SetDeadline(_ time.Time) error { return nil }

func (c *fakeConn) SetReadDeadline(_ time.Time) error { return nil }

func (c *fakeConn) SetWriteDeadline(_ time.Time) error { return nil }

type fakeListener struct {
	conn        func() net.Conn
	acceptError error
}

func (l *fakeListener) Accept() (net.Conn, error) {

	if l.acceptError != nil {
		return nil, l.acceptError
	}

	return l.conn(), nil
}

func (l *fakeListener) Addr() net.Addr {
	return nil
}

func (l *fakeListener) Close() error {
	return nil
}

func TestLimitLimiter(t *testing.T) {

	Convey("Given I call newLimitedListener", t, func() {

		l := &fakeListener{
			conn: func() net.Conn { return &fakeConn{} },
		}

		ll := newLimitedListener(l, 2.0, 1)

		Convey("When I call Accept and it works", func() {

			c, err := ll.Accept()

			Convey("Then err should be nil", func() {
				So(err, ShouldBeNil)
			})

			Convey("Then c should be correct", func() {
				So(c, ShouldNotBeNil)
			})
		})

		Convey("When I call Accept but underlying listener is returning an error", func() {

			l.acceptError = fmt.Errorf("boom")

			c, err := ll.Accept()

			Convey("Then err should be nil", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "boom")
			})

			Convey("Then c should be correct", func() {
				So(c, ShouldBeNil)
			})
		})

		Convey("When I spam Accept I should get rate limited", func() {

			// send a bunch of Accept to excite the rate limiter
			go func() { _, _ = ll.Accept() }()
			go func() { _, _ = ll.Accept() }()
			go func() { _, _ = ll.Accept() }()
			go func() { _, _ = ll.Accept() }()
			go func() { _, _ = ll.Accept() }()
			go func() { _, _ = ll.Accept() }()
			go func() { _, _ = ll.Accept() }()
			go func() { _, _ = ll.Accept() }()
			go func() { _, _ = ll.Accept() }()

			time.Sleep(300 * time.Millisecond)

			// this one should be closed because rate limited
			conn, _ := ll.Accept()

			Convey("Then err should be nil", func() {
				So(conn.(*fakeConn).closed, ShouldBeTrue)
			})
		})
	})
}
