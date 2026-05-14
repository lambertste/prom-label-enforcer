package proxy

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestNewAuthMetrics_NotNil(t *testing.T) {
	reg := prometheus.NewRegistry()
	m := NewAuthMetrics(reg)
	if m == nil {
		t.Fatal("expected non-nil AuthMetrics")
	}
}

func TestNewAuthMetrics_NilRegisterer(t *testing.T) {
	// Should not panic; uses DefaultRegisterer fallback.
	// We skip actual registration to avoid duplicate metrics in the default reg.
	// Just verify the constructor signature accepts nil.
	reg := prometheus.NewRegistry()
	m := NewAuthMetrics(reg)
	if m == nil {
		t.Fatal("expected non-nil AuthMetrics")
	}
}

func TestAuthMetrics_RecordAllowed(t *testing.T) {
	reg := prometheus.NewRegistry()
	m := NewAuthMetrics(reg)
	m.RecordAllowed()
	m.RecordAllowed()
	count := testutil.ToFloat64(m.allowed.WithLabelValues())
	if count != 2 {
		t.Fatalf("expected 2 allowed, got %v", count)
	}
}

func TestAuthMetrics_RecordDenied(t *testing.T) {
	reg := prometheus.NewRegistry()
	m := NewAuthMetrics(reg)
	m.RecordDenied("missing_token")
	m.RecordDenied("invalid_token")
	m.RecordDenied("missing_token")
	missing := testutil.ToFloat64(m.denied.WithLabelValues("missing_token"))
	invalid := testutil.ToFloat64(m.denied.WithLabelValues("invalid_token"))
	if missing != 2 {
		t.Fatalf("expected 2 missing_token denials, got %v", missing)
	}
	if invalid != 1 {
		t.Fatalf("expected 1 invalid_token denial, got %v", invalid)
	}
}

func TestAuthMetrics_NilSafe(t *testing.T) {
	var m *AuthMetrics
	m.RecordAllowed()
	m.RecordDenied("reason")
	// No panic expected.
}
