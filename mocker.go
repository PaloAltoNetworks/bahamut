package bahamut

import (
	"fmt"
	"sync"

	"github.com/robertkrimen/otto"

	"github.com/aporeto-inc/elemental"
)

type registryContent map[elemental.Operation]map[string]*Mock

type mocker struct {
	registry registryContent
	sync.Mutex
}

func newMocker() *mocker {
	return &mocker{
		registry: registryContent{
			elemental.OperationRetrieve:     map[string]*Mock{},
			elemental.OperationRetrieveMany: map[string]*Mock{},
			elemental.OperationCreate:       map[string]*Mock{},
			elemental.OperationUpdate:       map[string]*Mock{},
			elemental.OperationDelete:       map[string]*Mock{},
			elemental.OperationInfo:         map[string]*Mock{},
			elemental.OperationPatch:        map[string]*Mock{},
		},
	}
}

func (r *mocker) installMock(m *Mock) error {

	if m.Function != "" {
		vm := otto.New()
		if _, err := vm.Eval(m.Function); err != nil {
			return fmt.Errorf("invalid function: %s", err)
		}
	}

	if err := checkOperation(string(m.Operation)); err != nil {
		return err
	}

	if m.IdentityName == "" {
		return fmt.Errorf("invalid empty identity name")
	}

	r.Lock()
	defer r.Unlock()

	r.registry[m.Operation][m.IdentityName] = m

	return nil
}

func (r *mocker) uninstallMock(op elemental.Operation, identityName string) error {

	if err := checkOperation(string(op)); err != nil {
		return err
	}

	if identityName == "" {
		return fmt.Errorf("invalid empty identity name")
	}

	r.Lock()
	defer r.Unlock()

	if _, ok := r.registry[op][identityName]; !ok {
		return fmt.Errorf("no mock installed for operation '%s' and identity '%s'", op, identityName)
	}

	delete(r.registry[op], identityName)

	return nil
}

func (r *mocker) get(op elemental.Operation, identityName string) *Mock {

	r.Lock()
	defer r.Unlock()

	return r.registry[op][identityName]
}

func checkOperation(op string) error {
	switch op {
	case "":
		return fmt.Errorf("invalid empty operation")
	case "create", "update", "delete", "retrieve-many", "retrieve", "patch", "info":
		return nil
	}

	return fmt.Errorf("invalid operation: '%s'. Must be one of 'create', 'update', 'delete', 'retrieve-many', 'retrieve', 'patch' or 'info'", op)
}
