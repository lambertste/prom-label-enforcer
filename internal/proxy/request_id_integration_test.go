package proxy_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/prom-label-enforcer/internal/proxy"
)

func newInstrumentedRequestID(reg prometheus.Registerer) http.Handler {
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := proxy.RequestIDFromContext(r.Context())
		w.Header().Set("X-Request-Id", id)
		w.WriteHeader(http.StatusOK)
	})

	metrics := proxy.NewRequestIDMetrics(reg)
	cfg := proxy.DefaultRequestIDConfig()
	cfg.Metrics = metrics
	return proxy.NewRequestIDMiddleware(cfg, inner)
}

func TestInstrumentedRequestID_GeneratesAndRecords(t *testing.T) {
	reg := prometheus.NewPedanticRegistry()
	h := newInstrumentedRequestID(reg)

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	id := rec.Header().Get("X-Request-Id")
	if id == "" {
		t.Fatal("expected a generated request ID in response header")
	}

	mfs, err := reg.Gather()
	if err != nil {
		t.Fatalf("gather: %v", err)
	}
	if len(mfs) == 0 {
		t.Fatal("expected metrics to be recorded")
	}
}

func TestInstrumentedRequestID_PropagatesAndRecords(t *testing.T) {
	reg := prometheus.NewPedanticRegistry()
	h := newInstrumentedRequestID(reg)

	existingID := "test-trace-abc-123"
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	req.Header.Set("X-Request-Id", existingID)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	got := rec.Header().Get("X-Request-Id")
	if got != existingID {
		t.Fatalf("expected propagated ID %q, got %q", existingID, got)
	}

	mfs, err := reg.Gather()
	if err != nil {
		t.Fatalf("gather: %v", err)
	}
	if len(mfs) == 0 {
		t.Fatal("expected metrics to be recorded")
	}
}
