package proxy

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/prom-label-enforcer/internal/enforcer"
)

func okHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
}

func rejectHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	})
}

func TestNewAuditMiddleware_NilLogger(t *testing.T) {
	h := NewAuditMiddleware(okHandler(), nil, "rs")
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/receive", nil)
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec.Code)
	}
}

func TestAuditMiddleware_RecordsAllowed(t *testing.T) {
	var buf bytes.Buffer
	al := enforcer.NewAuditLogger(&buf)

	h := NewAuditMiddleware(okHandler(), al, "test-ruleset")
	req := httptest.NewRequest(http.MethodPost, "/receive?env=prod", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec.Code)
	}

	line := strings.TrimSpace(buf.String())
	if line == "" {
		t.Fatal("expected audit log output")
	}
	var event map[string]interface{}
	if err := json.Unmarshal([]byte(line), &event); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if event["allowed"] != true {
		t.Errorf("expected allowed=true in audit log")
	}
	if event["ruleset_id"] != "test-ruleset" {
		t.Errorf("unexpected ruleset_id: %v", event["ruleset_id"])
	}
}

func TestAuditMiddleware_RecordsRejected(t *testing.T) {
	var buf bytes.Buffer
	al := enforcer.NewAuditLogger(&buf)

	h := NewAuditMiddleware(rejectHandler(), al, "rs-reject")
	req := httptest.NewRequest(http.MethodPost, "/receive", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}

	var event map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(buf.String())), &event); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if event["allowed"] != false {
		t.Errorf("expected allowed=false in audit log")
	}
}
