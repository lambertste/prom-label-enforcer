package proxy

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestDefaultCircuitBreakerConfig(t *testing.T) {
	cfg := DefaultCircuitBreakerConfig()
	if cfg.FailureThreshold <= 0 {
		t.Error("expected positive FailureThreshold")
	}
	if cfg.SuccessThreshold <= 0 {
		t.Error("expected positive SuccessThreshold")
	}
	if cfg.Timeout <= 0 {
		t.Error("expected positive Timeout")
	}
}

func TestNewCircuitBreaker_ZeroValuesUseDefaults(t *testing.T) {
	cb := NewCircuitBreaker(CircuitBreakerConfig{})
	defaults := DefaultCircuitBreakerConfig()
	if cb.cfg.FailureThreshold != defaults.FailureThreshold {
		t.Errorf("expected %d, got %d", defaults.FailureThreshold, cb.cfg.FailureThreshold)
	}
}

func TestCircuitBreaker_InitiallyClosed(t *testing.T) {
	cb := NewCircuitBreaker(DefaultCircuitBreakerConfig())
	if cb.State() != StateClosed {
		t.Error("expected circuit to be closed initially")
	}
	if !cb.Allow() {
		t.Error("expected Allow() to return true when closed")
	}
}

func TestCircuitBreaker_OpensAfterThreshold(t *testing.T) {
	cb := NewCircuitBreaker(CircuitBreakerConfig{
		FailureThreshold: 3,
		SuccessThreshold: 2,
		Timeout:          100 * time.Millisecond,
	})
	for i := 0; i < 3; i++ {
		cb.RecordFailure()
	}
	if cb.State() != StateOpen {
		t.Errorf("expected StateOpen, got %v", cb.State())
	}
	if cb.Allow() {
		t.Error("expected Allow() to return false when open")
	}
}

func TestCircuitBreaker_TransitionsToHalfOpen(t *testing.T) {
	cb := NewCircuitBreaker(CircuitBreakerConfig{
		FailureThreshold: 1,
		SuccessThreshold: 1,
		Timeout:          10 * time.Millisecond,
	})
	cb.RecordFailure()
	time.Sleep(20 * time.Millisecond)
	if !cb.Allow() {
		t.Error("expected Allow() after timeout")
	}
	if cb.State() != StateHalfOpen {
		t.Errorf("expected StateHalfOpen, got %v", cb.State())
	}
}

func TestCircuitBreaker_ClosesAfterSuccessInHalfOpen(t *testing.T) {
	cb := NewCircuitBreaker(CircuitBreakerConfig{
		FailureThreshold: 1,
		SuccessThreshold: 2,
		Timeout:          10 * time.Millisecond,
	})
	cb.RecordFailure()
	time.Sleep(20 * time.Millisecond)
	cb.Allow() // transition to half-open
	cb.RecordSuccess()
	cb.RecordSuccess()
	if cb.State() != StateClosed {
		t.Errorf("expected StateClosed after successes, got %v", cb.State())
	}
}

func TestCircuitBreaker_Middleware_Allows(t *testing.T) {
	cb := NewCircuitBreaker(DefaultCircuitBreakerConfig())
	handler := cb.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestCircuitBreaker_Middleware_Rejects(t *testing.T) {
	cb := NewCircuitBreaker(CircuitBreakerConfig{
		FailureThreshold: 1,
		SuccessThreshold: 1,
		Timeout:          10 * time.Second,
	})
	cb.RecordFailure() // open the circuit
	handler := cb.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("expected 503, got %d", rec.Code)
	}
}
