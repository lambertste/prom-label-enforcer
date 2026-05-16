package proxy

import (
	"crypto/sha256"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// DefaultDedupConfig returns a DedupConfig with sensible defaults.
func DefaultDedupConfig() DedupConfig {
	return DedupConfig{
		TTL:    5 * time.Second,
		Header: "X-Idempotency-Key",
	}
}

// DedupConfig controls deduplication behaviour.
type DedupConfig struct {
	TTL        time.Duration
	Header     string
	Registerer interface {
		Register(interface{}) error
	}
}

type dedupEntry struct {
	expiry time.Time
	status int
	body   []byte
}

type dedupStore struct {
	mu      sync.Mutex
	entries map[string]*dedupEntry
	ttl     time.Duration
}

func newDedupStore(ttl time.Duration) *dedupStore {
	return &dedupStore{
		entries: make(map[string]*dedupEntry),
		ttl:     ttl,
	}
}

func (s *dedupStore) get(key string) (*dedupEntry, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	e, ok := s.entries[key]
	if !ok || time.Now().After(e.expiry) {
		delete(s.entries, key)
		return nil, false
	}
	return e, true
}

func (s *dedupStore) set(key string, status int, body []byte) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.entries[key] = &dedupEntry{
		expiry: time.Now().Add(s.ttl),
		status: status,
		body:   body,
	}
}

// NewDedupMiddleware returns middleware that deduplicates requests sharing
// the same idempotency key within the configured TTL window.
func NewDedupMiddleware(cfg DedupConfig, metrics *DedupMetrics, next http.Handler) http.Handler {
	if cfg.TTL == 0 {
		cfg.TTL = DefaultDedupConfig().TTL
	}
	if cfg.Header == "" {
		cfg.Header = DefaultDedupConfig().Header
	}
	store := newDedupStore(cfg.TTL)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.Header.Get(cfg.Header)
		if key == "" {
			next.ServeHTTP(w, r)
			return
		}
		hashed := fmt.Sprintf("%x", sha256.Sum256([]byte(key)))
		if entry, ok := store.get(hashed); ok {
			if metrics != nil {
				metrics.RecordDuplicate()
			}
			w.WriteHeader(entry.status)
			_, _ = w.Write(entry.body)
			return
		}
		if metrics != nil {
			metrics.RecordUnique()
		}
		rw := &dedupResponseWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rw, r)
		store.set(hashed, rw.status, rw.buf)
	})
}

type dedupResponseWriter struct {
	http.ResponseWriter
	status int
	buf    []byte
	wrote  bool
}

func (d *dedupResponseWriter) WriteHeader(code int) {
	if !d.wrote {
		d.status = code
		d.wrote = true
		d.ResponseWriter.WriteHeader(code)
	}
}

func (d *dedupResponseWriter) Write(b []byte) (int, error) {
	d.buf = append(d.buf, b...)
	return d.ResponseWriter.Write(b)
}
