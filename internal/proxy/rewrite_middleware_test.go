package proxy

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
)

func newRewriteTestRegistry() *prometheus.Registry {
	return prometheus.NewRegistry()
}

func TestDefaultRewriteConfig(t *testing.T) {
	cfg := DefaultRewriteConfig()
	if cfg.Rules == nil {
		t.Fatal("expected non-nil Rules map")
	}
	if cfg.Registerer == nil {
		t.Fatal("expected non-nil Registerer")
	}
}

func TestRewriteMiddleware_PassesThroughUnmatchedPath(t *testing.T) {
	reg := newRewriteTestRegistry()
	cfg := RewriteConfig{
		Rules:      map[string]string{"/api/v1": "/v1"},
		Registerer: reg,
	}

	var gotPath string
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusOK)
	})

	h := NewRewriteMiddleware(cfg, next)
	req := httptest.NewRequest(http.MethodGet, "/other/path", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if gotPath != "/other/path" {
		t.Errorf("expected /other/path, got %s", gotPath)
	}
}

func TestRewriteMiddleware_RewritesMatchingPrefix(t *testing.T) {
	reg := newRewriteTestRegistry()
	cfg := RewriteConfig{
		Rules:      map[string]string{"/api/v1": "/v1"},
		Registerer: reg,
	}

	var gotPath string
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusOK)
	})

	h := NewRewriteMiddleware(cfg, next)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/metrics", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if gotPath != "/v1/metrics" {
		t.Errorf("expected /v1/metrics, got %s", gotPath)
	}
}

func TestRewriteMiddleware_NilRegistererUsesDefault(t *testing.T) {
	cfg := RewriteConfig{
		Rules:      map[string]string{"/x": "/y"},
		Registerer: nil,
	}
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	// Should not panic
	h := NewRewriteMiddleware(cfg, next)
	req := httptest.NewRequest(http.MethodGet, "/x/foo", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
}
