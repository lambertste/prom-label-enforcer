package proxy

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// ServerConfig holds configuration for the HTTP proxy server.
type ServerConfig struct {
	ListenAddr      string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	RateLimit       float64 // requests per second per client (0 = disabled)
	RateLimitBurst  float64
}

// DefaultServerConfig returns a ServerConfig with sensible defaults.
func DefaultServerConfig() ServerConfig {
	return ServerConfig{
		ListenAddr:     ":9090",
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		IdleTimeout:    30 * time.Second,
		RateLimit:      100,
		RateLimitBurst: 20,
	}
}

// NewServer constructs an *http.Server wired with the enforcer handler,
// optional rate limiting, and a health endpoint.
func NewServer(cfg ServerConfig, handler http.Handler) *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", healthz)

	var root http.Handler = handler
	if cfg.RateLimit > 0 {
		rl := NewRateLimiter(cfg.RateLimit, cfg.RateLimitBurst)
		root = rl.Middleware(handler)
	}
	mux.Handle("/receive", root)

	return &http.Server{
		Addr:         cfg.ListenAddr,
		Handler:      mux,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	}
}

func healthz(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "ok")
}

// Shutdown gracefully stops the server within the provided context deadline.
func Shutdown(ctx context.Context, srv *http.Server) error {
	return srv.Shutdown(ctx)
}
