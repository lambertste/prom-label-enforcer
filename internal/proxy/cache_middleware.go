package proxy

import (
	"net/http"
	"sync"
	"time"
)

// CacheConfig holds configuration for the response cache middleware.
type CacheConfig struct {
	TTL     time.Duration
	MaxSize int
}

// DefaultCacheConfig returns a CacheConfig with sensible defaults.
func DefaultCacheConfig() CacheConfig {
	return CacheConfig{
		TTL:     30 * time.Second,
		MaxSize: 512,
	}
}

type cacheEntry struct {
	body      []byte
	status    int
	expires   time.Time
}

// ResponseCache is a simple in-memory HTTP response cache.
type ResponseCache struct {
	mu      sync.RWMutex
	entries map[string]cacheEntry
	cfg     CacheConfig
}

// NewResponseCache creates a new ResponseCache with the given config.
func NewResponseCache(cfg CacheConfig) *ResponseCache {
	if cfg.TTL <= 0 {
		cfg.TTL = DefaultCacheConfig().TTL
	}
	if cfg.MaxSize <= 0 {
		cfg.MaxSize = DefaultCacheConfig().MaxSize
	}
	return &ResponseCache{
		entries: make(map[string]cacheEntry, cfg.MaxSize),
		cfg:     cfg,
	}
}

// Get retrieves a cached response body and status for the given key.
// Returns (nil, 0, false) on miss or expiry.
func (c *ResponseCache) Get(key string) ([]byte, int, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	e, ok := c.entries[key]
	if !ok || time.Now().After(e.expires) {
		return nil, 0, false
	}
	return e.body, e.status, true
}

// Set stores a response body and status under the given key.
func (c *ResponseCache) Set(key string, body []byte, status int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.entries) >= c.cfg.MaxSize {
		// Evict one arbitrary entry to stay within MaxSize.
		for k := range c.entries {
			delete(c.entries, k)
			break
		}
	}
	c.entries[key] = cacheEntry{
		body:    body,
		status:  status,
		expires: time.Now().Add(c.cfg.TTL),
	}
}

// Middleware returns an http.Handler that serves cached GET responses.
func (c *ResponseCache) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			next.ServeHTTP(w, r)
			return
		}
		key := r.URL.RequestURI()
		if body, status, ok := c.Get(key); ok {
			w.WriteHeader(status)
			_, _ = w.Write(body)
			return
		}
		rec := &cachingResponseWriter{header: w.Header(), code: http.StatusOK}
		next.ServeHTTP(rec, r)
		c.Set(key, rec.body, rec.code)
		w.WriteHeader(rec.code)
		_, _ = w.Write(rec.body)
	})
}

type cachingResponseWriter struct {
	header http.Header
	code   int
	body   []byte
}

func (rw *cachingResponseWriter) Header() http.Header        { return rw.header }
func (rw *cachingResponseWriter) WriteHeader(code int)       { rw.code = code }
func (rw *cachingResponseWriter) Write(b []byte) (int, error) {
	rw.body = append(rw.body, b...)
	return len(b), nil
}
