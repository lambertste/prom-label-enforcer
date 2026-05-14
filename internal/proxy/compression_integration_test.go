package proxy

import (
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
)

// instrumentedCompressionHandler wraps the compression middleware and records
// metrics based on whether the response was compressed.
func instrumentedCompressionHandler(cfg CompressionConfig, metrics *CompressionMetrics, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		recorder := httptest.NewRecorder()
		NewCompressionMiddleware(cfg, next).ServeHTTP(recorder, r)

		if recorder.Header().Get("Content-Encoding") == "gzip" {
			metrics.RecordCompressed()
			metrics.RecordBytesSaved(float64(recorder.Body.Len()))
		} else {
			metrics.RecordUncompressed()
		}
		for k, vs := range recorder.Header() {
			for _, v := range vs {
				w.Header().Add(k, v)
			}
		}
		w.WriteHeader(recorder.Code)
		w.Write(recorder.Body.Bytes())
	})
}

func TestInstrumentedCompression_LargeBodyRecordsCompressed(t *testing.T) {
	reg := prometheus.NewRegistry()
	metrics := NewCompressionMetrics(reg)
	body := strings.Repeat("x", 4096)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(body))
	})
	h := instrumentedCompressionHandler(DefaultCompressionConfig(), metrics, handler)

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Header().Get("Content-Encoding") != "gzip" {
		t.Fatal("expected gzip Content-Encoding")
	}
	gr, err := gzip.NewReader(rec.Body)
	if err != nil {
		t.Fatalf("gzip reader error: %v", err)
	}
	decompressed, _ := io.ReadAll(gr)
	if string(decompressed) != body {
		t.Error("decompressed body mismatch")
	}

	mfs, _ := reg.Gather()
	for _, mf := range mfs {
		if mf.GetName() == "proxy_compression_compressed_total" {
			if v := mf.GetMetric()[0].GetCounter().GetValue(); v != 1 {
				t.Errorf("expected compressed_total=1, got %v", v)
			}
			return
		}
	}
	t.Error("compressed_total metric not found")
}

func TestInstrumentedCompression_SmallBodyRecordsUncompressed(t *testing.T) {
	reg := prometheus.NewRegistry()
	metrics := NewCompressionMetrics(reg)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("tiny"))
	})
	h := instrumentedCompressionHandler(DefaultCompressionConfig(), metrics, handler)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	mfs, _ := reg.Gather()
	for _, mf := range mfs {
		if mf.GetName() == "proxy_compression_uncompressed_total" {
			if v := mf.GetMetric()[0].GetCounter().GetValue(); v != 1 {
				t.Errorf("expected uncompressed_total=1, got %v", v)
			}
			return
		}
	}
	t.Error("uncompressed_total metric not found")
}
