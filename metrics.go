package bahamut

import (
	"net/http"

	opentracing "github.com/opentracing/opentracing-go"
)

// FinishMeasurementFunc is the kind of functinon returned by MetricsManager.MeasureRequest().
type FinishMeasurementFunc func(code int, span opentracing.Span)

// A MetricsManager handles Prometheus Metrics Management
type MetricsManager interface {
	MeasureRequest(method string, url string) FinishMeasurementFunc
	RegisterWSConnection()
	UnregisterWSConnection()
	Write(w http.ResponseWriter, r *http.Request)
}
