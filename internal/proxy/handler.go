package proxy

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/prom-label-enforcer/internal/enforcer"
)

// Handler wraps a reverse proxy with label enforcement.
type Handler struct {
	enforcer *enforcer.Enforcer
	proxy   *httputil.ReverseProxy
}

// NewHandler creates a new proxy handler that enforces labels before
// forwarding requests to the upstream Prometheus remote-write endpoint.
func NewHandler(upstream string, e *enforcer.Enforcer) (*Handler, error) {
	upstreamURL, err := url.Parse(upstream)
	if err != nil {
		return nil, err
	}

	rp := httputil.NewSingleHostReverseProxy(upstreamURL)

	return &Handler{
		enforcer: e,
		proxy:    rp,
	}, nil
}

// ServeHTTP reads the incoming remote-write body, validates labels,
// and either forwards or rejects the request.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	labels := extractLabels(r)

	if err := h.enforcer.Enforce(labels); err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	// Restore the body for the upstream proxy.
	r.Body = io.NopCloser(bytes.NewReader(body))
	r.ContentLength = int64(len(body))

	h.proxy.ServeHTTP(w, r)
}

// extractLabels pulls label key/value pairs from the request headers
// under the X-Prom-Label-* convention.
func extractLabels(r *http.Request) map[string]string {
	labels := make(map[string]string)
	const prefix = "X-Prom-Label-"
	for key, vals := range r.Header {
		if len(key) > len(prefix) && key[:len(prefix)] == prefix {
			labelName := key[len(prefix):]
			if len(vals) > 0 {
				labels[labelName] = vals[0]
			}
		}
	}
	return labels
}
