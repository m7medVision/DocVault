// Package sentry provides Sentry error tracking integration.
package sentry

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/getsentry/sentry-go"
)

// Init initializes Sentry error tracking.
func Init(ctx context.Context, dsn, environment, serviceName string) error {
	if dsn == "" {
		slog.Info("SENTRY_DSN not configured, skipping Sentry initialization")
		return nil
	}

	opts := sentry.ClientOptions{
		Dsn:              dsn,
		Environment:      environment,
		Release:          serviceName + "@1.0.0",
		TracesSampleRate: 1.0,
		EnableTracing:    true,
	}

	err := sentry.Init(opts)
	if err != nil {
		return err
	}

	slog.Info("Sentry initialized",
		"service", serviceName,
		"environment", environment,
	)

	return nil
}

// SetUser sets user context for Sentry.
func SetUser(userID, tenantID string) {
	if userID == "" {
		return
	}

	sentry.ConfigureScope(func(scope *sentry.Scope) {
		scope.SetUser(sentry.User{
			ID: userID,
		})
		if tenantID != "" {
			scope.SetTag("tenant_id", tenantID)
		}
	})
}

// AddTenant adds tenant context to Sentry scope.
func AddTenant(tenantID string) {
	if tenantID == "" {
		return
	}
	sentry.ConfigureScope(func(scope *sentry.Scope) {
		scope.SetTag("tenant_id", tenantID)
	})
}

// RecordError records an error to Sentry.
func RecordError(err error, tags map[string]string) {
	if err == nil {
		return
	}

	sentry.WithScope(func(scope *sentry.Scope) {
		for k, v := range tags {
			scope.SetTag(k, v)
		}
		sentry.CaptureException(err)
	})
}

// Flush flushes pending events to Sentry.
func Flush(ctx context.Context) bool {
	return sentry.Flush(2 * time.Second)
}

// Recover captures panics and reports them to Sentry.
func Recover() {
	if r := recover(); r != nil {
		sentry.CaptureMessage("panic recovered: " + fmt.Sprintf("%v", r))
	}
}
