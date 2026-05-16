package proxy

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
)

func newSanitizeRegistry() prometheus.Registerer {
	return prometheus.NewRegistry()
}

func TestDefaultSanitizeConfig(t *testing.T) {
	cfg := DefaultSanitizeConfig()
	if cfg.MaxHeaderValueLen == 0 {
		t.Fatal("expected non-zero MaxHeaderValueLen")
	}
	if !cfg.NormalizeHeaders {
		t.Fatal("expected NormalizeHeaders to be true")
	}
	if len(cfg.StripHeaders) == 0 {
		t.Fatal("expected at least one default strip header")
	}
}

func TestSanitizeMiddleware_StripsConfiguredHeaders(t *testing.T) {
	reg := newSanitizeRegistry()
	cfg := DefaultSanitizeConfig()
	cfg.Registerer = reg
	cfg.StripHeaders = []string{"X-Secret"}

	var capturedReq *http.Request
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedReq = r
		w.WriteHeader(http.StatusOK)
	})

	h := NewSanitizeMiddleware(cfg, next)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Secret", "sensitive")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if capturedReq.Header.Get("X-Secret") != "" {
		t.Errorf("expected X-Secret to be stripped")
	}
}

func TestSanitizeMiddleware_TruncatesLongHeaderValue(t *testing.T) {
	reg := newSanitizeRegistry()
	cfg := SanitizeConfig{
		NormalizeHeaders:  true,
		MaxHeaderValueLen: 10,
		Registerer:        reg,
	}

	var capturedReq *http.Request
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedReq = r
		w.WriteHeader(http.StatusOK)
	})

	h := NewSanitizeMiddleware(cfg, next)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Custom", strings.Repeat("a", 50))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	val := capturedReq.Header.Get("X-Custom")
	if len(val) != 10 {
		t.Errorf("expected header truncated to 10, got %d", len(val))
	}
}

func TestSanitizeMiddleware_CleanRequestPassesThrough(t *testing.T) {
	reg := newSanitizeRegistry()
	cfg := DefaultSanitizeConfig()
	cfg.Registerer = reg
	cfg.StripHeaders = []string{}

	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})

	h := NewSanitizeMiddleware(cfg, next)
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	req.Header.Set("Accept", "text/plain")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if !called {
		t.Fatal("expected next handler to be called")
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestSanitizeMiddleware_ZeroMaxHeaderValueLenUsesDefault(t *testing.T) {
	cfg := SanitizeConfig{MaxHeaderValueLen: 0, Registerer: newSanitizeRegistry()}
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(200) })
	// Should not panic
	h := NewSanitizeMiddleware(cfg, next)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
}
