package proxy

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDefaultCORSConfig(t *testing.T) {
	cfg := DefaultCORSConfig()
	if len(cfg.AllowedOrigins) == 0 {
		t.Fatal("expected non-empty AllowedOrigins")
	}
	if cfg.MaxAge <= 0 {
		t.Fatalf("expected positive MaxAge, got %d", cfg.MaxAge)
	}
}

func TestCORSMiddleware_WildcardOrigin(t *testing.T) {
	cfg := DefaultCORSConfig() // allows "*"
	mw := NewCORSMiddleware(cfg)

	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	req.Header.Set("Origin", "https://example.com")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "*" {
		t.Errorf("expected *, got %q", got)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestCORSMiddleware_SpecificOriginAllowed(t *testing.T) {
	cfg := CORSConfig{
		AllowedOrigins: []string{"https://trusted.example.com"},
		AllowedMethods: []string{"GET"},
		AllowedHeaders: []string{"Content-Type"},
		MaxAge:         600,
	}
	mw := NewCORSMiddleware(cfg)
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	req.Header.Set("Origin", "https://trusted.example.com")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "https://trusted.example.com" {
		t.Errorf("expected specific origin header, got %q", got)
	}
}

func TestCORSMiddleware_OriginNotAllowed(t *testing.T) {
	cfg := CORSConfig{
		AllowedOrigins: []string{"https://trusted.example.com"},
		AllowedMethods: []string{"GET"},
		AllowedHeaders: []string{"Content-Type"},
		MaxAge:         600,
	}
	mw := NewCORSMiddleware(cfg)
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	req.Header.Set("Origin", "https://evil.example.com")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Errorf("expected no CORS header for disallowed origin, got %q", got)
	}
}

func TestCORSMiddleware_PreflightReturns204(t *testing.T) {
	cfg := DefaultCORSConfig()
	mw := NewCORSMiddleware(cfg)
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodOptions, "/metrics", nil)
	req.Header.Set("Origin", "https://example.com")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204 for preflight, got %d", rec.Code)
	}
}

func TestCORSMiddleware_EmptyConfigUsesDefaults(t *testing.T) {
	mw := NewCORSMiddleware(CORSConfig{})
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	req.Header.Set("Origin", "https://any.example.com")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "*" {
		t.Errorf("expected wildcard origin from defaults, got %q", got)
	}
}
