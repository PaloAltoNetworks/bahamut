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
	conn               internalWSConn
	startTime          time.Time
	unregister         unregisterFunc
	tlsConnectionState *tls.ConnectionState
	span               opentracing.Span
	context            context.Context
	cancel             context.CancelFunc
	closeCh            chan struct{}
	closeLock          *sync.Mutex
}

func newWSSession(request *http.Request, config Config, unregister unregisterFunc, span opentracing.Span) *wsSession {

	id := uuid.Must(uuid.NewV4()).String()
	span.SetTag("bahamut.session.id", id)

	ctx, cancel := context.WithCancel(request.Context())

	return &wsSession{
		claims:             []string{},
		claimsMap:          map[string]string{},
		config:             config,
		headers:            request.Header,
		id:                 id,
		parameters:         request.URL.Query(),
		startTime:          time.Now(),
		closeCh:            make(chan struct{}),
		unregister:         unregister,
		span:               span,
		context:            ctx,
		cancel:             cancel,
		tlsConnectionState: request.TLS,
		remoteAddr:         request.RemoteAddr,
		closeLock:          &sync.Mutex{},
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

func (s *wsSession) GetClaims() []string {

	return s.claims
}

func (s *wsSession) GetClaimsMap() map[string]string {

	return s.claimsMap
}

func (s *wsSession) GetToken() string {

	return s.parameters.Get("token")
}

func (s *wsSession) GetContext() context.Context {

	return s.context
}

func (s *wsSession) TLSConnectionState() *tls.ConnectionState {

	return s.tlsConnectionState
}

func (s *wsSession) GetMetadata() interface{} {

	return s.metadata
}

func (s *wsSession) SetMetadata(m interface{}) {

	s.metadata = m
}

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

func (s *wsSession) setConn(conn internalWSConn) {

	s.conn = conn
}

func (s *wsSession) stop() {

	s.closeLock.Lock()
	defer s.closeLock.Unlock()

	if s.closeCh == nil {
		return
	}

	close(s.closeCh)
	s.closeCh = nil

	s.cancel()

	s.unregister(s)

	if s.conn != nil {
		s.conn.Close() // nolint: errcheck
	}
}

func (s *wsSession) listen() {}
