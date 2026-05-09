package enforcer

import (
	"encoding/json"
	"io"
	"log"
	"time"
)

// AuditEvent represents a single enforcement decision.
type AuditEvent struct {
	Timestamp  time.Time         `json:"timestamp"`
	Allowed    bool              `json:"allowed"`
	Reason     string            `json:"reason,omitempty"`
	Labels     map[string]string `json:"labels"`
	RuleSetID  string            `json:"ruleset_id"`
}

// AuditLogger writes enforcement decisions to an io.Writer as newline-delimited JSON.
type AuditLogger struct {
	writer  io.Writer
	enabled bool
	logger  *log.Logger
}

// NewAuditLogger creates an AuditLogger that writes to w.
// Pass nil to disable audit logging.
func NewAuditLogger(w io.Writer) *AuditLogger {
	if w == nil {
		return &AuditLogger{enabled: false}
	}
	return &AuditLogger{
		writer:  w,
		enabled: true,
		logger:  log.New(w, "", 0),
	}
}

// Record emits an AuditEvent. It is safe to call on a nil or disabled logger.
func (a *AuditLogger) Record(ruleSetID string, labels map[string]string, allowed bool, reason string) {
	if a == nil || !a.enabled {
		return
	}
	event := AuditEvent{
		Timestamp: time.Now().UTC(),
		Allowed:   allowed,
		Reason:    reason,
		Labels:    labels,
		RuleSetID: ruleSetID,
	}
	data, err := json.Marshal(event)
	if err != nil {
		return
	}
	a.logger.Println(string(data))
}
