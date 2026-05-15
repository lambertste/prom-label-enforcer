package proxy

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestDefaultRequestSizeMiddlewareConfig(t *testing.T) {
	cfg := DefaultRequestSizeMiddlewareConfig()
	if cfg.Registerer == nil {
		t.Fatal("expected non-nil Registerer")
	}
}

func TestRequestSizeMiddleware_ZeroContentLength(t *testing.T) {
	reg := prometheus.NewRegistry()
	mw := NewRequestSizeMiddleware(RequestSizeMiddlewareConfig{Registerer: reg}, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec := httptest.NewRecorder()
	mw.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestRequestSizeMiddleware_RecordsBodySize(t *testing.T) {
	reg := prometheus.NewRegistry()
	mw := NewRequestSizeMiddleware(RequestSizeMiddlewareConfig{Registerer: reg}, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	body := strings.Repeat("x", 512)
	req := httptest.NewRequest(http.MethodPost, "/write", bytes.NewBufferString(body))
	req.ContentLength = int64(len(body))
	rec := httptest.NewRecorder()
	mw.ServeHTTP(rec, req)

	count, err := testutil.GatherAndCount(reg)
	if err != nil {
		t.Fatalf("gather error: %v", err)
	}
	if count == 0 {
		t.Fatal("expected at least one metric to be recorded")
	}
}

func TestRequestSizeMiddleware_NilRegistererUsesDefault(t *testing.T) {
	// Should not panic with nil registerer.
	mw := NewRequestSizeMiddleware(RequestSizeMiddlewareConfig{Registerer: nil}, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	mw.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}
}

func TestRequestSizeMiddleware_PassesThroughToNext(t *testing.T) {
	reg := prometheus.NewRegistry()
	called := false
	mw := NewRequestSizeMiddleware(RequestSizeMiddlewareConfig{Registerer: reg}, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusAccepted)
	}))

	req := httptest.NewRequest(http.MethodPost, "/push", bytes.NewBufferString("data"))
	req.ContentLength = 4
	rec := httptest.NewRecorder()
	mw.ServeHTTP(rec, req)

	if !called {
		t.Fatal("expected next handler to be called")
	}
	if rec.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d", rec.Code)
	}
}
