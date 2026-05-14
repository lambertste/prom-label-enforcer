package proxy

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

func newInstrumentedIPFilter(t *testing.T, cfg IPFilterConfig) (http.Handler, *IPFilterMetrics) {
	t.Helper()
	reg := prometheus.NewRegistry()
	m := NewIPFilterMetrics(reg)
	mw, err := NewIPFilterMiddleware(cfg, m)
	if err != nil {
		t.Fatalf("NewIPFilterMiddleware: %v", err)
	}
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	return mw(inner), m
}

func TestInstrumentedIPFilter_AllowedRequestRecorded(t *testing.T) {
	cfg := DefaultIPFilterConfig()
	h, m := newInstrumentedIPFilter(t, cfg)

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	req.RemoteAddr = "192.0.2.1:4321"
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if got := testutil.ToFloat64(m.allowed); got != 1 {
		t.Fatalf("expected 1 allowed metric, got %v", got)
	}
	if got := testutil.ToFloat64(m.denied); got != 0 {
		t.Fatalf("expected 0 denied metric, got %v", got)
	}
}

func TestInstrumentedIPFilter_DeniedRequestRecorded(t *testing.T) {
	cfg := DefaultIPFilterConfig()
	cfg.DeniedCIDRs = []string{"192.0.2.0/24"}
	h, m := newInstrumentedIPFilter(t, cfg)

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	req.RemoteAddr = "192.0.2.99:1111"
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rr.Code)
	}
	if got := testutil.ToFloat64(m.denied); got != 1 {
		t.Fatalf("expected 1 denied metric, got %v", got)
	}
	if got := testutil.ToFloat64(m.allowed); got != 0 {
		t.Fatalf("expected 0 allowed metric, got %v", got)
	}
}
