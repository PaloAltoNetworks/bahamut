// Author: Antoine Mercadal
// See LICENSE file for full LICENSE
// Copyright 2016 Aporeto.

package bahamut

import (
	"fmt"
	"sync"

	"github.com/aporeto-inc/elemental"
	"golang.org/x/net/websocket"
)

type wsPushSession struct {
	events            chan *elemental.Event
	filters           chan *elemental.PushFilter
	filter            *elemental.PushFilter
	currentFilterLock *sync.Mutex

	*wsSession
}

func newWSPushSession(ws *websocket.Conn, config Config, unregister unregisterFunc) *wsPushSession {

	return &wsPushSession{
		wsSession:         newWSSession(ws, config, unregister),
		events:            make(chan *elemental.Event),
		filters:           make(chan *elemental.PushFilter, 8),
		currentFilterLock: &sync.Mutex{},
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

	return fmt.Sprintf("<pushsession id:%s parameters: %v>",
		s.id,
		s.parameters,
	)
}

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

func (s *wsPushSession) read() {

	for {
		var filter *elemental.PushFilter

		if err := websocket.JSON.Receive(s.socket, &filter); err != nil {
			s.stopAll <- true
			return
		}

		select {
		case s.filters <- filter:
		case <-s.stopRead:
			return
		}
	}
}

func (s *wsPushSession) write() {

	for {
		select {
		case event := <-s.events:

			f := s.currentFilter()
			if f != nil && f.IsFilteredOut(event.Identity, event.Type) {
				break
			}

			if err := websocket.JSON.Send(s.socket, event); err != nil {
				s.stopAll <- true
				return
			}

		case <-s.stopWrite:
			return
		}
	}
}

func (s *wsPushSession) stop() {

	s.stopRead <- true
	s.stopWrite <- true

	s.unregister(s)
	s.socket.Close() // nolint: errcheck
}

func (s *wsPushSession) listen() {

	go s.read()
	go s.write()
	defer s.stop()

	for {
		select {
		case filter := <-s.filters:
			s.setCurrentFilter(filter)

		case <-s.stopAll:
			return
		}
	}
}
