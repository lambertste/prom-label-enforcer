package proxy

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

func newInstrumentedAuth(tokens ...string) (func(http.Handler) http.Handler, *AuthMetrics) {
	cfg := DefaultAuthConfig()
	for _, t := range tokens {
		cfg.Tokens[t] = struct{}{}
	}
	reg := prometheus.NewRegistry()
	metrics := NewAuthMetrics(reg)
	mw := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := extractBearerToken(r)
			if _, ok := cfg.Tokens[token]; !ok {
				metrics.RecordDenied("invalid_token")
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			metrics.RecordAllowed()
			next.ServeHTTP(w, r)
		})
	}
	return mw, metrics
}

func TestInstrumentedAuth_AllowsAndRecords(t *testing.T) {
	mw, m := newInstrumentedAuth("tok1")
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	req.Header.Set("Authorization", "Bearer tok1")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if v := testutil.ToFloat64(m.allowed.WithLabelValues()); v != 1 {
		t.Fatalf("expected 1 allowed, got %v", v)
	}
}

func TestInstrumentedAuth_DeniesAndRecords(t *testing.T) {
	mw, m := newInstrumentedAuth("tok1")
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	req.Header.Set("Authorization", "Bearer bad")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
	if v := testutil.ToFloat64(m.denied.WithLabelValues("invalid_token")); v != 1 {
		t.Fatalf("expected 1 denied, got %v", v)
	}
}
