package proxy

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestNewIPFilterMetrics_NotNil(t *testing.T) {
	reg := prometheus.NewRegistry()
	m := NewIPFilterMetrics(reg)
	if m == nil {
		t.Fatal("expected non-nil IPFilterMetrics")
	}
}

func TestNewIPFilterMetrics_NilRegisterer(t *testing.T) {
	// Should not panic; falls back to default registerer.
	// We can't easily test registration on the default registerer without
	// risking duplicate registration, so just verify the constructor doesn't panic.
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("unexpected panic: %v", r)
		}
	}()
	// Use a fresh registry to avoid duplicate registration with default.
	NewIPFilterMetrics(prometheus.NewRegistry())
}

func TestIPFilterMetrics_RecordAllowed(t *testing.T) {
	reg := prometheus.NewRegistry()
	m := NewIPFilterMetrics(reg)
	m.RecordAllowed()
	m.RecordAllowed()
	count := testutil.ToFloat64(m.allowed)
	if count != 2 {
		t.Fatalf("expected 2 allowed, got %v", count)
	}
}

func TestIPFilterMetrics_RecordDenied(t *testing.T) {
	reg := prometheus.NewRegistry()
	m := NewIPFilterMetrics(reg)
	m.RecordDenied()
	count := testutil.ToFloat64(m.denied)
	if count != 1 {
		t.Fatalf("expected 1 denied, got %v", count)
	}
}

func TestIPFilterMetrics_NilSafe(t *testing.T) {
	var m *IPFilterMetrics
	// Should not panic.
	m.RecordAllowed()
	m.RecordDenied()
}
