package proxy

import (
	"github.com/prometheus/client_golang/prometheus"
)

// RedirectMetrics holds Prometheus counters for redirect middleware.
type RedirectMetrics struct {
	redirects   *prometheus.CounterVec
	passthroughs prometheus.Counter
}

// NewRedirectMetrics registers and returns RedirectMetrics.
// If reg is nil, prometheus.DefaultRegisterer is used.
func NewRedirectMetrics(reg prometheus.Registerer) *RedirectMetrics {
	if reg == nil {
		reg = prometheus.DefaultRegisterer
	}

	redirects := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "prom_label_enforcer",
		Subsystem: "redirect",
		Name:      "redirects_total",
		Help:      "Total number of redirects performed, labelled by reason.",
	}, []string{"reason"})

	passthroughs := prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "prom_label_enforcer",
		Subsystem: "redirect",
		Name:      "passthroughs_total",
		Help:      "Total number of requests passed through without redirect.",
	})

	_ = reg.Register(redirects)
	_ = reg.Register(passthroughs)

	return &RedirectMetrics{
		redirects:    redirects,
		passthroughs: passthroughs,
	}
}

// RecordRedirect increments the redirect counter for the given reason.
func (m *RedirectMetrics) RecordRedirect(reason string) {
	if m == nil {
		return
	}
	m.redirects.WithLabelValues(reason).Inc()
}

// RecordPassthrough increments the passthrough counter.
func (m *RedirectMetrics) RecordPassthrough() {
	if m == nil {
		return
	}
	m.passthroughs.Inc()
}
