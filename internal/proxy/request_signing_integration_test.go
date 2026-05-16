package proxy

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
)

func newInstrumentedSigning(secret string, reg prometheus.Registerer) http.Handler {
	cfg := DefaultRequestSigningConfig()
	cfg.SecretKey = secret
	cfg.Registerer = reg

	inner := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	return NewRequestSigningMiddleware(cfg, inner)
}

func TestInstrumentedRequestSigning_AllowsAndRecords(t *testing.T) {
	reg := prometheus.NewRegistry()
	h := newInstrumentedSigning("secret", reg)

	req := makeSignedRequest("secret", "/push")
	rw := httptest.NewRecorder()
	h.ServeHTTP(rw, req)

	if rw.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rw.Code)
	}

	mfs, err := reg.Gather()
	if err != nil {
		t.Fatalf("gather error: %v", err)
	}
	found := false
	for _, mf := range mfs {
		if mf.GetName() == "prom_label_enforcer_request_signing_allowed_total" {
			found = true
		}
	}
	if !found {
		t.Error("expected allowed_total metric to be registered")
	}
}

func TestInstrumentedRequestSigning_DeniesAndRecords(t *testing.T) {
	reg := prometheus.NewRegistry()
	h := newInstrumentedSigning("secret", reg)

	req := httptest.NewRequest(http.MethodGet, "/push", nil)
	rw := httptest.NewRecorder()
	h.ServeHTTP(rw, req)

	if rw.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rw.Code)
	}

	mfs, err := reg.Gather()
	if err != nil {
		t.Fatalf("gather error: %v", err)
	}
	found := false
	for _, mf := range mfs {
		if mf.GetName() == "prom_label_enforcer_request_signing_denied_total" {
			found = true
		}
	}
	if !found {
		t.Error("expected denied_total metric to be registered")
	}
}
