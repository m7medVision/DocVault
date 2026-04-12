// Package telemetry provides OpenTelemetry instrumentation for the worker service.
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

	appconfig "github.com/docvault/reminder/internal/config"
)

// Provider holds OpenTelemetry providers and instruments.
type Provider struct {
	TracerProvider *sdktrace.TracerProvider
	MeterProvider  *sdkmetric.MeterProvider
	Tracer         trace.Tracer
	Meter          otelpkg.Meter

	// Metrics
	JobCount        otelpkg.Int64Counter
	JobDuration     otelpkg.Float64Histogram
	JobErrorCount   otelpkg.Int64Counter
	ActiveJobs      otelpkg.Int64UpDownCounter
	DLQMessageCount otelpkg.Int64Counter
}

var globalProvider *Provider

// Init initializes OpenTelemetry with OTLP exporters.
func Init(ctx context.Context, cfg *appconfig.Config, serviceName string) (*Provider, error) {
	endpoint := cfg.OTELEndpoint
	if endpoint == "" {
		slog.Warn("OTEL_ENDPOINT not configured, skipping telemetry initialization")
		return &Provider{}, nil
	}

	// Create OTLP trace exporter
	traceExporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpoint(endpoint),
		otlptracegrpc.WithInsecure(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create trace exporter: %w", err)
	}

	// Create OTLP metric exporter
	metricExporter, err := otlpmetricgrpc.New(ctx,
		otlpmetricgrpc.WithEndpoint(endpoint),
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
	jobCount, err := meter.Int64Counter(
		"worker_jobs_total",
		otelpkg.WithDescription("Total number of jobs processed"),
		otelpkg.WithUnit("{job}"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create job count metric: %w", err)
	}

	jobDuration, err := meter.Float64Histogram(
		"worker_job_duration_seconds",
		otelpkg.WithDescription("Job processing duration in seconds"),
		otelpkg.WithUnit("s"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create job duration metric: %w", err)
	}

	jobErrorCount, err := meter.Int64Counter(
		"worker_jobs_errors_total",
		otelpkg.WithDescription("Total number of job errors"),
		otelpkg.WithUnit("{error}"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create job error metric: %w", err)
	}

	activeJobs, err := meter.Int64UpDownCounter(
		"worker_active_jobs",
		otelpkg.WithDescription("Number of active jobs being processed"),
		otelpkg.WithUnit("{job}"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create active jobs metric: %w", err)
	}

	dlqMessageCount, err := meter.Int64Counter(
		"worker_dlq_messages_total",
		otelpkg.WithDescription("Total number of messages sent to DLQ"),
		otelpkg.WithUnit("{message}"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create DLQ metric: %w", err)
	}

	globalProvider = &Provider{
		TracerProvider:  tp,
		MeterProvider:   mp,
		Tracer:          tracer,
		Meter:           meter,
		JobCount:        jobCount,
		JobDuration:     jobDuration,
		JobErrorCount:   jobErrorCount,
		ActiveJobs:      activeJobs,
		DLQMessageCount: dlqMessageCount,
	}

	slog.Info("OpenTelemetry initialized",
		"service", serviceName,
		"endpoint", endpoint,
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

// RecordJob records a job processing metric.
func (p *Provider) RecordJob(ctx context.Context, jobType string, success bool, duration time.Duration) {
	if p == nil {
		return
	}

	attrs := []attribute.KeyValue{
		attribute.String("job.type", jobType),
		attribute.Bool("job.success", success),
	}

	p.JobCount.Add(ctx, 1, otelpkg.WithAttributes(attrs...))
	p.JobDuration.Record(ctx, duration.Seconds(), otelpkg.WithAttributes(attrs...))

	if !success {
		p.JobErrorCount.Add(ctx, 1, otelpkg.WithAttributes(attrs...))
	}
}

// RecordDLQ records a DLQ message metric.
func (p *Provider) RecordDLQ(ctx context.Context, jobType string) {
	attrs := []attribute.KeyValue{
		attribute.String("job.type", jobType),
	}
	p.DLQMessageCount.Add(ctx, 1, otelpkg.WithAttributes(attrs...))
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
