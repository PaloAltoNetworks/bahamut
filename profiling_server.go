package bahamut

import (
	"net/http"
	"net/http/pprof"

	"github.com/go-zoo/bone"
)

// an profilingServer is the structure serving the profiling.
type profilingServer struct {
	config Config
	server *http.Server
}

// newProfilingServer returns a new profilingServer.
func newProfilingerver(config Config) *profilingServer {

	return &profilingServer{
		config: config,
	}
}

// start starts the profilingServer.
func (s *profilingServer) start() {

	address := s.config.ProfilingServer.ListenAddress

	s.server = &http.Server{Addr: address}
	log.WithField("address", address).Info("Starting profiling server.")

	mux := bone.New()
	mux.Handle("/debug/pprof/", http.HandlerFunc(pprof.Index))
	mux.Handle("/debug/pprof/cmdline", http.HandlerFunc(pprof.Cmdline))
	mux.Handle("/debug/pprof/profile", http.HandlerFunc(pprof.Profile))
	mux.Handle("/debug/pprof/symbol", http.HandlerFunc(pprof.Symbol))
	mux.Handle("/debug/pprof/trace", http.HandlerFunc(pprof.Trace))

	s.server.Handler = mux
	if err := s.server.ListenAndServe(); err != nil {
		log.WithError(err).Fatal("Unable to start profiling http server.")
	}
}

// stop stops the profilingServer.
func (s *profilingServer) stop() {

	// s.server.Shutdown() // Uncomment with Go 1.8
	// s.server = nil
}
