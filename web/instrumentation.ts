import { trace, SpanKind, context, propagation } from '@opentelemetry/api';
import * as Sentry from "@sentry/nextjs";

export async function register() {
  if (process.env.NEXT_RUNTIME === "nodejs") {
    await import("./sentry.server.config");
  }

  if (process.env.NEXT_RUNTIME === "edge") {
    await import("./sentry.edge.config");
  }
}

// Automatically captures all unhandled server-side request errors
export const onRequestError = Sentry.captureRequestError;

// Export trace utilities
export { trace, SpanKind, context, propagation };

// Helper function to get current span
export function getCurrentSpan() {
  return trace.getActiveSpan();
}

// Helper to add trace context to logs
export function getTraceContext() {
  const span = trace.getActiveSpan();
  if (span) {
    const spanContext = span.spanContext();
    return {
      trace_id: spanContext.traceId,
      span_id: spanContext.spanId,
      trace_flags: spanContext.traceFlags,
    };
  }
  return {};
}

// Helper to set user context for Sentry
export function setSentryUser(userId: string, tenantId?: string) {
  Sentry.setUser({
    id: userId,
    segment: tenantId,
  });
  if (tenantId) {
    Sentry.setTag('tenant_id', tenantId);
  }
}

// Helper to add tenant context
export function addSentryTenant(tenantId: string) {
  Sentry.setTag('tenant_id', tenantId);
}

// Helper to capture errors to Sentry
export function captureSentryError(error: Error, errorContext?: Record<string, string>) {
  Sentry.withScope((scope) => {
    if (errorContext) {
      Object.entries(errorContext).forEach(([key, value]) => {
        scope.setTag(key, value);
      });
    }
    Sentry.captureException(error);
  });
}
