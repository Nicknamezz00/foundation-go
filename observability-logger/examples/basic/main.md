```go
package main

import (
	"context"
)

func main() {
	// Example 1: Simple logger at INFO level
	log := logger.NewInfo("example-service")
	log.Info("application started")
	log.Debug("this won't be logged") // below INFO level

	// Example 2: Logger with custom configuration
	debugLog := logger.New("example-service", logger.Config{
		Level:  logger.LevelDebug,
		Format: "text", // human-readable format
	})
	debugLog.Debug("debug message", "key", "value")
	debugLog.Info("info message", "user_id", 123)

	// Example 3: Logger with source code location
	sourceLog := logger.New("example-service", logger.Config{
		Level:     logger.LevelInfo,
		Format:    "json",
		AddSource: true, // Enable source location (file:line)
	})
	sourceLog.Info("log with source location", "feature", "source_tracking")
	sourceLog.Warn("warning with location", "code", 1001)

	// Example 4: Context-based logging
	ctx := logger.WithContext(context.Background(), log)
	processRequest(ctx)

	// Example 5: Convenience functions
	logger.Info(ctx, "using convenience function", "status", "success")
	logger.Warn(ctx, "warning message", "code", 1001)
}

func processRequest(ctx context.Context) {
	// Retrieve logger from context
	log := logger.FromContext(ctx)
	log.Info("processing request", "step", "validation")
	log.Info("processing request", "step", "execution")
	log.Info("request completed", "duration_ms", 42)
}
```
