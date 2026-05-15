package proxy

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
)

// ResponseHeaderConfig holds configuration for the response header middleware.
type ResponseHeaderConfig struct {
	// Headers to inject into every response.
	StaticHeaders map[string]string
	// Headers to remove from upstream responses.
	StripHeaders []string
	Registerer   prometheus.Registerer
}

// DefaultResponseHeaderConfig returns a config with common security response headers.
func DefaultResponseHeaderConfig() ResponseHeaderConfig {
	return ResponseHeaderConfig{
		StaticHeaders: map[string]string{
			"X-Content-Type-Options":  "nosniff",
			"X-Frame-Options":         "DENY",
			"Referrer-Policy":         "strict-origin-when-cross-origin",
		},
		StripHeaders: []string{"Server", "X-Powered-By"},
		Registerer:   prometheus.DefaultRegisterer,
	}
}

type responseHeaderMetrics struct {
	injected  prometheus.Counter
	stripped  prometheus.Counter
}

func newResponseHeaderMetrics(reg prometheus.Registerer) *responseHeaderMetrics {
	if reg == nil {
		reg = prometheus.DefaultRegisterer
	}
	return &responseHeaderMetrics{
		injected: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "proxy_response_headers_injected_total",
			Help: "Total number of response headers injected.",
		}),
		stripped: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "proxy_response_headers_stripped_total",
			Help: "Total number of response headers stripped.",
		}),
	}
}

// NewResponseHeaderMiddleware returns middleware that injects and strips response headers.
func NewResponseHeaderMiddleware(cfg ResponseHeaderConfig, next http.Handler) http.Handler {
	if cfg.StaticHeaders == nil {
		cfg.StaticHeaders = map[string]string{}
	}
	m := newResponseHeaderMetrics(cfg.Registerer)
	_ = cfg.Registerer

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rw := &responseHeaderWriter{ResponseWriter: w, headers: cfg.StaticHeaders, strip: cfg.Strip, metrics: m}
		next.ServeHTTP(rw, r)
		if !rw.written {
			rw.applyHeaders()
		}
	})
}

type responseHeaderWriter struct {
	http.ResponseWriter
	headers map[string]string
	strip   []string
	metrics *responseHeaderMetrics
	written bool
}

func (rw *responseHeaderWriter) applyHeaders() {
	h := rw.ResponseWriter.Header()
	for _, key := range rw.strip {
		if h.Get(key) != "" {
			h.Del(key)
			rw.metrics.stripped.Inc()
		}
	}
	for k, v := range rw.headers {
		h.Set(k, v)
		rw.metrics.injected.Inc()
	}
}

func (rw *responseHeaderWriter) WriteHeader(code int) {
	if !rw.written {
		rw.written = true
		rw.applyHeaders()
	}
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseHeaderWriter) Write(b []byte) (int, error) {
	if !rw.written {
		rw.written = true
		rw.applyHeaders()
	}
	return rw.ResponseWriter.Write(b)
}
