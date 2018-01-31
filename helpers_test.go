package bahamut

import (
	"reflect"
	"sync"
)

type counter struct {
	count int
	sync.Mutex
}

func (c *counter) Add(i int) {
	c.Lock()
	defer c.Unlock()
	c.count += i
}

func (c *counter) Value() int {
	c.Lock()
	defer c.Unlock()
	return c.count
}

type stopper struct {
	stopped bool
	done    chan struct{}

	sync.Mutex
}

func newStopper() *stopper {
	return &stopper{
		done: make(chan struct{}, 1),
	}
}

func (s *stopper) Stop() {
	s.Lock()
	defer s.Unlock()

	select {
	case s.done <- struct{}{}:
	default:
	}

	s.stopped = true
}

func (s *stopper) isStopped() bool {
	s.Lock()
	defer s.Unlock()
	return s.stopped
}

func (s *stopper) Done() chan struct{} {
	return s.done
}

type mockWebsocket struct {
	writeErr error
	closeErr error
	outData  chan interface{}
	inData   chan interface{}

	sync.Mutex
}

func newMockWebsocket() *mockWebsocket {
	return &mockWebsocket{
		outData: make(chan interface{}),
		inData:  make(chan interface{}),
	}
}

func (s *mockWebsocket) setWriteErr(err error) {
	s.Lock()
	defer s.Unlock()
	s.writeErr = err
}

func (s *mockWebsocket) setCloseErr(err error) {
	s.Lock()
	defer s.Unlock()
	s.closeErr = err
}

func (s *mockWebsocket) setNextRead(i interface{}) {
	go func() { s.inData <- i }()
}

func (s *mockWebsocket) getLastWrite() <-chan interface{} {
	return s.outData
}

func (s *mockWebsocket) ReadJSON(data interface{}) error {

	s.Lock()
	defer s.Unlock()

	d := <-s.inData

	if err, ok := d.(error); ok {
		return err
	}

	reflect.ValueOf(data).Elem().Set(reflect.ValueOf(d))

	return nil
}

func (s *mockWebsocket) WriteJSON(data interface{}) error {

	s.Lock()
	defer s.Unlock()

	if s.writeErr != nil {
		return s.writeErr
	}

	s.outData <- data

	return nil
}

func (s *mockWebsocket) Close() error {

	s.Lock()
	defer s.Unlock()

	return s.closeErr
}
