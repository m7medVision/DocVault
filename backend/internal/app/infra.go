package app

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/docvault/backend/internal/config"
	"github.com/docvault/backend/internal/database"
	"github.com/docvault/backend/internal/migrate"
	"github.com/docvault/backend/internal/minio"
	"github.com/docvault/backend/internal/rabbitmq"
	appredis "github.com/docvault/backend/internal/redis"
	"github.com/docvault/backend/internal/sentry"
	"github.com/docvault/backend/internal/telemetry"
	"github.com/jackc/pgx/v5/pgxpool"
)

// infra holds the backing services the application is built on. The raw MinIO
// and RabbitMQ clients are not exposed: the only downstream-needed values
// (the object store and OCR dispatcher) are constructed here, and the
// connections that must be closed are captured in cleanup.
type infra struct {
	db            *pgxpool.Pool
	redis         *appredis.Client
	objectStore   *minio.ObjectStore
	ocrDispatcher *rabbitmq.OCRDispatcher
	cleanup       func()
}

// initInfra connects every backing service (Postgres + migrations, MinIO,
// RabbitMQ, Redis) and returns them bundled. On any failure it closes whatever
// was already opened so the caller never leaks a connection. The caller must
// defer the returned infra's cleanup.
func initInfra(ctx context.Context, cfg *config.Config) (*infra, error) {
	dbPool, err := database.NewConnection(ctx, cfg.DB.URL)
	if err != nil {
		return nil, fmt.Errorf("initialize database: %w", err)
	}
	slog.Info("database connection initialized")

	if err := migrate.Run(ctx, cfg.DB.URL); err != nil {
		dbPool.Close()
		return nil, fmt.Errorf("run database migrations: %w", err)
	}
	slog.Info("database migrations applied")

	minioClient, err := minio.NewClient(ctx, cfg.Storage.Endpoint, cfg.Storage.AccessKey, cfg.Storage.SecretKey, cfg.Storage.UseSSL)
	if err != nil {
		dbPool.Close()
		return nil, fmt.Errorf("initialize MinIO client: %w", err)
	}
	slog.Info("MinIO client initialized")

	if err := minio.EnsureBucket(ctx, minioClient, cfg.Storage.Bucket); err != nil {
		slog.Warn("failed to ensure MinIO bucket", "error", err)
	}

	rabbitConn, err := rabbitmq.NewConnection(ctx, cfg.Queue.URL)
	if err != nil {
		dbPool.Close()
		return nil, fmt.Errorf("initialize RabbitMQ connection: %w", err)
	}
	slog.Info("RabbitMQ connection initialized")

	redisClient, err := appredis.NewClient(&appredis.Config{
		Host:         cfg.Redis.Host,
		Port:         cfg.Redis.Port,
		Password:     cfg.Redis.Password,
		DB:           cfg.Redis.DB,
		PoolSize:     cfg.Redis.PoolSize,
		MaxRetries:   cfg.Redis.MaxRetries,
		DialTimeout:  cfg.Redis.DialTimeout,
		ReadTimeout:  cfg.Redis.ReadTimeout,
		WriteTimeout: cfg.Redis.WriteTimeout,
	})
	if err != nil {
		_ = rabbitConn.Close()
		dbPool.Close()
		return nil, fmt.Errorf("initialize redis client: %w", err)
	}

	return &infra{
		db:            dbPool,
		redis:         redisClient,
		objectStore:   minio.NewObjectStore(minioClient, cfg.Storage.Bucket),
		ocrDispatcher: rabbitmq.NewOCRDispatcher(rabbitConn, cfg.Queue.URL, cfg.Queue.OCRQueue),
		cleanup: func() {
			_ = redisClient.Close()
			_ = rabbitConn.Close()
			dbPool.Close()
		},
	}, nil
}

// initObservability initializes telemetry and Sentry and returns a shutdown
// closure the caller must defer. Both are best-effort: a failure to initialize
// is logged and the application continues without them.
func initObservability(cfg *config.Config) func() {
	telemetryCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if _, err := telemetry.Init(telemetryCtx, cfg, "docvault-backend"); err != nil {
		slog.Warn("failed to initialize telemetry, continuing without", "error", err)
	}
	if err := sentry.Init(telemetryCtx, cfg, "docvault-backend"); err != nil {
		slog.Warn("failed to initialize Sentry, continuing without", "error", err)
	}

	return func() {
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownCancel()
		if err := telemetry.Shutdown(shutdownCtx); err != nil {
			slog.Error("failed to shutdown telemetry", "error", err)
		}
		if !sentry.Flush(shutdownCtx) {
			slog.Error("failed to flush Sentry")
		}
	}
}
