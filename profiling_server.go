package bahamut

import (
	"context"

	"cloud.google.com/go/profiler"
	"go.uber.org/zap"

	gops "github.com/google/gops/agent"
)

// an profilingServer is the structure serving the profiling.
type profilingServer struct {
	config Config
}

// newProfilingServer returns a new profilingServer.
func newProfilingServer(config Config) *profilingServer {

	return &profilingServer{
		config: config,
	}
}

// start starts the profilingServer.
func (s *profilingServer) start(ctx context.Context) {

	if s.config.ProfilingServer.Mode == "gcp" {

		name := s.config.Meta.ServiceName
		if prfx := s.config.ProfilingServer.GCPServicePrefix; prfx != "" {
			name = prfx + "-" + name
		}

		if err := profiler.Start(profiler.Config{
			Service:        name,
			ServiceVersion: s.config.Meta.ServiceVersion,
			ProjectID:      s.config.ProfilingServer.GCPProjectID,
		}); err != nil {
			zap.L().Fatal("Unable to start gcp profile server", zap.Error(err))
		}

		projectID := s.config.ProfilingServer.GCPProjectID
		if projectID == "" {
			projectID = "auto"
		}

		zap.L().Info("GCP profiler started",
			zap.String("service", name),
			zap.String("project-id", projectID),
		)

	} else {

		if err := gops.Listen(gops.Options{
			Addr: s.config.ProfilingServer.ListenAddress,
		}); err != nil {
			zap.L().Fatal("Unable to start gops profile server", zap.Error(err))
		}

		zap.L().Info("GOPS profiler started", zap.String("address", s.config.ProfilingServer.ListenAddress))
	}

	<-ctx.Done()
}

// stop stops the profilingServer.
func (s *profilingServer) stop() {

	gops.Close()
	zap.L().Debug("Profile server stopped")
}
