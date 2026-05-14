package proxy

import "context"

type contextKey string

const traceIDKey contextKey = "trace_id"

// withTraceID stores a trace ID in the context.
func withTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, traceIDKey, traceID)
}

// TraceIDFromContext retrieves the trace ID from the context.
// Returns an empty string if no trace ID is present.
func TraceIDFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(traceIDKey).(string); ok {
		return v
	}
	return ""
}
