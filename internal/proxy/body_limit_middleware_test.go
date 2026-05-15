package proxy

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
)

func TestDefaultBodyLimitConfig(t *testing.T) {
	cfg := DefaultBodyLimitConfig()
	if cfg.MaxBytes != 1<<20 {
		t.Fatalf("expected 1 MiB default, got %d", cfg.MaxBytes)
	}
}

func TestBodyLimitMiddleware_ZeroMaxBytesUsesDefault(t *testing.T) {
	reg := prometheus.NewRegistry()
	h := NewBodyLimitMiddleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}), BodyLimitConfig{MaxBytes: 0}, reg)

	body := strings.NewReader("hello")
	req := httptest.NewRequest(http.MethodPost, "/", body)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestBodyLimitMiddleware_AllowsSmallBody(t *testing.T) {
	reg := prometheus.NewRegistry()
	h := NewBodyLimitMiddleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}), BodyLimitConfig{MaxBytes: 512}, reg)

	body := bytes.NewReader(make([]byte, 128))
	req := httptest.NewRequest(http.MethodPost, "/", body)
	req.ContentLength = 128
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestBodyLimitMiddleware_RejectsLargeBody(t *testing.T) {
	reg := prometheus.NewRegistry()
	h := NewBodyLimitMiddleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}), BodyLimitConfig{MaxBytes: 64}, reg)

	body := bytes.NewReader(make([]byte, 128))
	req := httptest.NewRequest(http.MethodPost, "/", body)
	req.ContentLength = 128
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("expected 413, got %d", rec.Code)
	}
}

func TestBodyLimitMiddleware_NilRegisterer(t *testing.T) {
	h := NewBodyLimitMiddleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}), BodyLimitConfig{MaxBytes: 512}, nil)

	body := bytes.NewReader(make([]byte, 32))
	req := httptest.NewRequest(http.MethodPost, "/", body)
	req.ContentLength = 32
	rec := httptest.NewRecorder()

	// Should not panic.
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}
