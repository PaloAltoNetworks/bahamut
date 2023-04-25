package bahamut

import (
	"context"
	"crypto/tls"
	"net/http"

	"go.aporeto.io/elemental"
)

var _ PushSession = &MockSession{}

// A MockSession can be used to mock a bahamut.Session.
type MockSession struct {
	MockMetadata           any
	MockClaimsMap          map[string]string
	MockCookies            map[string]*http.Cookie
	MockHeaders            map[string]string
	MockParameters         map[string]string
	MockPushConfig         *elemental.PushConfig
	MockTLSConnectionState *tls.ConnectionState
	MockDirectPush         func(...*elemental.Event)
	MockClientIP           string
	MockIdentifier         string
	MockToken              string
	MockClaims             []string
}

// NewMockSession returns a new MockSession.
func NewMockSession() *MockSession {
	return &MockSession{
		MockClaimsMap:  map[string]string{},
		MockCookies:    map[string]*http.Cookie{},
		MockHeaders:    map[string]string{},
		MockParameters: map[string]string{},
	}
}

// Cookie is part of the Session interface.
func (s *MockSession) Cookie(c string) (*http.Cookie, error) {

	v, ok := s.MockCookies[c]
	if !ok {
		return nil, http.ErrNoCookie
	}

	return v, nil
}

// DirectPush is part of the PushSession interface
func (s *MockSession) DirectPush(evts ...*elemental.Event) {
	if s.MockDirectPush != nil {
		s.MockDirectPush(evts...)
	}
}

// Identifier is part of the PushSession interface.
func (s *MockSession) Identifier() string { return s.MockIdentifier }

// Parameter is part of the PushSession interface.
func (s *MockSession) Parameter(k string) string { return s.MockParameters[k] }

// Header is part of the PushSession interface.
func (s *MockSession) Header(k string) string { return s.MockHeaders[k] }

// PushConfig is part of the PushSession interface.
func (s *MockSession) PushConfig() *elemental.PushConfig { return s.MockPushConfig }

// SetClaims is part of the PushSession interface.
func (s *MockSession) SetClaims(claims []string) { s.MockClaims = claims }

// Claims is part of the PushSession interface.
func (s *MockSession) Claims() []string { return s.MockClaims }

// ClaimsMap is part of the PushSession interface.
func (s *MockSession) ClaimsMap() map[string]string { return s.MockClaimsMap }

// Token is part of the PushSession interface.
func (s *MockSession) Token() string { return s.MockToken }

// TLSConnectionState is part of the PushSession interface.
func (s *MockSession) TLSConnectionState() *tls.ConnectionState { return s.MockTLSConnectionState }

// Metadata is part of the PushSession interface.
func (s *MockSession) Metadata() any { return s.MockMetadata }

// SetMetadata is part of the PushSession interface.
func (s *MockSession) SetMetadata(m any) { s.MockMetadata = m }

// Context is part of the PushSession interface.
func (s *MockSession) Context() context.Context { return context.Background() }

// ClientIP is part of the PushSession interface.
func (s *MockSession) ClientIP() string { return s.MockClientIP }
