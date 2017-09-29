package bahamut

import (
	"net/http"

	"github.com/aporeto-inc/elemental"
)

// MockAuditer mock the Auditer interface
type MockAuditer struct {
	Hits int
}

// Audit make the Auditer's job
func (p *MockAuditer) Audit(*Context, error) {
	p.Hits++
}

// MockAuth mocks an RequestAuthenticator and Authorizer
type MockAuth struct {
	DefinedHasError             bool
	ExpectedAuthenticatedResult bool
	ExpectedAuthorizedResult    bool
	ExpectedError               error
}

// AuthenticateRequest authenticate a given request
func (a *MockAuth) AuthenticateRequest(req *elemental.Request, ch elemental.ClaimsHolder) (bool, error) {

	if a.DefinedHasError {
		if a.ExpectedError == nil {
			a.ExpectedError = elemental.NewError("Error", "This is an error.", "bahamut-test", http.StatusInternalServerError)
		}
		return false, a.ExpectedError
	}

	return a.ExpectedAuthenticatedResult, nil
}

// IsAuthorized verifies the authentication
func (a *MockAuth) IsAuthorized(ctx *Context) (bool, error) {

	if a.DefinedHasError {
		if a.ExpectedError == nil {
			a.ExpectedError = elemental.NewError("Error", "This is an error.", "bahamut-test", http.StatusInternalServerError)
		}
		return false, a.ExpectedError
	}

	return a.ExpectedAuthorizedResult, nil
}

// MockProcessor is an empty processor
type MockProcessor struct{}

// MockCompleteProcessor defines all processor methods
type MockCompleteProcessor struct {
	ExpectedError error
}

// ProcessRetrieveMany retrieves many object
func (p *MockCompleteProcessor) ProcessRetrieveMany(*Context) error {
	return p.ExpectedError
}

// ProcessRetrieve retrieve a single object
func (p *MockCompleteProcessor) ProcessRetrieve(*Context) error {
	return p.ExpectedError
}

// ProcessCreate create a new object
func (p *MockCompleteProcessor) ProcessCreate(*Context) error {
	return p.ExpectedError
}

// ProcessUpdate update an existing object
func (p *MockCompleteProcessor) ProcessUpdate(*Context) error {
	return p.ExpectedError
}

// ProcessDelete delete an existing object
func (p *MockCompleteProcessor) ProcessDelete(*Context) error {
	return p.ExpectedError
}

// ProcessPatch patch an existing object
func (p *MockCompleteProcessor) ProcessPatch(*Context) error {
	return p.ExpectedError
}

// ProcessInfo returns info
func (p *MockCompleteProcessor) ProcessInfo(*Context) error {
	return p.ExpectedError
}

// // MockPushSessionHandler mocks PushSessionsHandler
// type MockPushSessionHandler struct {
// 	Count int
// 	Hits  int
// 	ShouldAcceptPush        bool
// }
//
// // OnPushSessionStart handles start session
// func (h *MockPushSessionHandler) OnPushSessionStart(session *PushSession) { h.Count++ }
//
// // OnPushSessionStop handles stop session
// func (h *MockPushSessionHandler) OnPushSessionStop(session *PushSession) { h.Count-- }
//
// // IsAuthenticated verify the authentication
// func (h *MockPushSessionHandler) IsAuthenticated(session *PushSession) (bool, error) {
// 	return true, nil
// }
//
// // ShouldPush decides if the event should be pushed
// func (h *MockPushSessionHandler) ShouldPush(session *PushSession, event *elemental.Event) (bool, error) {
// 	h.Hits++
// 	return !h.ShouldAcceptPush, nil
// }
