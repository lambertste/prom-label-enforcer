package proxy

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// SanitizeMetrics holds Prometheus counters for the sanitize middleware.
type SanitizeMetrics struct {
	requests *prometheus.CounterVec
}

// NewSanitizeMetrics registers and returns SanitizeMetrics.
// If reg is nil, prometheus.DefaultRegisterer is used.
func NewSanitizeMetrics(reg prometheus.Registerer) *SanitizeMetrics {
	if reg == nil {
		reg = prometheus.DefaultRegisterer
	}
	factory := promauto.With(reg)
	return &SanitizeMetrics{
		requests: factory.NewCounterVec(prometheus.CounterOpts{
			Name: "proxy_sanitize_requests_total",
			Help: "Total number of requests processed by the sanitize middleware, by outcome.",
		}, []string{"outcome"}),
	}
}

// RecordSanitized increments the counter for requests that required sanitization.
func (m *SanitizeMetrics) RecordSanitized() {
	if m == nil {
		return
	}
	m.requests.WithLabelValues("sanitized").Inc()
}

// RecordClean increments the counter for requests that needed no sanitization.
func (m *SanitizeMetrics) RecordClean() {
	if m == nil {
		return
	}
	m.requests.WithLabelValues("clean").Inc()
}
