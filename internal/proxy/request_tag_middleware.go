package proxy

import (
	"net/http"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

// DefaultRequestTagConfig returns a RequestTagConfig with sensible defaults.
func DefaultRequestTagConfig() RequestTagConfig {
	return RequestTagConfig{
		HeaderName:  "X-Request-Tags",
		ContextKey:  "request_tags",
		MaxTags:     10,
		Registerer:  prometheus.DefaultRegisterer,
	}
}

// RequestTagConfig configures the request tagging middleware.
type RequestTagConfig struct {
	// HeaderName is the HTTP header from which tags are read.
	HeaderName string
	// ContextKey is the key used to store tags in the request context.
	ContextKey string
	// MaxTags is the maximum number of tags accepted per request.
	MaxTags int
	// Registerer is the Prometheus registerer for metrics.
	Registerer prometheus.Registerer
}

// NewRequestTagMiddleware creates middleware that reads comma-separated tags
// from a request header, validates them, and stores them in the context.
func NewRequestTagMiddleware(cfg RequestTagConfig, next http.Handler) http.Handler {
	if cfg.HeaderName == "" {
		cfg.HeaderName = DefaultRequestTagConfig().HeaderName
	}
	if cfg.MaxTags <= 0 {
		cfg.MaxTags = DefaultRequestTagConfig().MaxTags
	}
	if cfg.Registerer == nil {
		cfg.Registerer = prometheus.DefaultRegisterer
	}

	metrics := NewRequestTagMetrics(cfg.Registerer)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw := r.Header.Get(cfg.HeaderName)
		var tags []string
		if raw != "" {
			for _, t := range strings.Split(raw, ",") {
				t = strings.TrimSpace(t)
				if t != "" {
					tags = append(tags, t)
				}
			}
			if len(tags) > cfg.MaxTags {
				tags = tags[:cfg.MaxTags]
				metrics.RecordTruncated()
			} else {
				metrics.RecordTagged(len(tags))
			}
		} else {
			metrics.RecordUntagged()
		}

		ctx := withRequestTags(r.Context(), tags)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
