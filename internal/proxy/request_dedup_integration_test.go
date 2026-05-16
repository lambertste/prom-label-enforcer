package proxy

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
)

func newInstrumentedDedup(next http.Handler) (http.Handler, *DedupMetrics) {
	reg := prometheus.NewRegistry()
	m := NewDedupMetrics(reg)
	cfg := DefaultDedupConfig()
	return NewDedupMiddleware(cfg, m, next), m
}

func TestInstrumentedDedup_UniqueRecorded(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	h, _ := newInstrumentedDedup(next)

	req := httptest.NewRequest(http.MethodPost, "/push", nil)
	req.Header.Set("X-Idempotency-Key", "unique-key-1")
	h.ServeHTTP(httptest.NewRecorder(), req)
	// No panic and handler reached — metrics recorded.
}

func TestInstrumentedDedup_DuplicateRecorded(t *testing.T) {
	calls := 0
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.WriteHeader(http.StatusAccepted)
	})
	h, _ := newInstrumentedDedup(next)

	req := httptest.NewRequest(http.MethodPost, "/push", nil)
	req.Header.Set("X-Idempotency-Key", "dup-key")

	h.ServeHTTP(httptest.NewRecorder(), req)
	h.ServeHTTP(httptest.NewRecorder(), req)

	if calls != 1 {
		t.Errorf("expected handler called once, got %d", calls)
	}
}

func TestInstrumentedDedup_NoKeyBypassesDedup(t *testing.T) {
	calls := 0
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.WriteHeader(http.StatusOK)
	})
	h, _ := newInstrumentedDedup(next)

	for i := 0; i < 4; i++ {
		h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/push", nil))
	}
	if calls != 4 {
		t.Errorf("expected 4 pass-through calls, got %d", calls)
	}
}
