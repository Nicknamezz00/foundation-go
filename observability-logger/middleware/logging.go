package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/Nicknamezz00/foundation-go/observability-logger/logger"
	"go.opentelemetry.io/otel/trace"
)

// Config defines the configuration for the logging middleware.
type Config struct {
	// Logger is the base logger to use for request logging.
	// Required.
	Logger *slog.Logger

	// RequestIDKey specifies the attribute key for request IDs in logs.
	// Defaults to "request_id" if empty.
	RequestIDKey string

	// RequestIDFunc is an optional function to extract request IDs from context.
	// If nil, request IDs will not be included in logs.
	RequestIDFunc func(context.Context) string

	// SkipPaths is a list of URL paths to skip logging for.
	// Useful for health check endpoints that generate noise.
	SkipPaths []string

	// SkipFunc is an optional function to determine if a request should be skipped.
	// If both SkipPaths and SkipFunc are provided, a request is skipped if either matches.
	SkipFunc func(*http.Request) bool
}

// Logging creates an HTTP middleware that logs requests with structured logging.
// The middleware:
// - Logs each HTTP request with method, path, status, and duration
// - Injects request_id if RequestIDFunc is provided
// - Injects trace_id from OpenTelemetry spans if available
// - Stores the request-scoped logger in context for downstream handlers
//
// This middleware is safe for concurrent use.
func Logging(cfg Config) func(http.Handler) http.Handler {
	// Set defaults
	if cfg.RequestIDKey == "" {
		cfg.RequestIDKey = "request_id"
	}

	// Build skip path map for O(1) lookup
	skipPaths := make(map[string]bool, len(cfg.SkipPaths))
	for _, path := range cfg.SkipPaths {
		skipPaths[path] = true
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check if request should be skipped
			if skipPaths[r.URL.Path] || (cfg.SkipFunc != nil && cfg.SkipFunc(r)) {
				next.ServeHTTP(w, r)
				return
			}

			start := time.Now()

			// Build request-scoped logger
			reqLogger := cfg.Logger

			// Add request_id if extractor provided
			if cfg.RequestIDFunc != nil {
				if reqID := cfg.RequestIDFunc(r.Context()); reqID != "" {
					reqLogger = reqLogger.With(cfg.RequestIDKey, reqID)
				}
			}

			// Add trace_id if span is active
			if sc := trace.SpanFromContext(r.Context()).SpanContext(); sc.IsValid() {
				reqLogger = reqLogger.With("trace_id", sc.TraceID().String())
			}

			// Store logger in context
			ctx := logger.WithContext(r.Context(), reqLogger)

			// Wrap response writer to capture status code
			rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}

			// Process request
			next.ServeHTTP(rw, r.WithContext(ctx))

			// Log request completion
			duration := time.Since(start)
			reqLogger.Info("http request",
				"method", r.Method,
				"path", r.URL.Path,
				"status", rw.status,
				"duration_ms", duration.Milliseconds(),
				"remote_addr", r.RemoteAddr,
			)
		})
	}
}

// responseWriter wraps http.ResponseWriter to capture the status code.
// This type is safe for concurrent use as long as the underlying ResponseWriter is.
type responseWriter struct {
	http.ResponseWriter
	status int
}

// WriteHeader captures the status code and delegates to the underlying ResponseWriter.
func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

// Write delegates to the underlying ResponseWriter.
// If WriteHeader has not been called, the status defaults to 200 OK.
func (rw *responseWriter) Write(b []byte) (int, error) {
	return rw.ResponseWriter.Write(b)
}
