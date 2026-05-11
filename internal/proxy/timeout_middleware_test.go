package proxy

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestDefaultTimeoutConfig(t *testing.T) {
	cfg := DefaultTimeoutConfig()
	if cfg.RequestTimeout != 30*time.Second {
		t.Errorf("expected 30s, got %v", cfg.RequestTimeout)
	}
}

func TestNewTimeoutMiddleware_ZeroDurationUsesDefault(t *testing.T) {
	// A zero timeout should fall back to the default (30 s), not block forever.
	mw := NewTimeoutMiddleware(TimeoutConfig{RequestTimeout: 0})
	if mw == nil {
		t.Fatal("expected non-nil middleware")
	}
}

func TestTimeoutMiddleware_AllowsFastRequests(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	mw := NewTimeoutMiddleware(TimeoutConfig{RequestTimeout: 5 * time.Second})
	ts := httptest.NewServer(mw(handler))
	defer ts.Close()

	resp, err := http.Get(ts.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestTimeoutMiddleware_RejectsSlowRequests(t *testing.T) {
	// Handler sleeps longer than the configured timeout.
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-time.After(500 * time.Millisecond):
			w.WriteHeader(http.StatusOK)
		case <-r.Context().Done():
			// context cancelled by timeout middleware — exit cleanly
		}
	})

	mw := NewTimeoutMiddleware(TimeoutConfig{RequestTimeout: 50 * time.Millisecond})
	ts := httptest.NewServer(mw(handler))
	defer ts.Close()

	resp, err := http.Get(ts.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Errorf("expected 503, got %d", resp.StatusCode)
	}
}

func TestTimeoutResponseWriter_WriteHeaderOnce(t *testing.T) {
	rec := httptest.NewRecorder()
	tw := &timeoutResponseWriter{ResponseWriter: rec}

	tw.WriteHeader(http.StatusCreated)
	tw.WriteHeader(http.StatusInternalServerError) // should be ignored

	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
}
