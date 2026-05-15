package proxy

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
)

func TestNewRewriteMetrics_NotNil(t *testing.T) {
	reg := prometheus.NewRegistry()
	m := NewRewriteMetrics(reg)
	if m == nil {
		t.Fatal("expected non-nil RewriteMetrics")
	}
}

func TestNewRewriteMetrics_NilRegisterer(t *testing.T) {
	// Should not panic when registerer is nil
	m := NewRewriteMetrics(nil)
	if m == nil {
		t.Fatal("expected non-nil RewriteMetrics")
	}
}

func TestRewriteMetrics_RecordRewritten(t *testing.T) {
	reg := prometheus.NewRegistry()
	m := NewRewriteMetrics(reg)
	// Should not panic
	m.RecordRewritten()
	m.RecordRewritten()

	mfs, err := reg.Gather()
	if err != nil {
		t.Fatalf("gather error: %v", err)
	}
	for _, mf := range mfs {
		if mf.GetName() == "proxy_rewrite_rewritten_total" {
			if got := mf.GetMetric()[0].GetCounter().GetValue(); got != 2 {
				t.Errorf("expected 2, got %v", got)
			}
			return
		}
	}
	t.Error("metric proxy_rewrite_rewritten_total not found")
}

func TestRewriteMetrics_RecordPassthrough(t *testing.T) {
	reg := prometheus.NewRegistry()
	m := NewRewriteMetrics(reg)
	m.RecordPassthrough()

	mfs, err := reg.Gather()
	if err != nil {
		t.Fatalf("gather error: %v", err)
	}
	for _, mf := range mfs {
		if mf.GetName() == "proxy_rewrite_passthrough_total" {
			if got := mf.GetMetric()[0].GetCounter().GetValue(); got != 1 {
				t.Errorf("expected 1, got %v", got)
			}
			return
		}
	}
	t.Error("metric proxy_rewrite_passthrough_total not found")
}

func TestRewriteMetrics_NilSafe(t *testing.T) {
	var m *RewriteMetrics
	// Neither call should panic on nil receiver
	m.RecordRewritten()
	m.RecordPassthrough()
}
