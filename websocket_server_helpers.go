package bahamut

import (
	"crypto/tls"
	"net/http"

	"github.com/aporeto-inc/elemental"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

// internalWSSession interface that enhance what a Session can do.
type internalWSSession interface {
	Session
	setRemoteAddress(string)
	setTLSConnectionState(*tls.ConnectionState)
	setSocket(*websocket.Conn)
	listen()
	close()
}

type unregisterFunc func(internalWSSession)

func writeWebSocketError(ws *websocket.Conn, response *elemental.Response, err error) {

	outError := processError(err, response.Request)

	response.StatusCode = outError.Code()
	if e := response.Encode(outError); e != nil {
		zap.L().Error("Unable to encode error", zap.Error(err))
		return
	}

	if e := ws.WriteJSON(response); e != nil {
		zap.L().Error("Unable to send error", zap.Error(err))
	}
}

func writeWebsocketResponse(ws *websocket.Conn, response *elemental.Response, c *Context) error {

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

	return ws.WriteJSON(response)
}

func runWSDispatcher(ctx *Context, s *websocket.Conn, r *elemental.Response, d func() error) {

	e := make(chan error, 1)

	go func() {
		e <- d()
	}()

	select {
	case <-ctx.Done():
		return
	case err := <-e:
		if err != nil {
			writeWebSocketError(s, r, err)
			return
		}

		if err = writeWebsocketResponse(s, r, ctx); err != nil {
			writeWebSocketError(s, r, err)
		}
	}
}
