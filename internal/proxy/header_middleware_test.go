package proxy

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDefaultHeaderMiddlewareConfig(t *testing.T) {
	cfg := DefaultHeaderMiddlewareConfig()
	if len(cfg.StaticHeaders) == 0 {
		t.Fatal("expected default static headers to be non-empty")
	}
	if _, ok := cfg.StaticHeaders["X-Content-Type-Options"]; !ok {
		t.Error("expected X-Content-Type-Options in default headers")
	}
	if len(cfg.RemoveRequestHeaders) == 0 {
		t.Error("expected at least one default remove request header")
	}
}

func TestHeaderMiddleware_InjectsStaticHeaders(t *testing.T) {
	cfg := HeaderMiddlewareConfig{
		StaticHeaders: map[string]string{
			"X-Custom-Header": "test-value",
		},
		RemoveRequestHeaders: nil,
	}
	handler := NewHeaderMiddleware(cfg, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	handler.ServeHTTP(rec, req)

	if got := rec.Header().Get("X-Custom-Header"); got != "test-value" {
		t.Errorf("expected X-Custom-Header=test-value, got %q", got)
	}
}

func TestHeaderMiddleware_StripsRequestHeaders(t *testing.T) {
	cfg := HeaderMiddlewareConfig{
		StaticHeaders:        map[string]string{},
		RemoveRequestHeaders: []string{"X-Secret-Token"},
	}
	var capturedHeader string
	handler := NewHeaderMiddleware(cfg, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedHeader = r.Header.Get("X-Secret-Token")
		w.WriteHeader(http.StatusOK)
	}))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Secret-Token", "supersecret")
	handler.ServeHTTP(rec, req)

	if capturedHeader != "" {
		t.Errorf("expected X-Secret-Token to be stripped, got %q", capturedHeader)
	}
}

func TestHeaderMiddleware_NilStaticHeadersDoesNotPanic(t *testing.T) {
	cfg := HeaderMiddlewareConfig{
		StaticHeaders:        nil,
		RemoveRequestHeaders: nil,
	}
	handler := NewHeaderMiddleware(cfg, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	// should not panic
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHeaderMiddleware_DefaultConfigAppliesSecurityHeaders(t *testing.T) {
	cfg := DefaultHeaderMiddlewareConfig()
	handler := NewHeaderMiddleware(cfg, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	handler.ServeHTTP(rec, req)

	expected := []string{"X-Content-Type-Options", "X-Frame-Options", "X-XSS-Protection", "Referrer-Policy"}
	for _, h := range expected {
		if rec.Header().Get(h) == "" {
			t.Errorf("expected security header %q to be set", h)
		}
	}
}
