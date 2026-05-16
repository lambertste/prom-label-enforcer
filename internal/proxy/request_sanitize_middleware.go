package proxy

import (
	"net/http"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

// DefaultSanitizeConfig returns a SanitizeConfig with sensible defaults.
func DefaultSanitizeConfig() SanitizeConfig {
	return SanitizeConfig{
		StripHeaders: []string{"X-Forwarded-For", "X-Real-IP"},
		NormalizeHeaders: true,
		MaxHeaderValueLen: 4096,
	}
}

// SanitizeConfig controls request sanitization behaviour.
type SanitizeConfig struct {
	StripHeaders      []string
	NormalizeHeaders  bool
	MaxHeaderValueLen int
	Registerer        prometheus.Registerer
}

// NewSanitizeMiddleware returns an http.Handler that strips and normalizes
// incoming request headers before delegating to next.
func NewSanitizeMiddleware(cfg SanitizeConfig, next http.Handler) http.Handler {
	if cfg.MaxHeaderValueLen == 0 {
		cfg.MaxHeaderValueLen = DefaultSanitizeConfig().MaxHeaderValueLen
	}
	metrics := NewSanitizeMetrics(cfg.Registerer)
	stripSet := make(map[string]struct{}, len(cfg.StripHeaders))
	for _, h := range cfg.StripHeaders {
		stripSet[http.CanonicalHeaderKey(h)] = struct{}{}
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sanitized := false
		for key, vals := range r.Header {
			canon := http.CanonicalHeaderKey(key)
			if _, strip := stripSet[canon]; strip {
				r.Header.Del(key)
				sanitized = true
				continue
			}
			if cfg.NormalizeHeaders {
				for i, v := range vals {
					truncated := v
					if len(v) > cfg.MaxHeaderValueLen {
						truncated = v[:cfg.MaxHeaderValueLen]
						sanitized = true
					}
					vals[i] = strings.TrimSpace(truncated)
				}
				r.Header[canon] = vals
			}
		}
		if sanitized {
			metrics.RecordSanitized()
		} else {
			metrics.RecordClean()
		}
		next.ServeHTTP(w, r)
	})
}
