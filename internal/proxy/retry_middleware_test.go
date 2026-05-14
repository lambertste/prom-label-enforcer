package proxy

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestDefaultRetryConfig(t *testing.T) {
	cfg := DefaultRetryConfig()
	if cfg.MaxAttempts != 3 {
		t.Errorf("expected MaxAttempts=3, got %d", cfg.MaxAttempts)
	}
	if cfg.Delay != 100*time.Millisecond {
		t.Errorf("expected Delay=100ms, got %s", cfg.Delay)
	}
	if !cfg.RetryableStatuses[http.StatusBadGateway] {
		t.Error("expected 502 to be retryable")
	}
}

func TestRetryMiddleware_NoRetryOnSuccess(t *testing.T) {
	var calls int32
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.WriteHeader(http.StatusOK)
	})

	cfg := DefaultRetryConfig()
	cfg.Delay = 0
	mw := NewRetryMiddleware(cfg, h)

	rec := httptest.NewRecorder()
	mw.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/receive", nil))

	if atomic.LoadInt32(&calls) != 1 {
		t.Errorf("expected 1 call, got %d", calls)
	}
}

func TestRetryMiddleware_RetriesOnRetryableStatus(t *testing.T) {
	var calls int32
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.WriteHeader(http.StatusServiceUnavailable)
	})

	cfg := DefaultRetryConfig()
	cfg.MaxAttempts = 3
	cfg.Delay = 0
	mw := NewRetryMiddleware(cfg, h)

	rec := httptest.NewRecorder()
	mw.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/receive", nil))

	if atomic.LoadInt32(&calls) != 3 {
		t.Errorf("expected 3 calls, got %d", calls)
	}
}

func TestRetryMiddleware_StopsAfterSuccess(t *testing.T) {
	var calls int32
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&calls, 1)
		if n < 2 {
			w.WriteHeader(http.StatusBadGateway)
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	cfg := DefaultRetryConfig()
	cfg.MaxAttempts = 5
	cfg.Delay = 0
	mw := NewRetryMiddleware(cfg, h)

	rec := httptest.NewRecorder()
	mw.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/receive", nil))

	if atomic.LoadInt32(&calls) != 2 {
		t.Errorf("expected 2 calls, got %d", calls)
	}
}

func TestRetryMiddleware_ZeroMaxAttemptsUsesDefault(t *testing.T) {
	var calls int32
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.WriteHeader(http.StatusBadGateway)
	})

	cfg := RetryConfig{MaxAttempts: 0, Delay: 0}
	mw := NewRetryMiddleware(cfg, h)

	rec := httptest.NewRecorder()
	mw.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/receive", nil))

	if atomic.LoadInt32(&calls) != int32(DefaultRetryConfig().MaxAttempts) {
		t.Errorf("expected %d calls, got %d", DefaultRetryConfig().MaxAttempts, calls)
	}
}
