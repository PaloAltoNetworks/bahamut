package bahamut

import (
	"context"
	"net/http"
	"net/http/pprof"
	"time"

	"cloud.google.com/go/profiler"
	"go.uber.org/zap"
)

// an profilingServer is the structure serving the profiling.
type profilingServer struct {
	cfg    config
	server *http.Server
}

// newProfilingServer returns a new profilingServer.
func newProfilingServer(cfg config) *profilingServer {

	return &profilingServer{
		cfg: cfg,
	}
}

// start starts the profilingServer.
func (s *profilingServer) start(ctx context.Context) {

	if s.cfg.profilingServer.mode == "gcp" {

		name := s.cfg.meta.serviceName
		if prfx := s.cfg.profilingServer.gcpServicePrefix; prfx != "" {
			name = prfx + "-" + name
		}

		if err := profiler.Start(profiler.Config{
			Service:        name,
			ServiceVersion: s.cfg.meta.serviceVersion,
			ProjectID:      s.cfg.profilingServer.gcpProjectID,
		}); err != nil {
			zap.L().Fatal("Unable to start gcp profile server", zap.Error(err))
		}

		projectID := s.cfg.profilingServer.gcpProjectID
		if projectID == "" {
			projectID = "auto"
		}

		zap.L().Info("GCP profiler started",
			zap.String("service", name),
			zap.String("project-id", projectID),
		)

	} else {

		mux := http.NewServeMux()
		mux.HandleFunc("/debug/pprof/", pprof.Index)
		mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
		mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
		mux.HandleFunc("/debug/pprof/trace", pprof.Trace)

		s.server = &http.Server{
			Addr:    s.cfg.profilingServer.listenAddress,
			Handler: mux,
		}

		go func() {
			if err := s.server.ListenAndServe(); err != nil {
				if err == http.ErrServerClosed {
					return
				}
				zap.L().Fatal("Unable to start profiling server", zap.Error(err))
			}
		}()

		zap.L().Info("Profiler profiler started", zap.String("address", s.cfg.profilingServer.listenAddress))
	}

	<-ctx.Done()
}

// stop stops the profilingServer.
func (s *profilingServer) stop() {

	if s.server == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)

	go func() {
		defer cancel()
		if err := s.server.Shutdown(ctx); err != nil {
			zap.L().Error("Could not gracefully stop profiling server", zap.Error(err))
		} else {
			zap.L().Debug("Profiling server stopped")
		}
	}()

	zap.L().Debug("Profile server stopped")
}
