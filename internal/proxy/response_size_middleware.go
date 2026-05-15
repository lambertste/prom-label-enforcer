package proxy

import (
	"bytes"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
)

// DefaultResponseSizeMiddlewareConfig returns a ResponseSizeMetrics backed by
// the default Prometheus registerer.
func DefaultResponseSizeMiddlewareConfig() *ResponseSizeMetrics {
	return NewResponseSizeMetrics(prometheus.DefaultRegisterer)
}

// capturingResponseWriter wraps http.ResponseWriter and captures the number of
// bytes written to the body.
type capturingResponseWriter struct {
	http.ResponseWriter
	buf         bytes.Buffer
	bytesWritten int
}

func (c *capturingResponseWriter) Write(b []byte) (int, error) {
	n, err := c.ResponseWriter.Write(b)
	c.bytesWritten += n
	return n, err
}

// NewResponseSizeMiddleware returns an http.Handler middleware that records the
// size of each response body using the supplied metrics. If metrics is nil a
// default instance is created.
func NewResponseSizeMiddleware(metrics *ResponseSizeMetrics, next http.Handler) http.Handler {
	if metrics == nil {
		metrics = DefaultResponseSizeMiddlewareConfig()
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cw := &capturingResponseWriter{ResponseWriter: w}
		next.ServeHTTP(cw, r)
		metrics.Record(cw.bytesWritten)
	})
}
