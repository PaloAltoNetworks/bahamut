package bahamut

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/aporeto-inc/elemental"
	"github.com/robertkrimen/otto"
)

type mock struct {
	Operation    elemental.Operation `json:"operation"`
	IdentityName string              `json:"identity"`
	Code         string              `json:"code"`

	vm *otto.Otto
}

type mockAction int

const (
	mockActionDone mockAction = iota
	mockActionContinue
	mockActionEOF
)

func (m *mock) execute(ctx *Context) (mockAction, error) {

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

type registryContent map[elemental.Operation]map[string]*mock

type mocker struct {
	registry registryContent
	sync.Mutex
}

func newMocker() *mocker {
	return &mocker{
		registry: registryContent{
			elemental.OperationRetrieve:     map[string]*mock{},
			elemental.OperationRetrieveMany: map[string]*mock{},
			elemental.OperationCreate:       map[string]*mock{},
			elemental.OperationUpdate:       map[string]*mock{},
			elemental.OperationDelete:       map[string]*mock{},
			elemental.OperationInfo:         map[string]*mock{},
			elemental.OperationPatch:        map[string]*mock{},
		},
	}
}

func (r *mocker) installMock(m *mock) {

	r.Lock()
	defer r.Unlock()

	r.registry[m.Operation][m.IdentityName] = m
}

func (r *mocker) uninstallMock(op elemental.Operation, identityName string) bool {

	r.Lock()
	defer r.Unlock()

	if _, ok := r.registry[op][identityName]; !ok {
		return false
	}

	delete(r.registry[op], identityName)

	return true
}

func (r *mocker) get(op elemental.Operation, identityName string) *mock {

	r.Lock()
	defer r.Unlock()

	return r.registry[op][identityName]
}
