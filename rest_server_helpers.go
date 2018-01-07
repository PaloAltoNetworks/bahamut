package bahamut

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"go.uber.org/zap"

	"github.com/aporeto-inc/elemental"
)

func setCommonHeader(w http.ResponseWriter, origin string) {

	if origin == "" {
		origin = "*"
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.Header().Set("X-Frame-Options", "SAMEORIGIN")
	w.Header().Set("Cache-control", "private, no-transform")
	w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
	w.Header().Set("Access-Control-Allow-Origin", origin)
	w.Header().Set("Access-Control-Expose-Headers", "X-Requested-With, X-Count-Total, X-Namespace")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, HEAD, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, Cache-Control, If-Modified-Since, X-Requested-With, X-Count-Total, X-Namespace, X-External-Tracking-Type, X-External-Tracking-ID")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
}

func writeHTTPError(w http.ResponseWriter, request *elemental.Request, err error) {

	outError := processError(err, request)

	setCommonHeader(w, request.Headers.Get("Origin"))
	w.WriteHeader(outError.Code())

	if e := json.NewEncoder(w).Encode(&outError); e != nil {
		zap.L().Error("Unable to encode error", zap.Error(e))
	}
}

func corsHandler(w http.ResponseWriter, r *http.Request) {
	setCommonHeader(w, r.Header.Get("Origin"))
	w.WriteHeader(http.StatusOK)
}

func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	writeHTTPError(w, fakeElementalRequest(r), elemental.NewError("Not Found", "Unable to find the requested resource", "bahamut", http.StatusNotFound))
}

func buildNameAndIPsToCertificate(certs []tls.Certificate) map[string]*tls.Certificate {

	out := map[string]*tls.Certificate{}

	for _, cert := range certs {

		x509Cert, err := x509.ParseCertificate(cert.Certificate[0])
		if err != nil {
			continue
		}
		if len(x509Cert.Subject.CommonName) > 0 {
			out[x509Cert.Subject.CommonName] = &cert
		}

		for _, san := range x509Cert.DNSNames {
			out[san] = &cert
		}

		for _, ipsan := range x509Cert.IPAddresses {
			out[ipsan.String()] = &cert
		}
	}

	return out
}

func writeHTTPResponse(w http.ResponseWriter, c *Context) {

	buffer := &bytes.Buffer{}

	if c.Redirect != "" {
		w.Header().Set("Location", c.Redirect)
		w.WriteHeader(http.StatusFound)
		_, _ = io.Copy(w, buffer)
		return
	}

	setCommonHeader(w, c.Request.Headers.Get("Origin"))

	if c.StatusCode == 0 {
		switch c.Request.Operation {
		case elemental.OperationCreate:
			c.StatusCode = http.StatusCreated
		case elemental.OperationInfo:
			c.StatusCode = http.StatusNoContent
		default:
			c.StatusCode = http.StatusOK
		}
	}

	if c.Request.Operation == elemental.OperationRetrieveMany || c.Request.Operation == elemental.OperationInfo {
		w.Header().Set("X-Count-Total", strconv.Itoa(c.CountTotal))
	}

	if c.OutputData != nil {
		if err := json.NewEncoder(buffer).Encode(c.OutputData); err != nil {
			writeHTTPError(w, c.Request, err)
		}
	} else {
		c.StatusCode = http.StatusNoContent
	}

	w.WriteHeader(c.StatusCode)

	if buffer != nil {
		if _, err := io.Copy(w, buffer); err != nil {
			writeHTTPError(w, c.Request, err)
		}
	}
}

func fakeElementalRequest(req *http.Request) *elemental.Request {
	r := elemental.NewRequest()
	r.Headers.Set("Origin", req.Header.Get("Origin"))

	return r
}

func handleEventualPanicHTTP(w http.ResponseWriter, request *elemental.Request, c chan error) {

	if err := handleRecoveredPanic(recover(), request); err != nil {
		c <- err
	}
}

func runHTTPDispatcher(ctx *Context, w http.ResponseWriter, d func() error) {

	e := make(chan error, 1)

	go func() {
		defer handleEventualPanicHTTP(w, ctx.Request, e)
		e <- d()

	}()

	select {
	case <-ctx.Done():
		return
	case err := <-e:
		if err != nil {
			writeHTTPError(w, ctx.Request, err)
			return
		}
		writeHTTPResponse(w, ctx)
	}
}
