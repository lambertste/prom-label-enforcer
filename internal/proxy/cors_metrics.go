package proxy

import "github.com/prometheus/client_golang/prometheus"

// CORSMetrics holds Prometheus counters for CORS middleware events.
type CORSMetrics struct {
	allowedRequests  prometheus.Counter
	rejectedRequests prometheus.Counter
	preflightRequests prometheus.Counter
}

// NewCORSMetrics creates and registers CORSMetrics with the given registerer.
// Falls back to prometheus.DefaultRegisterer if reg is nil.
func NewCORSMetrics(reg prometheus.Registerer) *CORSMetrics {
	if reg == nil {
		reg = prometheus.DefaultRegisterer
	}

	m := &CORSMetrics{
		allowedRequests: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: "prom_label_enforcer",
			Subsystem: "cors",
			Name:      "allowed_requests_total",
			Help:      "Total number of CORS requests with an allowed origin.",
		}),
		rejectedRequests: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: "prom_label_enforcer",
			Subsystem: "cors",
			Name:      "rejected_requests_total",
			Help:      "Total number of CORS requests with a disallowed origin.",
		}),
		preflightRequests: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: "prom_label_enforcer",
			Subsystem: "cors",
			Name:      "preflight_requests_total",
			Help:      "Total number of CORS preflight (OPTIONS) requests handled.",
		}),
	}

	reg.MustRegister(m.allowedRequests, m.rejectedRequests, m.preflightRequests)
	return m
}

func (m *CORSMetrics) recordAllowed() {
	if m != nil {
		m.allowedRequests.Inc()
	}
}

func (m *CORSMetrics) recordRejected() {
	if m != nil {
		m.rejectedRequests.Inc()
	}
}

func (m *CORSMetrics) recordPreflight() {
	if m != nil {
		m.preflightRequests.Inc()
	}
}
