package bahamut

import (
	"context"

	"cloud.google.com/go/profiler"
	"go.uber.org/zap"

	gops "github.com/google/gops/agent"
)

// an profilingServer is the structure serving the profiling.
type profilingServer struct {
	cfg config
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

		if err := gops.Listen(gops.Options{
			Addr: s.cfg.profilingServer.listenAddress,
		}); err != nil {
			zap.L().Fatal("Unable to start gops profile server", zap.Error(err))
		}

		zap.L().Info("GOPS profiler started", zap.String("address", s.cfg.profilingServer.listenAddress))
	}

	<-ctx.Done()
}

// stop stops the profilingServer.
func (s *profilingServer) stop() {

	gops.Close()
	zap.L().Debug("Profile server stopped")
}
