package logger

import (
	"context"
	"log/slog"

	"go.opentelemetry.io/otel/trace"
)

// TraceHandler wraps any slog.Handler and automatically injects trace_id
// from OpenTelemetry spans into log records.
//
// This handler is safe for concurrent use. The wrapped handler must also
// be safe for concurrent use.
type TraceHandler struct {
	handler    slog.Handler
	traceIDKey string
}

// NewTraceHandler creates a new TraceHandler that wraps the given base handler.
// The traceIDKey parameter specifies the attribute key for the trace ID.
//
// This function is safe for concurrent use.
func NewTraceHandler(base slog.Handler, traceIDKey string) *TraceHandler {
	if traceIDKey == "" {
		traceIDKey = "trace_id"
	}
	return &TraceHandler{
		handler:    base,
		traceIDKey: traceIDKey,
	}
}

// Enabled reports whether the handler handles records at the given level.
// This method is safe for concurrent use.
func (h *TraceHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.handler.Enabled(ctx, level)
}

// Handle processes a log record by injecting the trace_id if an active span exists,
// then delegates to the wrapped handler.
//
// This method is safe for concurrent use. The record is modified in-place before
// being passed to the wrapped handler.
func (h *TraceHandler) Handle(ctx context.Context, r slog.Record) error {
	// Extract trace ID from OpenTelemetry span context
	if sc := trace.SpanFromContext(ctx).SpanContext(); sc.IsValid() {
		// Add trace_id attribute to the record
		r.AddAttrs(slog.String(h.traceIDKey, sc.TraceID().String()))
	}
	return h.handler.Handle(ctx, r)
}

// WithAttrs returns a new handler with the given attributes added.
// The returned handler maintains the same trace_id injection behavior.
//
// This method is safe for concurrent use.
func (h *TraceHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &TraceHandler{
		handler:    h.handler.WithAttrs(attrs),
		traceIDKey: h.traceIDKey,
	}
}

// WithGroup returns a new handler with the given group name.
// The returned handler maintains the same trace_id injection behavior.
//
// This method is safe for concurrent use.
func (h *TraceHandler) WithGroup(name string) slog.Handler {
	return &TraceHandler{
		handler:    h.handler.WithGroup(name),
		traceIDKey: h.traceIDKey,
	}
}
