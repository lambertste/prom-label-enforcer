package proxy

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
)

func newTracingTestRegistry() prometheus.Registerer {
	return prometheus.NewRegistry()
}

func TestDefaultTracingConfig(t *testing.T) {
	cfg := DefaultTracingConfig()
	if cfg.HeaderName == "" {
		t.Fatal("expected non-empty HeaderName")
	}
	if !cfg.GenerateIfMissing {
		t.Fatal("expected GenerateIfMissing to be true")
	}
}

func TestTracingMiddleware_PropagatesExistingTraceID(t *testing.T) {
	cfg := DefaultTracingConfig()
	mw := NewTracingMiddleware(cfg, nil)

	var capturedID string
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedID = TraceIDFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	req.Header.Set(cfg.HeaderName, "test-trace-123")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if capturedID != "test-trace-123" {
		t.Fatalf("expected trace ID 'test-trace-123', got %q", capturedID)
	}
	if rec.Header().Get(cfg.HeaderName) != "test-trace-123" {
		t.Fatal("expected trace ID to be echoed in response header")
	}
}

func TestTracingMiddleware_GeneratesTraceIDWhenMissing(t *testing.T) {
	cfg := DefaultTracingConfig()
	mw := NewTracingMiddleware(cfg, nil)

	var capturedID string
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedID = TraceIDFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if capturedID == "" {
		t.Fatal("expected a generated trace ID, got empty string")
	}
}

func TestTracingMiddleware_RecordsMetrics(t *testing.T) {
	reg := prometheus.NewRegistry()
	metrics := NewTracingMetrics(reg)
	cfg := DefaultTracingConfig()
	mw := NewTracingMiddleware(cfg, metrics)

	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
	}))

	req := httptest.NewRequest(http.MethodPost, "/receive", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	mfs, err := reg.Gather()
	if err != nil {
		t.Fatalf("gather error: %v", err)
	}
	if len(mfs) == 0 {
		t.Fatal("expected metrics to be recorded")
	}
}

func TestTracingMiddleware_ZeroConfigUsesDefaults(t *testing.T) {
	mw := NewTracingMiddleware(TracingConfig{}, nil)

	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	// should not panic
	handler.ServeHTTP(rec, req)
}
