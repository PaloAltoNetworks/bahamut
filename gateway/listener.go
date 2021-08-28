package gateway

import (
	"net"

	"golang.org/x/time/rate"
)

type limitListener struct {
	net.Listener
	limiter       *rate.Limiter
	metricManager LimiterMetricManager
}

func newLimitedListener(l net.Listener, cps rate.Limit, burst int, metricManager LimiterMetricManager) net.Listener {

	return &limitListener{
		Listener:      l,
		limiter:       rate.NewLimiter(cps, burst),
		metricManager: metricManager,
	}

}

func (l *limitListener) Accept() (net.Conn, error) {

	c, err := l.Listener.Accept()
	if err != nil {
		return nil, err
	}

	if !l.limiter.Allow() {
		c.Close() // nolint
		if l.metricManager != nil {
			l.metricManager.RegisterLimitedConnection()
		}
	} else {
		if l.metricManager != nil {
			l.metricManager.RegisterAcceptedConnection()
		}
	}

	return c, nil
}
