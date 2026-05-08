package enforcer

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Metrics holds Prometheus counters for enforcer decisions.
type Metrics struct {
	Allowed  prometheus.Counter
	Rejected *prometheus.CounterVec
}

// NewMetrics registers and returns enforcer metrics using the given registerer.
// If reg is nil, prometheus.DefaultRegisterer is used.
func NewMetrics(reg prometheus.Registerer) *Metrics {
	if reg == nil {
		reg = prometheus.DefaultRegisterer
	}

	factory := promauto.With(reg)

	return &Metrics{
		Allowed: factory.NewCounter(prometheus.CounterOpts{
			Namespace: "prom_label_enforcer",
			Name:      "requests_allowed_total",
			Help:      "Total number of metric write requests that passed label enforcement.",
		}),
		Rejected: factory.NewCounterVec(prometheus.CounterOpts{
			Namespace: "prom_label_enforcer",
			Name:      "requests_rejected_total",
			Help:      "Total number of metric write requests rejected by label enforcement.",
		}, []string{"reason"}),
	}
}

// RecordAllowed increments the allowed counter.
func (m *Metrics) RecordAllowed() {
	if m == nil {
		return
	}
	m.Allowed.Inc()
}

// RecordRejected increments the rejected counter for the given reason.
// Common reasons: "missing_label", "disallowed_value".
func (m *Metrics) RecordRejected(reason string) {
	if m == nil {
		return
	}
	m.Rejected.WithLabelValues(reason).Inc()
}
