package bahamut

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

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

// A Mock represents a mocked action that you can install to run integration test with the bahamut server.
type Mock struct {
	// The operation to mock. Must be one of "retrieve", "retrieve-many", "create", "update", "delete", "info", "patch"
	Operation elemental.Operation `json:"operation"`

	// The name of the indentity you want to mock the response for.
	IdentityName string `json:"identity"`

	// Javascript function to execute in place for the processor.
	// The code MUST contain the at least the process function.
	// This function must returns an object that is the status code
	// and the json of the response.
	//
	//      function process(ctx) {
	//          return {code: 200, body: json.Stringify({name: "mocked name"})}
	//      }
	//
	// If the code cannot compile, the mock will not be installed. If you
	// have runtime error in your code, they will be visible only during execution.
	Function string `json:"function"`

	// If set to true, the processor will panic causing an EOF.
	// If panic is set, the Code is not executed.
	Panic bool `json:"panic"`

	// If set, the output (either panic or code return) will be delayed
	// by the given duration.
	Delay string `json:"delay"`

	vm *otto.Otto
}

func (m *Mock) execute(ctx *Context) (mockAction, error) {

	if m.Delay != "" {
		d, err := time.ParseDuration(m.Delay)
		if err != nil {
			return mockActionDone, fmt.Errorf("mock: unable to parse duration: %s", err)
		}
		time.Sleep(d)
	}

	if m.Panic {
		return mockActionDone, errMockPanicRequested{}
	}

	if m.vm == nil {
		m.vm = otto.New()
		if _, err := m.vm.Run(m.Function); err != nil {
			return mockActionDone, fmt.Errorf("mock: unable to parse function: %s", err)
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
	if codeValue.IsUndefined() {
		return mockActionDone, errors.New("mock: function returned undefined for 'code'")
	}
	if codeValue.IsString() {
		return mockActionDone, fmt.Errorf("mock: function returned a string for 'code': %s", codeValue.String())
	}
	code, err := codeValue.ToInteger()
	if err != nil {
		return mockActionDone, err
	}

	bodyValue, err := out.Get("data")
	if err != nil {
		return mockActionDone, err
	}
	if bodyValue.IsUndefined() {
		return mockActionDone, errors.New("mock: function returned undefined for 'data'")
	}
	if !bodyValue.IsString() {
		return mockActionDone, fmt.Errorf("mock: function returned a non string for 'data': %s", bodyValue.String())
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
		return mockActionDone, fmt.Errorf("mock: unable to decode provided data: %s", err)
	}

	ctx.StatusCode = int(code)
	ctx.OutputData = data

	return mockActionDone, nil
}
