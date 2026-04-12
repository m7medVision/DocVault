import * as Sentry from '@sentry/nextjs';

Sentry.init({
  dsn: process.env.SENTRY_DSN_WEB || '',
  environment: process.env.NODE_ENV || 'development',
  release: `docvault-web@1.0.0`,
  tracesSampleRate: 1.0,
  beforeSend(event) {
    if (event.request?.url?.includes('/health')) {
      return null;
    }
    return event;
  },
});
