package proxy

import (
	"github.com/prometheus/client_golang/prometheus"
)

// RequestSizeMetrics records metrics about incoming request body sizes.
type RequestSizeMetrics struct {
	bytesReceived prometheus.Histogram
	requestsTotal *prometheus.CounterVec
}

// NewRequestSizeMetrics creates a new RequestSizeMetrics, registering with reg.
// If reg is nil, prometheus.DefaultRegisterer is used.
func NewRequestSizeMetrics(reg prometheus.Registerer) *RequestSizeMetrics {
	if reg == nil {
		reg = prometheus.DefaultRegisterer
	}

	bytesReceived := prometheus.NewHistogram(prometheus.HistogramOpts{
		Namespace: "prom_label_enforcer",
		Subsystem: "request_size",
		Name:      "bytes",
		Help:      "Distribution of incoming request body sizes in bytes.",
		Buckets:   prometheus.ExponentialBuckets(64, 4, 8),
	})

	requestsTotal := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "prom_label_enforcer",
		Subsystem: "request_size",
		Name:      "requests_total",
		Help:      "Total number of requests categorised by size bucket.",
	}, []string{"bucket"})

	reg.MustRegister(bytesReceived, requestsTotal)

	return &RequestSizeMetrics{
		bytesReceived: bytesReceived,
		requestsTotal: requestsTotal,
	}
}

// Record observes the given byte count and increments the appropriate bucket counter.
func (m *RequestSizeMetrics) Record(bytes int64) {
	if m == nil {
		return
	}
	m.bytesReceived.Observe(float64(bytes))

	var bucket string
	switch {
	case bytes == 0:
		bucket = "empty"
	case bytes < 1024:
		bucket = "small"
	case bytes < 1024*1024:
		bucket = "medium"
	default:
		bucket = "large"
	}
	m.requestsTotal.WithLabelValues(bucket).Inc()
}
