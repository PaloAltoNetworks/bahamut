package gateway

import (
	"bufio"
	"errors"
	"net"
	"net/http"
)

type responseWriter interface {
	http.ResponseWriter
	http.Flusher
	status() int
}

type responseWriterWrapper struct {
	http.ResponseWriter
	code int
}

func newResponseWriter(rw http.ResponseWriter) responseWriter {
	nrw := &responseWriterWrapper{
		ResponseWriter: rw,
	}

	if _, ok := rw.(http.CloseNotifier); ok {
		return &responseWriterCloseNotifer{nrw}
	}

	return nrw
}

func (rw *responseWriterWrapper) status() int {
	return rw.code
}

func (rw *responseWriterWrapper) WriteHeader(s int) {
	rw.code = s
	rw.ResponseWriter.WriteHeader(s)
}

func (rw *responseWriterWrapper) Write(b []byte) (int, error) {
	return rw.ResponseWriter.Write(b)
}

func (rw *responseWriterWrapper) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hijacker, ok := rw.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, errors.New("the responseWriter doesn't support the Hijacker interface")
	}
	return hijacker.Hijack()
}

func (rw *responseWriterWrapper) Flush() {
	if flusher, ok := rw.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

type responseWriterCloseNotifer struct {
	*responseWriterWrapper
}

func (rw *responseWriterCloseNotifer) CloseNotify() <-chan bool {
	return rw.ResponseWriter.(http.CloseNotifier).CloseNotify()
}
