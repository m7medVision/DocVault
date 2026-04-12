import { trace, SpanKind, context, propagation } from '@opentelemetry/api';
import * as Sentry from '@sentry/nextjs';

// Initialize Sentry for error tracking (works on both client and server)
const sentryDsn = process.env.SENTRY_DSN_WEB || '';

if (sentryDsn) {
  Sentry.init({
    dsn: sentryDsn,
    environment: process.env.NODE_ENV || 'development',
    release: `docvault-web@1.0.0`,
    // Set sample rate for traces
    tracesSampleRate: 1.0,
    // Don't send health check failures to Sentry
    beforeSend(event) {
      if (event.request?.url?.includes('/health')) {
        return null;
      }
      return event;
    },
  });
  console.log('Sentry initialized for Next.js web app', { dsn: sentryDsn });
} else {
  console.warn('SENTRY_DSN not configured, skipping Sentry initialization');
}

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
