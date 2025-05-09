package log

import (
	"bytes"
	"fmt"
	"foundation-go/config"
	"foundation-go/tracer"
	"foundation-go/utility/envutil"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/sirupsen/logrus"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
)

const (
	Service       = "service_name"
	Host          = "host"
	Port          = "port"
	Errno         = "errno"
	Method        = "method"
	Args          = "args"
	Cost          = "cost_ms"
	Request       = "request"
	Response      = "response"
	RawBody       = "raw_body"
	RequestStatus = "request_status"
	URL           = "url"
	URI           = "uri"
	Error         = "error"
	Errmsg        = "errmsg"
	StatusCode    = "status_code"
)

const (
	HTTP      = "_http"
	WechatSDK = "_wechat_sdk"
	LarkSDK   = "_lark_sdk"
)

type Factory struct {
	Level  string `mapstructure:"level"`
	Output string `mapstructure:"output"`
}

var factory *Factory

func Init() {
	initFactory()
	initLog()
}

func initFactory() {
	if err := config.Sub("log").Unmarshal(&factory); err != nil {
		panic(err)
	}
}

func initLog() {
	_, err := os.Stat(factory.Output)
	if os.IsNotExist(err) {
		_ = os.MkdirAll(factory.Output, 0755)
	}
	infoWriter, err := rotatelogs.New(
		filepath.Join(factory.Output, "info.log.%Y%m%d%H"),
		rotatelogs.WithLinkName(filepath.Join(factory.Output, "info.log")),
		rotatelogs.WithRotationTime(time.Hour), // 每小时切割一次
		rotatelogs.WithMaxAge(7*24*time.Hour),  // 保存7天
	)
	if err != nil {
		logrus.Fatalf("failed to create info log file: %v", err)
	}

	warnWriter, err := rotatelogs.New(
		filepath.Join(factory.Output, "warn.log.%Y%m%d%H"),
		rotatelogs.WithLinkName(filepath.Join(factory.Output, "warn.log")),
		rotatelogs.WithRotationTime(time.Hour),
		rotatelogs.WithMaxAge(7*24*time.Hour),
	)
	if err != nil {
		logrus.Fatalf("failed to create warn log file: %v", err)
	}

	SetLevel()
	logrus.SetOutput(io.Discard) // 防止默认双打
	// hook
	logrus.AddHook(&traceHook{})
	addOutputHook(infoWriter, warnWriter)
}

func SetLevel() {
	level, err := logrus.ParseLevel(config.GetString("log.level"))
	if err != nil {
		panic(err)
	}
	logrus.SetLevel(level)
	//logrus.SetReportCaller(true)
}

func addOutputHook(infoWriter, warnWriter *rotatelogs.RotateLogs) {
	logrus.AddHook(&writerHook{
		WriterMap: map[logrus.Level]io.Writer{
			logrus.InfoLevel:  infoWriter,
			logrus.DebugLevel: infoWriter,
			logrus.WarnLevel:  warnWriter,
			logrus.ErrorLevel: warnWriter,
			logrus.FatalLevel: warnWriter,
			logrus.PanicLevel: warnWriter,
		},
		FormatterMap: map[logrus.Level]logrus.Formatter{
			logrus.InfoLevel:  getFormatter(),
			logrus.DebugLevel: getFormatter(),
			logrus.WarnLevel:  getFormatter(),
			logrus.ErrorLevel: getFormatter(),
			logrus.FatalLevel: getFormatter(),
			logrus.PanicLevel: getFormatter(),
		},
	})

	// 本地添加console
	if envutil.IsDev() {
		logrus.AddHook(&consoleHook{
			Writer: os.Stdout,
			Formatter: &prefixed.TextFormatter{
				ForceColors:     true,
				ForceFormatting: true,
				FullTimestamp:   true,
				TimestampFormat: time.DateTime,
			},
		})
	}
}

type OrderedTextFormatter struct{}

func (f *OrderedTextFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	var b bytes.Buffer
	// 固定字段顺序：时间、等级、消息
	timestamp := entry.Time.Format(time.RFC3339Nano)
	level := entry.Level.String()
	msg := entry.Message

	b.WriteString(fmt.Sprintf("[%s][%s] %s ", strings.ToUpper(level), timestamp, msg))

	// 固定附加字段顺序（可以手动指定）
	orderedFields := []string{Host, Port, Service, Method, Errno, Cost, Args, URI, URL, StatusCode, RawBody, Response, Error, Errmsg}
	var items []string
	for _, key := range orderedFields {
		if val, ok := entry.Data[key]; ok {
			items = append(items, fmt.Sprintf("%s=%v", key, val))
		}
	}

	b.WriteString(strings.Join(items, "||"))

	// 输出剩余未指定的字段
	for key, val := range entry.Data {
		skip := false
		for _, f := range orderedFields {
			if f == key {
				skip = true
				break
			}
		}
		if !skip {
			b.WriteString(fmt.Sprintf("||%s=%v", key, val))
		}
	}

	b.WriteByte('\n')
	return b.Bytes(), nil
}

type consoleHook struct {
	Writer    io.Writer
	Formatter logrus.Formatter
}

func (hook *consoleHook) Fire(entry *logrus.Entry) error {
	msg, err := hook.Formatter.Format(entry)
	if err != nil {
		return err
	}
	_, err = hook.Writer.Write(msg)
	return err
}

func (hook *consoleHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

type traceHook struct{}

func (t traceHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (t traceHook) Fire(entry *logrus.Entry) error {
	if entry.Context != nil {
		entry.Data["trace_id"] = tracer.TraceID(entry.Context)
		entry = entry.WithTime(time.Now()) //nolint
	}
	return nil
}

type writerHook struct {
	WriterMap    map[logrus.Level]io.Writer
	FormatterMap map[logrus.Level]logrus.Formatter
}

func (hook *writerHook) Fire(entry *logrus.Entry) error {
	writer, ok := hook.WriterMap[entry.Level]
	if !ok {
		return nil
	}
	formatter, ok := hook.FormatterMap[entry.Level]
	if !ok {
		formatter = getFormatter()
	}
	msg, err := formatter.Format(entry)
	if err != nil {
		return err
	}
	_, err = writer.Write(msg)
	return err
}

func (hook *writerHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func getFormatter() logrus.Formatter {
	return &OrderedTextFormatter{}
	//return &prefixed.TextFormatter{
	//	FullTimestamp:   true,
	//	TimestampFormat: time.RFC3339,
	//}
}
