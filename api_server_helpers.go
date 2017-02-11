package bahamut

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/Sirupsen/logrus"
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
	w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, Cache-Control, If-Modified-Since, X-Requested-With, X-Count-Total, X-Namespace")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
}

func writeHTTPError(w http.ResponseWriter, origin string, err error) {

	var outError elemental.Errors

	switch e := err.(type) {
	case elemental.Error:
		outError = elemental.NewErrors(e)
	case elemental.Errors:
		outError = e
	default:
		outError = elemental.NewErrors(elemental.NewError("Internal Server Error", e.Error(), "bahamut", http.StatusInternalServerError))
	}

	setCommonHeader(w, origin)
	w.WriteHeader(outError.Code())

	if e := json.NewEncoder(w).Encode(&outError); e != nil {
		log.WithFields(logrus.Fields{
			"error":         e.Error(),
			"originalError": err.Error(),
		}).Error("Unable to encode error.")
	}
}

func corsHandler(w http.ResponseWriter, r *http.Request) {
	setCommonHeader(w, r.Header.Get("Origin"))
	w.WriteHeader(http.StatusOK)
}

func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	writeHTTPError(w, r.Header.Get("Origin"), elemental.NewError("Not Found", "Unable to find the requested resource", "bahamut", http.StatusNotFound))
}

func writeHTTPResponse(w http.ResponseWriter, c *Context) {

	buffer := &bytes.Buffer{}

	if c.Redirect != "" {
		w.Header().Set("Location", c.Redirect)
		w.WriteHeader(http.StatusFound)
		io.Copy(w, buffer)
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
			writeHTTPError(w, c.Request.Headers.Get("Origin"), err)
		}
	}

	w.WriteHeader(c.StatusCode)

	if buffer != nil {
		if _, err := io.Copy(w, buffer); err != nil {
			writeHTTPError(w, c.Request.Headers.Get("Origin"), err)
		}
	}
}
