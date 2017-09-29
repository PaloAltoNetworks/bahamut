package bahamut

import (
	"net/http"

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

// start starts the healthServer.
func (s *healthServer) start() {

	s.server = &http.Server{Addr: s.config.HealthServer.ListenAddress}
	s.server.Handler = s
	s.server.SetKeepAlivesEnabled(true)

	if err := s.server.ListenAndServe(); err != nil {
		zap.L().Panic("Unable to start health server", zap.Error(err))
	}

	zap.L().Info("Health server started", zap.String("address", s.config.HealthServer.ListenAddress))
}

func (s *healthServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	if s.config.HealthServer.HealthHandler == nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if err := s.config.HealthServer.HealthHandler(w); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

// stop stops the healthServer.
func (s *healthServer) stop() {

	// a.server.Shutdown() // Uncomment with Go 1.8
	// a.server = nil
}
