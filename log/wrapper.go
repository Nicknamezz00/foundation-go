package log

import (
	"context"

	"github.com/sirupsen/logrus"
)

// tags
var Tags = struct {
	MQ            MQ
	RequestStatus requestStatus
}{
	MQ: MQ{
		MQPublishError:   mqPublishError,
		MQPublishSuccess: mqPublishSuccess,
		MQHandleSuccess:  mqHandleSuccess,
		MQHandleError:    mqHandleError,
		MQConsumeError:   mqConsumeError,
		MQConsumeSuccess: mqConsumeSuccess,
	},
	RequestStatus: requestStatus{
		Success:       "rpc_success",        // 成功,
		NetError:      "rpc_response_error", // 网络错误
		ClientError:   "rpc_response_error",
		RequestError:  "rpc_response_error",
		ResponseError: "rpc_response_error", // 响应正常，返回值不合法,
	},
}

type requestStatus struct {
	Success       string
	NetError      string
	ClientError   string
	RequestError  string
	ResponseError string
}

type MQ struct {
	MQPublishError   string
	MQPublishSuccess string
	MQHandleSuccess  string
	MQHandleError    string
	MQConsumeError   string
	MQConsumeSuccess string
}

const (
	RPCStatus = "_rpc_status"

	mqPublishError   = "_mq_publish_error"
	mqPublishSuccess = "_mq_publish_success"
	mqHandleSuccess  = "_mq_handle_success"
	mqHandleError    = "_mq_handle_error"
	mqConsumeError   = "_mq_consume_error"
	mqConsumeSuccess = "_mq_consume_success"
)

const (
	RPCSuccess       = "rpc_success"        // RPC成功
	RPCError         = "rpc_error"          // RPC异常
	RPCClientError   = "rpc_client_error"   // RPC client构造异常
	RPCRequestError  = "rpc_request_error"  // RPC请求异常
	RPCResponseError = "rpc_response_error" // RPC正常，返回值错误
)

func logf(ctx context.Context, level logrus.Level, fields logrus.Fields, format string, args ...any) {
	logger := logrus.WithContext(ctx)
	if len(fields) > 0 {
		logger = logger.WithFields(fields)
	}
	logger.Logf(level, format, args...)
}

func Infof(ctx context.Context, fields Fields, format string, args ...any) {
	logf(ctx, logrus.InfoLevel, fields, format, args...)
}

func Panic(ctx context.Context, fields Fields, args ...any) {
	logf(ctx, logrus.PanicLevel, fields, "%s", args...)
}

func Panicf(ctx context.Context, fields Fields, format string, args ...any) {
	logf(ctx, logrus.PanicLevel, fields, format, args...)
}

func Fatalf(ctx context.Context, fields Fields, format string, args ...any) {
	logf(ctx, logrus.FatalLevel, fields, format, args...)
}

func Warnf(ctx context.Context, fields Fields, format string, args ...any) {
	logf(ctx, logrus.WarnLevel, fields, format, args...)
}

func Errorf(ctx context.Context, fields Fields, format string, args ...any) {
	logf(ctx, logrus.ErrorLevel, fields, format, args...)
}
