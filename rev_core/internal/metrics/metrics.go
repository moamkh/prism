package metrics

import (
	"net/http"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// Request metrics
	RequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "rev_proxy_requests_total",
		Help: "Total number of HTTP requests handled by the proxy",
	}, []string{"method", "path", "status_code"})

	RequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "rev_proxy_request_duration_seconds",
		Help:    "HTTP request latency distribution",
		Buckets: prometheus.DefBuckets,
	}, []string{"method", "path"})

	// Auth metrics
	AuthFailuresTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "rev_proxy_auth_failures_total",
		Help: "Total number of authentication/authorization failures",
	}, []string{"reason"})

	// Usage metrics
	UsageDroppedTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "rev_proxy_usage_dropped_total",
		Help: "Total number of usage logs dropped due to full buffer",
	})

	UsageBufferLength = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "rev_proxy_usage_buffer_length",
		Help: "Current number of usage logs in the async buffer",
	})

	// Limiter metrics
	LimiterInflight = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "rev_proxy_limiter_inflight",
		Help: "Current number of in-flight requests per provider",
	}, []string{"provider_id"})

	LimiterMax = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "rev_proxy_limiter_max",
		Help: "Maximum allowed concurrent requests per provider",
	}, []string{"provider_id"})

	// DB metrics
	DBConnectionsInUse = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "rev_proxy_db_connections_in_use",
		Help: "Number of database connections currently in use",
	})

	DBConnectionsIdle = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "rev_proxy_db_connections_idle",
		Help: "Number of idle database connections",
	})
)

// Simple counters for JSON status endpoint.
var (
	RequestCount      int64
	RequestLatencySum int64 // milliseconds
)

// InstrumentHandler wraps an http.Handler to record request metrics.
func InstrumentHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(wrapped, r)
		duration := time.Since(start)
		path := cleanPath(r.URL.Path)
		RequestsTotal.WithLabelValues(r.Method, path, strconv.Itoa(wrapped.statusCode)).Inc()
		RequestDuration.WithLabelValues(r.Method, path).Observe(duration.Seconds())
		atomic.AddInt64(&RequestCount, 1)
		atomic.AddInt64(&RequestLatencySum, duration.Milliseconds())
	})
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func cleanPath(path string) string {
	switch path {
	case "/health", "/docs", "/swagger/doc.json", "/metrics", "/status", "/api/status":
		return path
	}
	if len(path) > 10 && path[:10] == "/v1/models" {
		return "/v1/models"
	}
	if len(path) > 3 && path[:3] == "/v1" {
		return "/v1/{path}"
	}
	return path
}
