package bahamut

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"runtime/debug"
	"strings"
	"time"

	"github.com/paulbellamy/ratecounter"
	"go.uber.org/zap"
)

type statsCounter struct {
	rps  *ratecounter.RateCounter
	wsps *ratecounter.RateCounter
	r    ratecounter.Counter
	ws   ratecounter.Counter
}

func newStatsCounter() *statsCounter {
	return &statsCounter{
		rps:  ratecounter.NewRateCounter(1 * time.Second),
		wsps: ratecounter.NewRateCounter(1 * time.Second),
	}
}

type runtimeStats struct {
	NumGoroutine     int
	NumCgoCall       int64
	NumCPU           int
	NumRequests      int64
	NumWSConnections int64
	MemStats         runtime.MemStats
	GCStats          debug.GCStats
}

// an healthServer is the structure serving the health check endpoint.
type healthServer struct {
	cfg          config
	server       *http.Server
	statsCounter *statsCounter
}

// newHealthServer returns a new healthServer.
func newHealthServer(cfg config, statsCounter *statsCounter) *healthServer {

	return &healthServer{
		cfg:          cfg,
		statsCounter: statsCounter,
	}
}

func (s *healthServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	switch r.URL.Path {

	case "/":

		if s.cfg.healthServer.healthHandler == nil {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		if err := s.cfg.healthServer.healthHandler(); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)

	case "/_rstats":

		stats := runtimeStats{
			NumGoroutine:     runtime.NumGoroutine(),
			NumCgoCall:       runtime.NumCgoCall(),
			NumCPU:           runtime.NumCPU(),
			NumRequests:      s.statsCounter.r.Value(),
			NumWSConnections: s.statsCounter.ws.Value(),
		}

		debug.ReadGCStats(&stats.GCStats)
		runtime.ReadMemStats(&stats.MemStats)

		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")

		if err := enc.Encode(stats); err != nil {
			http.Error(w, fmt.Sprintf("Unable to encode runtime stats: %s", err), http.StatusInternalServerError)
		}

	case "/metrics":

		w.Header().Set("content-type", "text/plain;")
		w.WriteHeader(http.StatusOK)

		var memStats runtime.MemStats
		runtime.ReadMemStats(&memStats)

		ts := time.Now().Unix() * 1000
		fmt.Fprintln(w, "# HELP http_requests_per_second The current number of requests per second.")
		fmt.Fprintln(w, "# TYPE http_requests_per_second gauge")
		fmt.Fprintf(w, "http_requests_per_second %d %d\n\n", s.statsCounter.rps.Rate(), ts)

		fmt.Fprintln(w, "# HELP http_requests_total The total number of requests.")
		fmt.Fprintln(w, "# TYPE http_requests_total counter")
		fmt.Fprintf(w, "http_requests_total %d %d\n\n", s.statsCounter.r.Value(), ts)

		fmt.Fprintln(w, "# HELP http_ws_connections_per_second The current number of ws connection per second.")
		fmt.Fprintln(w, "# TYPE http_ws_connections_per_second gauge")
		fmt.Fprintf(w, "http_ws_connections_per_second %d %d\n\n", s.statsCounter.wsps.Rate(), ts)

		fmt.Fprintln(w, "# HELP http_ws_connections_total The total number of websocket connections.")
		fmt.Fprintln(w, "# TYPE http_ws_connections_total counter")
		fmt.Fprintf(w, "http_ws_connections_total %d %d\n\n", s.statsCounter.ws.Value(), ts)

		fmt.Fprintln(w, "# HELP mem_sys The total bytes of memory obtained from the OS.")
		fmt.Fprintln(w, "# TYPE mem_sys gauge")
		fmt.Fprintf(w, "mem_sys %d %d\n\n", memStats.Sys, ts)

		fmt.Fprintln(w, "# HELP mem_heap_sys The number of bytes of heap memory obtained from the OS.")
		fmt.Fprintln(w, "# TYPE mem_heap_sys gauge")
		fmt.Fprintf(w, "mem_heap_sys %d %d\n\n", memStats.HeapSys, ts)

		fmt.Fprintln(w, "# HELP mem_heap_alloc The number of bytes of bytes of allocated heap objects.")
		fmt.Fprintln(w, "# TYPE mem_heap_alloc gauge")
		fmt.Fprintf(w, "mem_heap_alloc %d %d\n\n", memStats.HeapAlloc, ts)

		fmt.Fprintln(w, "# HELP mem_heap_idle The number of bytes in idle (unused) spans.")
		fmt.Fprintln(w, "# TYPE mem_heap_idle gauge")
		fmt.Fprintf(w, "mem_heap_idle %d %d\n\n", memStats.HeapIdle, ts)

		fmt.Fprintln(w, "# HELP mem_stack_sys The number of bytes of stack memory obtained from the OS.")
		fmt.Fprintln(w, "# TYPE mem_stack_sys gauge")
		fmt.Fprintf(w, "mem_stack_sys %d %d\n\n", memStats.StackSys, ts)

		fmt.Fprintln(w, "# HELP mem_stack_in_user The number of bytes in stack spans.")
		fmt.Fprintln(w, "# TYPE mem_stack_in_user gauge")
		fmt.Fprintf(w, "mem_stack_in_user %d %d\n\n", memStats.StackInuse, ts)

		fmt.Fprintln(w, "# HELP num_go_routines The current number of go routines.")
		fmt.Fprintln(w, "# TYPE num_go_routines gauge")
		fmt.Fprintf(w, "num_go_routines %d %d\n\n", runtime.NumGoroutine(), ts)

	default:

		if s.cfg.healthServer.customStats == nil {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}

		f := s.cfg.healthServer.customStats[strings.TrimPrefix(r.URL.Path, "/")]
		if f == nil {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}

		f(w, r)
	}
}

func (s *healthServer) start(ctx context.Context) {

	s.server = &http.Server{Addr: s.cfg.healthServer.listenAddress}
	s.server.Handler = s
	s.server.SetKeepAlivesEnabled(true)

	zap.L().Debug("Health server enabled", zap.String("listen", s.cfg.healthServer.listenAddress))

	go func() {
		if err := s.server.ListenAndServe(); err != nil {
			if err == http.ErrServerClosed {
				return
			}
			zap.L().Fatal("Unable to start health server", zap.Error(err))
		}
	}()

	zap.L().Info("Health server started", zap.String("address", s.cfg.healthServer.listenAddress))

	<-ctx.Done()
}

func (s *healthServer) stop() context.Context {

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)

	go func() {
		defer cancel()
		if err := s.server.Shutdown(ctx); err != nil {
			zap.L().Error("Could not gracefully stop health server", zap.Error(err))
		} else {
			zap.L().Debug("Health server stopped")
		}
	}()

	return ctx
}
