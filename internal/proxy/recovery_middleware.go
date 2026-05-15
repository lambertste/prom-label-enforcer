package proxy

import (
	"log"
	"net/http"
	"runtime/debug"

	"github.com/prometheus/client_golang/prometheus"
)

// RecoveryConfig holds configuration for the panic recovery middleware.
type RecoveryConfig struct {
	Logger     *log.Logger
	Metrics    *RecoveryMetrics
	PrintStack bool
}

// DefaultRecoveryConfig returns a RecoveryConfig with sensible defaults.
func DefaultRecoveryConfig(reg prometheus.Registerer) RecoveryConfig {
	return RecoveryConfig{
		Logger:     log.Default(),
		Metrics:    NewRecoveryMetrics(reg),
		PrintStack: true,
	}
}

type recoveryMiddleware struct {
	cfg     RecoveryConfig
	handler http.Handler
}

// NewRecoveryMiddleware wraps h with a middleware that recovers from panics,
// logs the error and stack trace, records a metric, and returns HTTP 500.
func NewRecoveryMiddleware(cfg RecoveryConfig, h http.Handler) http.Handler {
	if cfg.Logger == nil {
		cfg.Logger = log.Default()
	}
	return &recoveryMiddleware{cfg: cfg, handler: h}
}

func (m *recoveryMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if rec := recover(); rec != nil {
			if m.cfg.PrintStack {
				m.cfg.Logger.Printf("[recovery] panic: %v\n%s", rec, debug.Stack())
			} else {
				m.cfg.Logger.Printf("[recovery] panic: %v", rec)
			}
			if m.cfg.Metrics != nil {
				m.cfg.Metrics.RecordPanic()
			}
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
	}()
	m.handler.ServeHTTP(w, r)
}
