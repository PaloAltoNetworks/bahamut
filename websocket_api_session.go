// Author: Antoine Mercadal
// See LICENSE file for full LICENSE
// Copyright 2016 Aporeto.

package bahamut

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/aporeto-inc/elemental"
)

type wsAPISession struct {
	processorFinder processorFinderFunc
	pusherFunc      eventPusherFunc
	requests        chan *elemental.Request
	responses       chan *elemental.Response
	*wsSession
}

func newWSAPISession(request *http.Request, config Config, unregister unregisterFunc, processorFinder processorFinderFunc, pusherFunc eventPusherFunc) *wsAPISession {

	return &wsAPISession{
		wsSession:       newWSSession(request, config, unregister),
		processorFinder: processorFinder,
		pusherFunc:      pusherFunc,
		requests:        make(chan *elemental.Request),
		responses:       make(chan *elemental.Response),
	}
}

func (s *wsAPISession) String() string {

	return fmt.Sprintf("<apisession id:%s parameters: %v>",
		s.id,
		s.parameters,
	)
}

func (s *wsAPISession) read() {

	for {
		request := elemental.NewRequest()
		request.ClientIP = s.remoteAddr

		if err := s.conn.ReadJSON(request); err != nil {
			if _, ok := err.(*json.SyntaxError); !ok {
				s.stop()
				return
			}

			response := elemental.NewResponse()

			s.responses <- makeErrorResponse(
				s.context,
				response,
				elemental.NewError(
					"Bad Request",
					"Invalid JSON",
					"bahamut",
					http.StatusBadRequest,
				),
			)
		}

		s.requests <- request
	}
}

func (s *wsAPISession) write() {

	for {
		select {
		case resp := <-s.responses:

			if err := s.conn.WriteJSON(resp); err != nil {
				s.stop()
				return
			}

		case <-s.closeCh:
			return
		}
	}
}

func (s *wsAPISession) listen() {

	go s.read()
	go s.write()

	// TODO: this is here for backward compat.
	// we should remvove this when all enforcers
	// are switched to at least manipulate 2.x
	s.responses <- &elemental.Response{
		StatusCode: http.StatusOK,
	}

	for {
		select {
		case request := <-s.requests:

			ctx := traceRequest(s.context, request)
			bctx := NewContextWithRequest(request)
			bctx.ctx = ctx

			// We backport the token of the session into the request if we don't have an explicit one given in the request.
			if request.Password == "" {
				if t := s.GetToken(); t != "" {
					request.Username = "Bearer"
					request.Password = t
				}
			}

			// And we set the TLSConnectionState
			request.TLSConnectionState = s.TLSConnectionState()

			switch request.Operation {

			case elemental.OperationRetrieveMany:
				s.responses <- handleRetrieveMany(bctx, s.config, s.processorFinder, s.pusherFunc)

			case elemental.OperationRetrieve:
				s.responses <- handleRetrieve(bctx, s.config, s.processorFinder, s.pusherFunc)

			case elemental.OperationCreate:
				s.responses <- handleCreate(bctx, s.config, s.processorFinder, s.pusherFunc)

			case elemental.OperationUpdate:
				s.responses <- handleUpdate(bctx, s.config, s.processorFinder, s.pusherFunc)

			case elemental.OperationDelete:
				s.responses <- handleDelete(bctx, s.config, s.processorFinder, s.pusherFunc)

			case elemental.OperationInfo:
				s.responses <- handleInfo(bctx, s.config, s.processorFinder, s.pusherFunc)

			case elemental.OperationPatch:
				s.responses <- handlePatch(bctx, s.config, s.processorFinder, s.pusherFunc)
			}
			finishTracing(ctx)

		case <-s.closeCh:
			return
		}
	}
}
