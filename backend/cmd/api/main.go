// Package main is the entry point for the DocVault Core Platform Service.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	internalauth "github.com/docvault/backend/internal/auth"
	"github.com/docvault/backend/internal/authz"
	"github.com/docvault/backend/internal/config"
	"github.com/docvault/backend/internal/database"
	"github.com/docvault/backend/internal/handler"
	"github.com/docvault/backend/internal/middleware"
	"github.com/docvault/backend/internal/migrate"
	"github.com/docvault/backend/internal/minio"
	"github.com/docvault/backend/internal/rabbitmq"
	appredis "github.com/docvault/backend/internal/redis"
	"github.com/docvault/backend/internal/repository"
	"github.com/docvault/backend/internal/sentry"
	"github.com/docvault/backend/internal/telemetry"
	"github.com/joho/godotenv"
)

func main() {
	loadRepoEnv()

	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load configuration", "error", err)
		os.Exit(1)
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: cfg.Log.Level,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			return a
		},
	}))
	slog.SetDefault(logger)

	telemetryCtx, telemetryCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer telemetryCancel()

	tp, err := telemetry.Init(telemetryCtx, cfg, "docvault-backend")
	if err != nil {
		slog.Warn("failed to initialize telemetry, continuing without", "error", err)
	}
	defer func() {
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownCancel()
		if err := telemetry.Shutdown(shutdownCtx); err != nil {
			slog.Error("failed to shutdown telemetry", "error", err)
		}
		if !sentry.Flush(shutdownCtx) {
			slog.Error("failed to flush Sentry")
		}
	}()

	if err := sentry.Init(telemetryCtx, cfg, "docvault-backend"); err != nil {
		slog.Warn("failed to initialize Sentry, continuing without", "error", err)
	}

	_ = tp

	initCtx := context.Background()

	dbPool, err := database.NewConnection(initCtx, cfg.DB.URL)
	if err != nil {
		slog.Error("failed to initialize database", "error", err)
		os.Exit(1)
	}
	defer dbPool.Close()
	slog.Info("database connection initialized")

	if err := migrate.Run(initCtx, cfg.DB.URL); err != nil {
		slog.Error("failed to run database migrations", "error", err)
		os.Exit(1)
	}
	slog.Info("database migrations applied")

	minioClient, err := minio.NewClient(initCtx, cfg.Storage.Endpoint, cfg.Storage.AccessKey, cfg.Storage.SecretKey, cfg.Storage.UseSSL)
	if err != nil {
		slog.Error("failed to initialize MinIO client", "error", err)
		os.Exit(1)
	}
	slog.Info("MinIO client initialized")

	if err := minio.EnsureBucket(initCtx, minioClient, cfg.Storage.Bucket); err != nil {
		slog.Warn("failed to ensure MinIO bucket", "error", err)
	}

	rabbitConn, err := rabbitmq.NewConnection(initCtx, cfg.Queue.URL)
	if err != nil {
		slog.Error("failed to initialize RabbitMQ connection", "error", err)
		os.Exit(1)
	}
	defer rabbitConn.Close()
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
		slog.Error("failed to initialize redis client", "error", err)
		os.Exit(1)
	}
	defer redisClient.Close()

	jwtService, err := internalauth.NewJWTService(
		cfg.Auth.JWTSecret,
		cfg.Auth.JWTIssuer,
		cfg.Auth.JWTAudience,
		cfg.Auth.JWTAccessTokenTTL,
		cfg.Auth.JWTRefreshTokenTTL,
	)
	if err != nil {
		slog.Error("failed to initialize JWT service", "error", err)
		os.Exit(1)
	}

	tokenBlacklist := appredis.NewTokenBlacklist(redisClient)
	rateLimiter := appredis.NewRateLimiter(redisClient)
	jwtMiddleware := middleware.NewJWTMiddleware(jwtService, tokenBlacklist)
	authzEnforcer, err := authz.NewEnforcer(filepath.Join("internal", "authz", "model.conf"), cfg.DB.URL)
	if err != nil {
		slog.Error("failed to initialize authorization enforcer", "error", err)
		os.Exit(1)
	}
	slog.Info("JWT authentication initialized")
	slog.Info("Casbin authorization initialized")

	repos := repository.NewRepositories(dbPool, authzEnforcer)
	objectStore := minio.NewObjectStore(minioClient, cfg.Storage.Bucket)
	publisher := rabbitmq.NewPublisher(rabbitConn, cfg.Queue.URL, cfg.Queue.ProcessQueue)
	h := handler.New(cfg, repos, dbPool, objectStore, publisher, authzEnforcer)
	h.SetDB(dbPool)
	h.SetAuthorizationEnforcer(authzEnforcer)
	middleware.SetAuthorizationAuditLogger(h.AuditAuthorizationDecision)
	authHandler := handler.NewAuthHandler(dbPool, jwtService, tokenBlacklist, rateLimiter, authzEnforcer, logger, repos.User)

	mux := http.NewServeMux()
	handler.RegisterRoutes(h, authHandler, mux, jwtMiddleware, authzEnforcer)
	router := middleware.Chain(
		mux,
		middleware.Telemetry(),
		middleware.RequestID(),
		middleware.Logger(),
		middleware.Recoverer(),
		middleware.CORS(cfg.Server.CORSOrigins),
		jwtMiddleware.Authenticate,
	)

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		slog.Info("starting server", "port", cfg.Server.Port, "environment", cfg.Server.Environment)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server failed to start", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		slog.Error("server forced to shutdown", "error", err)
	}

	slog.Info("server stopped")
}

func loadRepoEnv() {
	root, err := findRepoRoot()
	if err != nil {
		return
	}

	_ = godotenv.Load(filepath.Join(root, ".env"))
}

func findRepoRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "justfile")); err == nil {
			if _, err := os.Stat(filepath.Join(dir, "docker-compose.yml")); err == nil {
				return dir, nil
			}
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("repo root not found")
		}
		dir = parent
	}
}
