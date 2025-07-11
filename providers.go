package tracex

import (
	"context"
	"os"

	"go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	oteltrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

func newTracerProvider(
	ctx context.Context,
	exp oteltrace.SpanExporter,
	res *resource.Resource,
) *oteltrace.TracerProvider {
	tp := oteltrace.NewTracerProvider(
		oteltrace.WithBatcher(exp),
		oteltrace.WithResource(res),
	)
	return tp
}

func newLoggerProvider(
	ctx context.Context,
	exp log.Exporter,
	res *resource.Resource,
) (*log.LoggerProvider, error) {
	loggerProvider := log.NewLoggerProvider(
		log.WithProcessor(log.NewBatchProcessor(exp)),
		log.WithResource(res),
	)
	return loggerProvider, nil
}

// newMeterProvider creates a new meter provider with the OTLP gRPC exporter.
func newMeterProvider(
	ctx context.Context,
	exp metric.Exporter,
	res *resource.Resource,
) (*metric.MeterProvider, error) {
	mp := metric.NewMeterProvider(
		metric.WithReader(metric.NewPeriodicReader(exp)),
		metric.WithResource(res),
	)

	return mp, nil
}

// newResource creates a new OTEL resource with the service name and version.
func newResource(serviceName string, serviceVersion string) *resource.Resource {
	hostName, _ := os.Hostname()

	return resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceName(serviceName),
		semconv.ServiceVersion(serviceVersion),
		semconv.HostName(hostName),
	)
}
