package proxy

import (
	"net/http"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

// DefaultRedirectConfig returns a RedirectConfig with sensible defaults.
func DefaultRedirectConfig() RedirectConfig {
	return RedirectConfig{
		HTTPSOnly:   false,
		StatusCode:  http.StatusMovedPermanently,
		StripPrefix: "",
	}
}

// RedirectConfig controls redirect middleware behaviour.
type RedirectConfig struct {
	// HTTPSOnly redirects all HTTP requests to HTTPS when true.
	HTTPSOnly bool
	// StatusCode is the HTTP status used for redirects (301 or 302).
	StatusCode int
	// StripPrefix removes a path prefix before redirecting.
	StripPrefix string
	// Registerer for metrics; defaults to prometheus.DefaultRegisterer.
	Registerer prometheus.Registerer
}

// NewRedirectMiddleware returns an http.Handler that performs path or scheme
// redirects according to cfg before delegating to next.
func NewRedirectMiddleware(cfg RedirectConfig, next http.Handler) http.Handler {
	if cfg.StatusCode == 0 {
		cfg.StatusCode = http.StatusMovedPermanently
	}
	if cfg.Registerer == nil {
		cfg.Registerer = prometheus.DefaultRegisterer
	}

	m := NewRedirectMetrics(cfg.Registerer)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if cfg.HTTPSOnly && r.TLS == nil && r.Header.Get("X-Forwarded-Proto") != "https" {
			target := "https://" + r.Host + r.RequestURI
			m.RecordRedirect("https_upgrade")
			http.Redirect(w, r, target, cfg.StatusCode)
			return
		}

		if cfg.StripPrefix != "" && strings.HasPrefix(r.URL.Path, cfg.StripPrefix) {
			newPath := strings.TrimPrefix(r.URL.Path, cfg.StripPrefix)
			if newPath == "" {
				newPath = "/"
			}
			target := newPath
			if r.URL.RawQuery != "" {
				target += "?" + r.URL.RawQuery
			}
			m.RecordRedirect("strip_prefix")
			http.Redirect(w, r, target, cfg.StatusCode)
			return
		}

		m.RecordPassthrough()
		next.ServeHTTP(w, r)
	})
}
