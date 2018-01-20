package bahamut

import (
	"context"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// an healthServer is the structure serving the health check endpoint.
type healthServer struct {
	config Config
	server *http.Server
}

// newHealthServer returns a new healthServer.
func newHealthServer(config Config) *healthServer {

	return &healthServer{
		config: config,
	}
}

func (s *healthServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	if s.config.HealthServer.HealthHandler == nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if err := s.config.HealthServer.HealthHandler(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *healthServer) start(ctx context.Context) {

	s.server = &http.Server{Addr: s.config.HealthServer.ListenAddress}
	s.server.Handler = s
	s.server.SetKeepAlivesEnabled(true)

	zap.L().Debug("Health server enabled", zap.String("listen", s.config.HealthServer.ListenAddress))

	go func() {
		if err := s.server.ListenAndServe(); err != nil {
			if err == http.ErrServerClosed {
				return
			}
			zap.L().Fatal("Unable to start health server", zap.Error(err))
		}
	}()

	zap.L().Info("Health server started", zap.String("address", s.config.HealthServer.ListenAddress))

	<-ctx.Done()
}

func (s *healthServer) stop() {

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := s.server.Shutdown(ctx); err != nil {
		zap.L().Error("Could not gracefuly stop health server", zap.Error(err))
	}

	zap.L().Debug("Health server stopped")
}
