package proxy

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDefaultHealthCheckConfig(t *testing.T) {
	cfg := DefaultHealthCheckConfig()
	if cfg.Path != "/healthz" {
		t.Errorf("expected /healthz, got %s", cfg.Path)
	}
	if cfg.ReadinessPath != "/readyz" {
		t.Errorf("expected /readyz, got %s", cfg.ReadinessPath)
	}
	if cfg.Version != "unknown" {
		t.Errorf("expected unknown, got %s", cfg.Version)
	}
}

func TestHealthCheckMiddleware_LivenessReturns200(t *testing.T) {
	cfg := DefaultHealthCheckConfig()
	cfg.Version = "v1.2.3"
	h := NewHealthCheckMiddleware(cfg, http.NotFoundHandler())

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var resp healthResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if resp.Status != "ok" {
		t.Errorf("expected status ok, got %s", resp.Status)
	}
	if resp.Version != "v1.2.3" {
		t.Errorf("expected version v1.2.3, got %s", resp.Version)
	}
}

func TestHealthCheckMiddleware_ReadinessReturns200WhenReady(t *testing.T) {
	h := NewHealthCheckMiddleware(DefaultHealthCheckConfig(), http.NotFoundHandler())

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var resp healthResponse
	_ = json.NewDecoder(rec.Body).Decode(&resp)
	if resp.Status != "ready" {
		t.Errorf("expected ready, got %s", resp.Status)
	}
}

func TestHealthCheckMiddleware_DelegatesOtherPaths(t *testing.T) {
	var called bool
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusTeapot)
	})
	h := NewHealthCheckMiddleware(DefaultHealthCheckConfig(), next)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	h.ServeHTTP(rec, req)

	if !called {
		t.Error("expected next handler to be called")
	}
	if rec.Code != http.StatusTeapot {
		t.Errorf("expected 418, got %d", rec.Code)
	}
}

func TestHealthCheckMiddleware_ZeroConfigUsesDefaults(t *testing.T) {
	h := NewHealthCheckMiddleware(HealthCheckConfig{}, http.NotFoundHandler())

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 with zero config, got %d", rec.Code)
	}
}
