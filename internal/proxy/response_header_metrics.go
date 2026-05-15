package proxy

import "github.com/prometheus/client_golang/prometheus"

// ResponseHeaderMetrics holds Prometheus metrics for the response header middleware.
type ResponseHeaderMetrics struct {
	InjectedTotal prometheus.Counter
	StrippedTotal prometheus.Counter
}

// NewResponseHeaderMetrics creates and registers ResponseHeaderMetrics.
// If reg is nil, prometheus.DefaultRegisterer is used.
func NewResponseHeaderMetrics(reg prometheus.Registerer) *ResponseHeaderMetrics {
	if reg == nil {
		reg = prometheus.DefaultRegisterer
	}

	injected := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "proxy_response_header_injected_total",
		Help: "Total number of static response headers injected.",
	})
	stripped := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "proxy_response_header_stripped_total",
		Help: "Total number of upstream response headers stripped.",
	})

	_ = reg.Register(injected)
	_ = reg.Register(stripped)

	return &ResponseHeaderMetrics{
		InjectedTotal: injected,
		StrippedTotal: stripped,
	}
}

func (m *ResponseHeaderMetrics) RecordInjected() {
	if m == nil {
		return
	}
	m.InjectedTotal.Add(1)
}

func (m *ResponseHeaderMetrics) RecordStripped() {
	if m == nil {
		return
	}
	m.StrippedTotal.Add(1)
}
