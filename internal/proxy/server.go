package proxy

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"
)

// ServerConfig holds HTTP server configuration.
type ServerConfig struct {
	ListenAddr      string
	UpstreamURL     string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	ShutdownTimeout time.Duration
}

// DefaultServerConfig returns a ServerConfig with sensible defaults.
func DefaultServerConfig() ServerConfig {
	return ServerConfig{
		ListenAddr:      ":9090",
		ReadTimeout:     10 * time.Second,
		WriteTimeout:    10 * time.Second,
		ShutdownTimeout: 5 * time.Second,
	}
}

// Server wraps the HTTP server and proxy handler.
type Server struct {
	httpServer *http.Server
	cfg        ServerConfig
}

// NewServer constructs a Server with the given handler and config.
func NewServer(h http.Handler, cfg ServerConfig) *Server {
	mux := http.NewServeMux()
	mux.Handle("/receive", h)
	mux.HandleFunc("/healthz", healthz)

	return &Server{
		cfg: cfg,
		httpServer: &http.Server{
			Addr:         cfg.ListenAddr,
			Handler:      mux,
			ReadTimeout:  cfg.ReadTimeout,
			WriteTimeout: cfg.WriteTimeout,
		},
	}
}

// Start begins listening and blocks until the server stops.
func (s *Server) Start() error {
	log.Printf("proxy server listening on %s", s.cfg.ListenAddr)
	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("server error: %w", err)
	}
	return nil
}

// Shutdown gracefully stops the server.
func (s *Server) Shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), s.cfg.ShutdownTimeout)
	defer cancel()
	return s.httpServer.Shutdown(ctx)
}

func healthz(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}
