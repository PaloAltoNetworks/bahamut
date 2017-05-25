package bahamut

import (
	"net"
	"net/http"
	"time"

	"github.com/aporeto-inc/addedeffect/cache"
)

type basicRateLimiter struct {
	cache cache.Cacher
	rps   int
}

// NewRateLimiter returns a new RateLimiter.
func NewRateLimiter(rps int) RateLimiter {

	return &basicRateLimiter{
		cache: cache.NewMemoryCache(),
		rps:   rps,
	}
}

func (r *basicRateLimiter) requestIP(req *http.Request) (string, error) {

	ip, _, err := net.SplitHostPort(req.RemoteAddr)
	if err != nil {
		return "", err
	}

	return ip, nil
}

func (r *basicRateLimiter) RateLimit(req *http.Request) (bool, error) {

	ip, err := r.requestIP(req)
	if err != nil {
		return false, err
	}

	var count int
	if c := r.cache.Get(ip); c != nil {
		count = c.(int)
	}

	count++

	r.cache.SetWithExpiration(ip, count, time.Second)

	return count > r.rps, nil
}

type rateLimiterWithBan struct {
	banCache cache.Cacher
	banTime  time.Duration

	basicRateLimiter
}

// NewBanRateLimiter returns a new RateLimiter.
func NewRateLimiterWithBan(rps int, banTime time.Duration) RateLimiter {

	return &rateLimiterWithBan{
		banCache: cache.NewMemoryCache(),
		banTime:  banTime,
		basicRateLimiter: basicRateLimiter{
			cache: cache.NewMemoryCache(),
			rps:   rps,
		},
	}
}

func (r *rateLimiterWithBan) RateLimit(req *http.Request) (bool, error) {

	ip, err := r.requestIP(req)
	if err != nil {
		return false, err
	}

	if r.banCache.Exists(ip) {
		return true, nil
	}

	limited, err := r.basicRateLimiter.RateLimit(req)
	if err != nil {
		return false, err
	}

	if limited {
		r.banCache.SetWithExpiration(ip, true, r.banTime)
	}

	return limited, nil
}
