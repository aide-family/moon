package conn

import (
	"os"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	traceSdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

type ITraceConfig interface {
	GetEndpoint() string
}

type ITraceEnv interface {
	GetName() string
	GetVersion() string
	GetNamespace() string
}

// NewTracerProvider new tracer provider
//
//	@param conf ITraceConfig
func NewTracerProvider(conf ITraceConfig, traceEnv ITraceEnv) *traceSdk.TracerProvider {
	exp, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(conf.GetEndpoint())))
	if err != nil {
		panic(err)
	}

	id, _ := os.Hostname()

	tp := traceSdk.NewTracerProvider(
		traceSdk.WithBatcher(exp),
		traceSdk.WithResource(resource.NewSchemaless(
			semconv.ServiceNameKey.String(traceEnv.GetName()),
			semconv.ServiceVersionKey.String(traceEnv.GetVersion()),
			semconv.ServiceInstanceIDKey.String(id),
			semconv.ServiceNamespaceKey.String(traceEnv.GetNamespace()),
			attribute.String("exporter", "jaeger"),
		)),
		traceSdk.WithSampler(traceSdk.ParentBased(traceSdk.TraceIDRatioBased(1.0))),
	)
	otel.SetTracerProvider(tp)
	return tp
}
