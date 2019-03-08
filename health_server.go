package bahamut

import (
	"context"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"
)

// an healthServer is the structure serving the health check endpoint.
type healthServer struct {
	cfg    config
	server *http.Server
}

// newHealthServer returns a new healthServer.
func newHealthServer(cfg config) *healthServer {

	s := &healthServer{
		cfg:    cfg,
		server: &http.Server{Addr: cfg.healthServer.listenAddress},
	}

	s.server.Handler = s

	return s
}

func (s *healthServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	switch r.URL.Path {

	case "/":

		if s.cfg.healthServer.healthHandler == nil {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		if err := s.cfg.healthServer.healthHandler(); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)

	case "/metrics":
		if s.cfg.healthServer.metricsManager == nil {
			w.WriteHeader(http.StatusNotImplemented)
			return
		}

		s.cfg.healthServer.metricsManager.Write(w, r)

	default:

		if s.cfg.healthServer.customStats == nil {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}

		f := s.cfg.healthServer.customStats[strings.TrimPrefix(r.URL.Path, "/")]
		if f == nil {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}

		f(w, r)
	}
}

func (s *healthServer) start(ctx context.Context) {

	zap.L().Debug("Health server enabled", zap.String("listen", s.cfg.healthServer.listenAddress))

	go func() {
		if err := s.server.ListenAndServe(); err != nil {
			if err == http.ErrServerClosed {
				return
			}
			zap.L().Fatal("Unable to start health server", zap.Error(err))
		}
	}()

	zap.L().Info("Health server started", zap.String("address", s.cfg.healthServer.listenAddress))

	<-ctx.Done()
}

func (s *healthServer) stop() context.Context {

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)

	go func() {
		defer cancel()
		if err := s.server.Shutdown(ctx); err != nil {
			zap.L().Error("Could not gracefully stop health server", zap.Error(err))
		} else {
			zap.L().Debug("Health server stopped")
		}
	}()

	return ctx
}
