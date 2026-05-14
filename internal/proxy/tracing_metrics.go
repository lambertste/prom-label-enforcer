package proxy

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// TracingMetrics holds Prometheus metrics for the tracing middleware.
type TracingMetrics struct {
	requestTotal    *prometheus.CounterVec
	requestDuration *prometheus.HistogramVec
}

// NewTracingMetrics creates and registers TracingMetrics with the given registerer.
// Falls back to prometheus.DefaultRegisterer if reg is nil.
func NewTracingMetrics(reg prometheus.Registerer) *TracingMetrics {
	if reg == nil {
		reg = prometheus.DefaultRegisterer
	}

	requestTotal := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "prom_label_enforcer",
		Subsystem: "proxy",
		Name:      "requests_total",
		Help:      "Total number of proxied requests partitioned by method and status code.",
	}, []string{"method", "status"})

	requestDuration := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "prom_label_enforcer",
		Subsystem: "proxy",
		Name:      "request_duration_seconds",
		Help:      "Histogram of proxied request durations.",
		Buckets:   prometheus.DefBuckets,
	}, []string{"method", "status"})

	reg.MustRegister(requestTotal, requestDuration)

	return &TracingMetrics{
		requestTotal:    requestTotal,
		requestDuration: requestDuration,
	}
}

// RecordRequest increments the request counter and observes the duration.
func (m *TracingMetrics) RecordRequest(method, status string, duration time.Duration) {
	if m == nil {
		return
	}
	m.requestTotal.WithLabelValues(method, status).Inc()
	m.requestDuration.WithLabelValues(method, status).Observe(duration.Seconds())
}
