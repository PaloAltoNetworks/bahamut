// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bahamut

import (
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

const defaultMaxOpenFiles = 256

func TestLimitListener(t *testing.T) {
	const max = 5

	attempts := (defaultMaxOpenFiles - max) / 2
	if attempts > 256 { // maximum length of accept queue is 128 by default
		attempts = 256
	}

	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close() // nolint
	l = newListener(l, max)

	var open int32
	// nolint
	go http.Serve(l, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if n := atomic.AddInt32(&open, 1); n > max {
			t.Errorf("%d open connections, want <= %d", n, max)
		}
		defer atomic.AddInt32(&open, -1)
		time.Sleep(10 * time.Millisecond)
		fmt.Fprint(w, "some body")
	}))

	var wg sync.WaitGroup
	var failed int32
	for i := 0; i < attempts; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c := http.Client{Timeout: 3 * time.Second}
			r, err := c.Get("http://" + l.Addr().String())
			if err != nil {
				if err == io.EOF {
					t.Log(err)
					atomic.AddInt32(&failed, 1)
				}
				return
			}
			defer r.Body.Close()        // nolint
			io.Copy(io.Discard, r.Body) // nolint
		}()
	}
	wg.Wait()

	// We expect some Gets to fail as the kernel's accept queue is filled,
	// but most should succeed.
	if int(failed) >= attempts/2 {
		t.Errorf("%d requests failed within %d attempts", failed, attempts)
	}
}

type errorListener struct {
	net.Listener
}

func (errorListener) Accept() (net.Conn, error) {
	return nil, errFake
}

var errFake = errors.New("fake error from errorListener")

// This used to hang.
func TestLimitListenerError(t *testing.T) {
	donec := make(chan bool, 1)

	go func() {
		const n = 2
		ll := newListener(errorListener{}, 2)
		for i := 0; i < n+1; i++ {
			_, err := ll.Accept()
			if err != errFake {
				panic(fmt.Sprintf("Accept error = %v; want errFake", err))
			}
		}
		donec <- true
	}()
	select {
	case <-donec:
	case <-time.After(5 * time.Second):
		t.Fatal("timeout. deadlock?")
	}
}

func TestLimitListenerClose(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close() // nolint
	ln = newListener(ln, 1)

	doneCh := make(chan struct{})
	defer close(doneCh)
	go func() {
		c, err := net.Dial("tcp", ln.Addr().String())
		if err != nil {
			panic(err)
		}
		defer c.Close() // nolint
		<-doneCh
	}()

	c, err := ln.Accept()
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close() // nolint

	acceptDone := make(chan struct{})
	go func() {
		c, err := ln.Accept()
		if err == nil {
			c.Close() // nolint
			t.Errorf("Unexpected successful Accept()")
		}
		close(acceptDone)
	}()

	// Wait a tiny bit to ensure the Accept() is blocking.
	time.Sleep(10 * time.Millisecond)
	ln.Close() // nolint

	select {
	case <-acceptDone:
	case <-time.After(5 * time.Second):
		t.Fatalf("Accept() still blocking")
	}
}
