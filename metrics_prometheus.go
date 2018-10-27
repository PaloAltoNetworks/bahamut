package bahamut

import (
	"net/http"
	"regexp"
	"strconv"
	"strings"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var vregexp = regexp.MustCompile(`^/v/\d+`)

func sanitizeURL(url string) string {

	url = vregexp.ReplaceAllString(url, "")

	parts := strings.Split(url, "/")
	if len(parts) <= 2 {
		return url
	}

	parts[2] = ":id"

	return strings.Join(parts, "/")
}

type prometheusMetricsManager struct {
	reqDurationMetric   *prometheus.SummaryVec
	reqTotalMetric      *prometheus.CounterVec
	errorMetric         *prometheus.CounterVec
	wsConnTotalMetric   prometheus.Counter
	wsConnCurrentMetric prometheus.Gauge

	handler http.Handler
}

// NewPrometheusMetricsManager returns a new MetricManager using the prometheus format.
func NewPrometheusMetricsManager() MetricsManager {
	mc := &prometheusMetricsManager{
		handler: promhttp.Handler(),
		reqTotalMetric: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_requests_total",
				Help: "The total number of requests.",
			},
			[]string{"method"},
		),
		reqDurationMetric: prometheus.NewSummaryVec(
			prometheus.SummaryOpts{
				Name: "http_requests_duration_seconds",
				Help: "The average duration of the requests",
			},
			[]string{"code", "method", "url"},
		),
		wsConnTotalMetric: prometheus.NewCounter(
			prometheus.CounterOpts{
				Name: "http_ws_connections_total",
				Help: "The total number of ws connection.",
			},
		),
		wsConnCurrentMetric: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "http_ws_connections_current",
				Help: "The current number of ws connection.",
			},
		),
		errorMetric: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_errors_5xx_total",
				Help: "The total number of 5xx errors.",
			},
			[]string{"trace", "method", "url"},
		),
	}

	prometheus.MustRegister(mc.reqTotalMetric)
	prometheus.MustRegister(mc.reqDurationMetric)
	prometheus.MustRegister(mc.wsConnTotalMetric)
	prometheus.MustRegister(mc.wsConnCurrentMetric)
	prometheus.MustRegister(mc.errorMetric)

	return mc
}

func (c *prometheusMetricsManager) MeasureRequest(code *int, method string, url string) func(Context) {

	c.reqTotalMetric.With(prometheus.Labels{
		"method": method,
	}).Inc()

	surl := sanitizeURL(url)

	timer := prometheus.NewTimer(
		prometheus.ObserverFunc(
			func(v float64) {
				c.reqDurationMetric.With(
					prometheus.Labels{
						"code":   strconv.Itoa(*code),
						"method": method,
						"url":    surl,
					},
				).Observe(v)
			},
		),
	)

	return func(ctx Context) {

		if *code >= http.StatusInternalServerError && ctx != nil {

			span := opentracing.SpanFromContext(ctx.Context())
			c.errorMetric.With(prometheus.Labels{
				"trace":  extractSpanID(span),
				"method": method,
				"url":    surl,
			}).Inc()
		}

		timer.ObserveDuration()
	}
}

func (c *prometheusMetricsManager) RegisterWSConnection() {
	c.wsConnTotalMetric.Inc()
	c.wsConnCurrentMetric.Inc()
}

func (c *prometheusMetricsManager) UnregisterWSConnection() {
	c.wsConnCurrentMetric.Dec()
}

func (c *prometheusMetricsManager) Write(w http.ResponseWriter, r *http.Request) {
	c.handler.ServeHTTP(w, r)
}
