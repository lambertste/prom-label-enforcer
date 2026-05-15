package proxy

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// DefaultUpstreamConfig returns a sensible default upstream proxy configuration.
func DefaultUpstreamConfig() UpstreamConfig {
	return UpstreamConfig{
		DialTimeout:  5 * time.Second,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		FlushInterval: 100 * time.Millisecond,
	}
}

// UpstreamConfig holds configuration for the reverse-proxy upstream middleware.
type UpstreamConfig struct {
	Target        string
	DialTimeout   time.Duration
	ReadTimeout   time.Duration
	WriteTimeout  time.Duration
	FlushInterval time.Duration
	Registerer    prometheus.Registerer
}

// NewUpstreamMiddleware creates an HTTP handler that reverse-proxies requests
// to the configured upstream target, recording metrics on each hop.
func NewUpstreamMiddleware(cfg UpstreamConfig) (http.Handler, error) {
	if cfg.DialTimeout == 0 {
		cfg = DefaultUpstreamConfig()
	}
	if cfg.Registerer == nil {
		cfg.Registerer = prometheus.DefaultRegisterer
	}

	target, err := url.Parse(cfg.Target)
	if err != nil {
		return nil, err
	}

	metrics := NewUpstreamMetrics(cfg.Registerer)

	proxy := httputil.NewSingleHostReverseProxy(target)
	proxy.FlushInterval = cfg.FlushInterval

	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		req.Header.Set("X-Forwarded-Host", req.Host)
	}

	proxy.ModifyResponse = func(resp *http.Response) error {
		if resp.StatusCode >= 500 {
			metrics.recordError(resp.Request.URL.Path)
		} else {
			metrics.recordSuccess(resp.Request.URL.Path)
		}
		return nil
	}

	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		metrics.recordError(r.URL.Path)
		http.Error(w, "upstream error: "+err.Error(), http.StatusBadGateway)
	}

	return proxy, nil
}
