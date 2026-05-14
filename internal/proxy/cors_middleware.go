package proxy

import (
	"net/http"
	"strconv"
	"strings"
)

// CORSConfig holds configuration for the CORS middleware.
type CORSConfig struct {
	AllowedOrigins []string
	AllowedMethods []string
	AllowedHeaders []string
	MaxAge         int
}

// DefaultCORSConfig returns a CORSConfig with sensible defaults.
func DefaultCORSConfig() CORSConfig {
	return CORSConfig{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders: []string{"Content-Type", "Authorization"},
		MaxAge:         3600,
	}
}

// NewCORSMiddleware returns an HTTP middleware that adds CORS headers.
func NewCORSMiddleware(cfg CORSConfig) func(http.Handler) http.Handler {
	if len(cfg.AllowedOrigins) == 0 {
		cfg = DefaultCORSConfig()
	}

	allowedOriginSet := make(map[string]struct{}, len(cfg.AllowedOrigins))
	for _, o := range cfg.AllowedOrigins {
		allowedOriginSet[o] = struct{}{}
	}

	methods := strings.Join(cfg.AllowedMethods, ", ")
	headers := strings.Join(cfg.AllowedHeaders, ", ")
	maxAge := strconv.Itoa(cfg.MaxAge)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			allowed := false

			if _, ok := allowedOriginSet["*"]; ok {
				w.Header().Set("Access-Control-Allow-Origin", "*")
				allowed = true
			} else if _, ok := allowedOriginSet[origin]; ok && origin != "" {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				allowed = true
			}

			if allowed {
				w.Header().Set("Access-Control-Allow-Methods", methods)
				w.Header().Set("Access-Control-Allow-Headers", headers)
				w.Header().Set("Access-Control-Max-Age", maxAge)
			}

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
