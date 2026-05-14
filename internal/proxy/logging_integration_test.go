package proxy

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func newInstrumentedLogging(slow time.Duration) http.Handler {
	reg := newLoggingTestRegistry()
	metrics := NewLoggingMetrics(reg)
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))
	cfg := LoggingConfig{
		Logger:        logger,
		SlowThreshold: slow,
	}
	mw := NewLoggingMiddleware(cfg, metrics)
	base := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	return mw(base)
}

func TestInstrumentedLogging_RecordsAndLogs(t *testing.T) {
	reg := newLoggingTestRegistry()
	metrics := NewLoggingMetrics(reg)
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))
	cfg := LoggingConfig{Logger: logger, SlowThreshold: time.Second}
	mw := NewLoggingMiddleware(cfg, metrics)
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if !strings.Contains(buf.String(), "/metrics") {
		t.Fatalf("expected path in log output, got: %s", buf.String())
	}
}

func TestInstrumentedLogging_SlowFlagSet(t *testing.T) {
	reg := newLoggingTestRegistry()
	metrics := NewLoggingMetrics(reg)
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))
	cfg := LoggingConfig{Logger: logger, SlowThreshold: 1 * time.Nanosecond}
	mw := NewLoggingMiddleware(cfg, metrics)
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(2 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest(http.MethodGet, "/slow", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if !strings.Contains(buf.String(), "true") {
		t.Fatalf("expected slow=true in log, got: %s", buf.String())
	}
}
