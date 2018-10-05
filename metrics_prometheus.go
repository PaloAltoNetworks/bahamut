package bahamut

import (
	"net/http"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type prometheusMetricsManager struct {
	reqDurationMetric   *prometheus.SummaryVec
	reqTotalMetric      *prometheus.CounterVec
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
			[]string{"code", "method"},
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
	}

	prometheus.MustRegister(mc.reqTotalMetric)
	prometheus.MustRegister(mc.reqDurationMetric)
	prometheus.MustRegister(mc.wsConnTotalMetric)
	prometheus.MustRegister(mc.wsConnCurrentMetric)

	return mc
}

func (c *prometheusMetricsManager) MeasureRequest(code *int, method string) func() {

	c.reqTotalMetric.With(prometheus.Labels{
		"method": method,
	}).Inc()

	timer := prometheus.NewTimer(
		prometheus.ObserverFunc(
			func(v float64) {
				c.reqDurationMetric.With(
					prometheus.Labels{
						"code":   strconv.Itoa(*code),
						"method": method,
					},
				).Observe(v)
			},
		),
	)

	return func() { timer.ObserveDuration() }
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
