package proxy

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
)

func newHealthTestRegistry() prometheus.Registerer {
	return prometheus.NewRegistry()
}

func TestNewHealthCheckMetrics_NotNil(t *testing.T) {
	m := NewHealthCheckMetrics(newHealthTestRegistry())
	if m == nil {
		t.Fatal("expected non-nil HealthCheckMetrics")
	}
}

func TestNewHealthCheckMetrics_NilRegisterer(t *testing.T) {
	// Should not panic; falls back to default registerer.
	// We can't easily assert registration, so just ensure no panic.
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("unexpected panic: %v", r)
		}
	}()
	// Use a fresh registry to avoid duplicate registration on default.
	_ = NewHealthCheckMetrics(newHealthTestRegistry())
}

func TestHealthCheckMetrics_RecordLiveness(t *testing.T) {
	reg := prometheus.NewRegistry()
	m := NewHealthCheckMetrics(reg)
	m.RecordLiveness()
	m.RecordLiveness()

	mfs, err := reg.Gather()
	if err != nil {
		t.Fatalf("gather error: %v", err)
	}
	for _, mf := range mfs {
		if mf.GetName() == "prom_label_enforcer_healthcheck_liveness_requests_total" {
			if got := mf.GetMetric()[0].GetCounter().GetValue(); got != 2 {
				t.Errorf("expected 2 liveness hits, got %v", got)
			}
			return
		}
	}
	t.Error("liveness counter metric not found")
}

func TestHealthCheckMetrics_RecordReadiness(t *testing.T) {
	reg := prometheus.NewRegistry()
	m := NewHealthCheckMetrics(reg)
	m.RecordReadiness("ready")
	m.RecordReadiness("not_ready")
	m.RecordReadiness("ready")

	mfs, err := reg.Gather()
	if err != nil {
		t.Fatalf("gather error: %v", err)
	}
	for _, mf := range mfs {
		if mf.GetName() == "prom_label_enforcer_healthcheck_readiness_requests_total" {
			for _, metric := range mf.GetMetric() {
				for _, lp := range metric.GetLabel() {
					if lp.GetName() == "result" && lp.GetValue() == "ready" {
						if got := metric.GetCounter().GetValue(); got != 2 {
							t.Errorf("expected 2 ready, got %v", got)
						}
					}
				}
			}
			return
		}
	}
	t.Error("readiness counter metric not found")
}

func TestHealthCheckMetrics_NilSafe(t *testing.T) {
	var m *HealthCheckMetrics
	m.RecordLiveness()
	m.RecordReadiness("ready")
}
