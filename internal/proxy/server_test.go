package proxy

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/prom-label-enforcer/internal/enforcer"
)

func TestDefaultServerConfig(t *testing.T) {
	cfg := DefaultServerConfig()
	if cfg.ListenAddr != ":9090" {
		t.Errorf("expected :9090, got %s", cfg.ListenAddr)
	}
	if cfg.ReadTimeout != 10*time.Second {
		t.Errorf("unexpected ReadTimeout: %v", cfg.ReadTimeout)
	}
}

func TestServer_Healthz(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer upstream.Close()

	cfg := &enforcer.Config{
		RequiredLabels: []enforcer.LabelRule{},
	}
	e, err := enforcer.New(cfg)
	if err != nil {
		t.Fatalf("enforcer.New: %v", err)
	}

	h, err := NewHandler(upstream.URL, e)
	if err != nil {
		t.Fatalf("NewHandler: %v", err)
	}

	svr := NewServer(h, DefaultServerConfig())

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	svr.httpServer.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	if rec.Body.String() != "ok" {
		t.Errorf("expected body 'ok', got %q", rec.Body.String())
	}
}

func TestServer_ReceiveRoute(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer upstream.Close()

	cfg := &enforcer.Config{
		RequiredLabels: []enforcer.LabelRule{},
	}
	e, err := enforcer.New(cfg)
	if err != nil {
		t.Fatalf("enforcer.New: %v", err)
	}

	h, err := NewHandler(upstream.URL, e)
	if err != nil {
		t.Fatalf("NewHandler: %v", err)
	}

	svr := NewServer(h, DefaultServerConfig())

	req := httptest.NewRequest(http.MethodPost, "/receive", nil)
	rec := httptest.NewRecorder()
	svr.httpServer.Handler.ServeHTTP(rec, req)

	// No required labels configured, so request should be forwarded.
	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec.Code)
	}
}
