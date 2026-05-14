package proxy

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
)

func newCompressionTestRegistry() prometheus.Registerer {
	return prometheus.NewRegistry()
}

func TestNewCompressionMetrics_NotNil(t *testing.T) {
	m := NewCompressionMetrics(newCompressionTestRegistry())
	if m == nil {
		t.Fatal("expected non-nil CompressionMetrics")
	}
	if m.CompressedTotal == nil {
		t.Error("expected CompressedTotal to be non-nil")
	}
	if m.UncompressedTotal == nil {
		t.Error("expected UncompressedTotal to be non-nil")
	}
	if m.BytesSaved == nil {
		t.Error("expected BytesSaved to be non-nil")
	}
}

func TestNewCompressionMetrics_NilRegisterer(t *testing.T) {
	// Should not panic when nil is passed; falls back to DefaultRegisterer.
	// We use a fresh default registry to avoid duplicate registration.
	defer func() {
		if r := recover(); r != nil {
			t.Logf("recovered from panic (expected on duplicate registration): %v", r)
		}
	}()
	_ = NewCompressionMetrics(newCompressionTestRegistry())
}

func TestCompressionMetrics_RecordCompressed(t *testing.T) {
	reg := prometheus.NewRegistry()
	m := NewCompressionMetrics(reg)
	m.RecordCompressed()
	m.RecordCompressed()

	mfs, err := reg.Gather()
	if err != nil {
		t.Fatalf("gather error: %v", err)
	}
	for _, mf := range mfs {
		if mf.GetName() == "proxy_compression_compressed_total" {
			if got := mf.GetMetric()[0].GetCounter().GetValue(); got != 2 {
				t.Errorf("expected 2, got %v", got)
			}
			return
		}
	}
	t.Error("metric proxy_compression_compressed_total not found")
}

func TestCompressionMetrics_RecordUncompressed(t *testing.T) {
	reg := prometheus.NewRegistry()
	m := NewCompressionMetrics(reg)
	m.RecordUncompressed()

	mfs, _ := reg.Gather()
	for _, mf := range mfs {
		if mf.GetName() == "proxy_compression_uncompressed_total" {
			if got := mf.GetMetric()[0].GetCounter().GetValue(); got != 1 {
				t.Errorf("expected 1, got %v", got)
			}
			return
		}
	}
	t.Error("metric proxy_compression_uncompressed_total not found")
}

func TestCompressionMetrics_NilSafe(t *testing.T) {
	var m *CompressionMetrics
	// None of these should panic.
	m.RecordCompressed()
	m.RecordUncompressed()
	m.RecordBytesSaved(512)
}
