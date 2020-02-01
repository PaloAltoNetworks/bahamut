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

const (
	// supportsErrorEventsQueryParam contains the name of the query parameter that can be passed in by the client to declare that
	// it can handle error events
	supportsErrorEventsQueryParam = "supportsErrors"
)

type unregisterFunc func(*wsPushSession)

type wsPushSession struct {
	dataCh                chan []byte
	pushConfig            *elemental.PushConfig
	currentPushConfigLock sync.RWMutex
	parametersLock        sync.RWMutex
	errorStateActive      bool
	errorStateLock        sync.RWMutex
	claims                []string
	claimsMap             map[string]string
	cfg                   config
	headers               http.Header
	id                    string
	metadata              interface{}
	parameters            url.Values
	remoteAddr            string
	conn                  wsc.Websocket
	startTime             time.Time
	unregister            unregisterFunc
	tlsConnectionState    *tls.ConnectionState
	ctx                   context.Context
	cancel                context.CancelFunc
	closeCh               chan struct{}
	encodingRead          elemental.EncodingType
	encodingWrite         elemental.EncodingType
	cookies               []*http.Cookie
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
		dataCh:             make(chan []byte, 64),
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

		f := s.currentPushConfig()
		if f != nil && f.IsFilteredOut(event.Identity, event.Type) {
			continue
		}

		// We convert the inner Entity to the requested encoding. We don't need additional
		// check as elemental.Convert will do anything if the EncodingTypes are identical.
		if err := event.Convert(s.encodingWrite); err != nil {
			zap.L().Error("Unable to convert event",
				zap.Stringer("event", event),
				zap.Error(err),
			)
			continue
		}

		data, err := elemental.Encode(s.encodingWrite, event)
		if err != nil {
			zap.L().Error("Unable to encode event",
				zap.Stringer("event", event),
				zap.Error(err),
			)
			continue
		}

		s.send(data)
	}
}

func (s *wsPushSession) String() string {

	return fmt.Sprintf("<pushsession id:%s>", s.id)
}

// SetClaims implements elemental.ClaimsHolder.
func (s *wsPushSession) SetClaims(claims []string) {

	s.claims = append([]string{}, claims...)
	s.claimsMap = claimsToMap(s.claims)
}

func (s *wsPushSession) ClaimsMap() map[string]string {

	copiedClaimsMap := map[string]string{}

	for k, v := range s.claimsMap {
		copiedClaimsMap[k] = v
	}

	return copiedClaimsMap
}

func (s *wsPushSession) Identifier() string                            { return s.id }
func (s *wsPushSession) Claims() []string                              { return append([]string{}, s.claims...) }
func (s *wsPushSession) Token() string                                 { return s.Parameter("token") }
func (s *wsPushSession) Context() context.Context                      { return s.ctx }
func (s *wsPushSession) TLSConnectionState() *tls.ConnectionState      { return s.tlsConnectionState }
func (s *wsPushSession) Metadata() interface{}                         { return s.metadata }
func (s *wsPushSession) SetMetadata(m interface{})                     { s.metadata = m }
func (s *wsPushSession) ClientIP() string                              { return s.remoteAddr }
func (s *wsPushSession) setRemoteAddress(addr string)                  { s.remoteAddr = addr }
func (s *wsPushSession) setConn(conn wsc.Websocket)                    { s.conn = conn }
func (s *wsPushSession) close(code int)                                { s.conn.Close(code) }
func (s *wsPushSession) setTLSConnectionState(st *tls.ConnectionState) { s.tlsConnectionState = st }
func (s *wsPushSession) Header(key string) string                      { return s.headers.Get(key) }
func (s *wsPushSession) PushConfig() *elemental.PushConfig             { return s.currentPushConfig() }
func (s *wsPushSession) Parameter(key string) string {
	s.parametersLock.RLock()
	defer s.parametersLock.RUnlock()
	return s.parameters.Get(key)
}

func (s *wsPushSession) inErrorState() bool {
	s.errorStateLock.RLock()
	defer s.errorStateLock.RUnlock()

	return s.errorStateActive
}

func (s *wsPushSession) setErrorState(on bool) {
	s.errorStateLock.RLock()
	defer s.errorStateLock.RUnlock()

	s.errorStateActive = on
}

func (s *wsPushSession) handlesErrorEvents() bool {
	_, ok := s.parameters[supportsErrorEventsQueryParam]
	return ok
}

func (s *wsPushSession) currentPushConfig() *elemental.PushConfig {
	s.currentPushConfigLock.RLock()
	defer s.currentPushConfigLock.RUnlock()

	if s.pushConfig == nil {
		return nil
	}

	return s.pushConfig.Duplicate()
}

func (s *wsPushSession) setCurrentPushConfig(f *elemental.PushConfig) {

	s.currentPushConfigLock.Lock()
	defer s.currentPushConfigLock.Unlock()

	s.pushConfig = f
	if f == nil {
		return
	}

	s.parametersLock.Lock()
	for k, v := range f.Parameters() {
		s.parameters[k] = v
	}
	s.parametersLock.Unlock()
}

func (s *wsPushSession) Cookie(name string) (*http.Cookie, error) {
	for _, cookie := range s.cookies {
		if cookie.Name == name {
			return cookie, nil
		}
	}
	return nil, http.ErrNoCookie
}

// send sends the given bytes as is, with no
// additional checks.
func (s *wsPushSession) send(data []byte) {

	select {
	case s.dataCh <- data:
	default:
		zap.L().Warn("Slow consumer. event dropped",
			zap.String("sessionID", s.id),
			zap.Strings("claims", s.claims),
		)
	}
}

func (s *wsPushSession) listen() {

	defer s.unregister(s)

	for {
		select {
		case data := <-s.dataCh:

			s.conn.Write(data)

		case data := <-s.conn.Read():

			pushConfig := elemental.NewPushConfig()
			if err := elemental.Decode(s.encodingRead, data, pushConfig); err != nil {
				if s.handlesErrorEvents() {
					s.setErrorState(true)
					msgpack, json, err := prepareEventData(elemental.NewErrorEvent(elemental.Error{
						Description: fmt.Sprintf("could not decode message into %T: %s", pushConfig, err),
						Subject:     "bahamut",
						Title:       "Bad request",
					}, s.encodingWrite))

					if err != nil {
						zap.L().Error("unable to prepare error event encoding, closing socket", zap.Error(err))
						s.close(websocket.CloseInternalServerErr)
						return
					}

					switch s.encodingWrite {
					case elemental.EncodingTypeMSGPACK:
						s.send(msgpack)
					case elemental.EncodingTypeJSON:
						s.send(json)
					}

					continue
				} else {
					s.close(websocket.CloseUnsupportedData)
					return
				}
			}

			if err := pushConfig.ParseIdentityFilters(); err != nil {
				zap.L().Debug("error parsing filter(s) in the received *elemental.PushConfig",
					zap.Error(err),
					zap.String("sessionID", s.id),
					zap.String("pushConfig", pushConfig.String()),
				)

				if s.handlesErrorEvents() {
					s.setErrorState(true)
					msgpack, json, err := prepareEventData(elemental.NewErrorEvent(elemental.Error{
						Title:       "Bad request",
						Subject:     "bahamut",
						Description: fmt.Sprintf("unable to parse identity filters: %s", err),
						Data: map[string]interface{}{
							"attribute": "filters",
						},
					}, s.encodingWrite))

					if err != nil {
						zap.L().Error("unable to prepare error event encoding, closing socket", zap.Error(err))
						s.close(websocket.CloseInternalServerErr)
						return
					}

					switch s.encodingWrite {
					case elemental.EncodingTypeMSGPACK:
						s.send(msgpack)
					case elemental.EncodingTypeJSON:
						s.send(json)
					}

					continue
				} else {
					s.close(websocket.CloseUnsupportedData)
					return
				}
			}

			s.setErrorState(false)
			s.setCurrentPushConfig(pushConfig)

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
