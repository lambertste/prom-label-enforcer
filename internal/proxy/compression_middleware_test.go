package proxy

import (
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func largeBody(n int) string {
	return strings.Repeat("a", n)
}

func TestDefaultCompressionConfig(t *testing.T) {
	cfg := DefaultCompressionConfig()
	if cfg.Level != gzip.DefaultCompression {
		t.Errorf("expected level %d, got %d", gzip.DefaultCompression, cfg.Level)
	}
	if cfg.MinSize != 1024 {
		t.Errorf("expected MinSize 1024, got %d", cfg.MinSize)
	}
}

func TestCompressionMiddleware_CompressesLargeResponse(t *testing.T) {
	body := largeBody(2048)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(body))
	})
	mw := NewCompressionMiddleware(DefaultCompressionConfig(), handler)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	rec := httptest.NewRecorder()
	mw.ServeHTTP(rec, req)

	if rec.Header().Get("Content-Encoding") != "gzip" {
		t.Fatal("expected Content-Encoding: gzip")
	}
	gr, err := gzip.NewReader(rec.Body)
	if err != nil {
		t.Fatalf("failed to create gzip reader: %v", err)
	}
	decompressed, _ := io.ReadAll(gr)
	if string(decompressed) != body {
		t.Error("decompressed body does not match original")
	}
}

func TestCompressionMiddleware_SkipsSmallResponse(t *testing.T) {
	body := "small"
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(body))
	})
	mw := NewCompressionMiddleware(DefaultCompressionConfig(), handler)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	rec := httptest.NewRecorder()
	mw.ServeHTTP(rec, req)

	if rec.Header().Get("Content-Encoding") == "gzip" {
		t.Error("expected no gzip encoding for small response")
	}
	if rec.Body.String() != body {
		t.Errorf("expected body %q, got %q", body, rec.Body.String())
	}
}

func TestCompressionMiddleware_SkipsWithoutAcceptEncoding(t *testing.T) {
	body := largeBody(2048)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(body))
	})
	mw := NewCompressionMiddleware(DefaultCompressionConfig(), handler)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	mw.ServeHTTP(rec, req)

	if rec.Header().Get("Content-Encoding") == "gzip" {
		t.Error("expected no gzip encoding when Accept-Encoding not set")
	}
}

func TestCompressionMiddleware_ZeroValuesUseDefaults(t *testing.T) {
	cfg := CompressionConfig{} // zero value
	body := largeBody(2048)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(body))
	})
	mw := NewCompressionMiddleware(cfg, handler)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	rec := httptest.NewRecorder()
	mw.ServeHTTP(rec, req)

	if rec.Header().Get("Content-Encoding") != "gzip" {
		t.Error("expected gzip encoding with zero-value config")
	}
}
