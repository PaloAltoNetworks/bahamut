package bahamut

import (
	"sync"

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

func (r *mocker) installMock(m *Mock) {

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

func (r *mocker) get(op elemental.Operation, identityName string) *Mock {

	r.Lock()
	defer r.Unlock()

	return r.registry[op][identityName]
}
