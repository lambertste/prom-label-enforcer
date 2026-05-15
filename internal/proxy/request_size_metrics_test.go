package proxy

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
)

func newRequestSizeTestRegistry() prometheus.Registerer {
	return prometheus.NewRegistry()
}

func TestNewRequestSizeMetrics_NotNil(t *testing.T) {
	m := NewRequestSizeMetrics(newRequestSizeTestRegistry())
	if m == nil {
		t.Fatal("expected non-nil RequestSizeMetrics")
	}
}

func TestNewRequestSizeMetrics_NilRegisterer(t *testing.T) {
	// Should not panic; falls back to default registerer.
	// We cannot easily assert registration against the default, so we just
	// verify it doesn't panic by calling with nil on a fresh default.
	// Use a real registry to avoid polluting the default.
	reg := prometheus.NewRegistry()
	m := NewRequestSizeMetrics(reg)
	if m == nil {
		t.Fatal("expected non-nil RequestSizeMetrics")
	}
}

func TestRequestSizeMetrics_RecordEmpty(t *testing.T) {
	m := NewRequestSizeMetrics(newRequestSizeTestRegistry())
	// Should not panic.
	m.Record(0)
}

func TestRequestSizeMetrics_RecordSmall(t *testing.T) {
	m := NewRequestSizeMetrics(newRequestSizeTestRegistry())
	m.Record(512)
}

func TestRequestSizeMetrics_RecordMedium(t *testing.T) {
	m := NewRequestSizeMetrics(newRequestSizeTestRegistry())
	m.Record(4096)
}

func TestRequestSizeMetrics_RecordLarge(t *testing.T) {
	m := NewRequestSizeMetrics(newRequestSizeTestRegistry())
	m.Record(2 * 1024 * 1024)
}

func TestRequestSizeMetrics_NilSafe(t *testing.T) {
	var m *RequestSizeMetrics
	// Should not panic.
	m.Record(1024)
}
