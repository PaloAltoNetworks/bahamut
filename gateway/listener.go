package gateway

import (
	"net"

	"golang.org/x/time/rate"
)

type limitListener struct {
	net.Listener
	limiter *rate.Limiter
}

func newLimitedListener(l net.Listener, cps rate.Limit, burst int) net.Listener {

	return &limitListener{
		Listener: l,
		limiter:  rate.NewLimiter(cps, burst),
	}

}

func (l *limitListener) Accept() (net.Conn, error) {

	c, err := l.Listener.Accept()
	if err != nil {
		return nil, err
	}

	if !l.limiter.Allow() {
		c.Close() //revive:disable-line:unhandled-error
	}

	return c, nil
}
