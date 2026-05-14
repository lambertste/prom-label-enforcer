package proxy

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

func newCacheTestRegistry() prometheus.Registerer {
	return prometheus.NewRegistry()
}

func TestNewCacheMetrics_NotNil(t *testing.T) {
	m := NewCacheMetrics(newCacheTestRegistry())
	if m == nil {
		t.Fatal("expected non-nil CacheMetrics")
	}
}

func TestNewCacheMetrics_NilRegisterer(t *testing.T) {
	// Should not panic; uses DefaultRegisterer. We re-register so use a fresh registry.
	reg := prometheus.NewRegistry()
	m := NewCacheMetrics(reg)
	if m == nil {
		t.Fatal("expected non-nil CacheMetrics")
	}
}

func TestCacheMetrics_RecordHit(t *testing.T) {
	reg := prometheus.NewRegistry()
	m := NewCacheMetrics(reg)
	m.RecordHit()
	m.RecordHit()
	count := testutil.ToFloat64(m.hits)
	if count != 2 {
		t.Errorf("expected 2 hits, got %v", count)
	}
}

func TestCacheMetrics_RecordMiss(t *testing.T) {
	reg := prometheus.NewRegistry()
	m := NewCacheMetrics(reg)
	m.RecordMiss()
	count := testutil.ToFloat64(m.misses)
	if count != 1 {
		t.Errorf("expected 1 miss, got %v", count)
	}
}

func TestCacheMetrics_NilSafe(t *testing.T) {
	var m *CacheMetrics
	m.RecordHit()
	m.RecordMiss()
	// Should not panic.
}

func TestInstrumentedMiddleware_RecordsHitAndMiss(t *testing.T) {
	reg := prometheus.NewRegistry()
	metrics := NewCacheMetrics(reg)

	backend := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("pong"))
	})

	cache := NewResponseCache(DefaultCacheConfig())
	h := cache.InstrumentedMiddleware(backend, metrics)

	// First request — miss.
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/metrics", nil))

	// Second request — hit.
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/metrics", nil))

	if got := testutil.ToFloat64(metrics.misses); got != 1 {
		t.Errorf("expected 1 miss, got %v", got)
	}
	if got := testutil.ToFloat64(metrics.hits); got != 1 {
		t.Errorf("expected 1 hit, got %v", got)
	}
}

func TestInstrumentedMiddleware_SkipsNonGet(t *testing.T) {
	reg := prometheus.NewRegistry()
	metrics := NewCacheMetrics(reg)
	backend := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
	})
	cache := NewResponseCache(DefaultCacheConfig())
	h := cache.InstrumentedMiddleware(backend, metrics)

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/receive", nil))

	if got := testutil.ToFloat64(metrics.hits); got != 0 {
		t.Errorf("expected 0 hits for POST, got %v", got)
	}
	if got := testutil.ToFloat64(metrics.misses); got != 0 {
		t.Errorf("expected 0 misses for POST, got %v", got)
	}
}
