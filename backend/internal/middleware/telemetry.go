// Package middleware provides HTTP middleware for the API server.
package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"

	"github.com/docvault/backend/internal/telemetry"
)

// Telemetry returns a middleware that instruments HTTP requests with OpenTelemetry.
func Telemetry() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			ctx := r.Context()

			// Extract trace context from incoming request headers using global propagator
			prop := otel.GetTextMapPropagator()
			ctx = prop.Extract(ctx, propagation.HeaderCarrier(r.Header))

			// Start a span for this request
			spanName := r.Method + " " + r.URL.Path
			ctx, span := telemetry.StartSpan(ctx, spanName,
				trace.WithSpanKind(trace.SpanKindServer),
				trace.WithAttributes(
					attribute.String("http.method", r.Method),
					attribute.String("http.url", r.URL.String()),
					attribute.String("http.route", r.URL.Path),
					attribute.String("http.host", r.Host),
					attribute.String("http.user_agent", r.UserAgent()),
					attribute.String("net.peer.ip", r.RemoteAddr),
				),
			)
			defer span.End()

			// Wrap response writer to capture status code
			wrapped := &telemetryResponseWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
				ctx:            ctx,
			}

			// Track active requests
			provider := telemetry.GetProvider()
			if provider != nil && provider.ActiveRequests != nil {
				provider.ActiveRequests.Add(ctx, 1)
				defer provider.ActiveRequests.Add(ctx, -1)
			}

			// Process request
			next.ServeHTTP(wrapped, r.WithContext(ctx))

			// Record metrics
			duration := time.Since(start)
			wrapped.statusCode = wrapped.statusCode // Ensure we capture the final status

			span.SetAttributes(
				attribute.Int("http.status_code", wrapped.statusCode),
				attribute.Int64("http.response_size", int64(wrapped.bytesWritten)),
				attribute.Float64("http.duration_ms", float64(duration.Milliseconds())),
			)

			// Record in metrics
			if provider != nil {
				provider.RecordRequest(ctx, r.Method, r.URL.Path, wrapped.statusCode, duration)
			}

			// Log with trace correlation
			requestID := GetRequestID(ctx)
			slog.Info("request completed",
				"method", r.Method,
				"path", r.URL.Path,
				"status", wrapped.statusCode,
				"duration_ms", duration.Milliseconds(),
				"request_id", requestID,
				"trace_id", span.SpanContext().TraceID().String(),
				"span_id", span.SpanContext().SpanID().String(),
			)

			// Add trace ID to response headers for correlation
			w.Header().Set("X-Trace-ID", span.SpanContext().TraceID().String())
		})
	}
}

// telemetryResponseWriter wraps http.ResponseWriter to capture metrics.
type telemetryResponseWriter struct {
	http.ResponseWriter
	statusCode   int
	bytesWritten int64
	ctx          context.Context
}

func (w *telemetryResponseWriter) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *telemetryResponseWriter) Write(b []byte) (int, error) {
	n, err := w.ResponseWriter.Write(b)
	w.bytesWritten += int64(n)
	return n, err
}
