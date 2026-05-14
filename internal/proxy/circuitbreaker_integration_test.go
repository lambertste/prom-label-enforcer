package proxy

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// instrumentedCircuitBreaker wraps CircuitBreaker with metrics recording.
type instrumentedCircuitBreaker struct {
	*CircuitBreaker
	metrics *CircuitBreakerMetrics
	name    string
}

func newInstrumentedCB(name string, cfg CircuitBreakerConfig, reg prometheus.Registerer) *instrumentedCircuitBreaker {
	return &instrumentedCircuitBreaker{
		CircuitBreaker: NewCircuitBreaker(cfg),
		metrics:        NewCircuitBreakerMetrics(reg),
		name:           name,
	}
}

func (icb *instrumentedCircuitBreaker) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !icb.Allow() {
			icb.metrics.RecordRejected()
			icb.metrics.RecordState(icb.name, icb.State())
			http.Error(w, "service unavailable: circuit open", http.StatusServiceUnavailable)
			return
		}
		icb.metrics.RecordAllowed()
		rw := &cbResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(rw, r)
		if rw.statusCode >= 500 {
			icb.RecordFailure()
			icb.metrics.RecordTrip()
		} else {
			icb.RecordSuccess()
		}
		icb.metrics.RecordState(icb.name, icb.State())
	})
}

func TestInstrumentedCircuitBreaker_AllowsAndRecords(t *testing.T) {
	reg := prometheus.NewRegistry()
	icb := newInstrumentedCB("test", DefaultCircuitBreakerConfig(), reg)
	handler := icb.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/receive", nil))
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	if icb.State() != StateClosed {
		t.Errorf("expected StateClosed, got %v", icb.State())
	}
}

func TestInstrumentedCircuitBreaker_OpensOnErrors(t *testing.T) {
	reg := prometheus.NewRegistry()
	icb := newInstrumentedCB("test", CircuitBreakerConfig{
		FailureThreshold: 2,
		SuccessThreshold: 1,
		Timeout:          10 * time.Millisecond,
	}, reg)
	errorHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	handler := icb.Middleware(errorHandler)
	for i := 0; i < 2; i++ {
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/receive", nil))
	}
	if icb.State() != StateOpen {
		t.Errorf("expected StateOpen after failures, got %v", icb.State())
	}
	// Next request should be rejected.
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/receive", nil))
	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("expected 503, got %d", rec.Code)
	}
}
