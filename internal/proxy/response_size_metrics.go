package proxy

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// ResponseSizeMetrics holds Prometheus metrics for tracking response body sizes.
type ResponseSizeMetrics struct {
	responseSizeBytes prometheus.Histogram
}

// NewResponseSizeMetrics creates a new ResponseSizeMetrics using the provided registerer.
// Falls back to the default prometheus registerer if reg is nil.
func NewResponseSizeMetrics(reg prometheus.Registerer) *ResponseSizeMetrics {
	if reg == nil {
		reg = prometheus.DefaultRegisterer
	}
	f := promauto.With(reg)
	return &ResponseSizeMetrics{
		responseSizeBytes: f.NewHistogram(prometheus.HistogramOpts{
			Namespace: "prom_label_enforcer",
			Subsystem: "proxy",
			Name:      "response_size_bytes",
			Help:      "Distribution of response body sizes in bytes.",
			Buckets:   prometheus.ExponentialBuckets(64, 4, 8),
		}),
	}
}

// Record observes the given response size in bytes.
// It is a no-op when the receiver is nil.
func (m *ResponseSizeMetrics) Record(sizeBytes int) {
	if m == nil {
		return
	}
	m.responseSizeBytes.Observe(float64(sizeBytes))
}
