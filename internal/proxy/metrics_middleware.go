package proxy

import (
	"net/http"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
)

// MetricsMiddlewareConfig holds configuration for the metrics middleware.
type MetricsMiddlewareConfig struct {
	Registerer prometheus.Registerer
	Namespace  string
}

// DefaultMetricsMiddlewareConfig returns a MetricsMiddlewareConfig with sensible defaults.
func DefaultMetricsMiddlewareConfig() MetricsMiddlewareConfig {
	return MetricsMiddlewareConfig{
		Registerer: prometheus.DefaultRegisterer,
		Namespace:  "prom_label_enforcer",
	}
}

type metricsMiddleware struct {
	requestsTotal   *prometheus.CounterVec
	requestDuration *prometheus.HistogramVec
}

func newMetricsMiddleware(cfg MetricsMiddlewareConfig) *metricsMiddleware {
	if cfg.Registerer == nil {
		cfg.Registerer = prometheus.DefaultRegisterer
	}
	if cfg.Namespace == "" {
		cfg.Namespace = "prom_label_enforcer"
	}

	requestsTotal := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: cfg.Namespace,
		Subsystem: "http",
		Name:      "requests_total",
		Help:      "Total number of HTTP requests handled by the proxy.",
	}, []string{"method", "path", "status"})

	requestDuration := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: cfg.Namespace,
		Subsystem: "http",
		Name:      "request_duration_seconds",
		Help:      "HTTP request duration in seconds.",
		Buckets:   prometheus.DefBuckets,
	}, []string{"method", "path"})

	cfg.Registerer.MustRegister(requestsTotal, requestDuration)

	return &metricsMiddleware{
		requestsTotal:   requestsTotal,
		requestDuration: requestDuration,
	}
}

// NewMetricsMiddleware returns an HTTP middleware that records request counts and durations.
func NewMetricsMiddleware(cfg MetricsMiddlewareConfig, next http.Handler) http.Handler {
	m := newMetricsMiddleware(cfg)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		timer := prometheus.NewTimer(prometheus.ObserverFunc(func(v float64) {
			m.requestDuration.WithLabelValues(r.Method, r.URL.Path).Observe(v)
		}))
		defer timer.ObserveDuration()

		rw := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rw, r)

		m.requestsTotal.WithLabelValues(
			r.Method,
			r.URL.Path,
			strconv.Itoa(rw.status),
		).Inc()
	})
}

type statusRecorder struct {
	http.ResponseWriter
	status  int
	wrote   bool
}

func (sr *statusRecorder) WriteHeader(code int) {
	if !sr.wrote {
		sr.status = code
		sr.wrote = true
		sr.ResponseWriter.WriteHeader(code)
	}
}
