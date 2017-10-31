package bahamut

import (
	"crypto/tls"
	"net/http"
	"net/url"
	"time"

	uuid "github.com/satori/go.uuid"

	"golang.org/x/net/websocket"
)

type wsSession struct {
	claims             []string
	claimsMap          map[string]string
	config             Config
	headers            http.Header
	id                 string
	parameters         url.Values
	remoteAddr         string
	socket             *websocket.Conn
	startTime          time.Time
	stopAll            chan bool
	stopRead           chan bool
	stopWrite          chan bool
	unregister         unregisterFunc
	tlsConnectionState *tls.ConnectionState
}

func newWSSession(ws *websocket.Conn, config Config, unregister unregisterFunc) *wsSession {

	var parameters url.Values
	if request := ws.Request(); request != nil {
		parameters = request.URL.Query()
	}

	var headers http.Header
	if config := ws.Config(); config != nil {
		headers = config.Header
	}

	return &wsSession{
		claims:     []string{},
		claimsMap:  map[string]string{},
		config:     config,
		headers:    headers,
		id:         uuid.NewV4().String(),
		parameters: parameters,
		socket:     ws,
		startTime:  time.Now(),
		stopAll:    make(chan bool, 2),
		stopRead:   make(chan bool, 2),
		stopWrite:  make(chan bool, 2),
		unregister: unregister,
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

// GetClaims implements elemental.ClaimsHolder.
func (s *wsSession) GetClaims() []string { return s.claims }

// GetClaimsMap implements elemental.ClaimsHolder.
func (s *wsSession) GetClaimsMap() map[string]string { return s.claimsMap }

// GetToken implements elemental.TokenHolder.
func (s *wsSession) GetToken() string { return s.parameters.Get("token") }

// TLSConnectionState implements elemental.TokenHolder.
func (s *wsSession) TLSConnectionState() *tls.ConnectionState { return s.tlsConnectionState }

// GetParameter implements the WebSocketSession interface.
func (s *wsSession) GetParameter(key string) string {
	return s.parameters.Get(key)
}

// setRemoteAddress implements the internalWSSession interface.
func (s *wsSession) setRemoteAddress(addr string) {
	s.remoteAddr = addr
}

// setTLSConnectionState implements internalWSSession.
func (s *wsSession) setTLSConnectionState(tlsConnectionState *tls.ConnectionState) {
	s.tlsConnectionState = tlsConnectionState
}

// Close implements the internalWSSession interface.
func (s *wsSession) close() {
	s.stopAll <- true
}

func (s *wsSession) stop() {

	s.stopRead <- true
	s.stopWrite <- true

	s.unregister(s)
	s.socket.Close() // nolint: errcheck
}

func (s *wsSession) listen() {}
