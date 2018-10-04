package bahamut

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"runtime/debug"
	"strings"
	"time"

	"go.uber.org/zap"
)

type runtimeStats struct {
	NumGoroutine     int
	NumCgoCall       int64
	NumCPU           int
	NumRequests      int64
	NumWSConnections int64
	MemStats         runtime.MemStats
	GCStats          debug.GCStats
}

// an healthServer is the structure serving the health check endpoint.
type healthServer struct {
	cfg    config
	server *http.Server
}

// newHealthServer returns a new healthServer.
func newHealthServer(cfg config) *healthServer {

	return &healthServer{
		cfg: cfg,
	}
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

	case "/_rstats":
		// this is deprecated

		stats := runtimeStats{
			NumGoroutine: runtime.NumGoroutine(),
			NumCgoCall:   runtime.NumCgoCall(),
			NumCPU:       runtime.NumCPU(),
		}

		debug.ReadGCStats(&stats.GCStats)
		runtime.ReadMemStats(&stats.MemStats)

		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")

		if err := enc.Encode(stats); err != nil {
			http.Error(w, fmt.Sprintf("Unable to encode runtime stats: %s", err), http.StatusInternalServerError)
		}

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

	s.server = &http.Server{Addr: s.cfg.healthServer.listenAddress}
	s.server.Handler = s

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
