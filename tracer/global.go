package tracer

import (
	"go.opentelemetry.io/otel"
)

var (
	tracer = otel.Tracer("default_tracer")
)
