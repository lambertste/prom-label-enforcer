package proxy

import (
	"log/slog"
	"net/http"
	"time"
)

// DefaultLoggingConfig returns a LoggingConfig with sensible defaults.
func DefaultLoggingConfig() LoggingConfig {
	return LoggingConfig{
		Logger:        slog.Default(),
		LogBody:       false,
		SlowThreshold: 500 * time.Millisecond,
	}
}

// LoggingConfig holds configuration for the logging middleware.
type LoggingConfig struct {
	Logger        *slog.Logger
	LogBody       bool
	SlowThreshold time.Duration
}

type loggingResponseWriter struct {
	http.ResponseWriter
	status int
	written bool
}

func (lw *loggingResponseWriter) WriteHeader(code int) {
	if !lw.written {
		lw.status = code
		lw.written = true
		lw.ResponseWriter.WriteHeader(code)
	}
}

// NewLoggingMiddleware returns an HTTP middleware that logs each request.
func NewLoggingMiddleware(cfg LoggingConfig, metrics *LoggingMetrics) func(http.Handler) http.Handler {
	if cfg.Logger == nil {
		cfg.Logger = slog.Default()
	}
	if cfg.SlowThreshold == 0 {
		cfg.SlowThreshold = DefaultLoggingConfig().SlowThreshold
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			lw := &loggingResponseWriter{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(lw, r)
			duration := time.Since(start)
			slow := duration >= cfg.SlowThreshold
			cfg.Logger.Info("request",
				"method", r.Method,
				"path", r.URL.Path,
				"status", lw.status,
				"duration_ms", duration.Milliseconds(),
				"slow", slow,
			)
			if metrics != nil {
				metrics.RecordRequest(r.Method, lw.status, duration, slow)
			}
		})
	}
}
