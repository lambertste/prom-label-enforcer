package proxy

import "github.com/prometheus/client_golang/prometheus"

// RequestSigningMetrics holds counters for signature validation outcomes.
type RequestSigningMetrics struct {
	allowed  prometheus.Counter
	denied   prometheus.Counter
	bypassed prometheus.Counter
}

// NewRequestSigningMetrics registers and returns RequestSigningMetrics.
func NewRequestSigningMetrics(reg prometheus.Registerer) *RequestSigningMetrics {
	if reg == nil {
		reg = prometheus.NewRegistry()
	}

	factory := func(name, help string) prometheus.Counter {
		c := prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: "prom_label_enforcer",
			Subsystem: "request_signing",
			Name:      name,
			Help:      help,
		})
		_ = reg.Register(c)
		return c
	}

	return &RequestSigningMetrics{
		allowed:  factory("allowed_total", "Total requests with a valid HMAC signature."),
		denied:   factory("denied_total", "Total requests rejected due to invalid or missing signature."),
		bypassed: factory("bypassed_total", "Total requests that bypassed signing (enforcement disabled or no key)."),
	}
}

func (m *RequestSigningMetrics) RecordAllowed() {
	if m == nil {
		return
	}
	m.allowed.Inc()
}

func (m *RequestSigningMetrics) RecordDenied() {
	if m == nil {
		return
	}
	m.denied.Inc()
}

func (m *RequestSigningMetrics) RecordBypassed() {
	if m == nil {
		return
	}
	m.bypassed.Inc()
}
