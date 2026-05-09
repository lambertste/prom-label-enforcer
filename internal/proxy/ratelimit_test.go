package proxy

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewRateLimiter_NotNil(t *testing.T) {
	rl := NewRateLimiter(10, 5)
	if rl == nil {
		t.Fatal("expected non-nil RateLimiter")
	}
}

func TestRateLimiter_Allow_WithinLimit(t *testing.T) {
	rl := NewRateLimiter(100, 5)
	for i := 0; i < 5; i++ {
		if !rl.Allow("client1") {
			t.Fatalf("expected allow on attempt %d", i+1)
		}
	}
}

func TestRateLimiter_Allow_ExceedsCapacity(t *testing.T) {
	rl := NewRateLimiter(0, 2) // rate=0 so no refill
	rl.Allow("client2")       // consume token 1
	rl.Allow("client2")       // consume token 2
	if rl.Allow("client2") {
		t.Fatal("expected deny after capacity exhausted")
	}
}

func TestRateLimiter_Allow_IndependentClients(t *testing.T) {
	rl := NewRateLimiter(0, 1)
	if !rl.Allow("clientA") {
		t.Fatal("clientA should be allowed")
	}
	if !rl.Allow("clientB") {
		t.Fatal("clientB should be allowed independently")
	}
	if rl.Allow("clientA") {
		t.Fatal("clientA should be denied after exhausting its bucket")
	}
}

func TestRateLimiter_Middleware_Allows(t *testing.T) {
	rl := NewRateLimiter(100, 10)
	handler := rl.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPost, "/receive", nil)
	req.RemoteAddr = "127.0.0.1:9000"
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestRateLimiter_Middleware_Rejects(t *testing.T) {
	rl := NewRateLimiter(0, 1) // capacity 1, no refill
	rl.Allow("10.0.0.1:1234") // exhaust

	handler := rl.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPost, "/receive", nil)
	req.RemoteAddr = "10.0.0.1:1234"
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", rec.Code)
	}
}
