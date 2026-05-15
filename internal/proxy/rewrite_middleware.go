package proxy

import (
	"net/http"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

// RewriteConfig holds configuration for the URL rewrite middleware.
type RewriteConfig struct {
	// Rules maps path prefixes to replacement prefixes.
	// e.g. "/api/v1" -> "/v1" rewrites "/api/v1/metrics" to "/v1/metrics"
	Rules map[string]string
	Registerer prometheus.Registerer
}

// DefaultRewriteConfig returns a RewriteConfig with sensible defaults.
func DefaultRewriteConfig() RewriteConfig {
	return RewriteConfig{
		Rules:      map[string]string{},
		Registerer: prometheus.DefaultRegisterer,
	}
}

// NewRewriteMiddleware returns middleware that rewrites request URL paths
// according to the configured prefix rules before passing to next.
func NewRewriteMiddleware(cfg RewriteConfig, next http.Handler) http.Handler {
	if cfg.Registerer == nil {
		cfg.Registerer = prometheus.DefaultRegisterer
	}
	metrics := NewRewriteMetrics(cfg.Registerer)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		original := r.URL.Path
		rewritten := false

		for prefix, replacement := range cfg.Rules {
			if strings.HasPrefix(original, prefix) {
				r2 := r.Clone(r.Context())
				r2.URL.Path = replacement + strings.TrimPrefix(original, prefix)
				if r2.URL.RawPath != "" {
					r2.URL.RawPath = replacement + strings.TrimPrefix(r2.URL.RawPath, prefix)
				}
				metrics.RecordRewritten()
				next.ServeHTTP(w, r2)
				rewritten = true
				break
			}
		}

		if !rewritten {
			metrics.RecordPassthrough()
			next.ServeHTTP(w, r)
		}
	})
}
