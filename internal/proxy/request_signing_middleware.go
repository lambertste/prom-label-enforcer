package proxy

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// DefaultRequestSigningConfig returns a RequestSigningConfig with safe defaults.
func DefaultRequestSigningConfig() RequestSigningConfig {
	return RequestSigningConfig{
		SecretKey:   "",
		HeaderName:  "X-Signature",
		TimestampHeader: "X-Timestamp",
		MaxSkew:     5 * time.Minute,
		Enforce:     true,
	}
}

// RequestSigningConfig configures HMAC-based request signature validation.
type RequestSigningConfig struct {
	SecretKey       string
	HeaderName      string
	TimestampHeader string
	MaxSkew         time.Duration
	Enforce         bool
	Registerer      prometheus.Registerer
}

// NewRequestSigningMiddleware returns middleware that validates HMAC signatures on incoming requests.
func NewRequestSigningMiddleware(cfg RequestSigningConfig, next http.Handler) http.Handler {
	if cfg.HeaderName == "" {
		cfg.HeaderName = DefaultRequestSigningConfig().HeaderName
	}
	if cfg.TimestampHeader == "" {
		cfg.TimestampHeader = DefaultRequestSigningConfig().TimestampHeader
	}
	if cfg.MaxSkew == 0 {
		cfg.MaxSkew = DefaultRequestSigningConfig().MaxSkew
	}

	metrics := NewRequestSigningMetrics(cfg.Registerer)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if cfg.SecretKey == "" || !cfg.Enforce {
			metrics.RecordBypassed()
			next.ServeHTTP(w, r)
			return
		}

		sig := r.Header.Get(cfg.HeaderName)
		ts := r.Header.Get(cfg.TimestampHeader)

		if sig == "" || ts == "" {
			metrics.RecordDenied()
			http.Error(w, "missing signature or timestamp", http.StatusUnauthorized)
			return
		}

		t, err := time.Parse(time.RFC3339, ts)
		if err != nil || time.Since(t).Abs() > cfg.MaxSkew {
			metrics.RecordDenied()
			http.Error(w, "timestamp out of range", http.StatusUnauthorized)
			return
		}

		expected := computeHMAC(cfg.SecretKey, ts+":"+r.URL.Path)
		if !hmac.Equal([]byte(sig), []byte(expected)) {
			metrics.RecordDenied()
			http.Error(w, "invalid signature", http.StatusUnauthorized)
			return
		}

		metrics.RecordAllowed()
		next.ServeHTTP(w, r)
	})
}

func computeHMAC(secret, data string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}
