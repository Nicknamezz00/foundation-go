package decorator

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"
)

type commandLoggingDecorator[C, R any] struct {
	base   CommandHandler[C, R]
	logger *logrus.Entry
}

func (d commandLoggingDecorator[C, R]) Handle(ctx context.Context, cmd C) (result R, err error) {
	handlerType := generateActionName(cmd)
	logger := d.logger.WithContext(ctx).WithFields(logrus.Fields{
		"command":      handlerType,
		"command_body": fmt.Sprintf("%+v", cmd),
	})

	defer func() {
		if err == nil {
			logger.Info("Command executed successfully")
		} else {
			logger.WithError(err).Error("Failed to execute command")
		}
	}()

	return d.base.Handle(ctx, cmd)
}

type queryLoggingDecorator[C any, R any] struct {
	base   QueryHandler[C, R]
	logger *logrus.Entry
}

func (d queryLoggingDecorator[C, R]) Handle(ctx context.Context, cmd C) (result R, err error) {
	logger := d.logger.WithFields(logrus.Fields{
		"query":      generateActionName(cmd),
		"query_body": fmt.Sprintf("%+v", cmd),
	})

	defer func() {
		if err == nil {
			logger.WithField("res", result).Info("Query executed successfully")
		} else {
			logger.WithError(err).Error("Failed to execute query")
		}
	}()

	return d.base.Handle(ctx, cmd)
}
