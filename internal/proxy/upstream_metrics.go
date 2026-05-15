package proxy

import "github.com/prometheus/client_golang/prometheus"

// upstreamMetrics holds Prometheus counters for upstream proxy traffic.
type upstreamMetrics struct {
	successes *prometheus.CounterVec
	errors    *prometheus.CounterVec
}

// NewUpstreamMetrics registers and returns upstream proxy metrics.
// A nil registerer falls back to prometheus.DefaultRegisterer.
func NewUpstreamMetrics(reg prometheus.Registerer) *upstreamMetrics {
	if reg == nil {
		reg = prometheus.DefaultRegisterer
	}

	successes := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "prom_label_enforcer",
		Subsystem: "upstream",
		Name:      "requests_total",
		Help:      "Total successful upstream requests.",
	}, []string{"path"})

	errors := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "prom_label_enforcer",
		Subsystem: "upstream",
		Name:      "errors_total",
		Help:      "Total failed upstream requests.",
	}, []string{"path"})

	for _, c := range []prometheus.Collector{successes, errors} {
		if err := reg.Register(c); err != nil {
			if are, ok := err.(prometheus.AlreadyRegisteredError); ok {
				switch c.(type) {
				case *prometheus.CounterVec:
					_ = are.ExistingCollector
				}
			}
		}
	}

	return &upstreamMetrics{successes: successes, errors: errors}
}

func (m *upstreamMetrics) recordSuccess(path string) {
	if m == nil {
		return
	}
	m.successes.WithLabelValues(path).Inc()
}

func (m *upstreamMetrics) recordError(path string) {
	if m == nil {
		return
	}
	m.errors.WithLabelValues(path).Inc()
}
