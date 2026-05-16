package proxy

import (
	"github.com/prometheus/client_golang/prometheus"
)

// DedupMetrics holds Prometheus counters for deduplication events.
type DedupMetrics struct {
	duplicates prometheus.Counter
	uniques     prometheus.Counter
}

// NewDedupMetrics registers and returns DedupMetrics.
// If reg is nil, prometheus.DefaultRegisterer is used.
func NewDedupMetrics(reg prometheus.Registerer) *DedupMetrics {
	if reg == nil {
		reg = prometheus.DefaultRegisterer
	}
	m := &DedupMetrics{
		duplicates: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "prom_label_enforcer_dedup_duplicates_total",
			Help: "Total number of deduplicated (replayed) requests.",
		}),
		uniques: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "prom_label_enforcer_dedup_unique_total",
			Help: "Total number of unique (first-seen) requests.",
		}),
	}
	_ = reg.Register(m.duplicates)
	_ = reg.Register(m.uniques)
	return m
}

// RecordDuplicate increments the duplicate counter.
func (m *DedupMetrics) RecordDuplicate() {
	if m == nil {
		return
	}
	m.duplicates.Inc()
}

// RecordUnique increments the unique counter.
func (m *DedupMetrics) RecordUnique() {
	if m == nil {
		return
	}
	m.uniques.Inc()
}
