package proxy

import (
	"net/http"

	"github.com/prom-label-enforcer/internal/enforcer"
)

// auditMiddleware wraps an http.Handler and emits an AuditEvent for every
// /receive request that passes through the enforcer.
type auditMiddleware struct {
	next        http.Handler
	auditLogger *enforcer.AuditLogger
	ruleSetID   string
}

// NewAuditMiddleware returns an http.Handler that records enforcement decisions
// via the provided AuditLogger before delegating to next.
func NewAuditMiddleware(next http.Handler, al *enforcer.AuditLogger, ruleSetID string) http.Handler {
	if al == nil {
		return next
	}
	return &auditMiddleware{
		next:        next,
		auditLogger: al,
		ruleSetID:   ruleSetID,
	}
}

func (m *auditMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rw := &responseCapture{ResponseWriter: w, statusCode: http.StatusOK}
	m.next.ServeHTTP(rw, r)

	labels := extractLabels(r)
	allowed := rw.statusCode < 400
	reason := ""
	if !allowed {
		reason = http.StatusText(rw.statusCode)
	}
	m.auditLogger.Record(m.ruleSetID, labels, allowed, reason)
}

// responseCapture captures the HTTP status code written by the wrapped handler.
type responseCapture struct {
	http.ResponseWriter
	statusCode int
}

func (rc *responseCapture) WriteHeader(code int) {
	rc.statusCode = code
	rc.ResponseWriter.WriteHeader(code)
}
