package tracer

import (
	"context"
	"net/http"
	"runtime"

	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"

	"go.opentelemetry.io/contrib/propagators/b3"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
)

func InitJaegerProvider(serviceName string, jaegerURL string) (func(ctx context.Context) error, error) {
	if jaegerURL == "" {
		panic("empty jaeger url")
	}
	tracer = otel.Tracer(serviceName)
	exp, err := otlptracehttp.New(context.Background(), otlptracehttp.WithEndpoint(jaegerURL))
	if err != nil {
		return nil, err
	}
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(resource.NewSchemaless(
			semconv.ServiceNameKey.String(serviceName),
		)),
	)
	otel.SetTracerProvider(tp)
	b3Propagator := b3.New(b3.WithInjectEncoding(b3.B3MultipleHeader))
	p := propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{}, propagation.Baggage{}, b3Propagator,
	)
	otel.SetTextMapPropagator(p)
	return tp.Shutdown, nil
}

func StartWithFuncName(ctx context.Context) (context.Context, trace.Span) {
	spanName := "unknown"
	pc, _, _, ok := runtime.Caller(1)
	if ok {
		callerFuncName := runtime.FuncForPC(pc).Name()
		if callerFuncName != "" {
			spanName = callerFuncName
		}
	}
	return Start(ctx, spanName)
}

func Start(ctx context.Context, name string) (context.Context, trace.Span) {
	return tracer.Start(ctx, name)
}

func Inject(ctx context.Context, req *http.Request) {
	propagator := otel.GetTextMapPropagator()
	propagator.Inject(ctx, propagation.HeaderCarrier(req.Header))
}

func TraceID(ctx context.Context) string {
	spanCtx := trace.SpanContextFromContext(ctx)
	return spanCtx.TraceID().String()
}
