package proxy_test

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/prom-label-enforcer/internal/proxy"
)

func newCORSTestRegistry() prometheus.Registerer {
	return prometheus.NewPedanticRegistry()
}

func TestNewCORSMetrics_NotNil(t *testing.T) {
	m := proxy.NewCORSMetrics(newCORSTestRegistry())
	if m == nil {
		t.Fatal("expected non-nil CORSMetrics")
	}
}

func TestNewCORSMetrics_NilRegisterer(t *testing.T) {
	m := proxy.NewCORSMetrics(nil)
	if m == nil {
		t.Fatal("expected non-nil CORSMetrics when registerer is nil")
	}
}

func TestCORSMetrics_RecordAllowed(t *testing.T) {
	reg := prometheus.NewPedanticRegistry()
	m := proxy.NewCORSMetrics(reg)

	if err := m.RecordAllowed("https://example.com"); err != nil {
		t.Fatalf("RecordAllowed returned error: %v", err)
	}

	mfs, err := reg.Gather()
	if err != nil {
		t.Fatalf("gather: %v", err)
	}
	if len(mfs) == 0 {
		t.Fatal("expected at least one metric family")
	}
}

func TestCORSMetrics_RecordDenied(t *testing.T) {
	reg := prometheus.NewPedanticRegistry()
	m := proxy.NewCORSMetrics(reg)

	if err := m.RecordDenied("https://evil.com"); err != nil {
		t.Fatalf("RecordDenied returned error: %v", err)
	}

	mfs, err := reg.Gather()
	if err != nil {
		t.Fatalf("gather: %v", err)
	}
	if len(mfs) == 0 {
		t.Fatal("expected at least one metric family")
	}
}

func TestCORSMetrics_NilSafe(t *testing.T) {
	var m *proxy.CORSMetrics

	if err := m.RecordAllowed("https://example.com"); err != nil {
		t.Fatalf("nil RecordAllowed should not error: %v", err)
	}
	if err := m.RecordDenied("https://evil.com"); err != nil {
		t.Fatalf("nil RecordDenied should not error: %v", err)
	}
}
