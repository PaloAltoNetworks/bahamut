package bahamut

import (
	"sync"
	"time"

	"go.uber.org/zap"
)

const (
	// PingStatusOK represents the status "ok"
	PingStatusOK = "ok"
	// PingStatusTimeout represents the status "timeout"
	PingStatusTimeout = "timeout"
	// PingStatusError represents the status "error"
	PingStatusError = "error"
)

// A Pinger is an interface for objects that implements a Ping method
type Pinger interface {
	Ping(timeout time.Duration) error
}

// RetrieveHealthStatus returns the status for each Pinger.
func RetrieveHealthStatus(timeout time.Duration, pingers map[string]Pinger) error {

	var firstError error

	var wg sync.WaitGroup
	wg.Add(len(pingers))
	m := &sync.Mutex{}
	for name, pinger := range pingers {
		go func(name string, pinger Pinger) {
			defer wg.Done()

			start := time.Now()
			err := pinger.Ping(timeout)
			status := stringifyStatus(err)
			duration := time.Since(start)

			zap.L().Info("Ping",
				zap.String("service", name),
				zap.String("status", status),
				zap.String("duration", duration.String()),
				zap.Error(err),
			)

			m.Lock()
			if err != nil && firstError == nil {
				firstError = err
			}
			m.Unlock()
		}(name, pinger)
	}

	wg.Wait()

	return firstError
}

// stringify status output
func stringifyStatus(err error) string {
	if err == nil {
		return PingStatusOK
	}

	errMsg := err.Error()
	if errMsg == PingStatusTimeout {
		return PingStatusTimeout
	}

	return PingStatusError
}
