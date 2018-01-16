package bahamut

import (
	"crypto/tls"
	"net/http"

	"github.com/aporeto-inc/elemental"
	"go.uber.org/zap"
)

// internalWSSession interface that enhance what a Session can do.
type internalWSSession interface {
	Session
	setRemoteAddress(string)
	setTLSConnectionState(*tls.ConnectionState)
	setConn(internalWSConn)
	listen()
	close()
}

type internalWSConn interface {
	ReadJSON(interface{}) error
	WriteJSON(interface{}) error
	Close() error
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
	response.Messages = c.messages()

	return response
}

func handleEventualPanicWebsocket(response *elemental.Response, c chan error, reco bool) {

	if err := handleRecoveredPanic(recover(), response.Request, reco); err != nil {
		c <- err
	}
}

func runWSDispatcher(ctx *Context, r *elemental.Response, d func() error, recover bool) *elemental.Response {

	e := make(chan error, 1)

	go func() {
		defer handleEventualPanicWebsocket(r, e, recover)
		e <- d()
	}()

	select {

	case <-ctx.Done():
		return nil

	case err := <-e:
		if err != nil {
			if _, ok := err.(errMockPanicRequested); ok {
				panic(err.Error())
			}
			return writeWebSocketError(r, err)
		}

		return writeWebsocketResponse(r, ctx)
	}
}
