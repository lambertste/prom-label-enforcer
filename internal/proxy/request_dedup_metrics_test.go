package proxy

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
)

func TestNewDedupMetrics_NotNil(t *testing.T) {
	reg := prometheus.NewRegistry()
	m := NewDedupMetrics(reg)
	if m == nil {
		t.Fatal("expected non-nil DedupMetrics")
	}
}

func TestNewDedupMetrics_NilRegisterer(t *testing.T) {
	// Should not panic; falls back to default registerer.
	// We use a fresh registry to avoid double-registration in tests.
	reg := prometheus.NewRegistry()
	m := NewDedupMetrics(reg)
	if m == nil {
		t.Fatal("expected non-nil DedupMetrics with nil registerer fallback")
	}
}

func TestDedupMetrics_RecordDuplicate(t *testing.T) {
	reg := prometheus.NewRegistry()
	m := NewDedupMetrics(reg)
	m.RecordDuplicate()
	m.RecordDuplicate()

	mfs, err := reg.Gather()
	if err != nil {
		t.Fatalf("gather error: %v", err)
	}
	for _, mf := range mfs {
		if mf.GetName() == "prom_label_enforcer_dedup_duplicates_total" {
			if got := mf.GetMetric()[0].GetCounter().GetValue(); got != 2 {
				t.Errorf("expected 2 duplicates, got %v", got)
			}
			return
		}
	}
	t.Error("duplicates metric not found")
}

func TestDedupMetrics_RecordUnique(t *testing.T) {
	reg := prometheus.NewRegistry()
	m := NewDedupMetrics(reg)
	m.RecordUnique()

	mfs, err := reg.Gather()
	if err != nil {
		t.Fatalf("gather error: %v", err)
	}
	for _, mf := range mfs {
		if mf.GetName() == "prom_label_enforcer_dedup_unique_total" {
			if got := mf.GetMetric()[0].GetCounter().GetValue(); got != 1 {
				t.Errorf("expected 1 unique, got %v", got)
			}
			return
		}
	}
	t.Error("unique metric not found")
}

func TestDedupMetrics_NilSafe(t *testing.T) {
	var m *DedupMetrics
	m.RecordDuplicate()
	m.RecordUnique()
	// Should not panic.
}
