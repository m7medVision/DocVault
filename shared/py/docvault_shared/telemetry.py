"""OpenTelemetry instrumentation for DocVault Python services."""

import logging
import os
import time
import structlog
from typing import Optional, Dict, Any
from functools import wraps

import sentry_sdk
from sentry_sdk.integrations.logging import LoggingIntegration

from opentelemetry import trace, metrics
from opentelemetry.sdk.trace import TracerProvider
from opentelemetry.sdk.trace.export import BatchSpanProcessor
from opentelemetry.sdk.metrics import MeterProvider
from opentelemetry.sdk.metrics.export import PeriodicExportingMetricReader
from opentelemetry.sdk.resources import Resource, SERVICE_NAME, SERVICE_VERSION
from opentelemetry.exporter.otlp.proto.grpc.trace_exporter import OTLPSpanExporter
from opentelemetry.exporter.otlp.proto.grpc.metric_exporter import OTLPMetricExporter
from opentelemetry.semconv.resource import ResourceAttributes
from opentelemetry.propagate import set_global_textmap
from opentelemetry.propagators.b3 import B3MultiFormat
from opentelemetry.trace import Status, StatusCode

logger = structlog.get_logger(__name__)


class TelemetryProvider:
    """Holds OpenTelemetry providers and instruments."""

    def __init__(self):
        self.tracer_provider: Optional[TracerProvider] = None
        self.meter_provider: Optional[MeterProvider] = None
        self.tracer = None
        self.meter = None

        self.job_counter = None
        self.job_duration = None
        self.job_error_counter = None
        self.active_jobs = None

        self._initialized = False


_global_provider: Optional[TelemetryProvider] = None


def get_logger():
    """Return the shared structured logger for telemetry-aware handlers."""
    return logger


def init_telemetry(service_name: str = "docvault-service") -> TelemetryProvider:
    """Initialize OpenTelemetry with OTLP exporters."""
    global _global_provider

    otel_endpoint = os.getenv("OTEL_EXPORTER_ENDPOINT", "")

    if not otel_endpoint:
        logger.warning("OTEL_EXPORTER_ENDPOINT not configured, skipping telemetry initialization")
        _global_provider = TelemetryProvider()
        return _global_provider

    environment = os.getenv("ENVIRONMENT", "development")

    resource = Resource.create(
        {
            SERVICE_NAME: service_name,
            SERVICE_VERSION: "1.0.0",
            ResourceAttributes.DEPLOYMENT_ENVIRONMENT: environment,
        }
    )

    span_exporter = OTLPSpanExporter(
        endpoint=otel_endpoint,
        insecure=True,
    )

    metric_exporter = OTLPMetricExporter(
        endpoint=otel_endpoint,
        insecure=True,
    )

    tracer_provider = TracerProvider(resource=resource)
    tracer_provider.add_span_processor(BatchSpanProcessor(span_exporter))

    metric_reader = PeriodicExportingMetricReader(metric_exporter, export_interval_millis=60000)
    meter_provider = MeterProvider(resource=resource, metric_readers=[metric_reader])

    trace.set_tracer_provider(tracer_provider)
    metrics.set_meter_provider(meter_provider)

    set_global_textmap(B3MultiFormat())

    tracer = trace.get_tracer(service_name)
    meter = metrics.get_meter(service_name)

    job_counter = meter.create_counter(
        name="jobs_total",
        description="Total number of jobs",
        unit="1",
    )

    job_duration = meter.create_histogram(
        name="job_duration_seconds",
        description="Job duration in seconds",
        unit="s",
    )

    job_error_counter = meter.create_counter(
        name="jobs_errors_total",
        description="Total number of job errors",
        unit="1",
    )

    active_jobs = meter.create_up_down_counter(
        name="active_jobs",
        description="Number of active jobs",
        unit="1",
    )

    _global_provider = TelemetryProvider()
    _global_provider.tracer_provider = tracer_provider
    _global_provider.meter_provider = meter_provider
    _global_provider.tracer = tracer
    _global_provider.meter = meter
    _global_provider.job_counter = job_counter
    _global_provider.job_duration = job_duration
    _global_provider.job_error_counter = job_error_counter
    _global_provider.active_jobs = active_jobs
    _global_provider._initialized = True

    logger.info(
        "OpenTelemetry initialized",
        service=service_name,
        endpoint=otel_endpoint,
        environment=environment,
    )

    init_sentry(service_name, environment)

    return _global_provider


def init_sentry(service_name: str, environment: str, sentry_dsn: str = "") -> None:
    """Initialize Sentry error tracking."""
    from .config import config as shared_config

    if not sentry_dsn:
        sentry_dsn = shared_config.sentry_dsn

    if not sentry_dsn:
        logger.warning("SENTRY_DSN not configured, skipping Sentry initialization")
        return

    sentry_logging = LoggingIntegration(
        level=logging.INFO,
        event_level=logging.ERROR,
    )

    sentry_sdk.init(
        dsn=sentry_dsn,
        environment=environment,
        release=f"{service_name}@1.0.0",
        integrations=[sentry_logging],
        traces_sample_rate=1.0,
        ignore_errors=[KeyboardInterrupt],
    )

    logger.info(
        "Sentry initialized",
        service=service_name,
        environment=environment,
    )


def shutdown_telemetry() -> None:
    """Gracefully shutdown the telemetry providers."""
    global _global_provider

    if _global_provider and _global_provider.tracer_provider:
        _global_provider.tracer_provider.shutdown()

    sentry_sdk.flush(timeout=5.0)

    logger.info("Telemetry shutdown complete")


def get_provider() -> TelemetryProvider:
    """Get the global telemetry provider."""
    global _global_provider
    return _global_provider


def start_span(
    name: str, kind=trace.SpanKind.INTERNAL, attributes: Optional[Dict[str, Any]] = None
):
    """Start a new span with the given name."""
    provider = get_provider()
    if provider and provider.tracer:
        return provider.tracer.start_as_current_span(
            name,
            kind=kind,
            attributes=attributes,
        )
    return trace.get_tracer("noop").start_as_current_span(name)


def record_job(
    job_type: str, success: bool, duration: float, attributes: Optional[Dict[str, Any]] = None
) -> None:
    """Record a job processing metric."""
    provider = get_provider()
    if not provider or not provider._initialized:
        return

    attrs = {
        "job.type": job_type,
        "job.success": success,
    }
    if attributes:
        attrs.update(attributes)

    provider.job_counter.add(1, attrs)
    provider.job_duration.record(duration, attrs)

    if not success:
        provider.job_error_counter.add(1, attrs)


def record_error(error: Exception, attributes: Optional[Dict[str, Any]] = None) -> None:
    """Record an error in the current span."""
    span = trace.get_current_span()
    if span and span.is_recording():
        span.set_status(Status(StatusCode.ERROR, str(error)))
        span.record_exception(error, attributes=attributes)

    with sentry_sdk.configure_scope() as scope:
        for key, value in (attributes or {}).items():
            scope.set_tag(key, str(value))
        sentry_sdk.capture_exception(error)


def set_user_context(user_id: str, tenant_id: str = None) -> None:
    """Set user context for Sentry error tracking."""
    if not user_id:
        return

    sentry_sdk.set_user({"id": user_id})
    if tenant_id:
        sentry_sdk.set_tag("tenant_id", tenant_id)


def add_tenant_context(tenant_id: str) -> None:
    """Add tenant context to Sentry scope."""
    if tenant_id:
        sentry_sdk.set_tag("tenant_id", tenant_id)


def traced(job_type: str):
    """Decorator to trace a function as a job."""

    def decorator(func):
        @wraps(func)
        def wrapper(*args, **kwargs):
            start_time = time.time()
            provider = get_provider()

            if provider and provider._initialized and provider.active_jobs:
                provider.active_jobs.add(1, {"job.type": job_type})

            with start_span(
                f"{job_type}_job", kind=trace.SpanKind.CONSUMER, attributes={"job.type": job_type}
            ):
                try:
                    result = func(*args, **kwargs)
                    duration = time.time() - start_time
                    record_job(job_type, True, duration)
                    return result
                except Exception as e:
                    duration = time.time() - start_time
                    record_job(job_type, False, duration)
                    record_error(e)
                    raise
                finally:
                    if provider and provider._initialized and provider.active_jobs:
                        provider.active_jobs.add(-1, {"job.type": job_type})

        return wrapper

    return decorator


class SpanContextLogger:
    """Context manager that adds trace context to logs."""

    def __init__(self, logger, span_name: str, attributes: Optional[Dict[str, Any]] = None):
        self.logger = logger
        self.span_name = span_name
        self.attributes = attributes or {}
        self.start_time = None

    def __enter__(self):
        self.start_time = time.time()
        span = trace.get_current_span()

        if span and span.is_recording():
            ctx = span.get_span_context()
            self.logger = self.logger.bind(
                trace_id=format(ctx.trace_id, "032x"),
                span_id=format(ctx.span_id, "016x"),
            )

        self.logger = self.logger.bind(**self.attributes)
        return self

    def __exit__(self, exc_type, exc_val, exc_tb):
        duration = time.time() - self.start_time

        if exc_type:
            self.logger.error(
                f"{self.span_name} failed",
                duration_ms=int(duration * 1000),
                error=str(exc_val),
            )
        else:
            self.logger.info(
                self.span_name,
                duration_ms=int(duration * 1000),
            )

        return False


def log_with_trace(logger, span_name: str, attributes: Optional[Dict[str, Any]] = None):
    """Create a logger context with trace correlation."""
    return SpanContextLogger(logger, span_name, attributes)
