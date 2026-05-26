// Package telemetry provides OpenTelemetry instrumentation for the backend service.
package telemetry

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	otelpkg "go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"

	appconfig "github.com/docvault/backend/internal/config"
)

// Provider holds OpenTelemetry providers and instruments.
type Provider struct {
	TracerProvider *sdktrace.TracerProvider
	MeterProvider  *sdkmetric.MeterProvider
	Tracer         trace.Tracer
	Meter          otelpkg.Meter

	// Metrics
	RequestCount    otelpkg.Int64Counter
	RequestDuration otelpkg.Float64Histogram
	ErrorCount      otelpkg.Int64Counter
	ActiveRequests  otelpkg.Int64UpDownCounter
}

var globalProvider *Provider

// Init initializes OpenTelemetry with OTLP exporters.
func Init(ctx context.Context, cfg *appconfig.Config, serviceName string) (*Provider, error) {
	if cfg.OTELEndpoint == "" {
		slog.Warn("OTEL_ENDPOINT not configured, skipping telemetry initialization")
		return &Provider{}, nil
	}

	// Create OTLP trace exporter
	traceExporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpoint(cfg.OTELEndpoint),
		otlptracegrpc.WithInsecure(), // Use OTEL_EXPORTER_ENDPOINT with insecure for development
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create trace exporter: %w", err)
	}

	// Create OTLP metric exporter
	metricExporter, err := otlpmetricgrpc.New(ctx,
		otlpmetricgrpc.WithEndpoint(cfg.OTELEndpoint),
		otlpmetricgrpc.WithInsecure(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create metric exporter: %w", err)
	}

	// Create resource with service information
	res, err := resource.New(ctx,
		resource.WithAttributes(
			attribute.String("service.name", serviceName),
			attribute.String("service.version", "1.0.0"),
			attribute.String("environment", cfg.Environment),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Create tracer provider
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(traceExporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)

	// Create meter provider
	mp := sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(res),
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(metricExporter)),
	)

	otel.SetTracerProvider(tp)
	otel.SetMeterProvider(mp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	// Create tracer and meter
	tracer := tp.Tracer(serviceName)
	meter := mp.Meter(serviceName)

	// Create metrics instruments
	requestCount, err := meter.Int64Counter(
		"http_requests_total",
		otelpkg.WithDescription("Total number of HTTP requests"),
		otelpkg.WithUnit("{request}"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create request count metric: %w", err)
	}

	requestDuration, err := meter.Float64Histogram(
		"http_request_duration_seconds",
		otelpkg.WithDescription("HTTP request duration in seconds"),
		otelpkg.WithUnit("s"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create request duration metric: %w", err)
	}

	errorCount, err := meter.Int64Counter(
		"http_errors_total",
		otelpkg.WithDescription("Total number of HTTP errors"),
		otelpkg.WithUnit("{error}"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create error count metric: %w", err)
	}

	activeRequests, err := meter.Int64UpDownCounter(
		"http_active_requests",
		otelpkg.WithDescription("Number of active HTTP requests"),
		otelpkg.WithUnit("{request}"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create active requests metric: %w", err)
	}

	globalProvider = &Provider{
		TracerProvider:  tp,
		MeterProvider:   mp,
		Tracer:          tracer,
		Meter:           meter,
		RequestCount:    requestCount,
		RequestDuration: requestDuration,
		ErrorCount:      errorCount,
		ActiveRequests:  activeRequests,
	}

	slog.Info("OpenTelemetry initialized",
		"service", serviceName,
		"endpoint", cfg.OTELEndpoint,
		"environment", cfg.Environment,
	)

	return globalProvider, nil
}

// Shutdown gracefully shuts down the telemetry providers.
func Shutdown(ctx context.Context) error {
	if globalProvider == nil {
		return nil
	}

	if globalProvider.TracerProvider != nil {
		if err := globalProvider.TracerProvider.Shutdown(ctx); err != nil {
			return fmt.Errorf("failed to shutdown tracer provider: %w", err)
		}
	}

	slog.Info("OpenTelemetry shutdown complete")
	return nil
}

// GetProvider returns the global telemetry provider.
func GetProvider() *Provider {
	return globalProvider
}

// StartSpan starts a new span with the given name.
func StartSpan(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	if globalProvider == nil || globalProvider.Tracer == nil {
		return ctx, trace.SpanFromContext(ctx)
	}
	return globalProvider.Tracer.Start(ctx, name, opts...)
}

// RecordRequest records an HTTP request metric.
func (p *Provider) RecordRequest(ctx context.Context, method, path string, status int, duration time.Duration) {
	attrs := []attribute.KeyValue{
		attribute.String("http.method", method),
		attribute.String("http.route", path),
		attribute.Int("http.status_code", status),
	}

	p.RequestCount.Add(ctx, 1, otelpkg.WithAttributes(attrs...))
	p.RequestDuration.Record(ctx, duration.Seconds(), otelpkg.WithAttributes(attrs...))

	if status >= 400 {
		p.ErrorCount.Add(ctx, 1, otelpkg.WithAttributes(attrs...))
	}
}

// AddSpanEvent adds an event to the current span.
func AddSpanEvent(ctx context.Context, name string, attrs ...attribute.KeyValue) {
	span := trace.SpanFromContext(ctx)
	span.AddEvent(name, trace.WithAttributes(attrs...))
}

// SetSpanAttributes sets attributes on the current span.
func SetSpanAttributes(ctx context.Context, attrs ...attribute.KeyValue) {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(attrs...)
}

// RecordError records an error on the current span.
func RecordError(ctx context.Context, err error, attrs ...attribute.KeyValue) {
	span := trace.SpanFromContext(ctx)
	span.RecordError(err, trace.WithAttributes(attrs...))
}
