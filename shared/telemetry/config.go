// Package telemetry provides shared OpenTelemetry configuration utilities.
package telemetry

import (
	"fmt"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

// Canonical naming convention:
//
//   - Spans: "service.operation" (e.g., "worker.process_reminder", "processing.run_ocr")
//   - Metrics:
//   - job.duration seconds (histogram)
//   - job.success (counter with labels: service, job_type)
//   - job.failure (counter with labels: service, job_type, error_type)
//   - queue.depth (gauge with labels: queue_name)
//
// Config holds shared telemetry configuration.
type Config struct {
	ServiceName    string
	ServiceVersion string
	Environment    string
	Endpoint       string
}

// InitResult holds the result of telemetry initialization.
type InitResult struct {
	TracerProvider *sdktrace.TracerProvider
	Tracer         trace.Tracer
}

// InitTrace initializes OpenTelemetry tracing with OTLP exporter.
func InitTrace(ctx context.Context, cfg Config) (*InitResult, error) {
	if cfg.Endpoint == "" {
		return &InitResult{}, nil
	}

	// Create OTLP trace exporter
	exporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpoint(cfg.Endpoint),
		otlptracegrpc.WithInsecure(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create trace exporter: %w", err)
	}

	// Create resource
	res, err := resource.New(ctx,
		resource.WithAttributes(
			attribute.String("service.name", cfg.ServiceName),
			attribute.String("service.version", cfg.ServiceVersion),
			attribute.String("environment", cfg.Environment),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Create tracer provider
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)

	return &InitResult{
		TracerProvider: tp,
		Tracer:         tp.Tracer(cfg.ServiceName),
	}, nil
}

// InitMetricExporter initializes an OTLP metric exporter.
func InitMetricExporter(ctx context.Context, endpoint string) (*otlpmetricgrpc.Exporter, error) {
	if endpoint == "" {
		return nil, nil
	}

	exporter, err := otlpmetricgrpc.New(ctx,
		otlpmetricgrpc.WithEndpoint(endpoint),
		otlpmetricgrpc.WithInsecure(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create metric exporter: %w", err)
	}

	return exporter, nil
}

// Propagator returns a text map propagator for trace context propagation.
func Propagator() propagation.TextMapPropagator {
	return propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
}

// SpanOptions holds options for starting a span.
type SpanOptions struct {
	Name       string
	Kind       trace.SpanKind
	Attributes []attribute.KeyValue
}

// StartSpan starts a new span with the given tracer.
func StartSpan(ctx context.Context, tracer trace.Tracer, opts SpanOptions) (context.Context, trace.Span) {
	if tracer == nil {
		return ctx, trace.SpanFromContext(ctx)
	}

	return tracer.Start(ctx, opts.Name,
		trace.WithSpanKind(opts.Kind),
		trace.WithAttributes(opts.Attributes...),
	)
}

// Common span attribute keys.
const (
	AttrServiceName    = attribute.Key("service.name")
	AttrServiceVersion = attribute.Key("service.version")
	AttrEnvironment    = attribute.Key("environment")
	AttrTenantID       = attribute.Key("tenant.id")
	AttrOrgID          = attribute.Key("org.id")
	AttrUserID         = attribute.Key("user.id")
	AttrDocumentID     = attribute.Key("document.id")
	AttrJobID          = attribute.Key("job.id")
	AttrJobType        = attribute.Key("job.type")
	AttrHTTPMethod     = attribute.Key("http.method")
	AttrHTTPURL        = attribute.Key("http.url")
	AttrHTTPRoute      = attribute.Key("http.route")
	AttrHTTPStatusCode = attribute.Key("http.status_code")
	AttrErrorType      = attribute.Key("error.type")
	AttrErrorMessage   = attribute.Key("error.message")
)

// SpanKind aliases for convenience.
type SpanKind = trace.SpanKind

const (
	SpanKindServer   = trace.SpanKindServer
	SpanKindClient   = trace.SpanKindClient
	SpanKindProducer = trace.SpanKindProducer
	SpanKindConsumer = trace.SpanKindConsumer
	SpanKindInternal = trace.SpanKindInternal
)

// RecordError records an error on a span.
func RecordError(span trace.Span, err error, attrs ...attribute.KeyValue) {
	if err == nil || !span.IsRecording() {
		return
	}

	allAttrs := append([]attribute.KeyValue{
		AttrErrorType(err),
	}, attrs...)
	span.RecordError(err, trace.WithAttributes(allAttrs...))
	span.SetStatus(trace.Status{
		Code:    trace.StatusCodeError,
		Message: err.Error(),
	})
}

// DurationMs returns duration in milliseconds as a float64 attribute value.
func DurationMs(d time.Duration) float64 {
	return float64(d.Milliseconds())
}
