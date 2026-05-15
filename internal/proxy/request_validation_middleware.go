package proxy

import (
	"net/http"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

// DefaultRequestValidationConfig returns a RequestValidationConfig with sensible defaults.
func DefaultRequestValidationConfig() RequestValidationConfig {
	return RequestValidationConfig{
		AllowedMethods:  []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodOptions},
		MaxHeaderBytes:  8192,
		RequireContentType: false,
	}
}

// RequestValidationConfig holds configuration for request validation middleware.
type RequestValidationConfig struct {
	AllowedMethods     []string
	MaxHeaderBytes     int
	RequireContentType bool
	Registerer         prometheus.Registerer
}

// NewRequestValidationMiddleware returns an HTTP middleware that validates incoming
// requests against method allowlists, header size limits, and optional content-type
// requirements. Rejected requests receive a 400 or 405 response.
func NewRequestValidationMiddleware(cfg RequestValidationConfig, next http.Handler) http.Handler {
	if cfg.AllowedMethods == nil {
		defaults := DefaultRequestValidationConfig()
		cfg.AllowedMethods = defaults.AllowedMethods
	}
	if cfg.MaxHeaderBytes == 0 {
		cfg.MaxHeaderBytes = DefaultRequestValidationConfig().MaxHeaderBytes
	}

	metrics := NewRequestValidationMetrics(cfg.Registerer)

	allowed := make(map[string]struct{}, len(cfg.AllowedMethods))
	for _, m := range cfg.AllowedMethods {
		allowed[strings.ToUpper(m)] = struct{}{}
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := allowed[strings.ToUpper(r.Method)]; !ok {
			metrics.RecordRejected("method_not_allowed")
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		headerSize := 0
		for name, vals := range r.Header {
			headerSize += len(name)
			for _, v := range vals {
				headerSize += len(v)
			}
		}
		if headerSize > cfg.MaxHeaderBytes {
			metrics.RecordRejected("headers_too_large")
			http.Error(w, "request headers too large", http.StatusBadRequest)
			return
		}

		if cfg.RequireContentType && r.Header.Get("Content-Type") == "" {
			metrics.RecordRejected("missing_content_type")
			http.Error(w, "Content-Type header required", http.StatusBadRequest)
			return
		}

		metrics.RecordAllowed()
		next.ServeHTTP(w, r)
	})
}
