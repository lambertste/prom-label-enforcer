package proxy

import "github.com/prometheus/client_golang/prometheus"

// RewriteMetrics holds counters for URL rewrite middleware activity.
type RewriteMetrics struct {
	rewritten   prometheus.Counter
	passthrough prometheus.Counter
}

// NewRewriteMetrics creates and registers rewrite middleware metrics.
// Falls back to prometheus.DefaultRegisterer if reg is nil.
func NewRewriteMetrics(reg prometheus.Registerer) *RewriteMetrics {
	if reg == nil {
		reg = prometheus.DefaultRegisterer
	}

	rewritten := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "proxy_rewrite_rewritten_total",
		Help: "Total number of requests whose URL path was rewritten.",
	})
	passthrough := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "proxy_rewrite_passthrough_total",
		Help: "Total number of requests that passed through without rewriting.",
	})

	_ = reg.Register(rewritten)
	_ = reg.Register(passthrough)

	return &RewriteMetrics{
		rewritten:   rewritten,
		passthrough: passthrough,
	}
}

// RecordRewritten increments the rewritten counter.
func (m *RewriteMetrics) RecordRewritten() {
	if m == nil {
		return
	}
	m.rewritten.Inc()
}

// RecordPassthrough increments the passthrough counter.
func (m *RewriteMetrics) RecordPassthrough() {
	if m == nil {
		return
	}
	m.passthrough.Inc()
}
