package gateway

import (
	"fmt"
	"net"
	"sync"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

type fakeConn struct {
	closed bool
}

func (c *fakeConn) Read(b []byte) (n int, err error) { return 0, nil }

func (c *fakeConn) Write(b []byte) (n int, err error) { return 0, nil }

func (c *fakeConn) Close() error { c.closed = true; return nil }

func (c *fakeConn) LocalAddr() net.Addr { return nil }

func (c *fakeConn) RemoteAddr() net.Addr { return nil }

func (c *fakeConn) SetDeadline(t time.Time) error { return nil }

func (c *fakeConn) SetReadDeadline(t time.Time) error { return nil }

func (c *fakeConn) SetWriteDeadline(t time.Time) error { return nil }

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

type fakeListenerLimiterMetricManager struct {
	accepted int
	rejected int
	total    int
	sync.Mutex
}

func (m *fakeListenerLimiterMetricManager) RegisterAcceptedConnection() {
	m.Lock()
	m.total = m.total + 1
	m.accepted = m.accepted + 1
	m.Unlock()
}

func (m *fakeListenerLimiterMetricManager) RegisterLimitedConnection() {
	m.Lock()
	m.total = m.total + 1
	m.rejected = m.rejected + 1
	m.Unlock()
}

func TestLimitLimiter(t *testing.T) {

	Convey("Given I call newLimitedListener", t, func() {

		l := &fakeListener{
			conn: func() net.Conn { return &fakeConn{} },
		}

		mm := &fakeListenerLimiterMetricManager{}

		ll := newLimitedListener(l, 2.0, 1, mm)

		Convey("When I call Accept and it works", func() {

			c, err := ll.Accept()

			Convey("Then err should be nil", func() {
				So(err, ShouldBeNil)
			})

			Convey("Then c should be correct", func() {
				So(c, ShouldNotBeNil)
				So(mm.total, ShouldBeGreaterThanOrEqualTo, 1)
				So(mm.accepted, ShouldBeGreaterThanOrEqualTo, 1)
				So(mm.rejected, ShouldEqual, 0)
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
				So(mm.total, ShouldBeGreaterThanOrEqualTo, 1)
				So(mm.accepted, ShouldBeGreaterThanOrEqualTo, 1)
				So(mm.rejected, ShouldBeGreaterThanOrEqualTo, 1)
			})
		})
	})
}
