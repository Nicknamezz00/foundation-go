package interceptor

import (
	"context"
	"log/slog"
	"time"

	"github.com/Nicknamezz00/foundation-go/observability-logger/logger"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Config defines the configuration for the gRPC logging interceptor.
type Config struct {
	// Logger is the base logger to use for RPC logging.
	// Required.
	Logger *slog.Logger

	// RequestIDKey specifies the attribute key for request IDs in logs.
	// Defaults to "request_id" if empty.
	RequestIDKey string

	// RequestIDFunc is an optional function to extract request IDs from context.
	// If nil, request IDs will not be included in logs.
	RequestIDFunc func(context.Context) string

	// SkipMethods is a list of gRPC method names to skip logging for.
	// Useful for health check methods that generate noise.
	// Example: []string{"/grpc.health.v1.Health/Check"}
	SkipMethods []string

	// SkipFunc is an optional function to determine if a request should be skipped.
	// If both SkipMethods and SkipFunc are provided, a request is skipped if either matches.
	SkipFunc func(ctx context.Context, method string) bool
}

// UnaryServerInterceptor creates a gRPC unary server interceptor that logs requests.
// The interceptor:
// - Logs each gRPC request with method, status code, and duration
// - Injects request_id if RequestIDFunc is provided
// - Injects trace_id from OpenTelemetry spans if available
// - Stores the request-scoped logger in context for downstream handlers
//
// This interceptor is safe for concurrent use.
func UnaryServerInterceptor(cfg Config) grpc.UnaryServerInterceptor {
	// Set defaults
	if cfg.RequestIDKey == "" {
		cfg.RequestIDKey = "request_id"
	}

	// Build skip method map for O(1) lookup
	skipMethods := make(map[string]bool, len(cfg.SkipMethods))
	for _, method := range cfg.SkipMethods {
		skipMethods[method] = true
	}

	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		// Check if request should be skipped
		if skipMethods[info.FullMethod] || (cfg.SkipFunc != nil && cfg.SkipFunc(ctx, info.FullMethod)) {
			return handler(ctx, req)
		}

		start := time.Now()

		// Build request-scoped logger
		reqLogger := cfg.Logger

		// Add request_id if extractor provided
		if cfg.RequestIDFunc != nil {
			if reqID := cfg.RequestIDFunc(ctx); reqID != "" {
				reqLogger = reqLogger.With(cfg.RequestIDKey, reqID)
			}
		}

		// Add trace_id if span is active
		if sc := trace.SpanFromContext(ctx).SpanContext(); sc.IsValid() {
			reqLogger = reqLogger.With("trace_id", sc.TraceID().String())
		}

		// Store logger in context
		ctx = logger.WithContext(ctx, reqLogger)

		// Process request
		resp, err := handler(ctx, req)

		// Log request completion
		duration := time.Since(start)
		code := status.Code(err)

		logAttrs := []any{
			"method", info.FullMethod,
			"code", code.String(),
			"duration_ms", duration.Milliseconds(),
		}

		// Log at appropriate level based on status code
		if err != nil {
			if code == codes.Internal || code == codes.Unknown || code == codes.DataLoss {
				reqLogger.Error("grpc request", append(logAttrs, "error", err.Error())...)
			} else {
				reqLogger.Warn("grpc request", append(logAttrs, "error", err.Error())...)
			}
		} else {
			reqLogger.Info("grpc request", logAttrs...)
		}

		return resp, err
	}
}

// StreamServerInterceptor creates a gRPC stream server interceptor that logs requests.
// The interceptor:
// - Logs each gRPC stream with method and duration
// - Injects request_id if RequestIDFunc is provided
// - Injects trace_id from OpenTelemetry spans if available
// - Stores the request-scoped logger in context for downstream handlers
//
// This interceptor is safe for concurrent use.
func StreamServerInterceptor(cfg Config) grpc.StreamServerInterceptor {
	// Set defaults
	if cfg.RequestIDKey == "" {
		cfg.RequestIDKey = "request_id"
	}

	// Build skip method map for O(1) lookup
	skipMethods := make(map[string]bool, len(cfg.SkipMethods))
	for _, method := range cfg.SkipMethods {
		skipMethods[method] = true
	}

	return func(
		srv interface{},
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		ctx := ss.Context()

		// Check if request should be skipped
		if skipMethods[info.FullMethod] || (cfg.SkipFunc != nil && cfg.SkipFunc(ctx, info.FullMethod)) {
			return handler(srv, ss)
		}

		start := time.Now()

		// Build request-scoped logger
		reqLogger := cfg.Logger

		// Add request_id if extractor provided
		if cfg.RequestIDFunc != nil {
			if reqID := cfg.RequestIDFunc(ctx); reqID != "" {
				reqLogger = reqLogger.With(cfg.RequestIDKey, reqID)
			}
		}

		// Add trace_id if span is active
		if sc := trace.SpanFromContext(ctx).SpanContext(); sc.IsValid() {
			reqLogger = reqLogger.With("trace_id", sc.TraceID().String())
		}

		// Store logger in context
		ctx = logger.WithContext(ctx, reqLogger)

		// Wrap the stream with the new context
		wrappedStream := &wrappedServerStream{
			ServerStream: ss,
			ctx:          ctx,
		}

		// Process stream
		err := handler(srv, wrappedStream)

		// Log stream completion
		duration := time.Since(start)
		code := status.Code(err)

		logAttrs := []any{
			"method", info.FullMethod,
			"code", code.String(),
			"duration_ms", duration.Milliseconds(),
			"is_client_stream", info.IsClientStream,
			"is_server_stream", info.IsServerStream,
		}

		// Log at appropriate level based on status code
		if err != nil {
			if code == codes.Internal || code == codes.Unknown || code == codes.DataLoss {
				reqLogger.Error("grpc stream", append(logAttrs, "error", err.Error())...)
			} else {
				reqLogger.Warn("grpc stream", append(logAttrs, "error", err.Error())...)
			}
		} else {
			reqLogger.Info("grpc stream", logAttrs...)
		}

		return err
	}
}

// wrappedServerStream wraps a grpc.ServerStream to inject a custom context.
type wrappedServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

// Context returns the wrapped context.
func (w *wrappedServerStream) Context() context.Context {
	return w.ctx
}
