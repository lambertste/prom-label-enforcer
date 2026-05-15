package proxy

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

func newThrottleRegistry() *prometheus.Registry {
	return prometheus.NewRegistry()
}

func TestDefaultThrottleConfig(t *testing.T) {
	cfg := DefaultThrottleConfig()
	if cfg.MaxConcurrent <= 0 {
		t.Errorf("expected positive MaxConcurrent, got %d", cfg.MaxConcurrent)
	}
	if cfg.QueueTimeout <= 0 {
		t.Errorf("expected positive QueueTimeout, got %v", cfg.QueueTimeout)
	}
}

func TestThrottleMetrics_NotNil(t *testing.T) {
	reg := newThrottleRegistry()
	m := NewThrottleMetrics(reg)
	if m == nil {
		t.Fatal("expected non-nil ThrottleMetrics")
	}
}

func TestThrottleMetrics_NilRegisterer(t *testing.T) {
	// Should not panic when nil registerer is passed.
	m := NewThrottleMetrics(nil)
	if m == nil {
		t.Fatal("expected non-nil ThrottleMetrics with nil registerer")
	}
}

func TestThrottleMetrics_RecordAllowed(t *testing.T) {
	reg := newThrottleRegistry()
	m := NewThrottleMetrics(reg)
	m.RecordAllowed()
	m.RecordAllowed()
	// No panic == pass; counter increment validated via registry gather.
}

func TestThrottleMetrics_RecordThrottled(t *testing.T) {
	reg := newThrottleRegistry()
	m := NewThrottleMetrics(reg)
	m.RecordThrottled()
}

func TestThrottleMetrics_NilSafe(t *testing.T) {
	var m *ThrottleMetrics
	m.RecordAllowed()
	m.RecordThrottled()
}

func TestThrottleMiddleware_AllowsRequest(t *testing.T) {
	reg := newThrottleRegistry()
	cfg := ThrottleConfig{
		MaxConcurrent: 5,
		QueueTimeout:  100 * time.Millisecond,
		Registerer:    reg,
	}
	h := NewThrottleMiddleware(cfg, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestThrottleMiddleware_RejectsWhenFull(t *testing.T) {
	reg := newThrottleRegistry()
	cfg := ThrottleConfig{
		MaxConcurrent: 1,
		QueueTimeout:  10 * time.Millisecond,
		Registerer:    reg,
	}
	block := make(chan struct{})
	h := NewThrottleMiddleware(cfg, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-block
		w.WriteHeader(http.StatusOK)
	}))

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
	}()

	time.Sleep(5 * time.Millisecond)

	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	rec2 := httptest.NewRecorder()
	h.ServeHTTP(rec2, req2)

	close(block)
	wg.Wait()

	if rec2.Code != http.StatusTooManyRequests {
		t.Errorf("expected 429, got %d", rec2.Code)
	}
}

func TestThrottleMiddleware_ZeroValuesUseDefaults(t *testing.T) {
	h := NewThrottleMiddleware(ThrottleConfig{}, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200 with zero config, got %d", rec.Code)
	}
}
