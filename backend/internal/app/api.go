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
	"github.com/docvault/backend/internal/middleware"
	appredis "github.com/docvault/backend/internal/redis"
	"github.com/joho/godotenv"
)

// Run is the composition root: it loads configuration, brings up observability
// and the backing infrastructure, wires the handlers, and serves until a
// shutdown signal. Each phase is delegated to a focused helper (initObservability,
// initInfra, buildHandlers, buildRouter, newServer, serve) so this function reads
// as the high-level startup sequence.
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

	defer initObservability(cfg)()

	inf, err := initInfra(context.Background(), cfg)
	if err != nil {
		return err
	}
	defer inf.cleanup()

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

	tokenBlacklist := appredis.NewTokenBlacklist(inf.redis)
	rateLimiter := appredis.NewRateLimiter(inf.redis)
	jwtMiddleware := middleware.NewJWTMiddleware(jwtService, tokenBlacklist)
	authzEnforcer, err := authz.NewEnforcer(filepath.Join("internal", "authz", "model.conf"), cfg.DB.URL)
	if err != nil {
		return fmt.Errorf("initialize authorization enforcer: %w", err)
	}
	slog.Info("JWT authentication initialized")
	slog.Info("Casbin authorization initialized")

	h, authHandler := buildHandlers(cfg, inf, jwtService, tokenBlacklist, rateLimiter, authzEnforcer, logger)
	router := buildRouter(cfg, h, authHandler, jwtMiddleware, authzEnforcer)
	return serve(cfg, newServer(cfg, router))
}

// serve starts the HTTP server in a goroutine and blocks until it fails or a
// SIGINT/SIGTERM arrives, then drains in-flight requests within a 30s deadline.
func serve(cfg *config.Config, server *http.Server) error {
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
