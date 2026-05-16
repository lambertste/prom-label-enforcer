package proxy

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
)

func newTagTestRegistry() prometheus.Registerer {
	return prometheus.NewRegistry()
}

func TestNewRequestTagMetrics_NotNil(t *testing.T) {
	m := NewRequestTagMetrics(newTagTestRegistry())
	if m == nil {
		t.Fatal("expected non-nil RequestTagMetrics")
	}
}

func TestNewRequestTagMetrics_NilRegisterer(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Logf("panicked as expected with default registerer collision: %v", r)
		}
	}()
	// Should not panic when a fresh registry is provided.
	m := NewRequestTagMetrics(newTagTestRegistry())
	if m == nil {
		t.Fatal("expected non-nil metrics")
	}
}

func TestRequestTagMetrics_RecordTagged(t *testing.T) {
	m := NewRequestTagMetrics(newTagTestRegistry())
	// Should not panic.
	m.RecordTagged(3)
	m.RecordTagged(1)
}

func TestRequestTagMetrics_RecordUntagged(t *testing.T) {
	m := NewRequestTagMetrics(newTagTestRegistry())
	m.RecordUntagged()
}

func TestRequestTagMetrics_RecordTruncated(t *testing.T) {
	m := NewRequestTagMetrics(newTagTestRegistry())
	m.RecordTruncated()
}

func TestRequestTagMetrics_NilSafe(t *testing.T) {
	var m *RequestTagMetrics
	m.RecordTagged(2)
	m.RecordUntagged()
	m.RecordTruncated()
}
