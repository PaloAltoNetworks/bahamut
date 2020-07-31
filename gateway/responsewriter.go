package gateway

import (
	"bufio"
	"errors"
	"net"
	"net/http"
)

type responseWriter struct {
	http.ResponseWriter
	code int
}

func newResponseWriter(rw http.ResponseWriter) *responseWriter {
	nrw := &responseWriter{
		ResponseWriter: rw,
	}

	return nrw
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.code = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	return rw.ResponseWriter.Write(b)
}

func (rw *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hijacker, ok := rw.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, errors.New("the responseWriter doesn't support the Hijacker interface")
	}
	return hijacker.Hijack()
}

func (rw *responseWriter) Flush() {
	if flusher, ok := rw.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}
