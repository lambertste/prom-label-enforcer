package proxy

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
)

func newIPFilterMetrics(t *testing.T) *IPFilterMetrics {
	t.Helper()
	return NewIPFilterMetrics(prometheus.NewRegistry())
}

func TestDefaultIPFilterConfig(t *testing.T) {
	cfg := DefaultIPFilterConfig()
	if len(cfg.AllowedCIDRs) != 0 || len(cfg.DeniedCIDRs) != 0 {
		t.Fatal("expected empty CIDR lists")
	}
	if cfg.TrustProxy {
		t.Fatal("expected TrustProxy false")
	}
}

func TestIPFilterMiddleware_AllowsAllByDefault(t *testing.T) {
	m := newIPFilterMetrics(t)
	mw, err := NewIPFilterMiddleware(DefaultIPFilterConfig(), m)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "192.168.1.5:1234"
	mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})).ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestIPFilterMiddleware_DeniedCIDRBlocked(t *testing.T) {
	m := newIPFilterMetrics(t)
	cfg := DefaultIPFilterConfig()
	cfg.DeniedCIDRs = []string{"10.0.0.0/8"}
	mw, err := NewIPFilterMiddleware(cfg, m)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "10.1.2.3:9999"
	mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})).ServeHTTP(rr, req)
	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rr.Code)
	}
}

func TestIPFilterMiddleware_AllowedCIDRPermits(t *testing.T) {
	m := newIPFilterMetrics(t)
	cfg := DefaultIPFilterConfig()
	cfg.AllowedCIDRs = []string{"172.16.0.0/12"}
	mw, err := NewIPFilterMiddleware(cfg, m)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "172.16.5.10:1234"
	mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})).ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestIPFilterMiddleware_AllowedCIDRBlocksOther(t *testing.T) {
	m := newIPFilterMetrics(t)
	cfg := DefaultIPFilterConfig()
	cfg.AllowedCIDRs = []string{"172.16.0.0/12"}
	mw, err := NewIPFilterMiddleware(cfg, m)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "8.8.8.8:53"
	mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})).ServeHTTP(rr, req)
	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rr.Code)
	}
}

func TestIPFilterMiddleware_InvalidCIDRReturnsError(t *testing.T) {
	m := newIPFilterMetrics(t)
	cfg := DefaultIPFilterConfig()
	cfg.AllowedCIDRs = []string{"not-a-cidr"}
	_, err := NewIPFilterMiddleware(cfg, m)
	if err == nil {
		t.Fatal("expected error for invalid CIDR")
	}
}

func TestIPFilterMiddleware_TrustProxyHeader(t *testing.T) {
	m := newIPFilterMetrics(t)
	cfg := DefaultIPFilterConfig()
	cfg.DeniedCIDRs = []string{"203.0.113.0/24"}
	cfg.TrustProxy = true
	mw, err := NewIPFilterMiddleware(cfg, m)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "127.0.0.1:1234"
	req.Header.Set("X-Forwarded-For", "203.0.113.5, 127.0.0.1")
	mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})).ServeHTTP(rr, req)
	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for denied forwarded IP, got %d", rr.Code)
	}
}
