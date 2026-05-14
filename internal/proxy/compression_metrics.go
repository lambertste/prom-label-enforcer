package proxy

import "github.com/prometheus/client_golang/prometheus"

// CompressionMetrics holds Prometheus metrics for the compression middleware.
type CompressionMetrics struct {
	CompressedTotal   prometheus.Counter
	UncompressedTotal prometheus.Counter
	BytesSaved        prometheus.Counter
}

// NewCompressionMetrics registers and returns CompressionMetrics using the given registerer.
// If reg is nil, prometheus.DefaultRegisterer is used.
func NewCompressionMetrics(reg prometheus.Registerer) *CompressionMetrics {
	if reg == nil {
		reg = prometheus.DefaultRegisterer
	}
	m := &CompressionMetrics{
		CompressedTotal: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "proxy_compression_compressed_total",
			Help: "Total number of responses that were gzip compressed.",
		}),
		UncompressedTotal: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "proxy_compression_uncompressed_total",
			Help: "Total number of responses that were NOT compressed (below min size or no accept-encoding).",
		}),
		BytesSaved: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "proxy_compression_bytes_saved_total",
			Help: "Approximate bytes saved by compression.",
		}),
	}
	reg.MustRegister(m.CompressedTotal, m.UncompressedTotal, m.BytesSaved)
	return m
}

// RecordCompressed increments the compressed counter.
func (m *CompressionMetrics) RecordCompressed() {
	if m == nil {
		return
	}
	m.CompressedTotal.Add(1)
}

// RecordUncompressed increments the uncompressed counter.
func (m *CompressionMetrics) RecordUncompressed() {
	if m == nil {
		return
	}
	m.UncompressedTotal.Add(1)
}

// RecordBytesSaved adds the given byte count to the bytes-saved counter.
func (m *CompressionMetrics) RecordBytesSaved(n float64) {
	if m == nil {
		return
	}
	m.BytesSaved.Add(n)
}
