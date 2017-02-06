package bahamut

import (
	"net/http"

	"github.com/Sirupsen/logrus"
	"github.com/aporeto-inc/elemental"
	"golang.org/x/net/websocket"
)

func writeWebSocketError(ws *websocket.Conn, response *elemental.Response, err error) {

	if !ws.IsServerConn() {
		return
	}

	var outError elemental.Errors

	switch e := err.(type) {
	case elemental.Error:
		outError = elemental.NewErrors(e)
	case elemental.Errors:
		outError = e
	default:
		outError = elemental.NewErrors(elemental.NewError("Internal Server Error", e.Error(), "bahamut", http.StatusInternalServerError))
	}

	response.StatusCode = outError.Code()
	response.Encode(outError)

	if e := websocket.JSON.Send(ws, response); e != nil {
		log.WithFields(logrus.Fields{
			"error":         e.Error(),
			"originalError": err.Error(),
		}).Error("Unable to encode error.")
	}
}

func writeWebsocketResponse(ws *websocket.Conn, response *elemental.Response, c *Context) error {

	if !ws.IsServerConn() {
		return nil
	}

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

		response.Total = c.CountTotal
	}

	if c.OutputData != nil {

		if err := response.Encode(c.OutputData); err != nil {
			return err
		}
	}

	response.StatusCode = c.StatusCode

	return websocket.JSON.Send(ws, response)
}
