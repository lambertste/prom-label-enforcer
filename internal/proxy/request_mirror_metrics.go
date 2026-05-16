package proxy

import "github.com/prometheus/client_golang/prometheus"

// MirrorMetrics holds Prometheus counters for the mirror middleware.
type MirrorMetrics struct {
	sentTotal   prometheus.Counter
	errorTotal  prometheus.Counter
}

// NewMirrorMetrics registers and returns MirrorMetrics.
// If reg is nil the global Prometheus registry is used.
func NewMirrorMetrics(reg prometheus.Registerer) *MirrorMetrics {
	if reg == nil {
		reg = prometheus.DefaultRegisterer
	}

	sent := prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "prom_label_enforcer",
		Subsystem: "mirror",
		Name:      "requests_sent_total",
		Help:      "Total number of requests successfully mirrored to the secondary target.",
	})
	errored := prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "prom_label_enforcer",
		Subsystem: "mirror",
		Name:      "requests_error_total",
		Help:      "Total number of mirror requests that failed.",
	})

	for _, c := range []prometheus.Collector{sent, errored} {
		if err := reg.Register(c); err != nil {
			if are, ok := err.(prometheus.AlreadyRegisteredError); ok {
				switch c.(type) {
				case prometheus.Counter:
					_ = are.ExistingCollector
				}
			}
		}
	}

	return &MirrorMetrics{sentTotal: sent, errorTotal: errored}
}

// RecordSent increments the sent counter.
func (m *MirrorMetrics) RecordSent() {
	if m == nil {
		return
	}
	m.sentTotal.Inc()
}

// RecordError increments the error counter.
func (m *MirrorMetrics) RecordError() {
	if m == nil {
		return
	}
	m.errorTotal.Inc()
}
