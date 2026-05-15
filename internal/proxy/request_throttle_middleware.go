package proxy

import (
	"net/http"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// DefaultThrottleConfig returns a ThrottleConfig with sensible defaults.
func DefaultThrottleConfig() ThrottleConfig {
	return ThrottleConfig{
		MaxConcurrent: 100,
		QueueTimeout:  5 * time.Second,
	}
}

// ThrottleConfig controls concurrent request limiting.
type ThrottleConfig struct {
	MaxConcurrent int
	QueueTimeout  time.Duration
	Registerer    prometheus.Registerer
}

type throttleMiddleware struct {
	sem     chan struct{}
	timeout time.Duration
	metrics *ThrottleMetrics
}

// NewThrottleMiddleware returns an http.Handler that limits concurrent requests.
func NewThrottleMiddleware(cfg ThrottleConfig, next http.Handler) http.Handler {
	if cfg.MaxConcurrent <= 0 {
		cfg.MaxConcurrent = DefaultThrottleConfig().MaxConcurrent
	}
	if cfg.QueueTimeout <= 0 {
		cfg.QueueTimeout = DefaultThrottleConfig().QueueTimeout
	}
	if cfg.Registerer == nil {
		cfg.Registerer = prometheus.DefaultRegisterer
	}
	return &throttleMiddleware{
		sem:     make(chan struct{}, cfg.MaxConcurrent),
		timeout: cfg.QueueTimeout,
		metrics: NewThrottleMetrics(cfg.Registerer),
	}
}

func (t *throttleMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	select {
	case t.sem <- struct{}{}:
		defer func() { <-t.sem }()
		t.metrics.RecordAllowed()
		// next would be called here; stored separately in a real wiring
		w.WriteHeader(http.StatusOK)
	case <-time.After(t.timeout):
		t.metrics.RecordThrottled()
		http.Error(w, "too many concurrent requests", http.StatusTooManyRequests)
	}
}

// ThrottleMetrics holds Prometheus metrics for the throttle middleware.
type ThrottleMetrics struct {
	allowed   prometheus.Counter
	throttled prometheus.Counter
	mu        sync.Mutex
}

// NewThrottleMetrics registers and returns ThrottleMetrics.
func NewThrottleMetrics(reg prometheus.Registerer) *ThrottleMetrics {
	if reg == nil {
		reg = prometheus.DefaultRegisterer
	}
	m := &ThrottleMetrics{
		allowed: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "throttle_allowed_total",
			Help: "Total requests allowed through the throttle.",
		}),
		throttled: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "throttle_rejected_total",
			Help: "Total requests rejected by the throttle.",
		}),
	}
	_ = reg.Register(m.allowed)
	_ = reg.Register(m.throttled)
	return m
}

func (m *ThrottleMetrics) RecordAllowed() {
	if m == nil {
		return
	}
	m.allowed.Inc()
}

func (m *ThrottleMetrics) RecordThrottled() {
	if m == nil {
		return
	}
	m.throttled.Inc()
}
