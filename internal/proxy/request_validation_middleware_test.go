package proxy

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func okValidationHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func TestDefaultRequestValidationConfig(t *testing.T) {
	cfg := DefaultRequestValidationConfig()
	if len(cfg.AllowedMethods) == 0 {
		t.Fatal("expected non-empty AllowedMethods")
	}
	if cfg.MaxHeaderBytes <= 0 {
		t.Fatalf("expected positive MaxHeaderBytes, got %d", cfg.MaxHeaderBytes)
	}
}

func TestRequestValidationMiddleware_AllowsValidRequest(t *testing.T) {
	cfg := DefaultRequestValidationConfig()
	h := NewRequestValidationMiddleware(cfg, http.HandlerFunc(okValidationHandler))

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestRequestValidationMiddleware_RejectsDisallowedMethod(t *testing.T) {
	cfg := RequestValidationConfig{
		AllowedMethods: []string{http.MethodGet},
		MaxHeaderBytes: 8192,
	}
	h := NewRequestValidationMiddleware(cfg, http.HandlerFunc(okValidationHandler))

	req := httptest.NewRequest(http.MethodDelete, "/metrics", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
}

func TestRequestValidationMiddleware_RejectsOversizedHeaders(t *testing.T) {
	cfg := RequestValidationConfig{
		AllowedMethods: []string{http.MethodGet},
		MaxHeaderBytes: 10, // very small limit
	}
	h := NewRequestValidationMiddleware(cfg, http.HandlerFunc(okValidationHandler))

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	req.Header.Set("X-Custom-Header", "this-value-is-definitely-longer-than-ten-bytes")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestRequestValidationMiddleware_RequiresContentType(t *testing.T) {
	cfg := RequestValidationConfig{
		AllowedMethods:     []string{http.MethodPost},
		MaxHeaderBytes:     8192,
		RequireContentType: true,
	}
	h := NewRequestValidationMiddleware(cfg, http.HandlerFunc(okValidationHandler))

	req := httptest.NewRequest(http.MethodPost, "/write", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for missing Content-Type, got %d", rec.Code)
	}
}

func TestRequestValidationMiddleware_ZeroConfigUsesDefaults(t *testing.T) {
	h := NewRequestValidationMiddleware(RequestValidationConfig{}, http.HandlerFunc(okValidationHandler))

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 with zero config, got %d", rec.Code)
	}
}
