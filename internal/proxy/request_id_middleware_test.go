package proxy

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
)

func newRequestIDTestRegistry() *prometheus.Registry {
	return prometheus.NewRegistry()
}

func makeRequestIDMiddleware(t *testing.T, cfg RequestIDConfig) (func(http.Handler) http.Handler, *RequestIDMetrics) {
	t.Helper()
	reg := newRequestIDTestRegistry()
	m := NewRequestIDMetrics(reg)
	return NewRequestIDMiddleware(cfg, m), m
}

func TestDefaultRequestIDConfig(t *testing.T) {
	cfg := DefaultRequestIDConfig()
	if cfg.Header != RequestIDHeader {
		t.Fatalf("expected header %q, got %q", RequestIDHeader, cfg.Header)
	}
	if cfg.ForceNew {
		t.Fatal("expected ForceNew to be false")
	}
}

func TestRequestIDMiddleware_GeneratesIDWhenAbsent(t *testing.T) {
	mw, _ := makeRequestIDMiddleware(t, DefaultRequestIDConfig())
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if RequestIDFromContext(r.Context()) == "" {
			t.Error("expected request ID in context")
		}
		w.WriteHeader(http.StatusOK)
	}))
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	handler.ServeHTTP(rec, req)
	if rec.Header().Get(RequestIDHeader) == "" {
		t.Error("expected X-Request-ID in response header")
	}
}

func TestRequestIDMiddleware_PropagatesExistingID(t *testing.T) {
	mw, _ := makeRequestIDMiddleware(t, DefaultRequestIDConfig())
	const existingID = "my-trace-123"
	var captured string
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		captured = RequestIDFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	}))
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(RequestIDHeader, existingID)
	handler.ServeHTTP(rec, req)
	if captured != existingID {
		t.Fatalf("expected %q, got %q", existingID, captured)
	}
	if rec.Header().Get(RequestIDHeader) != existingID {
		t.Error("expected response header to echo existing ID")
	}
}

func TestRequestIDMiddleware_ForceNewIgnoresExisting(t *testing.T) {
	cfg := RequestIDConfig{Header: RequestIDHeader, ForceNew: true}
	mw, _ := makeRequestIDMiddleware(t, cfg)
	const existingID = "old-id"
	var captured string
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		captured = RequestIDFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	}))
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(RequestIDHeader, existingID)
	handler.ServeHTTP(rec, req)
	if captured == existingID {
		t.Error("expected a new ID to be generated, got existing ID")
	}
}
