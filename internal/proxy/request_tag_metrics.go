package proxy

import "github.com/prometheus/client_golang/prometheus"

// RequestTagMetrics holds Prometheus counters for the tagging middleware.
type RequestTagMetrics struct {
	tagged    prometheus.Counter
	untagged  prometheus.Counter
	truncated prometheus.Counter
	tagCount  prometheus.Histogram
}

// NewRequestTagMetrics creates and registers metrics for request tagging.
func NewRequestTagMetrics(reg prometheus.Registerer) *RequestTagMetrics {
	if reg == nil {
		reg = prometheus.DefaultRegisterer
	}
	m := &RequestTagMetrics{
		tagged: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "proxy_request_tag_tagged_total",
			Help: "Total requests that carried at least one tag.",
		}),
		untagged: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "proxy_request_tag_untagged_total",
			Help: "Total requests with no tags.",
		}),
		truncated: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "proxy_request_tag_truncated_total",
			Help: "Total requests whose tag list was truncated to MaxTags.",
		}),
		tagCount: prometheus.NewHistogram(prometheus.HistogramOpts{
			Name:    "proxy_request_tag_count",
			Help:    "Distribution of tag counts per tagged request.",
			Buckets: []float64{1, 2, 5, 10},
		}),
	}
	reg.MustRegister(m.tagged, m.untagged, m.truncated, m.tagCount)
	return m
}

// RecordTagged increments the tagged counter and observes the tag count.
func (m *RequestTagMetrics) RecordTagged(n int) {
	if m == nil {
		return
	}
	m.tagged.Inc()
	m.tagCount.Observe(float64(n))
}

// RecordUntagged increments the untagged counter.
func (m *RequestTagMetrics) RecordUntagged() {
	if m == nil {
		return
	}
	m.untagged.Inc()
}

// RecordTruncated increments the truncated counter.
func (m *RequestTagMetrics) RecordTruncated() {
	if m == nil {
		return
	}
	m.truncated.Inc()
}
