package handler

import (
	"log/slog"

	"github.com/casbin/casbin/v3"
	"github.com/docvault/backend/internal/auth"
	sqldb "github.com/docvault/backend/internal/db"
	appredis "github.com/docvault/backend/internal/redis"
	"github.com/docvault/backend/internal/repository"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AuthHandler struct {
	db               *pgxpool.Pool
	jwtService       *auth.JWTService
	tokenBlacklist   *appredis.TokenBlacklist
	rateLimiter      *appredis.RateLimiter
	authzEnforcer    *casbin.Enforcer
	logger           *slog.Logger
	userRepo         repository.UserRepository
	registrationRepo repository.RegistrationRepository
	queries          *sqldb.Queries
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
		db:               db,
		jwtService:       jwtService,
		tokenBlacklist:   tokenBlacklist,
		rateLimiter:      rateLimiter,
		authzEnforcer:    authzEnforcer,
		logger:           logger,
		userRepo:         userRepo,
		registrationRepo: repository.NewRegistrationRepository(db),
		queries:          sqldb.New(db),
	}
}
