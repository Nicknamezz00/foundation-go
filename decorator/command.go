package decorator

import (
	"context"
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
)

func ApplyCommandDecorators[C, R any](handler CommandHandler[C, R], logger *logrus.Entry, metricsClient MetricsClient) CommandHandler[C, R] {
	return commandLoggingDecorator[C, R]{
		base: commandMetricsDecorator[C, R]{
			base:   handler,
			client: metricsClient,
		},
		logger: logger,
	}
}

type CommandHandler[C, R any] interface {
	Handle(ctx context.Context, cmd C) (R, error)
}

func generateActionName(handler any) string {
	return strings.Split(fmt.Sprintf("%T", handler), ".")[1]
}
