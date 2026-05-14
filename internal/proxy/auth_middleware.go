package proxy

import (
	"net/http"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// AuthConfig holds configuration for the authentication middleware.
type AuthConfig struct {
	// Tokens is the set of valid bearer tokens.
	Tokens map[string]struct{}
	// Realm is the WWW-Authenticate realm value.
	Realm string
}

// DefaultAuthConfig returns an AuthConfig with sensible defaults.
func DefaultAuthConfig() AuthConfig {
	return AuthConfig{
		Tokens: make(map[string]struct{}),
		Realm:  "prom-label-enforcer",
	}
}

// authMiddleware validates Bearer tokens on incoming requests.
type authMiddleware struct {
	cfg     AuthConfig
	metrics *AuthMetrics
}

// NewAuthMiddleware returns an HTTP middleware that enforces Bearer token auth.
func NewAuthMiddleware(cfg AuthConfig, reg prometheus.Registerer) func(http.Handler) http.Handler {
	if len(cfg.Tokens) == 0 {
		cfg.Tokens = make(map[string]struct{})
	}
	if cfg.Realm == "" {
		cfg.Realm = DefaultAuthConfig().Realm
	}
	am := &authMiddleware{
		cfg:     cfg,
		metrics: NewAuthMetrics(reg),
	}
	return am.handler
}

func (a *authMiddleware) handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := extractBearerToken(r)
		if token == "" {
			a.metrics.RecordDenied("missing_token")
			w.Header().Set("WWW-Authenticate", `Bearer realm="`+a.cfg.Realm+`"`)
			http.Error(w, "missing bearer token", http.StatusUnauthorized)
			return
		}
		if _, ok := a.cfg.Tokens[token]; !ok {
			a.metrics.RecordDenied("invalid_token")
			w.Header().Set("WWW-Authenticate", `Bearer realm="`+a.cfg.Realm+`", error="invalid_token"`)
			http.Error(w, "invalid bearer token", http.StatusUnauthorized)
			return
		}
		a.metrics.RecordAllowed()
		next.ServeHTTP(w, r)
	})
}

func extractBearerToken(r *http.Request) string {
	h := r.Header.Get("Authorization")
	if !strings.HasPrefix(h, "Bearer ") {
		return ""
	}
	return strings.TrimSpace(strings.TrimPrefix(h, "Bearer "))
}

// compile-time check
var _ = time.Now
