package log

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"
)

func CommandExecution(ctx context.Context, commandName string, cmd any, err error) {
	l := logrus.WithContext(ctx).WithFields(logrus.Fields{
		"commandName": commandName,
		"cmd":         cmd,
	})
	if err == nil {
		l.Info("_command_succeeded")
	} else {
		l.WithError(err).Error("_command_failed")
	}
}

func QueryExecution(ctx context.Context, queryName string, q any) (logrus.Fields, func(*error)) {
	fields := logrus.Fields{
		"queryName": queryName,
		"query":     q,
	}
	start := time.Now()
	return fields, func(err *error) {
		level, msg := logrus.InfoLevel, "_query_success"
		fields[Cost] = time.Since(start)

		if err != nil && (*err != nil) {
			level, msg = logrus.ErrorLevel, "_query_failed"
			fields[Error] = (*err).Error()
		}

		logf(ctx, level, fields, "%s", msg)
	}
}
