package proxy

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/prom-label-enforcer/internal/enforcer"
)

func newTestEnforcer(t *testing.T) *enforcer.Enforcer {
	t.Helper()
	cfg := &enforcer.Config{
		RequiredLabels: []enforcer.LabelRule{
			{Name: "env", AllowedValues: []string{"prod", "staging"}},
		},
	}
	e, err := enforcer.New(cfg)
	if err != nil {
		t.Fatalf("failed to create enforcer: %v", err)
	}
	return e
}

func TestHandler_Compliant(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer upstream.Close()

	e := newTestEnforcer(t)
	h, err := NewHandler(upstream.URL, e)
	if err != nil {
		t.Fatalf("NewHandler: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/receive", strings.NewReader("body"))
	req.Header.Set("X-Prom-Label-env", "prod")
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec.Code)
	}
}

func TestHandler_MissingLabel(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer upstream.Close()

	e := newTestEnforcer(t)
	h, err := NewHandler(upstream.URL, e)
	if err != nil {
		t.Fatalf("NewHandler: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/receive", strings.NewReader("body"))
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rec.Code)
	}
}

func TestHandler_DisallowedValue(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer upstream.Close()

	e := newTestEnforcer(t)
	h, err := NewHandler(upstream.URL, e)
	if err != nil {
		t.Fatalf("NewHandler: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/receive", strings.NewReader("body"))
	req.Header.Set("X-Prom-Label-env", "dev")
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rec.Code)
	}
}

func TestExtractLabels(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("X-Prom-Label-env", "prod")
	req.Header.Set("X-Prom-Label-region", "us-east-1")
	req.Header.Set("Content-Type", "application/json")

	labels := extractLabels(req)

	if labels["env"] != "prod" {
		t.Errorf("expected env=prod, got %q", labels["env"])
	}
	if labels["region"] != "us-east-1" {
		t.Errorf("expected region=us-east-1, got %q", labels["region"])
	}
	if _, ok := labels["Content-Type"]; ok {
		t.Error("non-label header should not be extracted")
	}
}
