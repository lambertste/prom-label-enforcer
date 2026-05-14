package proxy

import (
	"github.com/prometheus/client_golang/prometheus"
)

// CacheMetrics holds Prometheus counters for cache hits and misses.
type CacheMetrics struct {
	hits   prometheus.Counter
	misses prometheus.Counter
}

// NewCacheMetrics registers and returns a CacheMetrics instance.
// Passing a nil registerer falls back to prometheus.DefaultRegisterer.
func NewCacheMetrics(reg prometheus.Registerer) *CacheMetrics {
	if reg == nil {
		reg = prometheus.DefaultRegisterer
	}
	hits := prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "prom_label_enforcer",
		Subsystem: "cache",
		Name:      "hits_total",
		Help:      "Total number of response cache hits.",
	})
	misses := prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "prom_label_enforcer",
		Subsystem: "cache",
		Name:      "misses_total",
		Help:      "Total number of response cache misses.",
	})
	reg.MustRegister(hits, misses)
	return &CacheMetrics{hits: hits, misses: misses}
}

// RecordHit increments the cache hit counter.
func (m *CacheMetrics) RecordHit() {
	if m == nil {
		return
	}
	m.hits.Inc()
}

// RecordMiss increments the cache miss counter.
func (m *CacheMetrics) RecordMiss() {
	if m == nil {
		return
	}
	m.misses.Inc()
}

// InstrumentedMiddleware wraps a ResponseCache middleware and records
// hit/miss metrics based on whether the request was served from cache.
func (c *ResponseCache) InstrumentedMiddleware(next http.Handler, metrics *CacheMetrics) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			next.ServeHTTP(w, r)
			return
		}
		key := r.URL.RequestURI()
		if body, status, ok := c.Get(key); ok {
			metrics.RecordHit()
			w.WriteHeader(status)
			_, _ = w.Write(body)
			return
		}
		metrics.RecordMiss()
		rec := &cachingResponseWriter{header: w.Header(), code: http.StatusOK}
		next.ServeHTTP(rec, r)
		c.Set(key, rec.body, rec.code)
		w.WriteHeader(rec.code)
		_, _ = w.Write(rec.body)
	})
}
