package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests processed, partitioned by status code and method.",
		},
		[]string{"method", "status", "path"},
	)

	httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Histogram of response latency (seconds) of HTTP requests.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)
)

// metricsResponseWriter is a minimal wrapper for http.ResponseWriter that allows the
// written HTTP status code to be captured for metrics.
type metricsResponseWriter struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

func wrapResponseWriter(w http.ResponseWriter) *metricsResponseWriter {
	return &metricsResponseWriter{ResponseWriter: w, status: http.StatusOK}
}

func (rw *metricsResponseWriter) Status() int {
	return rw.status
}

func (rw *metricsResponseWriter) WriteHeader(code int) {
	if rw.wroteHeader {
		return
	}
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
	rw.wroteHeader = true
}

func (rw *metricsResponseWriter) Write(b []byte) (int, error) {
	if !rw.wroteHeader {
		// The status will be StatusOK if WriteHeader has not been called yet
		rw.WriteHeader(http.StatusOK)
	}
	return rw.ResponseWriter.Write(b)
}

// Flush propagates to the underlying writer so SSE/streaming handlers keep working
// through this wrapper.
func (rw *metricsResponseWriter) Flush() {
	if f, ok := rw.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

// Metrics is a middleware that captures HTTP metrics.
func Metrics(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		wrapped := wrapResponseWriter(w)

		next.ServeHTTP(wrapped, r)

		duration := time.Since(start).Seconds()

		// Do not track /metrics and /health calls to avoid pollution
		if r.URL.Path != "/metrics" && r.URL.Path != "/health" {
			status := strconv.Itoa(wrapped.status)
			httpRequestsTotal.WithLabelValues(r.Method, status, r.URL.Path).Inc()
			httpRequestDuration.WithLabelValues(r.Method, r.URL.Path).Observe(duration)
		}
	})
}
