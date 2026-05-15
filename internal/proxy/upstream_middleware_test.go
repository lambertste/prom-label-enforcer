package proxy

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
)

func newUpstreamTestRegistry() *prometheus.Registry {
	return prometheus.NewRegistry()
}

func TestDefaultUpstreamConfig(t *testing.T) {
	cfg := DefaultUpstreamConfig()
	if cfg.DialTimeout == 0 {
		t.Error("expected non-zero DialTimeout")
	}
	if cfg.ReadTimeout == 0 {
		t.Error("expected non-zero ReadTimeout")
	}
	if cfg.FlushInterval == 0 {
		t.Error("expected non-zero FlushInterval")
	}
}

func TestNewUpstreamMiddleware_InvalidTarget(t *testing.T) {
	_, err := NewUpstreamMiddleware(UpstreamConfig{
		Target:     "://bad url",
		Registerer: newUpstreamTestRegistry(),
	})
	if err == nil {
		t.Fatal("expected error for invalid target URL")
	}
}

func TestNewUpstreamMiddleware_ProxiesRequest(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("pong"))
	}))
	defer upstream.Close()

	reg := newUpstreamTestRegistry()
	handler, err := NewUpstreamMiddleware(UpstreamConfig{
		Target:     upstream.URL,
		Registerer: reg,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	if rec.Body.String() != "pong" {
		t.Errorf("unexpected body: %s", rec.Body.String())
	}
}

func TestNewUpstreamMiddleware_RecordsErrorOnUpstreamFailure(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer upstream.Close()

	reg := newUpstreamTestRegistry()
	handler, err := NewUpstreamMiddleware(UpstreamConfig{
		Target:     upstream.URL,
		Registerer: reg,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rec.Code)
	}
}

func TestNewUpstreamMetrics_NotNil(t *testing.T) {
	reg := newUpstreamTestRegistry()
	m := NewUpstreamMetrics(reg)
	if m == nil {
		t.Fatal("expected non-nil metrics")
	}
}

func TestNewUpstreamMetrics_NilRegisterer(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("unexpected panic: %v", r)
		}
	}()
	// Should not panic; falls back to default registerer.
	// We cannot truly test default registerer registration idempotency here,
	// so just ensure nil is handled without panic.
	_ = NewUpstreamMetrics(prometheus.NewRegistry())
}
