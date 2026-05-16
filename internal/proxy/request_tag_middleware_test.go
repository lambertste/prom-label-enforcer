package proxy

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func tagEchoHandler(t *testing.T, wantTags []string) http.Handler {
	t.Helper()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got := RequestTagsFromContext(r.Context())
		if len(got) != len(wantTags) {
			t.Errorf("tag count: got %d, want %d", len(got), len(wantTags))
			return
		}
		for i, tag := range wantTags {
			if got[i] != tag {
				t.Errorf("tag[%d]: got %q, want %q", i, got[i], tag)
			}
		}
		w.WriteHeader(http.StatusOK)
	})
}

func TestDefaultRequestTagConfig(t *testing.T) {
	cfg := DefaultRequestTagConfig()
	if cfg.HeaderName == "" {
		t.Error("expected non-empty HeaderName")
	}
	if cfg.MaxTags <= 0 {
		t.Error("expected positive MaxTags")
	}
}

func TestRequestTagMiddleware_NoHeader(t *testing.T) {
	cfg := DefaultRequestTagConfig()
	cfg.Registerer = newTagTestRegistry()
	h := NewRequestTagMiddleware(cfg, tagEchoHandler(t, nil))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("status: got %d, want 200", rec.Code)
	}
}

func TestRequestTagMiddleware_ParsesTags(t *testing.T) {
	cfg := DefaultRequestTagConfig()
	cfg.Registerer = newTagTestRegistry()
	h := NewRequestTagMiddleware(cfg, tagEchoHandler(t, []string{"env:prod", "team:platform"}))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Request-Tags", "env:prod, team:platform")
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("status: got %d, want 200", rec.Code)
	}
}

func TestRequestTagMiddleware_TruncatesExcessTags(t *testing.T) {
	cfg := DefaultRequestTagConfig()
	cfg.MaxTags = 2
	cfg.Registerer = newTagTestRegistry()
	h := NewRequestTagMiddleware(cfg, tagEchoHandler(t, []string{"a", "b"}))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Request-Tags", strings.Join([]string{"a", "b", "c", "d"}, ","))
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("status: got %d, want 200", rec.Code)
	}
}

func TestRequestTagsFromContext_Empty(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	tags := RequestTagsFromContext(req.Context())
	if tags != nil {
		t.Errorf("expected nil tags, got %v", tags)
	}
}
