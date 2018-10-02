package bahamut

import "net/http"

// A MetricsManager handles Prometheus Metrics Management
type MetricsManager interface {
	MeasureRequest(code *int, method string) func()
	RegisterWSConnection()
	UnregisterWSConnection()
	Write(w http.ResponseWriter, r *http.Request)
}
