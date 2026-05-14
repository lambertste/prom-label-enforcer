package proxy

import (
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
)

// TracingConfig holds configuration for the tracing middleware.
type TracingConfig struct {
	// HeaderName is the HTTP header used to propagate the trace ID.
	HeaderName string
	// GenerateIfMissing creates a new trace ID if none is present in the request.
	GenerateIfMissing bool
}

// DefaultTracingConfig returns a TracingConfig with sensible defaults.
func DefaultTracingConfig() TracingConfig {
	return TracingConfig{
		HeaderName:        "X-Trace-Id",
		GenerateIfMissing: true,
	}
}

type tracingResponseWriter struct {
	http.ResponseWriter
	status int
}

func (w *tracingResponseWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

// NewTracingMiddleware returns an HTTP middleware that attaches a trace ID to
// each request and records per-request metrics.
func NewTracingMiddleware(cfg TracingConfig, metrics *TracingMetrics) func(http.Handler) http.Handler {
	if cfg.HeaderName == "" {
		cfg = DefaultTracingConfig()
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			traceID := r.Header.Get(cfg.HeaderName)
			if traceID == "" && cfg.GenerateIfMissing {
				traceID = uuid.NewString()
			}
			if traceID != "" {
				w.Header().Set(cfg.HeaderName, traceID)
				r = r.WithContext(withTraceID(r.Context(), traceID))
			}

			tw := &tracingResponseWriter{ResponseWriter: w, status: http.StatusOK}
			start := time.Now()
			next.ServeHTTP(tw, r)
			duration := time.Since(start)

			if metrics != nil {
				metrics.RecordRequest(r.Method, strconv.Itoa(tw.status), duration)
			}
		})
	}
}
