package proxy

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func makeSignedRequest(secret, path string) *http.Request {
	ts := time.Now().UTC().Format(time.RFC3339)
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(ts + ":" + path))
	sig := hex.EncodeToString(h.Sum(nil))

	req := httptest.NewRequest(http.MethodGet, path, nil)
	req.Header.Set("X-Signature", sig)
	req.Header.Set("X-Timestamp", ts)
	return req
}

func TestDefaultRequestSigningConfig(t *testing.T) {
	cfg := DefaultRequestSigningConfig()
	if cfg.HeaderName != "X-Signature" {
		t.Errorf("unexpected HeaderName: %s", cfg.HeaderName)
	}
	if cfg.MaxSkew != 5*time.Minute {
		t.Errorf("unexpected MaxSkew: %v", cfg.MaxSkew)
	}
}

func TestRequestSigningMiddleware_AllowsValidSignature(t *testing.T) {
	cfg := DefaultRequestSigningConfig()
	cfg.SecretKey = "test-secret"

	handler := NewRequestSigningMiddleware(cfg, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := makeSignedRequest("test-secret", "/metrics")
	rw := httptest.NewRecorder()
	handler.ServeHTTP(rw, req)

	if rw.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rw.Code)
	}
}

func TestRequestSigningMiddleware_RejectsMissingSignature(t *testing.T) {
	cfg := DefaultRequestSigningConfig()
	cfg.SecretKey = "test-secret"

	handler := NewRequestSigningMiddleware(cfg, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rw := httptest.NewRecorder()
	handler.ServeHTTP(rw, req)

	if rw.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rw.Code)
	}
}

func TestRequestSigningMiddleware_RejectsWrongSignature(t *testing.T) {
	cfg := DefaultRequestSigningConfig()
	cfg.SecretKey = "test-secret"

	handler := NewRequestSigningMiddleware(cfg, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := makeSignedRequest("wrong-secret", "/metrics")
	rw := httptest.NewRecorder()
	handler.ServeHTTP(rw, req)

	if rw.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rw.Code)
	}
}

func TestRequestSigningMiddleware_BypassesWhenNoKey(t *testing.T) {
	cfg := DefaultRequestSigningConfig()
	// SecretKey intentionally empty

	handler := NewRequestSigningMiddleware(cfg, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rw := httptest.NewRecorder()
	handler.ServeHTTP(rw, req)

	if rw.Code != http.StatusOK {
		t.Errorf("expected 200 bypass, got %d", rw.Code)
	}
}

func TestRequestSigningMiddleware_RejectsExpiredTimestamp(t *testing.T) {
	cfg := DefaultRequestSigningConfig()
	cfg.SecretKey = "test-secret"
	cfg.MaxSkew = 1 * time.Second

	handler := NewRequestSigningMiddleware(cfg, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	old := time.Now().Add(-10 * time.Minute).UTC().Format(time.RFC3339)
	h := hmac.New(sha256.New, []byte("test-secret"))
	h.Write([]byte(old + ":/metrics"))
	sig := hex.EncodeToString(h.Sum(nil))

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	req.Header.Set("X-Signature", sig)
	req.Header.Set("X-Timestamp", old)

	rw := httptest.NewRecorder()
	handler.ServeHTTP(rw, req)

	if rw.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for old timestamp, got %d", rw.Code)
	}
}
