// Author: Antoine Mercadal
// See LICENSE file for full LICENSE
// Copyright 2016 Aporeto.

package bahamut

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/aporeto-inc/elemental"
)

type wsPushSession struct {
	events            chan *elemental.Event
	filters           chan *elemental.PushFilter
	filter            *elemental.PushFilter
	currentFilterLock *sync.Mutex

	*wsSession
}

func newWSPushSession(request *http.Request, config Config, unregister unregisterFunc) *wsPushSession {

	return &wsPushSession{
		wsSession:         newWSSession(request, config, unregister),
		events:            make(chan *elemental.Event),
		filters:           make(chan *elemental.PushFilter),
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

	return fmt.Sprintf("<pushsession id:%s parameters:%v>",
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

		if err := s.conn.ReadJSON(&filter); err != nil {
			s.stop()
			return
		}

		s.filters <- filter
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

			if err := s.conn.WriteJSON(event); err != nil {
				s.stop()
				return
			}

		case <-s.closeCh:
			return
		}
	}
}

func (s *wsPushSession) listen() {

	go s.read()
	go s.write()

	for {
		select {
		case filter := <-s.filters:
			s.setCurrentFilter(filter)

		case <-s.closeCh:
			return
		}
	}
}
