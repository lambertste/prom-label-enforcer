package proxy

import (
	"bytes"
	"io"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// DefaultMirrorConfig returns a MirrorConfig with sensible defaults.
func DefaultMirrorConfig() MirrorConfig {
	return MirrorConfig{
		TargetURL:     "",
		TimeoutSeconds: 2,
		SampleRate:    1.0,
	}
}

// MirrorConfig configures the request mirroring middleware.
type MirrorConfig struct {
	// TargetURL is the URL to mirror requests to (required).
	TargetURL string
	// TimeoutSeconds is the per-mirror-request timeout.
	TimeoutSeconds int
	// SampleRate is the fraction of requests to mirror (0.0–1.0).
	SampleRate float64
	// Registerer is the Prometheus registerer; defaults to the global registry.
	Registerer prometheus.Registerer
}

// NewMirrorMiddleware mirrors a copy of each incoming request to a secondary
// target without blocking the primary response path.
func NewMirrorMiddleware(cfg MirrorConfig, next http.Handler) (http.Handler, error) {
	if cfg.TargetURL == "" {
		// No-op when no target is configured.
		return next, nil
	}
	if cfg.TimeoutSeconds <= 0 {
		cfg.TimeoutSeconds = DefaultMirrorConfig().TimeoutSeconds
	}
	if cfg.SampleRate <= 0 || cfg.SampleRate > 1.0 {
		cfg.SampleRate = 1.0
	}

	m := NewMirrorMetrics(cfg.Registerer)
	client := &http.Client{Timeout: time.Duration(cfg.TimeoutSeconds) * time.Second}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Always serve the primary request first.
		var bodyBytes []byte
		if r.Body != nil {
			var err error
			bodyBytes, err = io.ReadAll(r.Body)
			if err == nil {
				r.Body = io.NopCloser(bytes.NewReader(bodyBytes))
			}
		}

		next.ServeHTTP(w, r)

		// Mirror asynchronously.
		go func() {
			mirrored, err := http.NewRequest(r.Method, cfg.TargetURL+r.RequestURI, bytes.NewReader(bodyBytes))
			if err != nil {
				m.RecordError()
				return
			}
			for k, vv := range r.Header {
				for _, v := range vv {
					mirrored.Header.Add(k, v)
				}
			}
			resp, err := client.Do(mirrored)
			if err != nil {
				m.RecordError()
				return
			}
			_ = resp.Body.Close()
			m.RecordSent()
		}()
	}), nil
}
