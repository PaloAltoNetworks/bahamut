// Copyright 2019 Aporeto Inc.
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//     http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package bahamut

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/gofrs/uuid"
	"github.com/gorilla/websocket"
	"go.aporeto.io/elemental"
	"go.aporeto.io/wsc"
	"go.uber.org/zap"
)

type unregisterFunc func(*wsPushSession)

type wsPushSession struct {
	events             chan *elemental.Event
	filter             *elemental.PushFilter
	currentFilterLock  sync.RWMutex
	parametersLock     sync.RWMutex
	claims             []string
	claimsMap          map[string]string
	cfg                config
	headers            http.Header
	id                 string
	metadata           interface{}
	parameters         url.Values
	remoteAddr         string
	conn               wsc.Websocket
	startTime          time.Time
	unregister         unregisterFunc
	tlsConnectionState *tls.ConnectionState
	ctx                context.Context
	cancel             context.CancelFunc
	closeCh            chan struct{}
	encodingRead       elemental.EncodingType
	encodingWrite      elemental.EncodingType
}

func newWSPushSession(
	request *http.Request,
	cfg config,
	unregister unregisterFunc,
	encodingRead elemental.EncodingType,
	encodingWrite elemental.EncodingType,
) *wsPushSession {

	id := uuid.Must(uuid.NewV4()).String()
	ctx, cancel := context.WithCancel(request.Context())

	return &wsPushSession{
		events:             make(chan *elemental.Event, 100),
		id:                 id,
		claims:             []string{},
		claimsMap:          map[string]string{},
		cfg:                cfg,
		headers:            request.Header,
		parameters:         request.URL.Query(),
		startTime:          time.Now(),
		closeCh:            make(chan struct{}),
		unregister:         unregister,
		ctx:                ctx,
		cancel:             cancel,
		tlsConnectionState: request.TLS,
		remoteAddr:         request.RemoteAddr,
		encodingRead:       encodingRead,
		encodingWrite:      encodingWrite,
	}
}

func (s *wsPushSession) DirectPush(events ...*elemental.Event) {

	for _, event := range events {

		if event.Timestamp.Before(s.startTime) {
			continue
		}

		select {
		case s.events <- event:
		default:
			zap.L().Warn("Slow consumer. event dropped",
				zap.String("sessionID", s.id),
				zap.Strings("claims", s.claims),
				zap.String("eventType", string(event.Type)),
				zap.String("eventIdentity", event.Identity),
			)
		}
	}
}

func (s *wsPushSession) String() string {

	return fmt.Sprintf("<pushsession id:%s>", s.id)
}

// SetClaims implements elemental.ClaimsHolder.
func (s *wsPushSession) SetClaims(claims []string) {

	s.claims = claims
	s.claimsMap = claimsToMap(claims)
}

func (s *wsPushSession) Identifier() string                            { return s.id }
func (s *wsPushSession) Claims() []string                              { return s.claims }
func (s *wsPushSession) ClaimsMap() map[string]string                  { return s.claimsMap }
func (s *wsPushSession) Token() string                                 { return s.Parameter("token") }
func (s *wsPushSession) Context() context.Context                      { return s.ctx }
func (s *wsPushSession) TLSConnectionState() *tls.ConnectionState      { return s.tlsConnectionState }
func (s *wsPushSession) Metadata() interface{}                         { return s.metadata }
func (s *wsPushSession) SetMetadata(m interface{})                     { s.metadata = m }
func (s *wsPushSession) setRemoteAddress(addr string)                  { s.remoteAddr = addr }
func (s *wsPushSession) setConn(conn wsc.Websocket)                    { s.conn = conn }
func (s *wsPushSession) close(code int)                                { s.conn.Close(code) }
func (s *wsPushSession) setTLSConnectionState(st *tls.ConnectionState) { s.tlsConnectionState = st }

func (s *wsPushSession) Parameter(key string) string {

	s.parametersLock.RLock()
	defer s.parametersLock.RUnlock()

	return s.parameters.Get(key)
}

func (s *wsPushSession) Header(key string) string {

	return s.headers.Get(key)
}

func (s *wsPushSession) currentFilter() *elemental.PushFilter {

	s.currentFilterLock.RLock()
	defer s.currentFilterLock.RUnlock()

	if s.filter == nil {
		return nil
	}

	return s.filter.Duplicate()
}

func (s *wsPushSession) setCurrentFilter(f *elemental.PushFilter) {

	s.currentFilterLock.Lock()
	defer s.currentFilterLock.Unlock()

	s.filter = f
	if s.filter == nil {
		return
	}

	s.parametersLock.Lock()
	defer s.parametersLock.Unlock()

	for k, v := range s.filter.Parameters() {
		s.parameters[k] = v
	}
}

func (s *wsPushSession) listen() {

	filter := elemental.NewPushFilter()

	defer s.unregister(s)

	for {
		select {
		case event := <-s.events:

			f := s.currentFilter()
			if f != nil && f.IsFilteredOut(event.Identity, event.Type) {
				break
			}

			// We convert the inner Entity to the requested encoding. We don't need additional
			// check as elemental.Convert will do anything if the EncodingTypes are identical.
			if err := event.Convert(s.encodingWrite); err != nil {
				zap.L().Error("Unable to convert event", zap.Error(err))
				s.close(websocket.CloseInternalServerErr)
				return
			}

			data, err := elemental.Encode(s.encodingWrite, event)
			if err != nil {
				zap.L().Error("Unable to encode event", zap.Error(err))
				s.close(websocket.CloseInternalServerErr)
				return
			}

			s.conn.Write(data)

		case data := <-s.conn.Read():

			if err := elemental.Decode(s.encodingRead, data, filter); err != nil {
				s.close(websocket.CloseUnsupportedData)
				return
			}

			s.setCurrentFilter(filter)

		case err := <-s.conn.Error():
			zap.L().Error("Error received from websocket", zap.String("session", s.id), zap.Error(err))

		case <-s.conn.Done():
			return

		case <-s.ctx.Done():
			s.close(websocket.CloseGoingAway)
			return
		}
	}
}
