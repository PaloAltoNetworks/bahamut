package bahamut

import (
	"os"
	"runtime/trace"

	"go.uber.org/zap"
)

// an tracingServer is the structure serving the tracing.
type tracingServer struct {
	config    Config
	traceFile *os.File
}

// newTracingServer returns a new tracingServer.
func newTracingServer(config Config) *tracingServer {

	return &tracingServer{
		config: config,
	}
}

// start starts the tracingServer.
func (s *tracingServer) start() {

	var err error

	s.traceFile, err = os.Create(s.config.TracingServer.OutFilePath)
	if err != nil {
		zap.L().Fatal("Unable to create trace out file",
			zap.String("path", s.config.TracingServer.OutFilePath),
			zap.Error(err),
		)
	}

	if err = trace.Start(s.traceFile); err != nil {
		zap.L().Fatal("Unable to start the trace server", zap.Error(err))
	}

	zap.L().Info("Trace server started", zap.String("out", s.config.TracingServer.OutFilePath))
}

// stop stops the tracingServer.
func (s *tracingServer) stop() {

	trace.Stop()
	s.traceFile.Close() // nolint: errcheck
}
