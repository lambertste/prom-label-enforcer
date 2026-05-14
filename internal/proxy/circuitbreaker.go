package proxy

import (
	"net/http"
	"sync"
	"time"
)

// CircuitState represents the state of the circuit breaker.
type CircuitState int

const (
	StateClosed CircuitState = iota
	StateOpen
	StateHalfOpen
)

// CircuitBreakerConfig holds configuration for the circuit breaker.
type CircuitBreakerConfig struct {
	FailureThreshold int
	SuccessThreshold int
	Timeout          time.Duration
}

// DefaultCircuitBreakerConfig returns sensible defaults.
func DefaultCircuitBreakerConfig() CircuitBreakerConfig {
	return CircuitBreakerConfig{
		FailureThreshold: 5,
		SuccessThreshold: 2,
		Timeout:          30 * time.Second,
	}
}

// CircuitBreaker implements the circuit breaker pattern.
type CircuitBreaker struct {
	mu       sync.Mutex
	cfg      CircuitBreakerConfig
	state    CircuitState
	failures int
	successes int
	openedAt time.Time
}

// NewCircuitBreaker creates a new CircuitBreaker with the given config.
func NewCircuitBreaker(cfg CircuitBreakerConfig) *CircuitBreaker {
	if cfg.FailureThreshold <= 0 {
		cfg.FailureThreshold = DefaultCircuitBreakerConfig().FailureThreshold
	}
	if cfg.SuccessThreshold <= 0 {
		cfg.SuccessThreshold = DefaultCircuitBreakerConfig().SuccessThreshold
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = DefaultCircuitBreakerConfig().Timeout
	}
	return &CircuitBreaker{cfg: cfg, state: StateClosed}
}

// Allow returns true if the request should be allowed through.
func (cb *CircuitBreaker) Allow() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	switch cb.state {
	case StateClosed:
		return true
	case StateOpen:
		if time.Since(cb.openedAt) >= cb.cfg.Timeout {
			cb.state = StateHalfOpen
			cb.successes = 0
			return true
		}
		return false
	case StateHalfOpen:
		return true
	}
	return false
}

// RecordSuccess records a successful call.
func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	if cb.state == StateHalfOpen {
		cb.successes++
		if cb.successes >= cb.cfg.SuccessThreshold {
			cb.state = StateClosed
			cb.failures = 0
		}
	} else {
		cb.failures = 0
	}
}

// RecordFailure records a failed call.
func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.failures++
	if cb.state == StateHalfOpen || cb.failures >= cb.cfg.FailureThreshold {
		cb.state = StateOpen
		cb.openedAt = time.Now()
		cb.failures = 0
	}
}

// State returns the current circuit state.
func (cb *CircuitBreaker) State() CircuitState {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return cb.state
}

// Middleware wraps an http.Handler with circuit breaker logic.
func (cb *CircuitBreaker) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !cb.Allow() {
			http.Error(w, "service unavailable: circuit open", http.StatusServiceUnavailable)
			return
		}
		rw := &cbResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(rw, r)
		if rw.statusCode >= 500 {
			cb.RecordFailure()
		} else {
			cb.RecordSuccess()
		}
	})
}

type cbResponseWriter struct {
	http.ResponseWriter
	statusCode int
	wroteHeader bool
}

func (rw *cbResponseWriter) WriteHeader(code int) {
	if !rw.wroteHeader {
		rw.statusCode = code
		rw.wroteHeader = true
		rw.ResponseWriter.WriteHeader(code)
	}
}
