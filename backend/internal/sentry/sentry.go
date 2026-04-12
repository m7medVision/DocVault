// Package sentry provides Sentry error tracking integration.
package sentry

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/getsentry/sentry-go"

	appconfig "github.com/docvault/backend/internal/config"
)

// Init initializes Sentry error tracking.
func Init(ctx context.Context, cfg *appconfig.Config, serviceName string) error {
	if cfg.SentryDSN == "" {
		slog.Info("SENTRY_DSN not configured, skipping Sentry initialization")
		return nil
	}

	opts := sentry.ClientOptions{
		Dsn:              cfg.SentryDSN,
		Environment:      cfg.Environment,
		Release:          "docvault-backend@1.0.0",
		TracesSampleRate: 1.0,
		// Enable performance monitoring
		EnableTracing: true,
		// Ignore health check paths in BeforeSend
		BeforeSend: func(event *sentry.Event, hint *sentry.EventHint) *sentry.Event {
			// Check if this is a health check
			if hint != nil && hint.Request != nil {
				path := hint.Request.URL.Path
				if path == "/health" || path == "/health/ready" {
					return nil
				}
			}
			return event
		},
	}

	err := sentry.Init(opts)
	if err != nil {
		return err
	}

	slog.Info("Sentry initialized",
		"service", serviceName,
		"environment", cfg.Environment,
	)

	return nil
}

// SetUser sets user context for Sentry.
func SetUser(ctx context.Context, userID, tenantID string) {
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
func RecordError(ctx context.Context, err error, tags map[string]string) {
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
func Recover(ctx context.Context) {
	if r := recover(); r != nil {
		sentry.CaptureMessage("panic recovered: " + fmt.Sprintf("%v", r))
	}
}
