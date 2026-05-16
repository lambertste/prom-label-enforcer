package proxy

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

func newMirrorRegistry() prometheus.Registerer {
	return prometheus.NewRegistry()
}

func TestDefaultMirrorConfig(t *testing.T) {
	cfg := DefaultMirrorConfig()
	if cfg.TimeoutSeconds != 2 {
		t.Errorf("expected TimeoutSeconds=2, got %d", cfg.TimeoutSeconds)
	}
	if cfg.SampleRate != 1.0 {
		t.Errorf("expected SampleRate=1.0, got %f", cfg.SampleRate)
	}
}

func TestMirrorMiddleware_NoTargetIsNoop(t *testing.T) {
	primary := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	h, err := NewMirrorMiddleware(MirrorConfig{}, primary)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestMirrorMiddleware_MirrorsRequest(t *testing.T) {
	var mirrorHits atomic.Int32
	mirrorSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mirrorHits.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer mirrorSrv.Close()

	primary := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
	})

	cfg := MirrorConfig{
		TargetURL:      mirrorSrv.URL,
		TimeoutSeconds: 2,
		SampleRate:     1.0,
		Registerer:     newMirrorRegistry(),
	}
	h, err := NewMirrorMiddleware(cfg, primary)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/metrics", strings.NewReader("hello")))
	if rec.Code != http.StatusAccepted {
		t.Errorf("expected 202, got %d", rec.Code)
	}

	// Allow goroutine to complete.
	time.Sleep(200 * time.Millisecond)
	if mirrorHits.Load() != 1 {
		t.Errorf("expected 1 mirror hit, got %d", mirrorHits.Load())
	}
}

func TestMirrorMiddleware_ErrorOnUnreachableTarget(t *testing.T) {
	primary := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	cfg := MirrorConfig{
		TargetURL:      "http://127.0.0.1:1", // unreachable
		TimeoutSeconds: 1,
		Registerer:     newMirrorRegistry(),
	}
	h, err := NewMirrorMiddleware(cfg, primary)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusOK {
		t.Errorf("primary should still return 200, got %d", rec.Code)
	}
	// Give the goroutine time to fail.
	time.Sleep(200 * time.Millisecond)
}

func TestMirrorMetrics_RecordSent(t *testing.T) {
	reg := prometheus.NewRegistry()
	m := NewMirrorMetrics(reg)
	m.RecordSent()
	m.RecordSent()
	mfs, _ := reg.Gather()
	for _, mf := range mfs {
		if mf.GetName() == "prom_label_enforcer_mirror_requests_sent_total" {
			if got := mf.GetMetric()[0].GetCounter().GetValue(); got != 2 {
				t.Errorf("expected 2 sent, got %f", got)
			}
			return
		}
	}
	t.Error("sent counter metric not found")
}

func TestMirrorMetrics_NilSafe(t *testing.T) {
	var m *MirrorMetrics
	m.RecordSent()
	m.RecordError()
	_ = io.Discard // suppress unused import
}
