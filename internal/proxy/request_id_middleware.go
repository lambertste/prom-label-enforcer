package proxy

import (
	"net/http"

	"github.com/google/uuid"
)

const RequestIDHeader = "X-Request-ID"

// DefaultRequestIDConfig returns a default RequestIDConfig.
type RequestIDConfig struct {
	// Header is the HTTP header used to read/write the request ID.
	Header string
	// ForceNew always generates a new ID even if one is present.
	ForceNew bool
}

// DefaultRequestIDConfig returns sensible defaults.
func DefaultRequestIDConfig() RequestIDConfig {
	return RequestIDConfig{
		Header:   RequestIDHeader,
		ForceNew: false,
	}
}

type requestIDMiddleware struct {
	cfg     RequestIDConfig
	metrics *RequestIDMetrics
}

// NewRequestIDMiddleware creates an HTTP middleware that ensures every request
// carries a unique request ID, propagating it to the response.
func NewRequestIDMiddleware(cfg RequestIDConfig, m *RequestIDMetrics) func(http.Handler) http.Handler {
	if cfg.Header == "" {
		cfg.Header = RequestIDHeader
	}
	mw := &requestIDMiddleware{cfg: cfg, metrics: m}
	return mw.handler
}

func (m *requestIDMiddleware) handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get(m.cfg.Header)
		propagated := id != "" && !m.cfg.ForceNew
		if !propagated {
			id = uuid.NewString()
			m.metrics.RecordGenerated()
		} else {
			m.metrics.RecordPropagated()
		}
		r = r.WithContext(withRequestID(r.Context(), id))
		w.Header().Set(m.cfg.Header, id)
		next.ServeHTTP(w, r)
	})
}
