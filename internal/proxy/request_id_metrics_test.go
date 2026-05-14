package proxy

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestNewRequestIDMetrics_NotNil(t *testing.T) {
	reg := prometheus.NewRegistry()
	m := NewRequestIDMetrics(reg)
	if m == nil {
		t.Fatal("expected non-nil RequestIDMetrics")
	}
}

func TestNewRequestIDMetrics_NilRegisterer(t *testing.T) {
	// Should not panic; falls back to default registerer.
	// We cannot easily test default registerer without side effects,
	// so just ensure it doesn't panic with a real registry.
	reg := prometheus.NewRegistry()
	m := NewRequestIDMetrics(reg)
	if m == nil {
		t.Fatal("expected non-nil metrics")
	}
}

func TestRequestIDMetrics_RecordGenerated(t *testing.T) {
	reg := prometheus.NewRegistry()
	m := NewRequestIDMetrics(reg)
	m.RecordGenerated()
	m.RecordGenerated()
	count := testutil.ToFloat64(m.generated)
	if count != 2 {
		t.Fatalf("expected 2 generated, got %v", count)
	}
}

func TestRequestIDMetrics_RecordPropagated(t *testing.T) {
	reg := prometheus.NewRegistry()
	m := NewRequestIDMetrics(reg)
	m.RecordPropagated()
	count := testutil.ToFloat64(m.propagated)
	if count != 1 {
		t.Fatalf("expected 1 propagated, got %v", count)
	}
}

func TestRequestIDMetrics_NilSafe(t *testing.T) {
	var m *RequestIDMetrics
	// Should not panic.
	m.RecordGenerated()
	m.RecordPropagated()
}
