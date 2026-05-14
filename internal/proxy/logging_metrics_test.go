package proxy

import (
	"net/http"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

func newLoggingTestRegistry() prometheus.Registerer {
	return prometheus.NewRegistry()
}

func TestNewLoggingMetrics_NotNil(t *testing.T) {
	m := NewLoggingMetrics(newLoggingTestRegistry())
	if m == nil {
		t.Fatal("expected non-nil metrics")
	}
}

func TestNewLoggingMetrics_NilRegisterer(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("unexpected panic: %v", r)
		}
	}()
	m := NewLoggingMetrics(nil)
	if m == nil {
		t.Fatal("expected non-nil metrics")
	}
}

func TestLoggingMetrics_RecordRequest(t *testing.T) {
	m := NewLoggingMetrics(newLoggingTestRegistry())
	m.RecordRequest(http.MethodGet, http.StatusOK, 100*time.Millisecond, false)
}

func TestLoggingMetrics_RecordSlowRequest(t *testing.T) {
	m := NewLoggingMetrics(newLoggingTestRegistry())
	m.RecordRequest(http.MethodPost, http.StatusAccepted, 600*time.Millisecond, true)
}

func TestLoggingMetrics_NilSafe(t *testing.T) {
	var m *LoggingMetrics
	m.RecordRequest(http.MethodGet, http.StatusOK, time.Second, false)
}
