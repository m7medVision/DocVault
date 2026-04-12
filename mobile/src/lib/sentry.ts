import * as Sentry from '@sentry/react-native';

// Initialize Sentry for error tracking
const sentryDsn = process.env.SENTRY_DSN_MOBILE || '';

if (sentryDsn) {
  Sentry.init({
    dsn: sentryDsn,
    environment: process.env.NODE_ENV || 'development',
    release: `docvault-mobile@1.0.0`,
    // Sample rate for traces (1.0 = 100%)
    tracesSampleRate: 1.0,
  });
  console.log('Sentry initialized for mobile app');
} else {
  console.warn('SENTRY_DSN not configured, skipping Sentry initialization');
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
export function captureSentryError(error: Error, context?: Record<string, string>) {
  Sentry.withScope((scope) => {
    if (context) {
      Object.entries(context).forEach(([key, value]) => {
        scope.setTag(key, value);
      });
    }
    Sentry.captureException(error);
  });
}

export { Sentry };
