package logger

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"sync"
)

// Log levels re-exported from slog for convenience.
const (
	LevelDebug = slog.LevelDebug // -4
	LevelInfo  = slog.LevelInfo  // 0
	LevelWarn  = slog.LevelWarn  // 4
	LevelError = slog.LevelError // 8
)

// contextKey is a private type for context keys to avoid collisions.
type contextKey struct{}

var loggerKey = contextKey{}

var (
	// repoRoot caches the repository root path for efficient path shortening.
	repoRoot     string
	repoRootOnce sync.Once
)

// findRepoRoot attempts to find the repository root by looking for common markers.
// It searches upward from the current working directory.
func findRepoRoot() string {
	// Try to get the caller's file path to start searching from
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		return ""
	}

	dir := filepath.Dir(file)

	// Walk up the directory tree looking for repository markers
	for {
		// Check for common repository markers
		markers := []string{".git", "go.mod", ".hg", ".svn"}
		for _, marker := range markers {
			if _, err := os.Stat(filepath.Join(dir, marker)); err == nil {
				return dir
			}
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached root without finding markers
			break
		}
		dir = parent
	}

	// Fallback: try current working directory
	if wd, err := os.Getwd(); err == nil {
		return wd
	}

	return ""
}

// shortenPath converts an absolute file path to a repository-relative path.
// Returns "file:line" format. Thread-safe through sync.Once initialization.
func shortenPath(file string, line int) string {
	repoRootOnce.Do(func() {
		repoRoot = findRepoRoot()
	})

	if repoRoot != "" {
		if rel, err := filepath.Rel(repoRoot, file); err == nil {
			// Successfully made relative path
			return fmt.Sprintf("%s:%d", rel, line)
		}
	}

	// Fallback: use just the filename if we can't make it relative
	return fmt.Sprintf("%s:%d", filepath.Base(file), line)
}

// Config defines the configuration for creating a logger.
type Config struct {
	// Level specifies the minimum log level (e.g., slog.LevelDebug, slog.LevelInfo).
	Level slog.Level

	// Format specifies the output format: "json" or "text".
	// Defaults to "json" if empty.
	Format string

	// Output specifies where logs should be written.
	// Defaults to os.Stdout if nil.
	Output io.Writer

	// AddTraceID enables automatic trace_id injection from OpenTelemetry spans.
	// Defaults to false.
	AddTraceID bool

	// TraceIDKey specifies the key name for trace_id in log records.
	// Defaults to "trace_id" if empty.
	TraceIDKey string

	// AddSource enables source code location (file:line) in log records.
	// The file path will be relative to the repository root.
	// Defaults to false.
	AddSource bool

	// SourceKey specifies the key name for source location in log records.
	// Defaults to "source" if empty.
	SourceKey string
}

// New creates a new slog.Logger with the specified service name and configuration.
// The logger is configured based on the provided Config and includes the service name
// as a default attribute in all log records.
//
// This function is safe for concurrent use.
func New(serviceName string, cfg Config) *slog.Logger {
	// Set defaults
	if cfg.Output == nil {
		cfg.Output = os.Stdout
	}
	if cfg.Format == "" {
		cfg.Format = "json"
	}
	if cfg.TraceIDKey == "" {
		cfg.TraceIDKey = "trace_id"
	}
	if cfg.SourceKey == "" {
		cfg.SourceKey = "source"
	}

	// Create base handler
	var base slog.Handler
	opts := &slog.HandlerOptions{
		Level:     cfg.Level,
		AddSource: cfg.AddSource,
	}

	// Use custom ReplaceAttr to shorten source paths
	if cfg.AddSource {
		opts.ReplaceAttr = func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.SourceKey {
				if source, ok := a.Value.Any().(*slog.Source); ok {
					// Shorten the file path to be relative to repository root
					a.Value = slog.StringValue(shortenPath(source.File, source.Line))
				}
			}
			return a
		}
	}

	if cfg.Format == "text" {
		base = slog.NewTextHandler(cfg.Output, opts)
	} else {
		base = slog.NewJSONHandler(cfg.Output, opts)
	}

	// Wrap with TraceHandler if enabled
	if cfg.AddTraceID {
		base = NewTraceHandler(base, cfg.TraceIDKey)
	}

	// Create logger with service name
	return slog.New(base).With("service", serviceName)
}

// NewDebug creates a logger at DEBUG level with default configuration.
// This is a convenience function equivalent to New(serviceName, Config{Level: LevelDebug}).
func NewDebug(serviceName string) *slog.Logger {
	return New(serviceName, Config{Level: LevelDebug})
}

// NewInfo creates a logger at INFO level with default configuration.
// This is a convenience function equivalent to New(serviceName, Config{Level: LevelInfo}).
func NewInfo(serviceName string) *slog.Logger {
	return New(serviceName, Config{Level: LevelInfo})
}

// NewWarn creates a logger at WARN level with default configuration.
// This is a convenience function equivalent to New(serviceName, Config{Level: LevelWarn}).
func NewWarn(serviceName string) *slog.Logger {
	return New(serviceName, Config{Level: LevelWarn})
}

// NewError creates a logger at ERROR level with default configuration.
// This is a convenience function equivalent to New(serviceName, Config{Level: LevelError}).
func NewError(serviceName string) *slog.Logger {
	return New(serviceName, Config{Level: LevelError})
}

// WithContext stores the logger in the context.
// This function is safe for concurrent use.
func WithContext(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, loggerKey, logger)
}

// FromContext retrieves the logger from the context.
// If no logger is found, it returns the default logger.
// This function is safe for concurrent use.
func FromContext(ctx context.Context) *slog.Logger {
	if logger, ok := ctx.Value(loggerKey).(*slog.Logger); ok {
		return logger
	}
	return slog.Default()
}

// Debug logs a debug message using the logger from context.
// This is a convenience function equivalent to FromContext(ctx).DebugContext(ctx, msg, args...).
func Debug(ctx context.Context, msg string, args ...any) {
	FromContext(ctx).DebugContext(ctx, msg, args...)
}

// Info logs an info message using the logger from context.
// This is a convenience function equivalent to FromContext(ctx).InfoContext(ctx, msg, args...).
func Info(ctx context.Context, msg string, args ...any) {
	FromContext(ctx).InfoContext(ctx, msg, args...)
}

// Warn logs a warning message using the logger from context.
// This is a convenience function equivalent to FromContext(ctx).WarnContext(ctx, msg, args...).
func Warn(ctx context.Context, msg string, args ...any) {
	FromContext(ctx).WarnContext(ctx, msg, args...)
}

// Error logs an error message using the logger from context.
// This is a convenience function equivalent to FromContext(ctx).ErrorContext(ctx, msg, args...).
func Error(ctx context.Context, msg string, args ...any) {
	FromContext(ctx).ErrorContext(ctx, msg, args...)
}
