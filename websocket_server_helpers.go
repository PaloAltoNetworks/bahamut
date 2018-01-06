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

func writeWebSocketError(response *elemental.Response, err error) *elemental.Response {

	outError := processError(err, response.Request)

	response.StatusCode = outError.Code()
	if e := response.Encode(outError); e != nil {
		zap.L().Panic("Unable to encode error", zap.Error(err))
	}

	return response
}

func writeWebsocketResponse(response *elemental.Response, c *Context) *elemental.Response {

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
			zap.L().Panic("Unable to encode output data", zap.Error(err))
		}
	} else {
		c.StatusCode = http.StatusNoContent
	}

	response.StatusCode = c.StatusCode

	return response
}

func runWSDispatcher(ctx *Context, s *websocket.Conn, r *elemental.Response, d func() error) *elemental.Response {

	e := make(chan error, 1)

	go func() {
		e <- d()
	}()

	select {

	case <-ctx.Done():
		return nil

	case err := <-e:

		if err != nil {
			return writeWebSocketError(r, err)
		}

		return writeWebsocketResponse(r, ctx)
	}
}
