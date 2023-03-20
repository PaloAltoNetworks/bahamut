// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bahamut

import (
	"net"
	"sync/atomic"
)

type limitListener struct {
	net.Listener
	nConn   int64
	maxConn int64
}

// newListener returns a Listener that uses the given semaphore to accept at most
// n simultaneous connections from the provided Listener where n is the size of
// the given channel.
func newListener(l net.Listener, n int) *limitListener {

	return &limitListener{
		Listener: l,
		maxConn:  int64(n),
	}
}

func (l *limitListener) release() {

	atomic.AddInt64(&l.nConn, -1)
}

func (l *limitListener) Accept() (net.Conn, error) {

	for {

		c, err := l.Listener.Accept()
		if err != nil {
			return nil, err
		}

		var currentConn int64
		if l.maxConn > 0 {
			currentConn = atomic.AddInt64(&l.nConn, 1)
		}

		if currentConn > l.maxConn {
			c.Close() // nolint: errcheck
			l.release()
			continue
		}

		return &limitListenerConn{Conn: c, release: l.release}, nil
	}
}

func (l *limitListener) Close() error {
	return l.Listener.Close()
}

type limitListenerConn struct {
	net.Conn
	release func()
}

func (c *limitListenerConn) Close() error {
	c.release()
	return c.Conn.Close()
}
