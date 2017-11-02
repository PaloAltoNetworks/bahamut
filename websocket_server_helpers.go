package bahamut

import (
	"crypto/tls"
	"net/http"

	"go.uber.org/zap"

	"github.com/aporeto-inc/elemental"
	"golang.org/x/net/websocket"
)

// internalWSSession interface that enhance what a Session can do.
type internalWSSession interface {
	Session
	setRemoteAddress(string)
	setTLSConnectionState(*tls.ConnectionState)
	listen()
	close()
}

type unregisterFunc func(internalWSSession)

func writeWebSocketError(ws *websocket.Conn, response *elemental.Response, err error) {

	if !ws.IsServerConn() {
		return
	}

	outError := processError(err, response.Request)

	response.StatusCode = outError.Code()
	if e := response.Encode(outError); e != nil {
		zap.L().Error("Unable to encode error", zap.Error(err))
		return
	}

	if e := websocket.JSON.Send(ws, response); e != nil {
		zap.L().Error("Unable to send error", zap.Error(err))
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
	} else {
		c.StatusCode = http.StatusNoContent
	}

	response.StatusCode = c.StatusCode

	return websocket.JSON.Send(ws, response)
}
