package httpapi

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/xcreativs/gigmann/internal/observability"
)

// metrics holds the Prometheus collectors for the HTTP layer.
type metrics struct {
	requests *prometheus.CounterVec
	duration *prometheus.HistogramVec
}

// newMetrics registers the HTTP collectors on a fresh registry (one per router,
// so tests can build many routers without duplicate-registration panics).
func newMetrics() (*metrics, *prometheus.Registry) {
	reg := prometheus.NewRegistry()
	m := &metrics{
		requests: prometheus.NewCounterVec(
			prometheus.CounterOpts{Name: "http_requests_total", Help: "Total HTTP requests by method and status."},
			[]string{"method", "status"},
		),
		duration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_request_duration_seconds",
				Help:    "HTTP request latency in seconds by method.",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method"},
		),
	}
	reg.MustRegister(m.requests, m.duration)
	return m, reg
}

// middleware records a count and latency observation for each request.
func (m *metrics) middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			start := time.Now()
			next.ServeHTTP(ww, r)
			m.requests.WithLabelValues(r.Method, strconv.Itoa(ww.Status())).Inc()
			m.duration.WithLabelValues(r.Method).Observe(time.Since(start).Seconds())
		})
	}
}

func metricsHandler(reg *prometheus.Registry) http.Handler {
	// Gather the per-router HTTP metrics plus the global AI usage metrics.
	return promhttp.HandlerFor(prometheus.Gatherers{reg, observability.AIRegistry()}, promhttp.HandlerOpts{})
}
