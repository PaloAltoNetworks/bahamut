package bahamut

import (
	"context"
	"net/http"
	"net/http/pprof"
	"time"

	"go.uber.org/zap"

	"github.com/go-zoo/bone"
	"github.com/google/gops/agent"
)

// an profilingServer is the structure serving the profiling.
type profilingServer struct {
	config Config
	server *http.Server
}

// newProfilingServer returns a new profilingServer.
func newProfilingServer(config Config) *profilingServer {

	return &profilingServer{
		config: config,
	}
}

// start starts the profilingServer.
func (s *profilingServer) start(ctx context.Context) {

	if err := agent.Listen(agent.Options{}); err != nil {
		zap.L().Fatal("Unable to start the gops agent", zap.Error(err))
	}

	mux := bone.New()
	mux.Handle("/debug/pprof/", http.HandlerFunc(pprof.Index))
	mux.Handle("/debug/pprof/cmdline", http.HandlerFunc(pprof.Cmdline))
	mux.Handle("/debug/pprof/profile", http.HandlerFunc(pprof.Profile))
	mux.Handle("/debug/pprof/symbol", http.HandlerFunc(pprof.Symbol))
	mux.Handle("/debug/pprof/trace", http.HandlerFunc(pprof.Trace))

	s.server = &http.Server{Addr: s.config.ProfilingServer.ListenAddress}
	s.server.Handler = mux

	go func() {
		if err := s.server.ListenAndServe(); err != nil {
			if err == http.ErrServerClosed {
				return
			}
			zap.L().Fatal("Unable to start profile server", zap.Error(err))
		}
	}()

	zap.L().Info("Profile server started", zap.String("address", s.config.ProfilingServer.ListenAddress))

	<-ctx.Done()
}

// stop stops the profilingServer.
func (s *profilingServer) stop() {

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := s.server.Shutdown(ctx); err != nil {
		zap.L().Error("Could not gracefully stop profile server", zap.Error(err))
	}

	zap.L().Debug("Profile server stopped")
}
