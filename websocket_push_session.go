// Author: Antoine Mercadal
// See LICENSE file for full LICENSE
// Copyright 2016 Aporeto.

package bahamut

import (
	"context"
	"crypto/tls"
	"encoding/json"
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
	filters            chan *elemental.PushFilter
	filter             *elemental.PushFilter
	currentFilterLock  sync.RWMutex
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
}

func newWSPushSession(request *http.Request, cfg config, unregister unregisterFunc) *wsPushSession {

	id := uuid.Must(uuid.NewV4()).String()
	ctx, cancel := context.WithCancel(request.Context())

	return &wsPushSession{
		events:             make(chan *elemental.Event),
		filters:            make(chan *elemental.PushFilter),
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
	}
}

func (s *wsPushSession) DirectPush(events ...*elemental.Event) {

	for _, event := range events {

		if event.Timestamp.Before(s.startTime) {
			continue
		}

		s.events <- event
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

	s.currentFilterLock.RLock()
	defer s.currentFilterLock.RUnlock()

	return s.parameters.Get(key)
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
	s.filter = f
	if s.filter != nil {
		for k, v := range s.filter.Parameters() {
			s.parameters[k] = v
		}
	}
	s.currentFilterLock.Unlock()
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

			data, err := json.Marshal(event)
			if err != nil {
				s.close(websocket.CloseInternalServerErr)
				return
			}

			s.conn.Write(data)

		case data := <-s.conn.Read():

			if err := json.Unmarshal(data, filter); err != nil {
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
