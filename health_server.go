package bahamut

import (
	"net/http"

	"github.com/Sirupsen/logrus"
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

	address := s.config.HealthServer.ListenAddress
	logrus.WithField("address", address).Info("Starting health server.")

	s.server = &http.Server{Addr: address}
	s.server.Handler = s
	s.server.SetKeepAlivesEnabled(true)

	if err := s.server.ListenAndServe(); err != nil {
		logrus.WithError(err).Fatal("Unable to start api server.")
	}
}

func (s *healthServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	if s.config.HealthServer.HealthHandler == nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if err := s.config.HealthServer.HealthHandler(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusNoContent)
}

// stop stops the healthServer.
func (s *healthServer) stop() {

	// a.server.Shutdown() // Uncomment with Go 1.8
	// a.server = nil
}
