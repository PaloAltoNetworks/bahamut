package bahamut

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

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

	if err := currentMocker.installMock(mock); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	zap.L().Info("mock installed",
		zap.String("identity", mock.IdentityName),
		zap.String("operation", string(mock.Operation)),
		zap.String("code", mock.Function),
	)

	w.WriteHeader(http.StatusCreated)
}

func (s *mockServer) handleUninstallMock(w http.ResponseWriter, req *http.Request) {

	op := bone.GetValue(req, "operation")
	identity := bone.GetValue(req, "identity")
	if err := currentMocker.uninstallMock(elemental.Operation(op), identity); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	zap.L().Info("mock uninstalled",
		zap.String("identity", identity),
		zap.String("operation", op),
	)

	w.WriteHeader(http.StatusOK)
}

func (s *mockServer) start(ctx context.Context) {

	s.server = &http.Server{Addr: s.config.MockServer.ListenAddress}

	mux := bone.New()
	mux.Post("/mock/install", http.HandlerFunc(s.handleInstallMock))
	mux.Delete("/mock/uninstall/:operation/:identity", http.HandlerFunc(s.handleUninstallMock))

	s.server.Handler = mux

	go func() {
		if err := s.server.ListenAndServe(); err != nil {
			if err == http.ErrServerClosed {
				return
			}
			zap.L().Fatal("Unable to start mock server", zap.Error(err))
		}
	}()

	zap.L().Warn("Mock server started", zap.String("listen", s.config.MockServer.ListenAddress))

	<-ctx.Done()
}

func (s *mockServer) stop() {

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := s.server.Shutdown(ctx); err != nil {
		zap.L().Error("Could not gracefuly stop mock server", zap.Error(err))
	}

	zap.L().Debug("Mock server stopped")
}
