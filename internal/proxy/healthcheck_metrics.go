package proxy

import "github.com/prometheus/client_golang/prometheus"

// HealthCheckMetrics holds Prometheus counters for health probe requests.
type HealthCheckMetrics struct {
	livenessTotal  prometheus.Counter
	readinessTotal *prometheus.CounterVec
}

// NewHealthCheckMetrics registers and returns HealthCheckMetrics.
// If reg is nil, prometheus.DefaultRegisterer is used.
func NewHealthCheckMetrics(reg prometheus.Registerer) *HealthCheckMetrics {
	if reg == nil {
		reg = prometheus.DefaultRegisterer
	}
	m := &HealthCheckMetrics{
		livenessTotal: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: "prom_label_enforcer",
			Subsystem: "healthcheck",
			Name:      "liveness_requests_total",
			Help:      "Total number of liveness probe requests handled.",
		}),
		readinessTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "prom_label_enforcer",
			Subsystem: "healthcheck",
			Name:      "readiness_requests_total",
			Help:      "Total number of readiness probe requests handled, by result.",
		}, []string{"result"}),
	}
	reg.MustRegister(m.livenessTotal, m.readinessTotal)
	return m
}

// RecordLiveness increments the liveness probe counter.
func (m *HealthCheckMetrics) RecordLiveness() {
	if m == nil {
		return
	}
	m.livenessTotal.Inc()
}

// RecordReadiness increments the readiness probe counter for the given result
// ("ready" or "not_ready").
func (m *HealthCheckMetrics) RecordReadiness(result string) {
	if m == nil {
		return
	}
	m.readinessTotal.WithLabelValues(result).Inc()
}
