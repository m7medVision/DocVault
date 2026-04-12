// Server-only OpenTelemetry initialization
// Note: For a simpler setup, we only initialize Sentry in instrumentation.ts
// OpenTelemetry can be configured separately if needed

import { trace, SpanKind } from '@opentelemetry/api';

// Export trace utilities for use in server components
export { trace, SpanKind };

// Helper function to get current span
export function getCurrentSpan() {
  return trace.getActiveSpan();
}
