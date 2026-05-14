package proxy

import "github.com/prometheus/client_golang/prometheus"

// IPFilterMetrics holds Prometheus counters for IP filter decisions.
type IPFilterMetrics struct {
	allowed prometheus.Counter
	denied  prometheus.Counter
}

// NewIPFilterMetrics registers and returns IPFilterMetrics.
// If reg is nil, prometheus.DefaultRegisterer is used.
func NewIPFilterMetrics(reg prometheus.Registerer) *IPFilterMetrics {
	if reg == nil {
		reg = prometheus.DefaultRegisterer
	}

	allowed := prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "prom_label_enforcer",
		Subsystem: "ip_filter",
		Name:      "allowed_total",
		Help:      "Total number of requests allowed by IP filter.",
	})
	denied := prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "prom_label_enforcer",
		Subsystem: "ip_filter",
		Name:      "denied_total",
		Help:      "Total number of requests denied by IP filter.",
	})

	reg.MustRegister(allowed, denied)

	return &IPFilterMetrics{
		allowed: allowed,
		denied:  denied,
	}
}

// RecordAllowed increments the allowed counter.
func (m *IPFilterMetrics) RecordAllowed() {
	if m == nil {
		return
	}
	m.allowed.Inc()
}

// RecordDenied increments the denied counter.
func (m *IPFilterMetrics) RecordDenied() {
	if m == nil {
		return
	}
	m.denied.Inc()
}
