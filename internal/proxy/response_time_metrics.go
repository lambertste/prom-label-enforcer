package proxy

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// ResponseTimeMetrics holds Prometheus metrics for tracking HTTP response times.
type ResponseTimeMetrics struct {
	latency *prometheus.HistogramVec
	slowRequests *prometheus.CounterVec
}

// NewResponseTimeMetrics creates a new ResponseTimeMetrics instance registered
// against the provided Prometheus registerer. Falls back to the default
// registerer when reg is nil.
func NewResponseTimeMetrics(reg prometheus.Registerer) *ResponseTimeMetrics {
	if reg == nil {
		reg = prometheus.DefaultRegisterer
	}
	factory := promauto.With(reg)
	return &ResponseTimeMetrics{
		latency: factory.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: "prom_label_enforcer",
				Subsystem: "proxy",
				Name:      "response_duration_seconds",
				Help:      "Histogram of HTTP response latencies in seconds.",
				Buckets:   prometheus.DefBuckets,
			},
			[]string{"method", "status"},
		),
		slowRequests: factory.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "prom_label_enforcer",
				Subsystem: "proxy",
				Name:      "slow_requests_total",
				Help:      "Total number of requests that exceeded the slow threshold.",
			},
			[]string{"method"},
		),
	}
}

// RecordLatency records the duration of a completed request.
func (m *ResponseTimeMetrics) RecordLatency(method, status string, d time.Duration) {
	if m == nil {
		return
	}
	m.latency.WithLabelValues(method, status).Observe(d.Seconds())
}

// RecordSlow increments the slow-request counter for the given method.
func (m *ResponseTimeMetrics) RecordSlow(method string) {
	if m == nil {
		return
	}
	m.slowRequests.WithLabelValues(method).Inc()
}
