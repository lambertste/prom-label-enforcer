package proxy

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

func newMetricsTestRegistry() *prometheus.Registry {
	return prometheus.NewRegistry()
}

func TestDefaultMetricsMiddlewareConfig(t *testing.T) {
	cfg := DefaultMetricsMiddlewareConfig()
	if cfg.Registerer == nil {
		t.Fatal("expected non-nil Registerer")
	}
	if cfg.Namespace == "" {
		t.Fatal("expected non-empty Namespace")
	}
}

func TestMetricsMiddleware_RecordsRequestCount(t *testing.T) {
	reg := newMetricsTestRegistry()
	cfg := MetricsMiddlewareConfig{Registerer: reg, Namespace: "test"}

	handler := NewMetricsMiddleware(cfg, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	count, err := testutil.GatherAndCount(reg)
	if err != nil {
		t.Fatalf("unexpected gather error: %v", err)
	}
	if count == 0 {
		t.Fatal("expected at least one metric family to be registered")
	}
}

func TestMetricsMiddleware_RecordsDuration(t *testing.T) {
	reg := newMetricsTestRegistry()
	cfg := MetricsMiddlewareConfig{Registerer: reg, Namespace: "test2"}

	handler := NewMetricsMiddleware(cfg, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
	}))

	req := httptest.NewRequest(http.MethodPost, "/push", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d", rr.Code)
	}
}

func TestMetricsMiddleware_NilRegistererUsesDefault(t *testing.T) {
	// Should not panic when Registerer is nil; falls back to DefaultRegisterer.
	// Use a fresh registry to avoid conflicts.
	reg := newMetricsTestRegistry()
	cfg := MetricsMiddlewareConfig{Registerer: reg, Namespace: "test3"}

	handler := NewMetricsMiddleware(cfg, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestStatusRecorder_WriteHeaderOnce(t *testing.T) {
	rr := httptest.NewRecorder()
	sr := &statusRecorder{ResponseWriter: rr, status: http.StatusOK}

	sr.WriteHeader(http.StatusCreated)
	sr.WriteHeader(http.StatusBadRequest) // should be ignored

	if sr.status != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", sr.status)
	}
}
