package proxy

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
)

func newResponseHeaderRegistry() *prometheus.Registry {
	return prometheus.NewRegistry()
}

func baseResponseHeaderHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Server", "upstream-server")
		w.WriteHeader(http.StatusOK)
	})
}

func TestDefaultResponseHeaderConfig(t *testing.T) {
	cfg := DefaultResponseHeaderConfig()
	if len(cfg.StaticHeaders) == 0 {
		t.Fatal("expected default static headers to be non-empty")
	}
	if len(cfg.StripHeaders) == 0 {
		t.Fatal("expected default strip headers to be non-empty")
	}
	if _, ok := cfg.StaticHeaders["X-Content-Type-Options"]; !ok {
		t.Error("expected X-Content-Type-Options in default static headers")
	}
}

func TestNewResponseHeaderMetrics_NotNil(t *testing.T) {
	reg := newResponseHeaderRegistry()
	m := NewResponseHeaderMetrics(reg)
	if m == nil {
		t.Fatal("expected non-nil metrics")
	}
	if m.InjectedTotal == nil || m.StrippedTotal == nil {
		t.Error("expected all metric fields to be initialised")
	}
}

func TestNewResponseHeaderMetrics_NilRegisterer(t *testing.T) {
	m := NewResponseHeaderMetrics(nil)
	if m == nil {
		t.Fatal("expected non-nil metrics when registerer is nil")
	}
}

func TestResponseHeaderMetrics_RecordInjected(t *testing.T) {
	reg := newResponseHeaderRegistry()
	m := NewResponseHeaderMetrics(reg)
	m.RecordInjected()
	m.RecordInjected()
	// no panic == pass
}

func TestResponseHeaderMetrics_RecordStripped(t *testing.T) {
	reg := newResponseHeaderRegistry()
	m := NewResponseHeaderMetrics(reg)
	m.RecordStripped()
	// no panic == pass
}

func TestResponseHeaderMetrics_NilSafe(t *testing.T) {
	var m *ResponseHeaderMetrics
	m.RecordInjected()
	m.RecordStripped()
	// no panic == pass
}

func TestResponseHeaderMiddleware_InjectsHeaders(t *testing.T) {
	cfg := ResponseHeaderConfig{
		StaticHeaders: map[string]string{"X-Custom": "value"},
		StripHeaders:  []string{},
		Registerer:    newResponseHeaderRegistry(),
	}
	h := NewResponseHeaderMiddleware(cfg, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if got := rec.Header().Get("X-Custom"); got != "value" {
		t.Errorf("expected X-Custom=value, got %q", got)
	}
}

func TestResponseHeaderMiddleware_StripsUpstreamHeaders(t *testing.T) {
	cfg := ResponseHeaderConfig{
		StaticHeaders: map[string]string{},
		StripHeaders:  []string{"Server"},
		Registerer:    newResponseHeaderRegistry(),
	}
	h := NewResponseHeaderMiddleware(cfg, baseResponseHeaderHandler())

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if got := rec.Header().Get("Server"); got != "" {
		t.Errorf("expected Server header to be stripped, got %q", got)
	}
}

func TestResponseHeaderMiddleware_NilStaticHeaders(t *testing.T) {
	cfg := ResponseHeaderConfig{
		StaticHeaders: nil,
		StripHeaders:  nil,
		Registerer:    newResponseHeaderRegistry(),
	}
	h := NewResponseHeaderMiddleware(cfg, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec.Code)
	}
}
