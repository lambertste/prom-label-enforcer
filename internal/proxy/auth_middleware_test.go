package proxy

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
)

func newAuthTestRegistry() *prometheus.Registry {
	return prometheus.NewRegistry()
}

func makeAuthMiddleware(tokens ...string) func(http.Handler) http.Handler {
	cfg := DefaultAuthConfig()
	for _, t := range tokens {
		cfg.Tokens[t] = struct{}{}
	}
	return NewAuthMiddleware(cfg, newAuthTestRegistry())
}

var authOKHandler = http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
})

func TestDefaultAuthConfig(t *testing.T) {
	cfg := DefaultAuthConfig()
	if cfg.Realm == "" {
		t.Fatal("expected non-empty realm")
	}
	if cfg.Tokens == nil {
		t.Fatal("expected non-nil token map")
	}
}

func TestAuthMiddleware_MissingToken(t *testing.T) {
	mw := makeAuthMiddleware("secret")
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	mw(authOKHandler).ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
	mw := makeAuthMiddleware("secret")
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	req.Header.Set("Authorization", "Bearer wrong")
	mw(authOKHandler).ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}

func TestAuthMiddleware_ValidToken(t *testing.T) {
	mw := makeAuthMiddleware("secret")
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	req.Header.Set("Authorization", "Bearer secret")
	mw(authOKHandler).ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestExtractBearerToken_NoHeader(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	if tok := extractBearerToken(req); tok != "" {
		t.Fatalf("expected empty token, got %q", tok)
	}
}

func TestExtractBearerToken_Valid(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer mytoken")
	if tok := extractBearerToken(req); tok != "mytoken" {
		t.Fatalf("expected 'mytoken', got %q", tok)
	}
}
