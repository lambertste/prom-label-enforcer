package enforcer

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

func newTestRegistry() prometheus.Registerer {
	return prometheus.NewRegistry()
}

func TestNewMetrics_NotNil(t *testing.T) {
	reg := prometheus.NewRegistry()
	m := NewMetrics(reg)
	if m == nil {
		t.Fatal("expected non-nil Metrics")
	}
	if m.Allowed == nil {
		t.Error("expected non-nil Allowed counter")
	}
	if m.Rejected == nil {
		t.Error("expected non-nil Rejected counter vec")
	}
}

func TestMetrics_RecordAllowed(t *testing.T) {
	reg := prometheus.NewRegistry()
	m := NewMetrics(reg)

	m.RecordAllowed()
	m.RecordAllowed()

	count := testutil.ToFloat64(m.Allowed)
	if count != 2 {
		t.Errorf("expected allowed count 2, got %v", count)
	}
}

func TestMetrics_RecordRejected(t *testing.T) {
	reg := prometheus.NewRegistry()
	m := NewMetrics(reg)

	m.RecordRejected("missing_label")
	m.RecordRejected("missing_label")
	m.RecordRejected("disallowed_value")

	missingCount := testutil.ToFloat64(m.Rejected.WithLabelValues("missing_label"))
	if missingCount != 2 {
		t.Errorf("expected missing_label count 2, got %v", missingCount)
	}

	disallowedCount := testutil.ToFloat64(m.Rejected.WithLabelValues("disallowed_value"))
	if disallowedCount != 1 {
		t.Errorf("expected disallowed_value count 1, got %v", disallowedCount)
	}
}

func TestMetrics_NilSafe(t *testing.T) {
	var m *Metrics
	// Should not panic
	m.RecordAllowed()
	m.RecordRejected("missing_label")
}
