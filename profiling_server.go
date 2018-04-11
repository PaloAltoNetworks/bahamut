package bahamut

import (
	"context"

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

	if err := gops.Listen(gops.Options{
		Addr: s.config.ProfilingServer.ListenAddress,
	}); err != nil {
		zap.L().Fatal("Unable to start profile server", zap.Error(err))
	}

	zap.L().Info("Profile server started", zap.String("address", s.config.ProfilingServer.ListenAddress))

	<-ctx.Done()
}

// stop stops the profilingServer.
func (s *profilingServer) stop() {

	gops.Close()
	zap.L().Debug("Profile server stopped")
}
