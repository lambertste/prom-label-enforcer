package proxy

import (
	"encoding/json"
	"net/http"
	"sync/atomic"
	"time"
)

// HealthCheckConfig holds configuration for the health check middleware.
type HealthCheckConfig struct {
	// Path is the endpoint path for health checks (default: /healthz).
	Path string
	// ReadinessPath is the endpoint for readiness probes (default: /readyz).
	ReadinessPath string
	// Version is an optional build/version string included in responses.
	Version string
}

// DefaultHealthCheckConfig returns a HealthCheckConfig with sensible defaults.
func DefaultHealthCheckConfig() HealthCheckConfig {
	return HealthCheckConfig{
		Path:          "/healthz",
		ReadinessPath: "/readyz",
		Version:       "unknown",
	}
}

type healthCheckMiddleware struct {
	cfg   HealthCheckConfig
	ready atomic.Bool
	start time.Time
}

type healthResponse struct {
	Status  string `json:"status"`
	Version string `json:"version,omitempty"`
	Uptime  string `json:"uptime,omitempty"`
}

// NewHealthCheckMiddleware returns an http.Handler that intercepts liveness
// and readiness probe paths, delegating all other requests to next.
func NewHealthCheckMiddleware(cfg HealthCheckConfig, next http.Handler) http.Handler {
	if cfg.Path == "" {
		cfg.Path = DefaultHealthCheckConfig().Path
	}
	if cfg.ReadinessPath == "" {
		cfg.ReadinessPath = DefaultHealthCheckConfig().ReadinessPath
	}
	h := &healthCheckMiddleware{cfg: cfg, start: time.Now()}
	h.ready.Store(true)
	return h.handler(next)
}

func (h *healthCheckMiddleware) handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case h.cfg.Path:
			h.writeLiveness(w)
		case h.cfg.ReadinessPath:
			h.writeReadiness(w)
		default:
			next.ServeHTTP(w, r)
		}
	})
}

func (h *healthCheckMiddleware) writeLiveness(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(healthResponse{
		Status:  "ok",
		Version: h.cfg.Version,
		Uptime:  time.Since(h.start).Round(time.Second).String(),
	})
}

func (h *healthCheckMiddleware) writeReadiness(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	if h.ready.Load() {
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(healthResponse{Status: "ready"})
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
		_ = json.NewEncoder(w).Encode(healthResponse{Status: "not ready"})
	}
}
