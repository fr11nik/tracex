# tracex

Библиотека для быстрой инициализации OpenTelemetry в Go-сервисах. Объединяет трассировку, метрики и логирование в одном месте.

## Установка

```bash
go get github.com/fr11nik/tracex
```

## Быстрый старт

```go
ctx := context.Background()

tel, err := tracex.NewTelemetryProvider(ctx, "my-service", "1.0.0",
    tracex.WithDefaultExporters(),
)
if err != nil {
    log.Fatal(err)
}
defer tel.Shutdown(ctx)
```

После этого глобальные провайдеры OTel (`otel.GetTracerProvider()`, `otel.GetMeterProvider()`, `global.GetLoggerProvider()`) зарегистрированы и готовы к использованию.

## Конструкторы

### `NewTelemetryProvider` (рекомендуется)

```go
tel, err := tracex.NewTelemetryProvider(ctx, serviceName, version, opts...)
```

Требует хотя бы один экспортер или опцию `WithDefaultExporters`. Если ни один не задан — возвращает `ErrAtLeastOneExporter`.

### `NewTelemetry` (устарел)

```go
tel, err := tracex.NewTelemetry(ctx, serviceName, version, opts...)
```

Автоматически включает консольные экспортеры, если ни один не задан. Оставлен для обратной совместимости.

## Опции

### Экспортеры трассировки

| Опция | Описание |
|---|---|
| `WithSpanConsoleExporter()` | Вывод спанов в stdout |
| `WithSpanGrpcExporter(ctx, endpoint)` | Отправка спанов по gRPC (OTLP) |
| `WithSpanHTTPExporter(ctx, endpoint)` | Отправка спанов по HTTP (OTLP) |

### Экспортеры метрик

| Опция | Описание |
|---|---|
| `WithMetricConsoleExporter()` | Вывод метрик в stdout |
| `WithMetricGrpcExporter(ctx, endpoint)` | Отправка метрик по gRPC (OTLP) |

### Экспортеры логов

| Опция | Описание |
|---|---|
| `WithLogConsoleExporter()` | Вывод логов в stdout |
| `WithLogGrpcExporter(ctx, endpoint)` | Отправка логов по gRPC (OTLP) |

### Прочее

| Опция | Описание |
|---|---|
| `WithDefaultExporters()` | Включает консольные экспортеры для всех сигналов |
| `WithSlogHandler(mh)` | Добавляет OTel bridge в существующий `slogx.MultiHandler` |

## Примеры

### Отправка в коллектор по gRPC

```go
endpoint := "localhost:4317"

tel, err := tracex.NewTelemetryProvider(ctx, "my-service", "1.0.0",
    tracex.WithSpanGrpcExporter(ctx, endpoint),
    tracex.WithMetricGrpcExporter(ctx, endpoint),
    tracex.WithLogGrpcExporter(ctx, endpoint),
)
if err != nil {
    log.Fatal(err)
}
defer tel.Shutdown(ctx)
```

### Интеграция с существующим slog-логгером

Если у вас уже есть `slogx.MultiHandler`, OTel bridge добавится в него — отдельный логгер создаваться не будет:

```go
handler := slogx.NewMultiHandler(jsonHandler)

tel, err := tracex.NewTelemetryProvider(ctx, "my-service", "1.0.0",
    tracex.WithSlogHandler(handler),
    tracex.WithLogGrpcExporter(ctx, "localhost:4317"),
    tracex.WithSpanGrpcExporter(ctx, "localhost:4317"),
)
```

### Использование трассировки и метрик

После инициализации работайте с OTel напрямую через глобальные провайдеры:

```go
tracer := otel.Tracer("my-service")
ctx, span := tracer.Start(ctx, "operation-name")
defer span.End()

meter := otel.Meter("my-service")
counter, _ := meter.Int64Counter("requests.total")
counter.Add(ctx, 1)
```

### Корректное завершение

```go
tel, err := tracex.NewTelemetryProvider(...)
if err != nil {
    log.Fatal(err)
}

// Обязательно вызывайте Shutdown — это гарантирует отправку буферизованных данных
defer tel.Shutdown(ctx)
```

## Архитектура

```
NewTelemetryProvider
├── setupLogging   → LoggerProvider → global.SetLoggerProvider + slog bridge
├── setupMetrics   → MeterProvider  → otel.SetMeterProvider
└── setupTracing   → TracerProvider → otel.SetTracerProvider + propagation
```

Resource автоматически включает имя сервиса, версию и hostname.

gRPC-экспортеры работают без TLS (`WithInsecure`) и без повторных попыток — настройте это на стороне коллектора или оберните собственным экспортером.
