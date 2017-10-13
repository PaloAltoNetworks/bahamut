package bahamut

import (
	"net/http"
	"net/http/pprof"

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
func (s *profilingServer) start() {

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
	if err := s.server.ListenAndServe(); err != nil {
		zap.L().Panic("Unable to start profiling http server", zap.Error(err))
	}

	zap.L().Info("Profiling server started", zap.String("address", s.config.ProfilingServer.ListenAddress))
}

// stop stops the profilingServer.
func (s *profilingServer) stop() {

	// s.server.Shutdown() // Uncomment with Go 1.8
	// s.server = nil
}
