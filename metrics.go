package bahamut

import "net/http"

// A MetricsManager handles Prometheus Metrics Management
type MetricsManager interface {
	MeasureRequest(code *int, method string, url string) func(Context)
	RegisterWSConnection()
	UnregisterWSConnection()
	Write(w http.ResponseWriter, r *http.Request)
}
