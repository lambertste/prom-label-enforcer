package proxy

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
)

// HeaderMiddlewareConfig holds configuration for the header injection middleware.
type HeaderMiddlewareConfig struct {
	// StaticHeaders are unconditionally added to every response.
	StaticHeaders map[string]string
	// RemoveRequestHeaders lists request header names to strip before forwarding.
	RemoveRequestHeaders []string
	// Registerer is used for metrics; defaults to prometheus.DefaultRegisterer.
	Registerer prometheus.Registerer
}

// DefaultHeaderMiddlewareConfig returns a config with sensible security defaults.
func DefaultHeaderMiddlewareConfig() HeaderMiddlewareConfig {
	return HeaderMiddlewareConfig{
		StaticHeaders: map[string]string{
			"X-Content-Type-Options":  "nosniff",
			"X-Frame-Options":         "DENY",
			"X-XSS-Protection":        "1; mode=block",
			"Referrer-Policy":         "strict-origin-when-cross-origin",
		},
		RemoveRequestHeaders: []string{"X-Forwarded-For"},
		Registerer:           prometheus.DefaultRegisterer,
	}
}

type headerMetrics struct {
	injected prometheus.Counter
	stripped prometheus.Counter
}

func newHeaderMetrics(reg prometheus.Registerer) *headerMetrics {
	if reg == nil {
		reg = prometheus.DefaultRegisterer
	}
	return &headerMetrics{
		injected: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "prom_enforcer_header_injected_total",
			Help: "Total number of response headers injected.",
		}),
		stripped: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "prom_enforcer_header_stripped_total",
			Help: "Total number of request headers stripped.",
		}),
	}
}

// NewHeaderMiddleware returns an http.Handler that injects static response
// headers and strips configured request headers on every request.
func NewHeaderMiddleware(cfg HeaderMiddlewareConfig, next http.Handler) http.Handler {
	if cfg.StaticHeaders == nil {
		cfg.StaticHeaders = map[string]string{}
	}
	m := newHeaderMetrics(cfg.Registerer)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for _, h := range cfg.RemoveRequestHeaders {
			if r.Header.Get(h) != "" {
				r.Header.Del(h)
				m.stripped.Inc()
			}
		}
		for k, v := range cfg.StaticHeaders {
			w.Header().Set(k, v)
			m.injected.Inc()
		}
		next.ServeHTTP(w, r)
	})
}
