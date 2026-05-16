package proxy

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func newDedupRegistry() *testRegistry {
	return newTestRegistryForDedup()
}

func newTestRegistryForDedup() *testRegistry {
	return &testRegistry{}
}

func TestDefaultDedupConfig(t *testing.T) {
	cfg := DefaultDedupConfig()
	if cfg.TTL != 5*time.Second {
		t.Errorf("expected TTL 5s, got %v", cfg.TTL)
	}
	if cfg.Header != "X-Idempotency-Key" {
		t.Errorf("unexpected header: %s", cfg.Header)
	}
}

func TestDedupMiddleware_PassesThroughWithoutKey(t *testing.T) {
	calls := 0
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.WriteHeader(http.StatusOK)
	})
	h := NewDedupMiddleware(DefaultDedupConfig(), nil, next)
	for i := 0; i < 3; i++ {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/", nil))
	}
	if calls != 3 {
		t.Errorf("expected 3 calls without key, got %d", calls)
	}
}

func TestDedupMiddleware_DeduplicatesOnSameKey(t *testing.T) {
	calls := 0
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.WriteHeader(http.StatusCreated)
	})
	cfg := DefaultDedupConfig()
	h := NewDedupMiddleware(cfg, nil, next)

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("X-Idempotency-Key", "key-abc")

	rec1 := httptest.NewRecorder()
	h.ServeHTTP(rec1, req)

	rec2 := httptest.NewRecorder()
	h.ServeHTTP(rec2, req)

	if calls != 1 {
		t.Errorf("expected handler called once, got %d", calls)
	}
	if rec2.Code != http.StatusCreated {
		t.Errorf("expected replayed status 201, got %d", rec2.Code)
	}
}

func TestDedupMiddleware_UniqueKeysCallHandlerEachTime(t *testing.T) {
	calls := 0
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.WriteHeader(http.StatusOK)
	})
	h := NewDedupMiddleware(DefaultDedupConfig(), nil, next)
	for _, key := range []string{"k1", "k2", "k3"} {
		req := httptest.NewRequest(http.MethodPost, "/", nil)
		req.Header.Set("X-Idempotency-Key", key)
		h.ServeHTTP(httptest.NewRecorder(), req)
	}
	if calls != 3 {
		t.Errorf("expected 3 calls for 3 unique keys, got %d", calls)
	}
}

func TestDedupMiddleware_ZeroTTLUsesDefault(t *testing.T) {
	cfg := DedupConfig{TTL: 0, Header: ""}
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
	h := NewDedupMiddleware(cfg, nil, next)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}
