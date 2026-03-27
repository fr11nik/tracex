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
	"go.opentelemetry.io/otel/sdk/trace"
	oteltrace "go.opentelemetry.io/otel/trace"
)

const (
	defaultScrapeInterval = time.Second * 5
)

type Telemetry struct {
	lp               *log.LoggerProvider
	mp               *metric.MeterProvider
	mpScrapeInterval time.Duration
	tp               *trace.TracerProvider
	meter            otelmetric.Meter
	tracer           oteltrace.Tracer
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
	rp := newResource(serviceName, version)

	telemetry := &Telemetry{}
	for _, opt := range opts {
		err := opt(telemetry)
		if err != nil {
			return nil, err
		}
	}

	if telemetry.logExp == nil {
		telemetry.logExp, _ = logConsoleExporter()
	}
	if telemetry.traceExp == nil {
		telemetry.traceExp, _ = spanConsoleExporter()
	}
	if telemetry.metricExp == nil {
		telemetry.metricExp, _ = metricConsoleExporter()
	}
	if telemetry.mpScrapeInterval == time.Duration(0) {
		telemetry.mpScrapeInterval = defaultScrapeInterval
	}

	lp, err := newLoggerProvider(ctx, telemetry.logExp, rp)
	if err != nil {
		return nil, fmt.Errorf("failed to create logger: %w", err)
	}
	global.SetLoggerProvider(lp)

	// Inject OTel slog bridge
	otelHandler := otelslog.NewHandler(serviceName, otelslog.WithLoggerProvider(lp))

	if telemetry.slogMultiHandler != nil {
		// Inject в уже существующий логгер
		telemetry.slogMultiHandler.AddHandler(otelHandler)
	} else {
		// Fallback: создаём логгер с JSON + OTel (старое поведение)
		slogx.InitLoggingJSON(nil, slogx.WithRawHandler(otelHandler))
	}

	mp, err := newMeterProvider(ctx, telemetry.metricExp, rp, telemetry.mpScrapeInterval)
	if err != nil {
		return nil, fmt.Errorf("failed to create meter: %w", err)
	}
	otel.SetMeterProvider(mp)
	meter := mp.Meter(serviceName)

	tp := newTracerProvider(ctx, telemetry.traceExp, rp)
	if err != nil {
		return nil, fmt.Errorf("failed to create tracer: %w", err)
	}
	otel.SetTracerProvider(tp)
	tracer := tp.Tracer(serviceName)

	tc := propagation.TraceContext{}
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
