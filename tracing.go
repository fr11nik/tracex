package tracex

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/fr11nik/slogx"
	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/log/global"
	otelmetric "go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/trace"
	oteltrace "go.opentelemetry.io/otel/trace"
)

type Telemetry struct {
	lp     *log.LoggerProvider
	mp     *metric.MeterProvider
	tp     *trace.TracerProvider
	meter  otelmetric.Meter
	tracer oteltrace.Tracer
	exporters
}

type exporters struct {
	logExp    log.Exporter
	metricExp metric.Exporter
	traceExp  trace.SpanExporter
}

// NewTelemetry creates a new telemetry instance.
func NewTelemetry(
	ctx context.Context,
	serviceName, version string,
	opts ...Option,
) (*Telemetry, error) {
	rp := newResource(serviceName, version)
	telemetry := &Telemetry{}
	for _, opt := range opts {
		err := opt(telemetry)
		if err != nil {
			return nil, err
		}
	}

	if telemetry.exporters.logExp == nil {
		telemetry.exporters.logExp, _ = logConsoleExporter()
	}
	if telemetry.exporters.traceExp == nil {
		telemetry.exporters.traceExp, _ = spanConsoleExporter()
	}
	if telemetry.exporters.metricExp == nil {
		telemetry.exporters.metricExp, _ = metricConsoleExporter()
	}

	lp, err := newLoggerProvider(ctx, telemetry.exporters.logExp, rp)
	if err != nil {
		return nil, fmt.Errorf("failed to create logger: %w", err)
	}

	global.SetLoggerProvider(lp)
	slogx.InitLogging(os.Stdout, slogx.WithOtelSlogOption(
		serviceName, otelslog.WithLoggerProvider(lp),
	))

	mp, err := newMeterProvider(ctx, telemetry.exporters.metricExp, rp)
	if err != nil {
		return nil, fmt.Errorf("failed to create meter: %w", err)
	}
	otel.SetMeterProvider(mp)
	meter := mp.Meter(serviceName)

	tp := newTracerProvider(ctx, telemetry.exporters.traceExp, rp)
	if err != nil {
		return nil, fmt.Errorf("failed to create tracer: %w", err)
	}
	otel.SetTracerProvider(tp)
	tracer := tp.Tracer(serviceName)

	tc := propagation.TraceContext{}
	// Register the TraceContext propagator globally.
	otel.SetTextMapPropagator(tc)

	otel.SetErrorHandler(&localErrorHandler{ctx})
	return &Telemetry{
		lp:     lp,
		mp:     mp,
		tp:     tp,
		meter:  meter,
		tracer: tracer,
	}, nil
}

// Shutdown shuts down the logger, meter, and tracer.
func (t *Telemetry) Shutdown(ctx context.Context) {
	t.lp.Shutdown(ctx)
	t.mp.Shutdown(ctx)
	t.tp.Shutdown(ctx)
}

type localErrorHandler struct {
	ctx context.Context
}

func (ler *localErrorHandler) Handle(err error) {
	slog.ErrorContext(ler.ctx, err.Error(), "handler", "Otel.Tracer")
}
