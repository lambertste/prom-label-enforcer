package proxy

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
)

// DefaultBodyLimitConfig returns a BodyLimitConfig with sensible defaults.
func DefaultBodyLimitConfig() BodyLimitConfig {
	return BodyLimitConfig{
		MaxBytes: 1 << 20, // 1 MiB
	}
}

// BodyLimitConfig holds configuration for the body limit middleware.
type BodyLimitConfig struct {
	// MaxBytes is the maximum number of bytes allowed in a request body.
	// Requests exceeding this limit receive a 413 response.
	MaxBytes int64
}

// NewBodyLimitMiddleware returns an http.Handler that rejects requests whose
// body exceeds cfg.MaxBytes. A zero MaxBytes falls back to the default.
func NewBodyLimitMiddleware(next http.Handler, cfg BodyLimitConfig, reg prometheus.Registerer) http.Handler {
	if cfg.MaxBytes <= 0 {
		cfg = DefaultBodyLimitConfig()
	}

	m := NewBodyLimitMetrics(reg)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.ContentLength > cfg.MaxBytes {
			m.RecordRejected()
			http.Error(w, "request body too large", http.StatusRequestEntityTooLarge)
			return
		}

		// Cap the body reader even when Content-Length is absent or untrustworthy.
		r.Body = http.MaxBytesReader(w, r.Body, cfg.MaxBytes)
		m.RecordAllowed()
		next.ServeHTTP(w, r)
	})
}
