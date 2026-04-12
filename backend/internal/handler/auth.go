package handler

import (
	"log/slog"

	"github.com/casbin/casbin/v3"
	"github.com/docvault/backend/internal/auth"
	appredis "github.com/docvault/backend/internal/redis"
	"github.com/docvault/backend/internal/repository"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AuthHandler struct {
	db             *pgxpool.Pool
	jwtService     *auth.JWTService
	tokenBlacklist *appredis.TokenBlacklist
	rateLimiter    *appredis.RateLimiter
	authzEnforcer  *casbin.Enforcer
	logger         *slog.Logger
	userRepo       repository.UserRepository
}

func NewAuthHandler(
	db *pgxpool.Pool,
	jwtService *auth.JWTService,
	tokenBlacklist *appredis.TokenBlacklist,
	rateLimiter *appredis.RateLimiter,
	authzEnforcer *casbin.Enforcer,
	logger *slog.Logger,
	userRepo repository.UserRepository,
) *AuthHandler {
	return &AuthHandler{
		db:             db,
		jwtService:     jwtService,
		tokenBlacklist: tokenBlacklist,
		rateLimiter:    rateLimiter,
		authzEnforcer:  authzEnforcer,
		logger:         logger,
		userRepo:       userRepo,
	}
}
