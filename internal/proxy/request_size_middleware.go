package proxy

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
)

// RequestSizeMiddlewareConfig holds configuration for the request size tracking middleware.
type RequestSizeMiddlewareConfig struct {
	Registerer prometheus.Registerer
}

// DefaultRequestSizeMiddlewareConfig returns a config with default values.
func DefaultRequestSizeMiddlewareConfig() RequestSizeMiddlewareConfig {
	return RequestSizeMiddlewareConfig{
		Registerer: prometheus.DefaultRegisterer,
	}
}

type requestSizeMiddleware struct {
	metrics *RequestSizeMetrics
	next    http.Handler
}

// NewRequestSizeMiddleware returns an http.Handler that observes the size of
// each incoming request body and records it via RequestSizeMetrics.
func NewRequestSizeMiddleware(cfg RequestSizeMiddlewareConfig, next http.Handler) http.Handler {
	if cfg.Registerer == nil {
		cfg.Registerer = prometheus.DefaultRegisterer
	}
	return &requestSizeMiddleware{
		metrics: NewRequestSizeMetrics(cfg.Registerer),
		next:    next,
	}
}

func (m *requestSizeMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	size := float64(0)
	if r.ContentLength > 0 {
		size = float64(r.ContentLength)
	}
	if m.metrics != nil {
		m.metrics.RequestSize.Observe(size)
	}
	m.next.ServeHTTP(w, r)
}
