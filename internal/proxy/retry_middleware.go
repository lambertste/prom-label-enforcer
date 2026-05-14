package proxy

import (
	"log"
	"net/http"
	"time"
)

// RetryConfig holds configuration for the retry middleware.
type RetryConfig struct {
	// MaxAttempts is the total number of attempts (including the first).
	MaxAttempts int
	// Delay is the wait time between consecutive attempts.
	Delay time.Duration
	// RetryableStatuses is the set of HTTP status codes that trigger a retry.
	RetryableStatuses map[int]bool
}

// DefaultRetryConfig returns a RetryConfig with sensible defaults.
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxAttempts: 3,
		Delay:       100 * time.Millisecond,
		RetryableStatuses: map[int]bool{
			http.StatusBadGateway:         true,
			http.StatusServiceUnavailable: true,
			http.StatusGatewayTimeout:     true,
		},
	}
}

// retryResponseWriter captures the status code written by the inner handler.
type retryResponseWriter struct {
	http.ResponseWriter
	statusCode int
	written    bool
}

func (rw *retryResponseWriter) WriteHeader(code int) {
	if !rw.written {
		rw.statusCode = code
		rw.written = true
		rw.ResponseWriter.WriteHeader(code)
	}
}

// NewRetryMiddleware returns an http.Handler that retries requests to next
// when the upstream responds with a configured retryable status code.
func NewRetryMiddleware(cfg RetryConfig, next http.Handler) http.Handler {
	if cfg.MaxAttempts <= 0 {
		cfg.MaxAttempts = DefaultRetryConfig().MaxAttempts
	}
	if cfg.Delay <= 0 {
		cfg.Delay = DefaultRetryConfig().Delay
	}
	if len(cfg.RetryableStatuses) == 0 {
		cfg.RetryableStatuses = DefaultRetryConfig().RetryableStatuses
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for attempt := 1; attempt <= cfg.MaxAttempts; attempt++ {
			rw := &retryResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}
			next.ServeHTTP(rw, r)

			if !cfg.RetryableStatuses[rw.statusCode] {
				return
			}

			if attempt < cfg.MaxAttempts {
				log.Printf("retry_middleware: attempt %d/%d failed with status %d, retrying after %s",
					attempt, cfg.MaxAttempts, rw.statusCode, cfg.Delay)
				time.Sleep(cfg.Delay)
			}
		}
	})
}
