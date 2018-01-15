package bahamut

import (
	"net"
	"net/http"
	"time"

	"github.com/bluele/gcache"
)

type basicRateLimiter struct {
	cache gcache.Cache
	rps   int
}

// NewRateLimiter returns a new RateLimiter.
func NewRateLimiter(rps int) RateLimiter {

	return &basicRateLimiter{
		cache: gcache.New(1024).LRU().Build(),
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

	return r.rateLimitWithIP(ip)
}

func (r *basicRateLimiter) rateLimitWithIP(ip string) (bool, error) {

	var count int
	if c, _ := r.cache.Get(ip); c != nil {
		count = c.(int)
	}

	count++

	r.cache.SetWithExpire(ip, count, time.Second) // nolint: errcheck

	return count > r.rps, nil
}

type rateLimiterWithBan struct {
	banCache gcache.Cache
	banTime  time.Duration

	basicRateLimiter
}

// NewRateLimiterWithBan returns a new RateLimiter that bans for the
// given duration any IP that exceed the rate limit.
func NewRateLimiterWithBan(rps int, banTime time.Duration) RateLimiter {

	return &rateLimiterWithBan{
		banCache: gcache.New(1024).LRU().Build(),
		banTime:  banTime,
		basicRateLimiter: basicRateLimiter{
			cache: gcache.New(1024).LRU().Build(),
			rps:   rps,
		},
	}
}

func (r *rateLimiterWithBan) RateLimit(req *http.Request) (bool, error) {

	ip, err := r.requestIP(req)
	if err != nil {
		return false, err
	}

	if ok, _ := r.banCache.Get(ip); ok != nil {
		return true, nil
	}

	limited, err := r.rateLimitWithIP(ip)
	if err != nil {
		return false, err
	}

	if limited {
		r.banCache.SetWithExpire(ip, true, r.banTime) // nolint: errcheck
	}

	return limited, nil
}
