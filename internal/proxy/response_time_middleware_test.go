package proxy

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

func newRTMiddlewareRegistry() prometheus.Registerer {
	return prometheus.NewRegistry()
}

func TestResponseTimeMiddleware_NilMetrics_PassesThrough(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mw := NewResponseTimeMiddleware(nil)(handler)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	mw.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestResponseTimeMiddleware_RecordsLatency(t *testing.T) {
	reg := newRTMiddlewareRegistry()
	m := NewResponseTimeMetrics(reg)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(5 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	})

	mw := NewResponseTimeMiddleware(m)(handler)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	mw.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestResponseTimeMiddleware_PropagatesStatusCode(t *testing.T) {
	reg := newRTMiddlewareRegistry()
	m := NewResponseTimeMetrics(reg)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	})

	mw := NewResponseTimeMiddleware(m)(handler)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	mw.ServeHTTP(rec, req)

	if rec.Code != http.StatusTeapot {
		t.Fatalf("expected 418, got %d", rec.Code)
	}
}

func TestResponseTimeWriter_WriteHeaderOnce(t *testing.T) {
	rec := httptest.NewRecorder()
	rw := &responseTimeWriter{ResponseWriter: rec}

	rw.WriteHeader(http.StatusAccepted)
	rw.WriteHeader(http.StatusInternalServerError) // should be ignored

	if rw.statusCode() != http.StatusAccepted {
		t.Fatalf("expected 202, got %d", rw.statusCode())
	}
	if rec.Code != http.StatusAccepted {
		t.Fatalf("underlying writer expected 202, got %d", rec.Code)
	}
}

func TestDefaultResponseTimeMiddlewareConfig_NotNil(t *testing.T) {
	m := DefaultResponseTimeMiddlewareConfig()
	if m == nil {
		t.Fatal("expected non-nil ResponseTimeMetrics")
	}
}
