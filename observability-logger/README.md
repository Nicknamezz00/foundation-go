# Foundation Go Logger

A lightweight, production-ready logging SDK for Go applications with OpenTelemetry integration.

## Features

- **Structured Logging**: Built on Go's standard `log/slog` package
- **OpenTelemetry Integration**: Automatic trace_id injection from active spans
- **HTTP Middleware**: Request logging with customizable request ID extraction
- **gRPC Interceptors**: Unary and stream interceptors for gRPC services
- **Context-Aware**: Store and retrieve loggers from context
- **Flexible Configuration**: JSON or text output, configurable log levels
- **Source Location**: Repository-relative file paths and line numbers
- **Concurrency-Safe**: All components are safe for concurrent use

## Installation

```bash
go get github.com/Nicknamezz00/foundation-go/log
```

## Quick Start

### Basic Logger

```go
package main

import (
    "github.com/Nicknamezz00/foundation-go/log/logger"
)

func main() {
    // Create a logger at INFO level
    log := logger.NewInfo("my-service")

    log.Info("application started")
    log.Debug("this won't be logged") // below INFO level
}
```

### With Custom Configuration

```go
log := logger.New("my-service", logger.Config{
    Level:      logger.LevelDebug,
    Format:     "json",
    AddTraceID: true,
})

log.Debug("debug message", "key", "value")
log.Info("info message", "user_id", 123)
log.Warn("warning message")
log.Error("error message", "error", err)
```

### With Source Code Location

Enable source location to include file and line number in logs:

```go
log := logger.New("my-service", logger.Config{
    Level:     logger.LevelDebug,
    AddSource: true, // Enable source location (file:line)
})

log.Info("processing request", "user_id", 123)
// Output includes: "source":"internal/handler/user.go:42"
```

**How it works:**
- Automatically finds your repository root (looks for `.git`, `go.mod`, etc.)
- Converts absolute paths to repository-relative paths
- Thread-safe with cached repository root lookup
- Falls back to filename only if repository root cannot be determined

**Example output:**
```json
{
  "time": "2026-02-27T10:30:00Z",
  "level": "INFO",
  "msg": "processing request",
  "service": "my-service",
  "source": "internal/handler/user.go:42",
  "user_id": 123
}
```

**Performance note:** Source location adds a small overhead (runtime.Caller). Enable it for debugging or in development, but consider disabling in high-throughput production environments.

### With OpenTelemetry Trace ID

```go
import (
    "context"
    "github.com/Nicknamezz00/foundation-go/log/logger"
    "go.opentelemetry.io/otel"
)

func main() {
    log := logger.New("my-service", logger.Config{
        Level:      logger.LevelInfo,
        AddTraceID: true, // Enable trace_id injection
    })

    ctx := context.Background()
    ctx, span := otel.Tracer("my-service").Start(ctx, "operation")
    defer span.End()

    // Logs will include trace_id from the active span
    log.InfoContext(ctx, "processing request")
}
```

### Context-Based Logging

```go
import (
    "context"
    "github.com/Nicknamezz00/foundation-go/log/logger"
)

func main() {
    log := logger.NewInfo("my-service")

    // Store logger in context
    ctx := logger.WithContext(context.Background(), log)

    // Retrieve and use logger from context
    processRequest(ctx)
}

func processRequest(ctx context.Context) {
    // Get logger from context
    log := logger.FromContext(ctx)
    log.Info("processing request")

    // Or use convenience functions
    logger.Info(ctx, "processing request")
    logger.Error(ctx, "failed to process", "error", err)
}
```

### HTTP Middleware

```go
import (
    "net/http"
    "github.com/Nicknamezz00/foundation-go/log/logger"
    "github.com/Nicknamezz00/foundation-go/log/middleware"
)

func main() {
    log := logger.NewInfo("my-service")

    // Create middleware with request ID extraction
    loggingMiddleware := middleware.Logging(middleware.Config{
        Logger:       log,
        RequestIDFunc: func(ctx context.Context) string {
            // Extract request ID from your middleware
            return getRequestID(ctx)
        },
        SkipPaths: []string{"/health", "/metrics"},
    })

    mux := http.NewServeMux()
    mux.HandleFunc("/api/users", handleUsers)

    // Wrap with logging middleware
    http.ListenAndServe(":8080", loggingMiddleware(mux))
}

func handleUsers(w http.ResponseWriter, r *http.Request) {
    // Logger is available in context with request_id and trace_id
    logger.Info(r.Context(), "handling user request")
    w.WriteHeader(http.StatusOK)
}
```

### gRPC Interceptor

```go
import (
    "context"
    "github.com/Nicknamezz00/foundation-go/log/logger"
    "github.com/Nicknamezz00/foundation-go/log/interceptor"
    "google.golang.org/grpc"
)

func main() {
    log := logger.NewInfo("my-service")

    // Create gRPC server with logging interceptor
    server := grpc.NewServer(
        grpc.UnaryInterceptor(
            interceptor.UnaryServerInterceptor(interceptor.Config{
                Logger: log,
                RequestIDFunc: func(ctx context.Context) string {
                    // Extract request ID from gRPC metadata
                    return getRequestIDFromMetadata(ctx)
                },
                SkipMethods: []string{
                    "/grpc.health.v1.Health/Check",
                },
            }),
        ),
        grpc.StreamInterceptor(
            interceptor.StreamServerInterceptor(interceptor.Config{
                Logger: log,
            }),
        ),
    )

    // Register your gRPC services
    pb.RegisterUserServiceServer(server, &userService{})

    lis, _ := net.Listen("tcp", ":50051")
    server.Serve(lis)
}

// In your gRPC handler
func (s *userService) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.User, error) {
    // Logger is available in context with request_id and trace_id
    logger.Info(ctx, "getting user", "user_id", req.Id)

    // Or retrieve explicitly
    log := logger.FromContext(ctx)
    log.Info("processing request")

    return &pb.User{Id: req.Id}, nil
}
```

### gRPC Features

The gRPC interceptor provides:

- **Automatic logging**: Logs method name, status code, and duration for every RPC
- **Smart log levels**: Errors logged at ERROR level (Internal, Unknown, DataLoss) or WARN level (other errors)
- **Stream support**: Both unary and streaming RPCs are supported
- **Context injection**: Logger is automatically stored in context for handlers
- **Skip methods**: Skip logging for health checks or other noisy methods

**Log output example:**
```json
{
  "time": "2026-02-27T10:30:00Z",
  "level": "INFO",
  "msg": "grpc request",
  "service": "my-service",
  "method": "/user.v1.UserService/GetUser",
  "code": "OK",
  "duration_ms": 42,
  "trace_id": "abc123...",
  "request_id": "req-456"
}
```

## Log Levels

The SDK provides four log levels (re-exported from `slog`):

- `logger.LevelDebug` (-4): Detailed debugging information
- `logger.LevelInfo` (0): General informational messages
- `logger.LevelWarn` (4): Warning messages
- `logger.LevelError` (8): Error messages

### Convenience Constructors

```go
logger.NewDebug("service")  // DEBUG level
logger.NewInfo("service")   // INFO level
logger.NewWarn("service")   // WARN level
logger.NewError("service")  // ERROR level
```

## Configuration

### Logger Config

```go
type Config struct {
    // Level specifies the minimum log level
    Level slog.Level

    // Format: "json" or "text" (default: "json")
    Format string

    // Output: where to write logs (default: os.Stdout)
    Output io.Writer

    // AddTraceID: enable trace_id injection (default: false)
    AddTraceID bool

    // TraceIDKey: attribute key for trace_id (default: "trace_id")
    TraceIDKey string

    // AddSource: enable source code location (file:line) (default: false)
    AddSource bool

    // SourceKey: attribute key for source location (default: "source")
    SourceKey string
}
```

### Middleware Config

```go
type Config struct {
    // Logger: base logger for request logging (required)
    Logger *slog.Logger

    // RequestIDKey: attribute key for request IDs (default: "request_id")
    RequestIDKey string

    // RequestIDFunc: extract request ID from context (optional)
    RequestIDFunc func(context.Context) string

    // SkipPaths: paths to skip logging (e.g., health checks)
    SkipPaths []string

    // SkipFunc: custom skip logic (optional)
    SkipFunc func(*http.Request) bool
}
```

### gRPC Interceptor Config

```go
type Config struct {
    // Logger: base logger for RPC logging (required)
    Logger *slog.Logger

    // RequestIDKey: attribute key for request IDs (default: "request_id")
    RequestIDKey string

    // RequestIDFunc: extract request ID from context (optional)
    RequestIDFunc func(context.Context) string

    // SkipMethods: gRPC methods to skip logging (e.g., health checks)
    // Example: []string{"/grpc.health.v1.Health/Check"}
    SkipMethods []string

    // SkipFunc: custom skip logic (optional)
    SkipFunc func(ctx context.Context, method string) bool
}
```

## Concurrency Safety

All components are designed for safe concurrent use:

- **Logger**: `slog.Logger` is safe for concurrent use
- **TraceHandler**: Immutable after creation, safe for concurrent use
- **Middleware**: Each request gets its own logger instance
- **Context functions**: Safe for concurrent access

## Usage Patterns

### ⚠️ Important: Initialize Once, Reuse Everywhere

**DO NOT** create a new logger instance every time you need to log. The logger should be initialized once at application startup and reused throughout your application.

### Recommended Pattern: Initialize Once in Main

```go
func main() {
    // ✅ Initialize once with your configuration
    log := logger.New("my-service", logger.Config{
        Level:      logger.LevelInfo,
        AddTraceID: true,
        AddSource:  true,
    })

    // Store in context for use throughout the application
    ctx := logger.WithContext(context.Background(), log)

    // Pass context to your handlers
    handleRequest(ctx)
}

func handleRequest(ctx context.Context) {
    // ✅ Retrieve from context when needed
    log := logger.FromContext(ctx)
    log.Info("processing request", "user_id", 123)

    // ✅ Or use convenience functions
    logger.Info(ctx, "processing request", "user_id", 123)
}

// ❌ WRONG: Don't do this
func handleRequest(ctx context.Context) {
    log := logger.NewInfo("my-service") // Creates new logger every call!
    log.Info("processing request")
}
```

### Pattern for HTTP Servers

The middleware automatically stores the logger in the request context:

```go
func main() {
    // Initialize once
    log := logger.New("my-service", logger.Config{
        Level:      logger.LevelInfo,
        AddTraceID: true,
        AddSource:  true,
    })

    // Middleware stores logger in request context
    loggingMW := middleware.Logging(middleware.Config{
        Logger: log,
        RequestIDFunc: func(ctx context.Context) string {
            // Extract request ID from your middleware
            return getRequestID(ctx)
        },
        SkipPaths: []string{"/health", "/metrics"},
    })

    mux := http.NewServeMux()
    mux.HandleFunc("/api/users", handleUsers)

    http.ListenAndServe(":8080", loggingMW(mux))
}

func handleUsers(w http.ResponseWriter, r *http.Request) {
    // Logger is already in r.Context() from middleware
    logger.Info(r.Context(), "handling user request")

    // Or retrieve explicitly
    log := logger.FromContext(r.Context())
    log.Info("processing", "user_id", 123)
}
```

### Two Ways to Log from Context

Once the logger is stored in context, you can use either approach:

**Option 1: Retrieve and use the logger**
```go
log := logger.FromContext(ctx)
log.Info("message", "key", "value")
log.Error("error occurred", "error", err)
```

**Option 2: Use convenience functions**
```go
logger.Info(ctx, "message", "key", "value")
logger.Error(ctx, "error occurred", "error", err)
```

Both approaches are equivalent - the convenience functions internally call `FromContext(ctx)`.

## Best Practices

1. **Initialize once**: Create your logger in `main()` with your desired configuration
2. **Store in context**: Use `WithContext` to make it available throughout your request/operation
3. **Reuse everywhere**: Retrieve with `FromContext` or use convenience functions
4. **Never create repeatedly**: Don't call `logger.New()` in handlers or business logic
5. **Structured logging**: Use key-value pairs instead of string formatting
6. **Appropriate levels**: Use DEBUG for development, INFO for production
7. **Enable source location**: Use `AddSource: true` for debugging (adds small overhead)
8. **Trace integration**: Enable `AddTraceID: true` when using OpenTelemetry

### Why Initialize Once?

- **Performance**: Creating loggers is not free - do it once, not per request
- **Configuration**: Your logger configuration should be consistent across the application
- **Concurrency-safe**: A single logger instance is safe for concurrent use across goroutines
- **Memory efficient**: Reusing the same logger reduces allocations

## Examples

See the `examples/` directory for complete working examples:

- `basic/` - Simple logger usage with different configurations

### Complete Example: Production Setup

```go
package main

import (
    "context"
    "net/http"
    "os"

    "github.com/Nicknamezz00/foundation-go/log/logger"
    "github.com/Nicknamezz00/foundation-go/log/middleware"
    "go.opentelemetry.io/otel"
)

func main() {
    // Initialize logger once with production configuration
    log := logger.New("my-service", logger.Config{
        Level:      getLogLevel(),
        Format:     "json",
        AddTraceID: true,
        AddSource:  false, // Disable in production for performance
    })

    // Set up HTTP server with logging middleware
    loggingMW := middleware.Logging(middleware.Config{
        Logger:        log,
        RequestIDFunc: extractRequestID,
        SkipPaths:     []string{"/health", "/metrics"},
    })

    mux := http.NewServeMux()
    mux.HandleFunc("/api/users", handleUsers)
    mux.HandleFunc("/health", handleHealth)

    log.Info("starting server", "port", 8080)
    if err := http.ListenAndServe(":8080", loggingMW(mux)); err != nil {
        log.Error("server failed", "error", err)
        os.Exit(1)
    }
}

func handleUsers(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()

    // Start OpenTelemetry span
    ctx, span := otel.Tracer("my-service").Start(ctx, "handleUsers")
    defer span.End()

    // Log with context - includes request_id and trace_id
    logger.Info(ctx, "handling user request", "method", r.Method)

    // Business logic here
    users := fetchUsers(ctx)

    logger.Info(ctx, "users fetched", "count", len(users))
    w.WriteHeader(http.StatusOK)
}

func fetchUsers(ctx context.Context) []User {
    // Logger is available throughout the call chain
    logger.Debug(ctx, "fetching users from database")

    // Simulate database call
    users := queryDatabase(ctx)

    logger.Debug(ctx, "users retrieved", "count", len(users))
    return users
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
    // Health checks are skipped by middleware (SkipPaths)
    w.WriteHeader(http.StatusOK)
}

func extractRequestID(ctx context.Context) string {
    // Extract from your request ID middleware
    if reqID, ok := ctx.Value("request_id").(string); ok {
        return reqID
    }
    return ""
}

func getLogLevel() logger.Level {
    env := os.Getenv("LOG_LEVEL")
    switch env {
    case "debug":
        return logger.LevelDebug
    case "warn":
        return logger.LevelWarn
    case "error":
        return logger.LevelError
    default:
        return logger.LevelInfo
    }
}
```

## License

MIT
