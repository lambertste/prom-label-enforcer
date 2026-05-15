package proxy

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
)

func newResponseSizeTestRegistry() *prometheus.Registry {
	return prometheus.NewRegistry()
}

func TestNewResponseSizeMetrics_NotNil(t *testing.T) {
	reg := newResponseSizeTestRegistry()
	m := NewResponseSizeMetrics(reg)
	if m == nil {
		t.Fatal("expected non-nil ResponseSizeMetrics")
	}
}

func TestNewResponseSizeMetrics_NilRegisterer(t *testing.T) {
	// Should not panic when nil is passed.
	m := NewResponseSizeMetrics(nil)
	if m == nil {
		t.Fatal("expected non-nil ResponseSizeMetrics with nil registerer")
	}
}

func TestResponseSizeMetrics_NilSafe(t *testing.T) {
	var m *ResponseSizeMetrics
	// Must not panic.
	m.Record(1024)
}

func TestResponseSizeMiddleware_ZeroBytesBody(t *testing.T) {
	reg := newResponseSizeTestRegistry()
	m := NewResponseSizeMetrics(reg)

	emptyHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	h := NewResponseSizeMiddleware(m, emptyHandler)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}
}

func TestResponseSizeMiddleware_RecordsBodySize(t *testing.T) {
	reg := newResponseSizeTestRegistry()
	m := NewResponseSizeMetrics(reg)

	body := strings.Repeat("x", 512)
	handle := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, body)
	})

	h := NewResponseSizeMiddleware(m, handle)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/metrics", nil))

	if rec.Body.Len() != 512 {
		t.Fatalf("expected body length 512, got %d", rec.Body.Len())
	}
}

func TestResponseSizeMiddleware_NilMetricsUsesDefault(t *testing.T) {
	handle := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "hello")
	})
	// Should not panic with nil metrics.
	h := NewResponseSizeMiddleware(nil, handle)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}
