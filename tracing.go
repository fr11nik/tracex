package tracex

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/fr11nik/slogx"
	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/log/global"
	otelmetric "go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	oteltrace "go.opentelemetry.io/otel/trace"
)

type Telemetry struct {
	lp                   *log.LoggerProvider
	mp                   *metric.MeterProvider
	mpScrapeInterval     time.Duration
	tp                   *trace.TracerProvider
	meter                otelmetric.Meter
	tracer               oteltrace.Tracer
	withDefaultExporters bool
	exporters

	// slogMultiHandler — если передан, OTel bridge inject'ится в него.
	slogMultiHandler *slogx.MultiHandler
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
	t := &Telemetry{}
	for _, opt := range opts {
		if err := opt(t); err != nil {
			return nil, err
		}
	}

	if t.withDefaultExporters {
		t.logExp, _ = logConsoleExporter()
		t.metricExp, _ = metricConsoleExporter()
		t.traceExp, _ = spanConsoleExporter()
	}

	res := newResource(serviceName, version)

	if err := t.setupLogging(ctx, serviceName, res); err != nil {
		return nil, err
	}
	if err := t.setupMetrics(ctx, serviceName, res); err != nil {
		return nil, err
	}
	if err := t.setupTracing(ctx, serviceName, res); err != nil {
		return nil, err
	}

	return t, nil
}

func (t *Telemetry) setupLogging(ctx context.Context, serviceName string, res *resource.Resource) error {
	if t.logExp == nil {
		return nil
	}

	lp, err := newLoggerProvider(ctx, t.logExp, res)
	if err != nil {
		return fmt.Errorf("failed to create logger provider: %w", err)
	}
	global.SetLoggerProvider(lp)
	t.lp = lp

	otelHandler := otelslog.NewHandler(serviceName, otelslog.WithLoggerProvider(lp))
	if t.slogMultiHandler != nil {
		t.slogMultiHandler.AddHandler(otelHandler)
	} else {
		slogx.InitLoggingJSON(nil, slogx.WithRawHandler(otelHandler))
	}
	return nil
}

func (t *Telemetry) setupMetrics(ctx context.Context, serviceName string, res *resource.Resource) error {
	if t.metricExp == nil {
		return nil
	}

	mp, err := newMeterProvider(ctx, t.metricExp, res, t.mpScrapeInterval)
	if err != nil {
		return fmt.Errorf("failed to create meter provider: %w", err)
	}
	otel.SetMeterProvider(mp)
	t.mp = mp
	t.meter = mp.Meter(serviceName)
	return nil
}

func (t *Telemetry) setupTracing(ctx context.Context, serviceName string, res *resource.Resource) error {
	if t.traceExp == nil {
		return nil
	}

	tp := newTracerProvider(ctx, t.traceExp, res)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.TraceContext{})
	otel.SetErrorHandler(&localErrorHandler{ctx})
	t.tp = tp
	t.tracer = tp.Tracer(serviceName)
	return nil
}

// Shutdown shuts down the logger, meter, and tracer.
func (t *Telemetry) Shutdown(ctx context.Context) {
	if t.lp != nil {
		t.lp.Shutdown(ctx)
	}
	if t.mp != nil {
		t.mp.Shutdown(ctx)
	}
	if t.tp != nil {
		t.tp.Shutdown(ctx)
	}
}

type localErrorHandler struct {
	ctx context.Context
}

func (ler *localErrorHandler) Handle(err error) {
	slog.ErrorContext(ler.ctx, err.Error(), "handler", "Otel.Tracer")
}
