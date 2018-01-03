package bahamut

import (
	"context"
	"crypto/tls"
	"net/http"
	"net/url"
	"sync"
	"time"

	opentracing "github.com/opentracing/opentracing-go"
	uuid "github.com/satori/go.uuid"

	"golang.org/x/net/websocket"
)

type wsSession struct {
	claims             []string
	claimsMap          map[string]string
	config             Config
	headers            http.Header
	id                 string
	metadata           interface{}
	parameters         url.Values
	remoteAddr         string
	socket             *websocket.Conn
	startTime          time.Time
	unregister         unregisterFunc
	tlsConnectionState *tls.ConnectionState
	span               opentracing.Span
	context            context.Context
	cancel             context.CancelFunc
	closeCh            chan struct{}
	isClosed           bool
	isClosedLock       sync.Mutex
}

func newWSSession(ws *websocket.Conn, config Config, unregister unregisterFunc, span opentracing.Span) *wsSession {

	id := uuid.Must(uuid.NewV4()).String()
	span.SetTag("bahamut.session.id", id)

	var parameters url.Values
	if request := ws.Request(); request != nil {
		parameters = request.URL.Query()
	}

	var headers http.Header
	if config := ws.Config(); config != nil {
		headers = config.Header
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &wsSession{
		claims:       []string{},
		claimsMap:    map[string]string{},
		config:       config,
		headers:      headers,
		id:           id,
		parameters:   parameters,
		socket:       ws,
		startTime:    time.Now(),
		closeCh:      make(chan struct{}),
		unregister:   unregister,
		span:         span,
		context:      ctx,
		cancel:       cancel,
		isClosedLock: sync.Mutex{},
	}
}

// Identifier returns the identifier of the push session.
// implements the Session interface.
func (s *wsSession) Identifier() string {
	return s.id
}

// SetClaims implements elemental.ClaimsHolder.
func (s *wsSession) SetClaims(claims []string) {
	s.claims = claims
	s.claimsMap = claimsToMap(claims)
}

func (s *wsSession) GetClaims() []string { return s.claims }

func (s *wsSession) GetClaimsMap() map[string]string { return s.claimsMap }

func (s *wsSession) GetToken() string { return s.parameters.Get("token") }

func (s *wsSession) TLSConnectionState() *tls.ConnectionState { return s.tlsConnectionState }

func (s *wsSession) GetMetadata() interface{} { return s.metadata }

func (s *wsSession) SetMetadata(m interface{}) { s.metadata = m }

func (s *wsSession) GetParameter(key string) string {
	return s.parameters.Get(key)
}

func (s *wsSession) Span() opentracing.Span {
	return s.span
}

func (s *wsSession) NewChildSpan(name string) opentracing.Span {

	return opentracing.StartSpan(name, opentracing.ChildOf(s.span.Context()))
}

// setRemoteAddress implements the internalWSSession interface.
func (s *wsSession) setRemoteAddress(addr string) {
	s.remoteAddr = addr
}

// setTLSConnectionState implements internalWSSession.
func (s *wsSession) setTLSConnectionState(tlsConnectionState *tls.ConnectionState) {
	s.tlsConnectionState = tlsConnectionState
}

func (s *wsSession) close() {

	s.isClosedLock.Lock()
	defer s.isClosedLock.Unlock()
	if s.isClosed {
		return
	}

	s.isClosed = true
	close(s.closeCh)
}

func (s *wsSession) listen() {}
