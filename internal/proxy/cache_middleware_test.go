package proxy

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestDefaultCacheConfig(t *testing.T) {
	cfg := DefaultCacheConfig()
	if cfg.TTL <= 0 {
		t.Errorf("expected positive TTL, got %v", cfg.TTL)
	}
	if cfg.MaxSize <= 0 {
		t.Errorf("expected positive MaxSize, got %d", cfg.MaxSize)
	}
}

func TestNewResponseCache_ZeroValuesUseDefaults(t *testing.T) {
	c := NewResponseCache(CacheConfig{})
	if c.cfg.TTL != DefaultCacheConfig().TTL {
		t.Errorf("expected default TTL")
	}
	if c.cfg.MaxSize != DefaultCacheConfig().MaxSize {
		t.Errorf("expected default MaxSize")
	}
}

func TestResponseCache_MissOnEmpty(t *testing.T) {
	c := NewResponseCache(DefaultCacheConfig())
	_, _, ok := c.Get("/metrics")
	if ok {
		t.Error("expected cache miss on empty cache")
	}
}

func TestResponseCache_HitAfterSet(t *testing.T) {
	c := NewResponseCache(DefaultCacheConfig())
	c.Set("/metrics", []byte("data"), http.StatusOK)
	body, status, ok := c.Get("/metrics")
	if !ok {
		t.Fatal("expected cache hit")
	}
	if status != http.StatusOK {
		t.Errorf("expected 200, got %d", status)
	}
	if string(body) != "data" {
		t.Errorf("expected 'data', got %s", body)
	}
}

func TestResponseCache_ExpiredEntry(t *testing.T) {
	cfg := CacheConfig{TTL: 10 * time.Millisecond, MaxSize: 10}
	c := NewResponseCache(cfg)
	c.Set("/metrics", []byte("data"), http.StatusOK)
	time.Sleep(20 * time.Millisecond)
	_, _, ok := c.Get("/metrics")
	if ok {
		t.Error("expected cache miss after TTL expiry")
	}
}

func TestCacheMiddleware_CachesGetRequests(t *testing.T) {
	var calls atomic.Int32
	backend := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("hello"))
	})

	c := NewResponseCache(DefaultCacheConfig())
	h := c.Middleware(backend)

	for i := 0; i < 3; i++ {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
		h.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
	}
	if calls.Load() != 1 {
		t.Errorf("expected backend called once, got %d", calls.Load())
	}
}

func TestCacheMiddleware_SkipsNonGet(t *testing.T) {
	var calls atomic.Int32
	backend := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusNoContent)
	})

	c := NewResponseCache(DefaultCacheConfig())
	h := c.Middleware(backend)

	for i := 0; i < 3; i++ {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/receive", nil)
		h.ServeHTTP(rec, req)
	}
	if calls.Load() != 3 {
		t.Errorf("expected backend called 3 times for POST, got %d", calls.Load())
	}
}

func TestResponseCache_MaxSizeEviction(t *testing.T) {
	c := NewResponseCache(CacheConfig{TTL: time.Minute, MaxSize: 2})
	c.Set("/a", []byte("a"), 200)
	c.Set("/b", []byte("b"), 200)
	c.Set("/c", []byte("c"), 200) // should evict one entry
	count := 0
	for _, k := range []string{"/a", "/b", "/c"} {
		if _, _, ok := c.Get(k); ok {
			count++
		}
	}
	if count > 2 {
		t.Errorf("expected at most 2 entries after eviction, got %d", count)
	}
}
