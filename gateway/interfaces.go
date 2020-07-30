package gateway

import (
	"errors"
	"net/http"
	"time"

	"github.com/vulcand/oxy/ratelimit"
)

// ErrUpstreamerTooManyRequests can be returned to
// instruct the bahamut.Gateway to return to stop
// routing and return a a 429 Too Many Request error to
// the client.
var ErrUpstreamerTooManyRequests = errors.New("Please retry in a moment")

// An Upstreamer is the interface that can compute upstreams.
type Upstreamer interface {

	// Upstream is called by the bahamut.Gateway for each incoming request
	// in order to find which upstream to forward the request to, based
	// on the incoming http.Request and any other details the implementation
	// whishes to. Needless to say, it must be fast or it would severely degrade
	// the performances of the bahamut.Gateway.
	//
	// The request state must not be changed from this function.
	//
	// The returned upstream is a string in the form "https://10.3.19.4".
	// If it is empty, the bahamut.Gayeway will return a
	// 503 Service Unavailable error.
	//
	// If Upstream returns an error, the bahamut.Gayeway will check for a
	// known ErrUpstreamerX and will act accordingly. Otherwise it will
	// return the error as a 500 Internal Server Error.
	Upstream(req *http.Request) (upstream string, err error)
}

// A SourceLimiter can be used to decide rates per token.
// This allows to perform advanced computation to determine how
// to rate limit one unique client.
type SourceLimiter interface {

	// DefaultRates returns the default rate limiting.
	DefaultRates() *ratelimit.RateSet

	// ExtractRates will be called to decide what would be the rate to
	// given a request.
	ExtractRates(r *http.Request) (*ratelimit.RateSet, error)

	// ExtractSource will be called to decide what would be the rate to
	// given a request.
	ExtractSource(req *http.Request) (token string, amount int64, err error)
}

// A LatencyBasedUpstreamer is the interface that can circle back
// response time as an input for Upstreamer decision.
type LatencyBasedUpstreamer interface {
	CollectLatency(address string, responseTime time.Duration)
	Upstreamer
}

// A Gateway can be used as an api gateway.
type Gateway interface {
	Start()
	Stop()
}
