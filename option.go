package tracex

import "context"

type Option func(t *Telemetry) error

func WithLogConsoleExporter() Option {
	return func(t *Telemetry) error {
		var err error
		t.exporters.logExp, err = logConsoleExporter()
		return err
	}
}

func WithLogGrpcExporter(ctx context.Context, otlpEndpoint string) Option {
	return func(t *Telemetry) error {
		var err error
		t.exporters.logExp, err = logGrpcExporter(ctx, otlpEndpoint)
		return err
	}
}

func WithMetricConsoleExporter() Option {
	return func(t *Telemetry) error {
		var err error
		t.exporters.metricExp, err = metricConsoleExporter()
		return err
	}
}

func WithMetricGrpcExporter(ctx context.Context, otlpEndpoint string) Option {
	return func(t *Telemetry) error {
		var err error
		t.exporters.metricExp, err = metricGrpcExporter(ctx, otlpEndpoint)
		return err
	}
}

func WithSpanConsoleExporter() Option {
	return func(t *Telemetry) error {
		var err error
		t.exporters.traceExp, err = spanConsoleExporter()
		return err
	}
}

func WithSpanGrpcExporter(ctx context.Context, otlpEndpoint string) Option {
	return func(t *Telemetry) error {
		var err error
		t.exporters.traceExp, err = spanGrpcExporter(ctx, otlpEndpoint)
		return err
	}
}

func WithSpanHttpExporter(ctx context.Context, otlpEndpoint string) Option {
	return func(t *Telemetry) error {
		var err error
		t.exporters.traceExp, err = spanHttpExporter(ctx, otlpEndpoint)
		return err
	}
}
