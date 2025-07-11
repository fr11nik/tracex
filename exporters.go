package tracex

import (
	"context"

	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutlog"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/metric"
	oteltrace "go.opentelemetry.io/otel/sdk/trace"
)

// Span exportters
func spanConsoleExporter() (oteltrace.SpanExporter, error) {
	return stdouttrace.New()
}

func spanGrpcExporter(ctx context.Context, otlpEndpoint string) (oteltrace.SpanExporter, error) {
	includetlsOpt := otlptracegrpc.WithInsecure()
	retryOpt := otlptracegrpc.WithRetry(otlptracegrpc.RetryConfig{Enabled: false})
	endpointOpt := otlptracegrpc.WithEndpoint(otlpEndpoint)
	return otlptracegrpc.New(ctx, includetlsOpt, endpointOpt, retryOpt)
}

func spanHttpExporter(ctx context.Context, otlpEndpoint string) (oteltrace.SpanExporter, error) {
	insecureopt := otlptracehttp.WithInsecure()
	endpointOpt := otlptracehttp.WithEndpoint(otlpEndpoint)
	return otlptracehttp.New(ctx, insecureopt, endpointOpt)
}

// Log exporter
func logGrpcExporter(ctx context.Context, otlpEndpoint string) (log.Exporter, error) {
	includetlsOpt := otlploggrpc.WithInsecure()
	endpointOpt := otlploggrpc.WithEndpoint(otlpEndpoint)
	retryOpt := otlploggrpc.WithRetry(otlploggrpc.RetryConfig{Enabled: false})
	return otlploggrpc.New(ctx, includetlsOpt, endpointOpt, retryOpt)
}

func logConsoleExporter() (log.Exporter, error) {
	return stdoutlog.New()
}

// Metric exporters
func metricGrpcExporter(ctx context.Context, otlpEndpoint string) (metric.Exporter, error) {
	includetlsOpt := otlpmetricgrpc.WithInsecure()
	endpointOpt := otlpmetricgrpc.WithEndpoint(otlpEndpoint)
	retryOpt := otlpmetricgrpc.WithRetry(otlpmetricgrpc.RetryConfig{Enabled: false})
	return otlpmetricgrpc.New(ctx, includetlsOpt, endpointOpt, retryOpt)
}

func metricConsoleExporter() (metric.Exporter, error) {
	return stdoutmetric.New()
}
