package proxy

import (
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// DefaultResponseTimeMiddlewareConfig returns a ResponseTimeMetrics using the
// default (global) Prometheus registerer.
func DefaultResponseTimeMiddlewareConfig() *ResponseTimeMetrics {
	return NewResponseTimeMetrics(prometheus.DefaultRegisterer)
}

// responseTimeWriter wraps http.ResponseWriter to capture the status code.
type responseTimeWriter struct {
	http.ResponseWriter
	status  int
	written bool
}

func (w *responseTimeWriter) WriteHeader(code int) {
	if !w.written {
		w.status = code
		w.written = true
		w.ResponseWriter.WriteHeader(code)
	}
}

func (w *responseTimeWriter) statusCode() int {
	if !w.written {
		return http.StatusOK
	}
	return w.status
}

// NewResponseTimeMiddleware returns an HTTP middleware that measures the
// latency of every request and records it via m. If m is nil the middleware
// is a no-op pass-through.
//
// The recorded latency includes the full time spent in downstream handlers,
// including any time waiting for the proxied backend to respond.
func NewResponseTimeMiddleware(m *ResponseTimeMetrics) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if m == nil {
				next.ServeHTTP(w, r)
				return
			}

			rw := &responseTimeWriter{ResponseWriter: w}
			start := time.Now()
			next.ServeHTTP(rw, r)
			latency := time.Since(start)

			m.RecordLatency(latency)
		})
	}
}
