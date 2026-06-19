package app

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
	identitypg "github.com/docvault/backend/internal/identity/adapter/postgres"
	"github.com/docvault/backend/internal/middleware"
	"github.com/docvault/backend/internal/migrate"
	"github.com/docvault/backend/internal/minio"
	"github.com/docvault/backend/internal/platform/cache"
	"github.com/docvault/backend/internal/rabbitmq"
	appredis "github.com/docvault/backend/internal/redis"
	"github.com/docvault/backend/internal/repository"
	"github.com/docvault/backend/internal/search"
	"github.com/docvault/backend/internal/sentry"
	"github.com/docvault/backend/internal/telemetry"
	handler "github.com/docvault/backend/internal/transport/http"
	"github.com/docvault/backend/internal/usecase"
	"github.com/joho/godotenv"
)

func Run() error {
	loadRepoEnv()

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load configuration: %w", err)
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
	_ = tp

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

	initCtx := context.Background()

	dbPool, err := database.NewConnection(initCtx, cfg.DB.URL)
	if err != nil {
		return fmt.Errorf("initialize database: %w", err)
	}
	defer dbPool.Close()
	slog.Info("database connection initialized")

	if err := migrate.Run(initCtx, cfg.DB.URL); err != nil {
		return fmt.Errorf("run database migrations: %w", err)
	}
	slog.Info("database migrations applied")

	minioClient, err := minio.NewClient(initCtx, cfg.Storage.Endpoint, cfg.Storage.AccessKey, cfg.Storage.SecretKey, cfg.Storage.UseSSL)
	if err != nil {
		return fmt.Errorf("initialize MinIO client: %w", err)
	}
	slog.Info("MinIO client initialized")

	if err := minio.EnsureBucket(initCtx, minioClient, cfg.Storage.Bucket); err != nil {
		slog.Warn("failed to ensure MinIO bucket", "error", err)
	}

	rabbitConn, err := rabbitmq.NewConnection(initCtx, cfg.Queue.URL)
	if err != nil {
		return fmt.Errorf("initialize RabbitMQ connection: %w", err)
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
		return fmt.Errorf("initialize redis client: %w", err)
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
		return fmt.Errorf("initialize JWT service: %w", err)
	}

	tokenBlacklist := appredis.NewTokenBlacklist(redisClient)
	rateLimiter := appredis.NewRateLimiter(redisClient)
	jwtMiddleware := middleware.NewJWTMiddleware(jwtService, tokenBlacklist)
	authzEnforcer, err := authz.NewEnforcer(filepath.Join("internal", "authz", "model.conf"), cfg.DB.URL)
	if err != nil {
		return fmt.Errorf("initialize authorization enforcer: %w", err)
	}
	slog.Info("JWT authentication initialized")
	slog.Info("Casbin authorization initialized")

	repos := buildRepositories(dbPool, authzEnforcer)
	// One Redis-backed cache shared by the repository decorators below; each uses
	// a distinct key prefix so namespaces never collide.
	appCache := cache.NewRedis(redisClient.Client)
	// Cache the per-(org,user) group-membership lookup the authorizer runs on
	// nearly every non-admin request. Membership mutations bust the affected
	// keys; a short TTL bounds staleness. Wrapped once here so every consumer
	// (document/folder services, the handler's authorizer) shares the cache.
	aclRepo := repository.NewCachingACL(repos.ACL, appCache)
	// Cache the per-org document-stats aggregate; document create/delete bust it,
	// a short TTL covers status changes made by the processing workers.
	documentRepo := repository.NewCachingDocuments(repos.Document, appCache)
	// Cache the per-org folder tree read on every navigation render; structural
	// folder mutations bust it, a short TTL covers is_restricted display toggles.
	folderRepo := repository.NewCachingFolders(repos.Folder, appCache)
	objectStore := minio.NewObjectStore(minioClient, cfg.Storage.Bucket)
	ocrDispatcher := rabbitmq.NewOCRDispatcher(rabbitConn, cfg.Queue.URL, cfg.Queue.OCRQueue)
	embedder := search.NewOpenRouterEmbedder(cfg.Search.EmbeddingAPIKey, cfg.Search.EmbeddingModel, cfg.Search.EmbeddingDim)
	// Cache query embeddings in Redis: identical /search and /chat queries reuse
	// the vector instead of re-calling the external embedding API. The key is
	// content-addressed (model+dim+text) and carries no tenant data, so it is
	// safe to share across tenants. Best-effort: a cache failure falls through
	// to the embedder.
	queryEmbedder := search.NewCachingEmbedder(embedder, cache.NewRedis(redisClient.Client), cfg.Search.EmbeddingModel, cfg.Search.EmbeddingDim, 7*24*time.Hour)
	h := handler.New(cfg, handler.Dependencies{
		DB:              dbPool,
		AuthzEnforcer:   authzEnforcer,
		DocumentSvc:     usecase.NewDocumentService(documentRepo, aclRepo, objectStore, ocrDispatcher),
		FolderSvc:       usecase.NewFolderService(folderRepo, aclRepo),
		TagSvc:          usecase.NewTagService(repos.Tag),
		AuditSvc:        usecase.NewAuditService(repos.Audit),
		ReminderSvc:     usecase.NewReminderService(repos.Reminder),
		NotificationSvc: usecase.NewNotificationService(repos.Notification),
		SearchSvc:       usecase.NewSearchService(queryEmbedder, repos.Search),
		ChatSvc:         usecase.NewChatService(queryEmbedder, repos.Search),
		UserRepo:        repos.User,
		MembershipRepo:  repos.Membership,
		PolicyRepo:      repos.Policy,
		ACLRepo:         aclRepo,
	})
	middleware.SetAuthorizationAuditLogger(h.AuditAuthorizationDecision)
	authHandler := handler.NewAuthHandler(dbPool, jwtService, tokenBlacklist, rateLimiter, authzEnforcer, logger, repos.User, identitypg.NewRegistrationRepository(dbPool))

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
		Addr:    fmt.Sprintf(":%d", cfg.Server.Port),
		Handler: router,
		// ReadHeaderTimeout guards against slow-loris header attacks. We
		// deliberately do NOT set an absolute WriteTimeout: /chat streams
		// Server-Sent Events that routinely outlast any fixed deadline, and a
		// global WriteTimeout silently truncates the stream mid-response.
		// Per-request bounds come from the handler context instead. ReadTimeout
		// is generous enough to read large (up to 50MB) multipart uploads over
		// slow mobile links.
		ReadHeaderTimeout: 15 * time.Second,
		ReadTimeout:       5 * time.Minute,
		WriteTimeout:      0,
		IdleTimeout:       60 * time.Second,
	}

	serverErr := make(chan error, 1)
	go func() {
		slog.Info("starting server", "port", cfg.Server.Port, "environment", cfg.Server.Environment)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErr <- err
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverErr:
		return fmt.Errorf("server failed: %w", err)
	case <-quit:
	}

	slog.Info("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		return fmt.Errorf("server forced to shutdown: %w", err)
	}

	slog.Info("server stopped")
	return nil
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
