// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bahamut

import (
	"net"
	"sync/atomic"
)

type negociation struct {
	conn net.Conn
	err  error
}

type limitListener struct {
	net.Listener
	connCh        chan negociation
	metricManager MetricsManager
	nConn         int64
	maxConn       int64
}

// NewListener returns a Listener that uses the given semaphore to accept at most
// n simultaneous connections from the provided Listener where n is the size of
// the given channel.
func NewListener(l net.Listener, n int, metricManager MetricsManager) *limitListener {

	return &limitListener{
		Listener:      l,
		maxConn:       int64(n),
		metricManager: metricManager,
	}
}

func (l *limitListener) release() {

	atomic.AddInt64(&l.nConn, -1)

	if l.metricManager != nil {
		l.metricManager.UnregisterTCPConnection()
	}
}

func (l *limitListener) Accept() (net.Conn, error) {

	for {

		c, err := l.Listener.Accept()
		if err != nil {
			return nil, err
		}

		var new int64
		if l.maxConn > 0 {
			new = atomic.AddInt64(&l.nConn, 1)
		}

		if l.metricManager != nil {
			l.metricManager.RegisterTCPConnection()
		}

		if new > l.maxConn {
			c.Close()
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
