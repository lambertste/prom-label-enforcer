package proxy

import (
	"net"
	"net/http"
	"strings"
)

// IPFilterConfig holds configuration for IP allowlist/denylist filtering.
type IPFilterConfig struct {
	// AllowedCIDRs is a list of CIDR blocks that are permitted. If non-empty,
	// only requests from matching IPs are allowed.
	AllowedCIDRs []string
	// DeniedCIDRs is a list of CIDR blocks that are always rejected.
	DeniedCIDRs []string
	// TrustProxy indicates whether to read the real IP from X-Forwarded-For.
	TrustProxy bool
}

// DefaultIPFilterConfig returns an IPFilterConfig with no restrictions.
func DefaultIPFilterConfig() IPFilterConfig {
	return IPFilterConfig{
		AllowedCIDRs: []string{},
		DeniedCIDRs:  []string{},
		TrustProxy:   false,
	}
}

type ipFilterMiddleware struct {
	cfg     IPFilterConfig
	allowed []*net.IPNet
	denied  []*net.IPNet
	metrics *IPFilterMetrics
}

// NewIPFilterMiddleware constructs an IP filtering middleware.
func NewIPFilterMiddleware(cfg IPFilterConfig, m *IPFilterMetrics) (func(http.Handler) http.Handler, error) {
	if cfg.AllowedCIDRs == nil {
		cfg.AllowedCIDRs = []string{}
	}
	if cfg.DeniedCIDRs == nil {
		cfg.DeniedCIDRs = []string{}
	}

	mw := &ipFilterMiddleware{cfg: cfg, metrics: m}

	for _, cidr := range cfg.AllowedCIDRs {
		_, network, err := net.ParseCIDR(cidr)
		if err != nil {
			return nil, err
		}
		mw.allowed = append(mw.allowed, network)
	}
	for _, cidr := range cfg.DeniedCIDRs {
		_, network, err := net.ParseCIDR(cidr)
		if err != nil {
			return nil, err
		}
		mw.denied = append(mw.denied, network)
	}

	return mw.handler, nil
}

func (m *ipFilterMiddleware) handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := realIP(r, m.cfg.TrustProxy)
		parsed := net.ParseIP(ip)

		for _, network := range m.denied {
			if parsed != nil && network.Contains(parsed) {
				m.metrics.RecordDenied()
				http.Error(w, "forbidden", http.StatusForbidden)
				return
			}
		}

		if len(m.allowed) > 0 {
			for _, network := range m.allowed {
				if parsed != nil && network.Contains(parsed) {
					m.metrics.RecordAllowed()
					next.ServeHTTP(w, r)
					return
				}
			}
			m.metrics.RecordDenied()
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}

		m.metrics.RecordAllowed()
		next.ServeHTTP(w, r)
	})
}

func realIP(r *http.Request, trustProxy bool) string {
	if trustProxy {
		if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
			parts := strings.SplitN(xff, ",", 2)
			return strings.TrimSpace(parts[0])
		}
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
