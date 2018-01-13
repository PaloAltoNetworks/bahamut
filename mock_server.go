package bahamut

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/aporeto-inc/elemental"
	"github.com/go-zoo/bone"

	"go.uber.org/zap"
)

var currentMocker *mocker

// an mockServer is the structure serving the health check endpoint.
type mockServer struct {
	config Config
	server *http.Server
}

// newMockServer returns a new mockServer.
func newMockServer(config Config) *mockServer {

	// Install the shared mocker.
	currentMocker = newMocker()

	return &mockServer{
		config: config,
	}
}

// start starts the mockServer.
func (s *mockServer) start() {

	s.server = &http.Server{Addr: s.config.MockServer.ListenAddress}

	mux := bone.New()
	mux.Post("/mock/install", http.HandlerFunc(s.handleInstallMock))
	mux.Delete("/mock/uninstall/:operation/:identity", http.HandlerFunc(s.handleUninstallMock))

	s.server.Handler = mux

	zap.L().Warn("Mock server enabled", zap.String("listen", s.config.MockServer.ListenAddress))

	if err := s.server.ListenAndServe(); err != nil {
		zap.L().Panic("Unable to start mock server", zap.Error(err))
	}
}

func (s *mockServer) handleInstallMock(w http.ResponseWriter, req *http.Request) {

	mock := &Mock{}
	if err := json.NewDecoder(req.Body).Decode(mock); err != nil {
		http.Error(w, fmt.Sprintf("Unable to decode provided mock: %s", err), http.StatusBadRequest)
		return
	}

	if err := checkOperation(string(mock.Operation)); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	currentMocker.installMock(mock)

	zap.L().Info("mock installed",
		zap.String("identity", mock.IdentityName),
		zap.String("operation", string(mock.Operation)),
		zap.String("code", mock.Code),
	)

	w.WriteHeader(http.StatusCreated)
}

func (s *mockServer) handleUninstallMock(w http.ResponseWriter, req *http.Request) {

	op := bone.GetValue(req, "operation")
	if err := checkOperation(op); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	identity := bone.GetValue(req, "identity")
	ok := currentMocker.uninstallMock(elemental.Operation(op), identity)
	if !ok {
		http.Error(w, fmt.Sprintf("No mock installed for operation %s and identity %s", op, identity), http.StatusNotFound)
		return
	}

	zap.L().Info("mock uninstalled",
		zap.String("identity", identity),
		zap.String("operation", op),
	)

	w.WriteHeader(http.StatusOK)
}

// stop stops the mockServer.
func (s *mockServer) stop() {

	// a.server.Shutdown() // Uncomment with Go 1.8
	// a.server = nil
}

func checkOperation(op string) error {
	switch op {
	case "create", "update", "delete", "retrieve-many", "retrieve", "patch", "info":
		return nil
	}

	return errors.New("Invalid operation: %s. Must be one of 'create', 'update', 'delete', 'retrieve-many', 'retrieve', 'patch' or 'info'")
}
