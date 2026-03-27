package tracex

import (
	"context"

	"github.com/fr11nik/slogx"
)

type Option func(t *Telemetry) error

// WithSlogHandler позволяет передать уже инициализированный MultiHandler.
// OTel log bridge будет inject'нут в него, вместо создания нового логгера.
func WithSlogHandler(mh *slogx.MultiHandler) Option {
	return func(t *Telemetry) error {
		t.slogMultiHandler = mh
		return nil
	}
}

func WithLogConsoleExporter() Option {
	return func(t *Telemetry) error {
		var err error
		t.logExp, err = logConsoleExporter()
		return err
	}
}

func WithLogGrpcExporter(ctx context.Context, otlpEndpoint string) Option {
	return func(t *Telemetry) error {
		var err error
		t.logExp, err = logGrpcExporter(ctx, otlpEndpoint)
		return err
	}
}

func WithMetricConsoleExporter() Option {
	return func(t *Telemetry) error {
		var err error
		t.metricExp, err = metricConsoleExporter()
		return err
	}
}

func WithMetricGrpcExporter(ctx context.Context, otlpEndpoint string) Option {
	return func(t *Telemetry) error {
		var err error
		t.metricExp, err = metricGrpcExporter(ctx, otlpEndpoint)
		return err
	}
}

func WithSpanConsoleExporter() Option {
	return func(t *Telemetry) error {
		var err error
		t.traceExp, err = spanConsoleExporter()
		return err
	}
}

func WithSpanGrpcExporter(ctx context.Context, otlpEndpoint string) Option {
	return func(t *Telemetry) error {
		var err error
		t.traceExp, err = spanGrpcExporter(ctx, otlpEndpoint)
		return err
	}
}

func WithSpanHTTPExporter(ctx context.Context, otlpEndpoint string) Option {
	return func(t *Telemetry) error {
		var err error
		t.traceExp, err = spanHttpExporter(ctx, otlpEndpoint)
		return err
	}
}

func WithDefaultExporters() Option {
	return func(t *Telemetry) error {
		t.withDefaultExporters = true
		return nil
	}
}
