package proxy

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
	"sync"
)

// CompressionConfig holds configuration for the compression middleware.
type CompressionConfig struct {
	// Level is the gzip compression level (1-9). Defaults to gzip.DefaultCompression.
	Level int
	// MinSize is the minimum response size in bytes to compress. Defaults to 1024.
	MinSize int
}

// DefaultCompressionConfig returns a CompressionConfig with sensible defaults.
func DefaultCompressionConfig() CompressionConfig {
	return CompressionConfig{
		Level:   gzip.DefaultCompression,
		MinSize: 1024,
	}
}

var gzipPool = sync.Pool{
	New: func() any {
		w, _ := gzip.NewWriterLevel(io.Discard, gzip.DefaultCompression)
		return w
	},
}

type gzipResponseWriter struct {
	http.ResponseWriter
	gz     *gzip.Writer
	buf    []byte
	minSize int
	wroteHeader bool
	status      int
}

func (g *gzipResponseWriter) WriteHeader(code int) {
	g.status = code
}

func (g *gzipResponseWriter) Write(b []byte) (int, error) {
	g.buf = append(g.buf, b...)
	return len(b), nil
}

func (g *gzipResponseWriter) flush() {
	if len(g.buf) >= g.minSize {
		g.ResponseWriter.Header().Set("Content-Encoding", "gzip")
		g.ResponseWriter.Header().Del("Content-Length")
		if g.status != 0 {
			g.ResponseWriter.WriteHeader(g.status)
		}
		g.gz.Write(g.buf)
		g.gz.Close()
	} else {
		if g.status != 0 {
			g.ResponseWriter.WriteHeader(g.status)
		}
		g.ResponseWriter.Write(g.buf)
	}
}

// NewCompressionMiddleware returns an http.Handler that gzip-compresses responses
// when the client accepts gzip encoding and the response meets the minimum size.
func NewCompressionMiddleware(cfg CompressionConfig, next http.Handler) http.Handler {
	if cfg.Level == 0 {
		cfg.Level = gzip.DefaultCompression
	}
	if cfg.MinSize <= 0 {
		cfg.MinSize = DefaultCompressionConfig().MinSize
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}
		gz := gzipPool.Get().(*gzip.Writer)
		gz.Reset(w)
		defer gzipPool.Put(gz)

		grw := &gzipResponseWriter{
			ResponseWriter: w,
			gz:             gz,
			minSize:        cfg.MinSize,
		}
		next.ServeHTTP(grw, r)
		grw.flush()
	})
}
