package bahamut

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/aporeto-inc/elemental"

	log "github.com/Sirupsen/logrus"
)

func setCommonHeader(w http.ResponseWriter, origin string) {

	if origin == "" {
		origin = "*"
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.Header().Set("Access-Control-Allow-Origin", origin)
	w.Header().Set("Access-Control-Expose-Headers", "X-Requested-With, X-Count-Local, X-Count-Total, X-PageCurrent, X-Page-Size, X-Page-Prev, X-Page-Next, X-Page-First, X-Page-Last, X-Namespace")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, HEAD, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, Cache-Control, If-Modified-Since, X-Requested-With, X-Count-Local, X-Count-Total, X-PageCurrent, X-Page-Size, X-Page-Prev, X-Page-Next, X-Page-First, X-Page-Last, X-Namespace")
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
		log.WithFields(log.Fields{
			"package":       "bahamut",
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

	setCommonHeader(w, c.Request.Headers.Get("Origin"))

	buffer := &bytes.Buffer{}

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

		c.Page.compute(c.Count.Total)

		w.Header().Set("X-Page-Current", strconv.Itoa(c.Page.Current))
		w.Header().Set("X-Page-Size", strconv.Itoa(c.Page.Size))

		w.Header().Set("X-Page-First", strconv.Itoa(c.Page.First))
		w.Header().Set("X-Page-Last", strconv.Itoa(c.Page.Last))
		w.Header().Set("X-Page-Prev", strconv.Itoa(c.Page.Prev))
		w.Header().Set("X-Page-Next", strconv.Itoa(c.Page.Next))
		w.Header().Set("X-Count-Local", strconv.Itoa(c.Count.Current))
		w.Header().Set("X-Count-Total", strconv.Itoa(c.Count.Total))
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
