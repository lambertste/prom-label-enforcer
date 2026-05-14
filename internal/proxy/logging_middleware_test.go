package proxy

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestDefaultLoggingConfig(t *testing.T) {
	cfg := DefaultLoggingConfig()
	if cfg.Logger == nil {
		t.Fatal("expected non-nil logger")
	}
	if cfg.SlowThreshold != 500*time.Millisecond {
		t.Fatalf("unexpected slow threshold: %v", cfg.SlowThreshold)
	}
}

func TestLoggingMiddleware_LogsRequest(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, nil))
	cfg := LoggingConfig{Logger: logger, SlowThreshold: time.Second}
	mw := NewLoggingMiddleware(cfg, nil)
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if buf.Len() == 0 {
		t.Fatal("expected log output")
	}
}

func TestLoggingMiddleware_NilLoggerUsesDefault(t *testing.T) {
	cfg := LoggingConfig{Logger: nil, SlowThreshold: time.Second}
	mw := NewLoggingMiddleware(cfg, nil)
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	req := httptest.NewRequest(http.MethodPost, "/push", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}
}

func TestLoggingMiddleware_RecordsMetrics(t *testing.T) {
	reg := newLoggingTestRegistry()
	metrics := NewLoggingMetrics(reg)
	cfg := DefaultLoggingConfig()
	mw := NewLoggingMiddleware(cfg, metrics)
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
}

func TestLoggingMiddleware_SlowRequest(t *testing.T) {
	reg := newLoggingTestRegistry()
	metrics := NewLoggingMetrics(reg)
	cfg := LoggingConfig{Logger: slog.Default(), SlowThreshold: 1 * time.Nanosecond}
	mw := NewLoggingMiddleware(cfg, metrics)
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(2 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest(http.MethodGet, "/slow", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}
