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

	"github.com/aporeto-inc/addedeffect/wsc"
	"github.com/aporeto-inc/elemental"
	"github.com/gorilla/websocket"

	uuid "github.com/satori/go.uuid"
)

type unregisterFunc func(*wsPushSession)

type wsPushSession struct {
	events             chan *elemental.Event
	filters            chan *elemental.PushFilter
	filter             *elemental.PushFilter
	currentFilterLock  *sync.Mutex
	claims             []string
	claimsMap          map[string]string
	config             Config
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
	closeLock          *sync.Mutex
}

func newWSPushSession(request *http.Request, config Config, unregister unregisterFunc) *wsPushSession {

	id := uuid.Must(uuid.NewV4()).String()
	ctx, cancel := context.WithCancel(request.Context())

	return &wsPushSession{
		events:             make(chan *elemental.Event),
		filters:            make(chan *elemental.PushFilter),
		currentFilterLock:  &sync.Mutex{},
		id:                 id,
		claims:             []string{},
		claimsMap:          map[string]string{},
		config:             config,
		headers:            request.Header,
		parameters:         request.URL.Query(),
		startTime:          time.Now(),
		closeCh:            make(chan struct{}),
		unregister:         unregister,
		ctx:                ctx,
		cancel:             cancel,
		tlsConnectionState: request.TLS,
		remoteAddr:         request.RemoteAddr,
		closeLock:          &sync.Mutex{},
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

	return fmt.Sprintf("<pushsession id:%s parameters:%v>",
		s.id,
		s.parameters,
	)
}

// SetClaims implements elemental.ClaimsHolder.
func (s *wsPushSession) SetClaims(claims []string) {

	s.claims = claims
	s.claimsMap = claimsToMap(claims)
}

func (s *wsPushSession) Identifier() string                            { return s.id }
func (s *wsPushSession) GetClaims() []string                           { return s.claims }
func (s *wsPushSession) GetClaimsMap() map[string]string               { return s.claimsMap }
func (s *wsPushSession) GetToken() string                              { return s.parameters.Get("token") }
func (s *wsPushSession) GetContext() context.Context                   { return s.ctx }
func (s *wsPushSession) TLSConnectionState() *tls.ConnectionState      { return s.tlsConnectionState }
func (s *wsPushSession) GetMetadata() interface{}                      { return s.metadata }
func (s *wsPushSession) SetMetadata(m interface{})                     { s.metadata = m }
func (s *wsPushSession) GetParameter(key string) string                { return s.parameters.Get(key) }
func (s *wsPushSession) setRemoteAddress(addr string)                  { s.remoteAddr = addr }
func (s *wsPushSession) setConn(conn wsc.Websocket)                    { s.conn = conn }
func (s *wsPushSession) close(code int)                                { s.conn.Close(code) }
func (s *wsPushSession) setTLSConnectionState(st *tls.ConnectionState) { s.tlsConnectionState = st }

func (s *wsPushSession) currentFilter() *elemental.PushFilter {

	s.currentFilterLock.Lock()
	defer s.currentFilterLock.Unlock()

	if s.filter == nil {
		return nil
	}

	return s.filter.Duplicate()
}

func (s *wsPushSession) setCurrentFilter(f *elemental.PushFilter) {

	s.currentFilterLock.Lock()
	s.filter = f
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

		case <-s.conn.Done():
			return

		case <-s.ctx.Done():
			s.close(websocket.CloseGoingAway)
			return
		}
	}
}
