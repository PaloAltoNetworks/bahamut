package gateway

import (
	"context"
	"net"
	"time"

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

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	lerr := l.limiter.Wait(ctx)

	c, err := l.Listener.Accept()
	if err != nil {
		return nil, err
	}

	if lerr != nil {
		c.Close() // nolint
	}

	return c, nil
}
