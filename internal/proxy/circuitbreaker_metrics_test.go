package proxy

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
)

func newCBTestRegistry() prometheus.Registerer {
	return prometheus.NewRegistry()
}

func TestNewCircuitBreakerMetrics_NotNil(t *testing.T) {
	m := NewCircuitBreakerMetrics(newCBTestRegistry())
	if m == nil {
		t.Fatal("expected non-nil CircuitBreakerMetrics")
	}
}

func TestNewCircuitBreakerMetrics_NilRegisterer(t *testing.T) {
	// Should not panic with nil registerer.
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("unexpected panic: %v", r)
		}
	}()
	m := NewCircuitBreakerMetrics(nil)
	if m == nil {
		t.Fatal("expected non-nil CircuitBreakerMetrics")
	}
}

func TestCircuitBreakerMetrics_RecordState(t *testing.T) {
	m := NewCircuitBreakerMetrics(newCBTestRegistry())
	// Should not panic.
	m.RecordState("upstream", StateClosed)
	m.RecordState("upstream", StateOpen)
	m.RecordState("upstream", StateHalfOpen)
}

func TestCircuitBreakerMetrics_RecordTrip(t *testing.T) {
	m := NewCircuitBreakerMetrics(newCBTestRegistry())
	// Should not panic.
	m.RecordTrip()
	m.RecordTrip()
}

func TestCircuitBreakerMetrics_RecordAllowed(t *testing.T) {
	m := NewCircuitBreakerMetrics(newCBTestRegistry())
	m.RecordAllowed()
}

func TestCircuitBreakerMetrics_RecordRejected(t *testing.T) {
	m := NewCircuitBreakerMetrics(newCBTestRegistry())
	m.RecordRejected()
}

func TestCircuitBreakerMetrics_NilSafe(t *testing.T) {
	var m *CircuitBreakerMetrics
	// All methods should be nil-safe.
	m.RecordState("x", StateClosed)
	m.RecordTrip()
	m.RecordAllowed()
	m.RecordRejected()
}
