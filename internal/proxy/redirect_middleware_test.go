package proxy

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
)

func newRedirectTestRegistry() *prometheus.Registry {
	return prometheus.NewRegistry()
}

func TestDefaultRedirectConfig(t *testing.T) {
	cfg := DefaultRedirectConfig()
	if cfg.StatusCode != http.StatusMovedPermanently {
		t.Fatalf("expected 301, got %d", cfg.StatusCode)
	}
	if cfg.HTTPSOnly {
		t.Fatal("expected HTTPSOnly=false")
	}
	if cfg.StripPrefix != "" {
		t.Fatal("expected empty StripPrefix")
	}
}

func TestRedirectMiddleware_PassesThroughByDefault(t *testing.T) {
	reg := newRedirectTestRegistry()
	cfg := DefaultRedirectConfig()
	cfg.Registerer = reg

	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mw := NewRedirectMiddleware(cfg, next)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	mw.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestRedirectMiddleware_HTTPSUpgrade(t *testing.T) {
	reg := newRedirectTestRegistry()
	cfg := DefaultRedirectConfig()
	cfg.HTTPSOnly = true
	cfg.Registerer = reg

	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mw := NewRedirectMiddleware(cfg, next)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "http://example.com/metrics", nil)
	req.Host = "example.com"
	mw.ServeHTTP(rec, req)

	if rec.Code != http.StatusMovedPermanently {
		t.Fatalf("expected 301, got %d", rec.Code)
	}
	loc := rec.Header().Get("Location")
	if loc != "https://example.com/metrics" {
		t.Fatalf("unexpected Location: %s", loc)
	}
}

func TestRedirectMiddleware_StripPrefix(t *testing.T) {
	reg := newRedirectTestRegistry()
	cfg := DefaultRedirectConfig()
	cfg.StripPrefix = "/old"
	cfg.Registerer = reg

	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mw := NewRedirectMiddleware(cfg, next)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/old/metrics", nil)
	mw.ServeHTTP(rec, req)

	if rec.Code != http.StatusMovedPermanently {
		t.Fatalf("expected 301, got %d", rec.Code)
	}
	loc := rec.Header().Get("Location")
	if loc != "/metrics" {
		t.Fatalf("unexpected Location: %s", loc)
	}
}

func TestRedirectMiddleware_ZeroStatusUsesDefault(t *testing.T) {
	reg := newRedirectTestRegistry()
	cfg := RedirectConfig{
		HTTPSOnly:  true,
		StatusCode: 0,
		Registerer: reg,
	}
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {})
	mw := NewRedirectMiddleware(cfg, next)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "http://example.com/", nil)
	req.Host = "example.com"
	mw.ServeHTTP(rec, req)

	if rec.Code != http.StatusMovedPermanently {
		t.Fatalf("expected 301, got %d", rec.Code)
	}
}
