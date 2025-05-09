package log

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"
)

type Fields = logrus.Fields

func WhenRequest(ctx context.Context, action string, url string) (Fields, func(*error)) {
	fields, startTime := logrus.Fields{
		URL:           url,
		RequestStatus: Tags.RequestStatus.Success,
	}, time.Now()

	return fields, func(err *error) {
		level, info := logrus.InfoLevel, action+"_success"
		fields[Cost] = time.Since(startTime)

		if err != nil && (*err != nil) {
			fields[Error] = (*err).Error()
			level, info = logrus.ErrorLevel, action+"_failure"
		}

		logf(ctx, level, fields, "%s", info)
	}
}
