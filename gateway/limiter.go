package gateway

import (
	"errors"
	"net/http"
	"time"

	"github.com/karlseguin/ccache/v2"
	"golang.org/x/time/rate"
)

const maxCacheSize = 65536

var errTooManyRequest = errors.New("Please retry in a moment")

type sourceLimiter struct {
	nextHTTP        http.Handler
	nextWS          http.Handler
	rls             *ccache.Cache
	sourceExtractor SourceExtractor
	rateExtractor   RateExtractor
	errorHandler    *errorHandler
	defaultLimit    rate.Limit
	defaultBurst    int
	metricManager   LimiterMetricManager
}

func newSourceLimiter(
	nextHTTP http.Handler,
	nextWS http.Handler,
	defaultLimit rate.Limit,
	defaultBurst int,
	sourceExtractor SourceExtractor,
	rateExtractor RateExtractor,
	errorHandler *errorHandler,
	metricManager LimiterMetricManager,
) *sourceLimiter {

	if errorHandler == nil {
		panic("errorHandler must not be nil")
	}

	if sourceExtractor == nil {
		panic("sourceExtractor must not be nil")
	}

	return &sourceLimiter{
		nextHTTP:        nextHTTP,
		nextWS:          nextWS,
		defaultLimit:    defaultLimit,
		defaultBurst:    defaultBurst,
		sourceExtractor: sourceExtractor,
		rateExtractor:   rateExtractor,
		errorHandler:    errorHandler,
		rls:             ccache.New(ccache.Configure().MaxSize(maxCacheSize)),
		metricManager:   metricManager,
	}
}

func (l *sourceLimiter) ServeHTTP(w http.ResponseWriter, req *http.Request) {

	key, err := l.sourceExtractor.ExtractSource(req)
	if err != nil {
		l.errorHandler.ServeHTTP(w, req, errTooManyRequest)
		return
	}

	var rl *rate.Limiter

	var limit rate.Limit
	var burst int

	if l.rateExtractor != nil {
		limit, burst, err = l.rateExtractor.ExtractRates(req)
		if err != nil {
			l.errorHandler.ServeHTTP(w, req, errTooManyRequest)
			return
		}
	} else {
		limit = l.defaultLimit
		burst = l.defaultBurst
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
		if l.metricManager != nil {
			l.metricManager.RegisterLimitedConnection()
		}
		return
	}

	if l.metricManager != nil {
		l.metricManager.RegisterAcceptedConnection()
	}

	if req.Header.Get(internalWSMarkingHeader) != "" {
		l.nextWS.ServeHTTP(w, req)
	} else {
		l.nextHTTP.ServeHTTP(w, req)
	}
}
