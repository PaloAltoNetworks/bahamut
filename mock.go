package bahamut

import (
	"encoding/json"
	"fmt"

	"github.com/aporeto-inc/elemental"
	"github.com/robertkrimen/otto"
)

type mockAction int

const (
	mockActionDone mockAction = iota
	mockActionContinue
)

type errMockPanicRequested struct{}

func (e errMockPanicRequested) Error() string { return "Panic requested by mock" }

// A Mock represents a mocked action that you can install
// to run integration test with the bahamut server.
type Mock struct {
	Operation    elemental.Operation `json:"operation"`
	IdentityName string              `json:"identity"`
	Code         string              `json:"code"`
	Panic        bool                `json:"panic"`

	vm *otto.Otto
}

func (m *Mock) execute(ctx *Context) (mockAction, error) {

	if m.Panic {
		return mockActionDone, errMockPanicRequested{}
	}

	if m.vm == nil {
		m.vm = otto.New()
		if _, err := m.vm.Run(m.Code); err != nil {
			return mockActionDone, fmt.Errorf("mock: unable to parse code: %s", err)
		}
	}

	v, err := m.vm.Call(`process`, nil, ctx)
	if err != nil {
		return mockActionDone, fmt.Errorf("mock: unable to call 'process': %s", err)
	}

	out := v.Object()
	if out == nil {
		return mockActionContinue, nil
	}

	codeValue, err := out.Get("code")
	if err != nil {
		return mockActionDone, err
	}
	code, err := codeValue.ToInteger()
	if err != nil {
		return mockActionDone, err
	}

	bodyValue, err := out.Get("data")
	if err != nil {
		return mockActionDone, err
	}
	body, err := bodyValue.ToString()
	if err != nil {
		return mockActionDone, err
	}

	var data interface{}
	if ctx.Request.Operation == elemental.OperationRetrieveMany {
		data = []map[string]interface{}{}
	} else {
		data = map[string]interface{}{}
	}

	if err := json.Unmarshal([]byte(body), &data); err != nil {
		return mockActionDone, err
	}

	ctx.StatusCode = int(code)
	ctx.OutputData = data

	return mockActionDone, nil
}
