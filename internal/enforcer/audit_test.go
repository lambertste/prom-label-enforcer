package enforcer

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestNewAuditLogger_NilWriter(t *testing.T) {
	al := NewAuditLogger(nil)
	if al == nil {
		t.Fatal("expected non-nil AuditLogger")
	}
	if al.enabled {
		t.Error("expected logger to be disabled when writer is nil")
	}
}

func TestAuditLogger_Record_Disabled(t *testing.T) {
	// Should not panic when disabled
	al := NewAuditLogger(nil)
	al.Record("rs1", map[string]string{"env": "prod"}, true, "")
}

func TestAuditLogger_Record_NilReceiver(t *testing.T) {
	// Should not panic on nil receiver
	var al *AuditLogger
	al.Record("rs1", map[string]string{"env": "prod"}, false, "missing label")
}

func TestAuditLogger_Record_WritesJSON(t *testing.T) {
	var buf bytes.Buffer
	al := NewAuditLogger(&buf)

	labels := map[string]string{"env": "staging", "team": "infra"}
	al.Record("ruleset-a", labels, true, "")

	line := strings.TrimSpace(buf.String())
	if line == "" {
		t.Fatal("expected output, got empty")
	}

	var event AuditEvent
	if err := json.Unmarshal([]byte(line), &event); err != nil {
		t.Fatalf("failed to unmarshal audit event: %v", err)
	}
	if event.RuleSetID != "ruleset-a" {
		t.Errorf("expected ruleset_id 'ruleset-a', got %q", event.RuleSetID)
	}
	if !event.Allowed {
		t.Error("expected allowed=true")
	}
	if event.Labels["env"] != "staging" {
		t.Errorf("expected label env=staging, got %q", event.Labels["env"])
	}
	if event.Timestamp.IsZero() {
		t.Error("expected non-zero timestamp")
	}
}

func TestAuditLogger_Record_Rejected(t *testing.T) {
	var buf bytes.Buffer
	al := NewAuditLogger(&buf)

	al.Record("ruleset-b", map[string]string{"env": "unknown"}, false, "disallowed value for env")

	var event AuditEvent
	if err := json.Unmarshal([]byte(strings.TrimSpace(buf.String())), &event); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if event.Allowed {
		t.Error("expected allowed=false")
	}
	if event.Reason != "disallowed value for env" {
		t.Errorf("unexpected reason: %q", event.Reason)
	}
}

func TestAuditLogger_Record_MultipleEvents(t *testing.T) {
	var buf bytes.Buffer
	al := NewAuditLogger(&buf)

	al.Record("rs", map[string]string{"a": "1"}, true, "")
	al.Record("rs", map[string]string{"a": "2"}, false, "bad")

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 log lines, got %d", len(lines))
	}
}
