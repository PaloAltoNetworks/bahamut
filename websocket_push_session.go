// Author: Antoine Mercadal
// See LICENSE file for full LICENSE
// Copyright 2016 Aporeto.

package bahamut

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/aporeto-inc/elemental"

	opentracing "github.com/opentracing/opentracing-go"
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
		wsSession:         newWSSession(request, config, unregister, opentracing.StartSpan("bahamut.session.push")),
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

		if err := s.conn.ReadJSON(&filter); err != nil {
			s.close()
			return
		}

		select {
		case s.filters <- filter:
		case <-s.closeCh:
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

			if err := s.conn.WriteJSON(event); err != nil {
				s.close()
				return
			}

		case <-s.closeCh:
			return
		}
	}
}

// while this function is the same for wsAPISession and wsPushSession
// it has to be written in both of the struc instead of wsSession as
// if would call s.unregister using *wsSession and not a *wsPushSession
func (s *wsPushSession) stop() {

	s.close()
	s.unregister(s)
	s.conn.Close() // nolint: errcheck
}

func (s *wsPushSession) listen() {

	go s.read()
	go s.write()
	defer s.stop()

	for {
		select {
		case filter := <-s.filters:
			s.setCurrentFilter(filter)

		case <-s.closeCh:
			return
		}
	}
}
