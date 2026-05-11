package proxy

import (
	"context"
	"net/http"
	"time"
)

// TimeoutConfig holds configuration for the timeout middleware.
type TimeoutConfig struct {
	// RequestTimeout is the maximum duration allowed for a single request.
	RequestTimeout time.Duration
}

// DefaultTimeoutConfig returns a TimeoutConfig with sensible defaults.
func DefaultTimeoutConfig() TimeoutConfig {
	return TimeoutConfig{
		RequestTimeout: 30 * time.Second,
	}
}

// NewTimeoutMiddleware returns an HTTP middleware that cancels requests
// exceeding the configured timeout duration. When the deadline is exceeded
// the middleware responds with 503 Service Unavailable.
func NewTimeoutMiddleware(cfg TimeoutConfig) func(http.Handler) http.Handler {
	if cfg.RequestTimeout <= 0 {
		cfg = DefaultTimeoutConfig()
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), cfg.RequestTimeout)
			defer cancel()

			done := make(chan struct{})
			var panicVal interface{}

			rw := &timeoutResponseWriter{ResponseWriter: w, code: http.StatusOK}

			go func() {
				defer func() {
					panicVal = recover()
					close(done)
				}()
				next.ServeHTTP(rw, r.WithContext(ctx))
			}()

			select {
			case <-done:
				if panicVal != nil {
					panic(panicVal)
				}
			case <-ctx.Done():
				w.Header().Set("Content-Type", "text/plain; charset=utf-8")
				w.WriteHeader(http.StatusServiceUnavailable)
				_, _ = w.Write([]byte("request timeout\n"))
			}
		})
	}
}

// timeoutResponseWriter wraps http.ResponseWriter to capture the status code
// and prevent writes after the timeout has fired.
type timeoutResponseWriter struct {
	http.ResponseWriter
	code    int
	wroteHeader bool
}

func (tw *timeoutResponseWriter) WriteHeader(code int) {
	if !tw.wroteHeader {
		tw.code = code
		tw.wroteHeader = true
		tw.ResponseWriter.WriteHeader(code)
	}
}

func (tw *timeoutResponseWriter) Write(b []byte) (int, error) {
	if !tw.wroteHeader {
		tw.WriteHeader(http.StatusOK)
	}
	return tw.ResponseWriter.Write(b)
}
