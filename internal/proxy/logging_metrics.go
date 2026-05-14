package proxy

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// LoggingMetrics holds Prometheus metrics for the logging middleware.
type LoggingMetrics struct {
	requestTotal   *prometheus.CounterVec
	durationSeconds *prometheus.HistogramVec
	slowTotal      prometheus.Counter
}

// NewLoggingMetrics registers and returns LoggingMetrics.
func NewLoggingMetrics(reg prometheus.Registerer) *LoggingMetrics {
	if reg == nil {
		reg = prometheus.NewRegistry()
	}
	m := &LoggingMetrics{
		requestTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "proxy_logged_requests_total",
			Help: "Total number of logged HTTP requests.",
		}, []string{"method", "status"}),
		durationSeconds: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "proxy_request_duration_seconds",
			Help:    "Histogram of HTTP request durations.",
			Buckets: prometheus.DefBuckets,
		}, []string{"method"}),
		slowTotal: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "proxy_slow_requests_total",
			Help: "Total number of slow HTTP requests.",
		}),
	}
	reg.MustRegister(m.requestTotal, m.durationSeconds, m.slowTotal)
	return m
}

// RecordRequest records a completed request into the metrics.
func (m *LoggingMetrics) RecordRequest(method string, status int, d time.Duration, slow bool) {
	if m == nil {
		return
	}
	m.requestTotal.WithLabelValues(method, strconv.Itoa(status)).Inc()
	m.durationSeconds.WithLabelValues(method).Observe(d.Seconds())
	if slow {
		m.slowTotal.Inc()
	}
}

// StatusText is a helper used in tests.
func StatusText(code int) string { return http.StatusText(code) }
