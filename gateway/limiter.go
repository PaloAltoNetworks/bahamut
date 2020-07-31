package gateway

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/cespare/xxhash"
	"github.com/karlseguin/ccache/v2"
	"golang.org/x/time/rate"
)

const maxCacheSize = 65536

var errTooManyRequest = errors.New("Please retry in a moment")

type defaultExtractorFunc struct {
}

func (f defaultExtractorFunc) ExtractSource(r *http.Request) (string, error) {

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "default", nil
	}

	return fmt.Sprintf("%d", xxhash.Sum64([]byte(authHeader))), nil
}

type sourceLimiter struct {
	next            http.Handler
	rls             *ccache.Cache
	sourceExtractor SourceExtractor
	rateExtractor   RateExtractor
	errorHandler    *errorHandler
}

func newSourceLimiter(
	next http.Handler,
	extractor SourceExtractor,
	rateExtractor RateExtractor,
	errorHandler *errorHandler,
) *sourceLimiter {

	if rateExtractor == nil {
		panic("rateExtractor must not be nil")
	}

	if errorHandler == nil {
		panic("errorHandler must not be nil")
	}

	if extractor == nil {
		extractor = defaultExtractorFunc{}
	}

	return &sourceLimiter{
		next:            next,
		sourceExtractor: extractor,
		rateExtractor:   rateExtractor,
		errorHandler:    errorHandler,
		rls:             ccache.New(ccache.Configure().MaxSize(maxCacheSize)),
	}
}

func (l *sourceLimiter) ServeHTTP(w http.ResponseWriter, req *http.Request) {

	key, err := l.sourceExtractor.ExtractSource(req)
	if err != nil {
		l.errorHandler.ServeHTTP(w, req, errTooManyRequest)
		return
	}

	var rl *rate.Limiter

	limit, burst, err := l.rateExtractor.ExtractRates(req)
	if err != nil {
		l.errorHandler.ServeHTTP(w, req, errTooManyRequest)
		return
	}

	if item := l.rls.Get(key); item == nil || item.Value() == nil || item.Expired() {
		rl = rate.NewLimiter(limit, burst)
		l.rls.Set(key, rl, time.Hour)
	} else {
		rl = item.Value().(*rate.Limiter)
	}

	if rl.Limit() != limit {
		rl.SetLimit(limit)
	}
	if rl.Burst() != burst {
		rl.SetBurst(burst)
	}

	if !rl.Allow() {
		l.errorHandler.ServeHTTP(w, req, errTooManyRequest)
		return
	}

	l.next.ServeHTTP(w, req)
}
