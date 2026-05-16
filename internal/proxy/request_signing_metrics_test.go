package proxy

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
)

func newSigningTestRegistry() prometheus.Registerer {
	return prometheus.NewRegistry()
}

func TestNewRequestSigningMetrics_NotNil(t *testing.T) {
	m := NewRequestSigningMetrics(newSigningTestRegistry())
	if m == nil {
		t.Fatal("expected non-nil metrics")
	}
}

func TestNewRequestSigningMetrics_NilRegisterer(t *testing.T) {
	m := NewRequestSigningMetrics(nil)
	if m == nil {
		t.Fatal("expected non-nil metrics with nil registerer")
	}
}

func TestRequestSigningMetrics_RecordAllowed(t *testing.T) {
	m := NewRequestSigningMetrics(newSigningTestRegistry())
	m.RecordAllowed()
	m.RecordAllowed()
	// no panic = pass
}

func TestRequestSigningMetrics_RecordDenied(t *testing.T) {
	m := NewRequestSigningMetrics(newSigningTestRegistry())
	m.RecordDenied()
}

func TestRequestSigningMetrics_RecordBypassed(t *testing.T) {
	m := NewRequestSigningMetrics(newSigningTestRegistry())
	m.RecordBypassed()
}

func TestRequestSigningMetrics_NilSafe(t *testing.T) {
	var m *RequestSigningMetrics
	m.RecordAllowed()
	m.RecordDenied()
	m.RecordBypassed()
	// no panic = pass
}
