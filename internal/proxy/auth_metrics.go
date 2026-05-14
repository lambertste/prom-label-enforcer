package proxy

import "github.com/prometheus/client_golang/prometheus"

// AuthMetrics holds Prometheus counters for authentication events.
type AuthMetrics struct {
	allowed *prometheus.CounterVec
	denied  *prometheus.CounterVec
}

// NewAuthMetrics registers and returns AuthMetrics using reg.
// If reg is nil, prometheus.DefaultRegisterer is used.
func NewAuthMetrics(reg prometheus.Registerer) *AuthMetrics {
	if reg == nil {
		reg = prometheus.DefaultRegisterer
	}
	allowed := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "prom_label_enforcer",
		Subsystem: "auth",
		Name:      "requests_allowed_total",
		Help:      "Total number of requests that passed authentication.",
	}, []string{})
	denied := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "prom_label_enforcer",
		Subsystem: "auth",
		Name:      "requests_denied_total",
		Help:      "Total number of requests denied by authentication.",
	}, []string{"reason"})
	reg.MustRegister(allowed, denied)
	return &AuthMetrics{allowed: allowed, denied: denied}
}

// RecordAllowed increments the allowed counter.
func (m *AuthMetrics) RecordAllowed() {
	if m == nil {
		return
	}
	m.allowed.WithLabelValues().Inc()
}

// RecordDenied increments the denied counter with the given reason.
func (m *AuthMetrics) RecordDenied(reason string) {
	if m == nil {
		return
	}
	m.denied.WithLabelValues(reason).Inc()
}
