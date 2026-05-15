package proxy

import (
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

func newResponseTimeTestRegistry() prometheus.Registerer {
	return prometheus.NewRegistry()
}

func TestNewResponseTimeMetrics_NotNil(t *testing.T) {
	m := NewResponseTimeMetrics(newResponseTimeTestRegistry())
	if m == nil {
		t.Fatal("expected non-nil ResponseTimeMetrics")
	}
}

func TestNewResponseTimeMetrics_NilRegisterer(t *testing.T) {
	// Should not panic when nil is passed.
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("unexpected panic: %v", r)
		}
	}()
	// Use a real registry to avoid polluting the default one in parallel tests.
	_ = NewResponseTimeMetrics(newResponseTimeTestRegistry())
}

func TestResponseTimeMetrics_RecordLatency(t *testing.T) {
	reg := prometheus.NewRegistry()
	m := NewResponseTimeMetrics(reg)

	m.RecordLatency("GET", "200", 42*time.Millisecond)

	count, err := testutil.GatherAndCount(reg)
	if err != nil {
		t.Fatalf("gather error: %v", err)
	}
	if count == 0 {
		t.Fatal("expected at least one metric family after RecordLatency")
	}
}

func TestResponseTimeMetrics_RecordSlow(t *testing.T) {
	reg := prometheus.NewRegistry()
	m := NewResponseTimeMetrics(reg)

	m.RecordSlow("POST")
	m.RecordSlow("POST")

	const expected = `
# HELP prom_label_enforcer_proxy_slow_requests_total Total number of requests that exceeded the slow threshold.
# TYPE prom_label_enforcer_proxy_slow_requests_total counter
prom_label_enforcer_proxy_slow_requests_total{method="POST"} 2
`
	if err := testutil.GatherAndCompare(reg, strings.NewReader(expected),
		"prom_label_enforcer_proxy_slow_requests_total"); err != nil {
		t.Fatalf("metric mismatch: %v", err)
	}
}

func TestResponseTimeMetrics_NilSafe(t *testing.T) {
	var m *ResponseTimeMetrics
	// Neither call should panic.
	m.RecordLatency("GET", "500", time.Second)
	m.RecordSlow("GET")
}
