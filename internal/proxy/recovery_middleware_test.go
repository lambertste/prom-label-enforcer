package proxy

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
)

func newRecoveryTestRegistry() *prometheus.Registry {
	return prometheus.NewRegistry()
}

func panicHandler(w http.ResponseWriter, r *http.Request) {
	panic("test panic")
}

func okRecoveryHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func TestDefaultRecoveryConfig(t *testing.T) {
	reg := newRecoveryTestRegistry()
	cfg := DefaultRecoveryConfig(reg)
	if cfg.Logger == nil {
		t.Fatal("expected non-nil logger")
	}
	if cfg.Metrics == nil {
		t.Fatal("expected non-nil metrics")
	}
	if !cfg.PrintStack {
		t.Fatal("expected PrintStack to be true")
	}
}

func TestRecoveryMiddleware_PassesThroughNormalRequest(t *testing.T) {
	reg := newRecoveryTestRegistry()
	cfg := DefaultRecoveryConfig(reg)
	h := NewRecoveryMiddleware(cfg, http.HandlerFunc(okRecoveryHandler))

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestRecoveryMiddleware_RecoversPanic(t *testing.T) {
	var buf bytes.Buffer
	logger := log.New(&buf, "", 0)
	reg := newRecoveryTestRegistry()
	cfg := RecoveryConfig{
		Logger:     logger,
		Metrics:    NewRecoveryMetrics(reg),
		PrintStack: false,
	}
	h := NewRecoveryMiddleware(cfg, http.HandlerFunc(panicHandler))

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
	if buf.Len() == 0 {
		t.Fatal("expected panic to be logged")
	}
}

func TestRecoveryMiddleware_NilLoggerUsesDefault(t *testing.T) {
	reg := newRecoveryTestRegistry()
	cfg := RecoveryConfig{
		Logger:  nil,
		Metrics: NewRecoveryMetrics(reg),
	}
	// Should not panic on construction.
	h := NewRecoveryMiddleware(cfg, http.HandlerFunc(panicHandler))

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
}
