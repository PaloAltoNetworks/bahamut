// Author: Antoine Mercadal
// See LICENSE file for full LICENSE
// Copyright 2016 Aporeto.

package bahamut

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/aporeto-inc/elemental"
)

func setCommonHeader(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Expose-Headers", "X-Requested-With, X-Count-Local, X-Count-Total, X-PageCurrent, X-Page-Size, X-Page-Prev, X-Page-Next, X-Page-First, X-Page-Last")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, HEAD, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, Cache-Control, If-Modified-Since, X-Requested-With, X-Count-Local, X-Count-Total, X-PageCurrent, X-Page-Size, X-Page-Prev, X-Page-Next, X-Page-First, X-Page-Last")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
}

// WriteHTTPError write a Error into a http.ResponseWriter.
func WriteHTTPError(w http.ResponseWriter, code int, errs ...*elemental.Error) {
	setCommonHeader(w)
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(errs)
}

// Context contains information about a current operation
type Context struct {
	Info        *Info
	Page        *Page
	Count       *Count
	InputData   interface{}
	OutputData  interface{}
	Errors      elemental.Errors
	StatusCode  int
	Operation   Operation
	EventsQueue elemental.EventsList
}

// NewContext creates a new *Context for the given Operation.
func NewContext(operation Operation) *Context {

	return &Context{
		Info:        NewInfo(),
		Errors:      elemental.Errors{},
		Page:        NewPage(),
		Count:       NewCount(),
		Operation:   operation,
		EventsQueue: elemental.EventsList{},
	}
}

// ReadRequest reads information from the given http.Request and populate the Context.
func (c *Context) ReadRequest(req *http.Request) error {

	c.Info.FromRequest(req)
	c.Page.FromValues(req.URL.Query())

	return nil
}

// EnqueueEvents enqueues the given event to the Context.
func (c *Context) EnqueueEvents(events ...*elemental.Event) {

	c.EventsQueue = append(c.EventsQueue, events...)
}

// AddErrors inserts a new Error in the Context.
func (c *Context) AddErrors(err ...*elemental.Error) {

	c.Errors = append(c.Errors, err...)
}

// HasErrors returns true if the context has some errors.
func (c *Context) HasErrors() bool {

	return len(c.Errors) > 0
}

// WriteResponse writes the final response to the given http.ResponseWriter.
func (c *Context) WriteResponse(w http.ResponseWriter) error {

	setCommonHeader(w)

	buffer := &bytes.Buffer{}

	if c.HasErrors() {

		if c.StatusCode == 0 {
			c.StatusCode = http.StatusInternalServerError
		}

		if err := json.NewEncoder(buffer).Encode(c.Errors); err != nil {
			return err
		}

	} else {

		if c.StatusCode == 0 {
			if c.Operation == OperationCreate {
				c.StatusCode = http.StatusCreated
			} else {
				c.StatusCode = http.StatusOK
			}
		}

		if c.Operation == OperationRetrieveMany {

			c.Page.compute(c.Info.BaseRawURL, c.Info.Parameters, c.Count.Total)

			w.Header().Set("X-Page-Current", strconv.Itoa(c.Page.Current))
			w.Header().Set("X-Page-Size", strconv.Itoa(c.Page.Size))

			w.Header().Set("X-Page-First", c.Page.First)
			w.Header().Set("X-Page-Last", c.Page.Last)

			if pageLink := c.Page.Prev; pageLink != "" {
				w.Header().Set("X-Page-Prev", pageLink)
			}

			if pageLink := c.Page.Next; pageLink != "" {
				w.Header().Set("X-Page-Next", pageLink)
			}

			w.Header().Set("X-Count-Local", strconv.Itoa(c.Count.Current))
			w.Header().Set("X-Count-Total", strconv.Itoa(c.Count.Total))
		}

		if c.OutputData != nil {
			if err := json.NewEncoder(buffer).Encode(c.OutputData); err != nil {
				return err
			}
		}
	}

	w.WriteHeader(c.StatusCode)

	var err error
	if buffer != nil {
		_, err = io.Copy(w, buffer)
	}

	return err
}
