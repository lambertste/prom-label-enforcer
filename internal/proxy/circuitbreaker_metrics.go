package proxy

import (
	"github.com/prometheus/client_golang/prometheus"
)

// CircuitBreakerMetrics holds Prometheus metrics for the circuit breaker.
type CircuitBreakerMetrics struct {
	stateGauge    *prometheus.GaugeVec
	tripTotal     prometheus.Counter
	allowedTotal  prometheus.Counter
	rejectedTotal prometheus.Counter
}

// NewCircuitBreakerMetrics creates and registers circuit breaker metrics.
func NewCircuitBreakerMetrics(reg prometheus.Registerer) *CircuitBreakerMetrics {
	if reg == nil {
		reg = prometheus.NewRegistry()
	}
	m := &CircuitBreakerMetrics{
		stateGauge: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "prom_label_enforcer",
			Subsystem: "circuit_breaker",
			Name:      "state",
			Help:      "Current circuit breaker state (0=closed, 1=open, 2=half-open).",
		}, []string{"name"}),
		tripTotal: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: "prom_label_enforcer",
			Subsystem: "circuit_breaker",
			Name:      "trips_total",
			Help:      "Total number of times the circuit has opened.",
		}),
		allowedTotal: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: "prom_label_enforcer",
			Subsystem: "circuit_breaker",
			Name:      "allowed_total",
			Help:      "Total requests allowed through the circuit breaker.",
		}),
		rejectedTotal: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: "prom_label_enforcer",
			Subsystem: "circuit_breaker",
			Name:      "rejected_total",
			Help:      "Total requests rejected by the circuit breaker.",
		}),
	}
	reg.MustRegister(m.stateGauge, m.tripTotal, m.allowedTotal, m.rejectedTotal)
	return m
}

// RecordState updates the state gauge for the named circuit.
func (m *CircuitBreakerMetrics) RecordState(name string, state CircuitState) {
	if m == nil {
		return
	}
	m.stateGauge.WithLabelValues(name).Set(float64(state))
}

// RecordTrip increments the trip counter.
func (m *CircuitBreakerMetrics) RecordTrip() {
	if m == nil {
		return
	}
	m.tripTotal.Inc()
}

// RecordAllowed increments the allowed counter.
func (m *CircuitBreakerMetrics) RecordAllowed() {
	if m == nil {
		return
	}
	m.allowedTotal.Inc()
}

// RecordRejected increments the rejected counter.
func (m *CircuitBreakerMetrics) RecordRejected() {
	if m == nil {
		return
	}
	m.rejectedTotal.Inc()
}
