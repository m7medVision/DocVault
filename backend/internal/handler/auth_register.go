package handler

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/docvault/backend/internal/auth"
	"github.com/docvault/backend/internal/authz"
	sqldb "github.com/docvault/backend/internal/db"
	appredis "github.com/docvault/backend/internal/redis"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type RegisterRequest struct {
	Email       string `json:"email"`
	Password    string `json:"password"`
	DisplayName string `json:"display_name"`
	TenantName  string `json:"tenant_name"`
	OrgName     string `json:"org_name"`
	Locale      string `json:"locale"`
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	allowed, _, _, err := h.rateLimiter.Allow(
		ctx,
		appredis.BuildRateLimitKey("auth:register", clientIP(r)),
		appredis.RegisterRateLimit,
	)
	if err != nil {
		h.logger.Error("registration rate limit failed", "error", err)
		respondError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if !allowed {
		respondError(w, http.StatusTooManyRequests, "too many registration attempts")
		return
	}

	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	req.DisplayName = strings.TrimSpace(req.DisplayName)
	req.TenantName = strings.TrimSpace(req.TenantName)
	req.OrgName = strings.TrimSpace(req.OrgName)

	if err := auth.ValidateEmail(req.Email); err != nil {
		respondError(w, http.StatusBadRequest, "invalid email")
		return
	}
	if err := auth.ValidatePasswordStrength(req.Password); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	if req.DisplayName == "" {
		respondError(w, http.StatusBadRequest, "display_name is required")
		return
	}
	if req.Locale == "" {
		req.Locale = "en"
	}
	if req.Locale != "en" && req.Locale != "ar" {
		respondError(w, http.StatusBadRequest, "locale must be 'en' or 'ar'")
		return
	}
	if req.TenantName == "" {
		req.TenantName = req.DisplayName + "'s Workspace"
	}
	if req.OrgName == "" {
		req.OrgName = "Default Organization"
	}

	if _, err := h.userRepo.FindByEmail(ctx, req.Email); err == nil {
		respondError(w, http.StatusConflict, "email already registered")
		return
	} else if err != nil && err != pgx.ErrNoRows {
		h.logger.Error("failed to check existing user", "error", err)
		respondError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	passwordHash, err := auth.HashPassword(req.Password)
	if err != nil {
		h.logger.Error("failed to hash password", "error", err)
		respondError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	tx, err := h.db.Begin(ctx)
	if err != nil {
		h.logger.Error("failed to begin registration transaction", "error", err)
		respondError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	defer tx.Rollback(ctx)

	now := time.Now().UTC()
	createdAt := pgtype.Timestamptz{Time: now, Valid: true}
	tenantID := uuid.NewString()
	orgID := uuid.NewString()
	userID := uuid.NewString()
	membershipID := uuid.NewString()
	txQueries := h.queries.WithTx(tx)

	if err := txQueries.CreateTenant(ctx, sqldb.CreateTenantParams{
		ID:        tenantID,
		Name:      req.TenantName,
		Plan:      "free",
		CreatedAt: createdAt,
	}); err != nil {
		h.logger.Error("failed to create tenant", "error", err)
		respondError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	if err := txQueries.CreateOrganization(ctx, sqldb.CreateOrganizationParams{
		ID:        orgID,
		TenantID:  tenantID,
		Name:      req.OrgName,
		CreatedAt: createdAt,
	}); err != nil {
		h.logger.Error("failed to create organization", "error", err)
		respondError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	if err := txQueries.CreateUser(ctx, sqldb.CreateUserParams{
		ID:                  userID,
		TenantID:            tenantID,
		Email:               req.Email,
		PasswordHash:        &passwordHash,
		DisplayName:         req.DisplayName,
		Locale:              req.Locale,
		EmailVerified:       false,
		FailedLoginAttempts: 0,
		CreatedAt:           createdAt,
	}); err != nil {
		h.logger.Error("failed to create user", "error", err)
		respondError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	if err := txQueries.CreateMembership(ctx, sqldb.CreateMembershipParams{
		ID:        membershipID,
		UserID:    userID,
		OrgID:     orgID,
		Role:      sqldb.MembershipRoleOwner,
		CreatedAt: createdAt,
	}); err != nil {
		h.logger.Error("failed to create membership", "error", err)
		respondError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	if err := tx.Commit(ctx); err != nil {
		h.logger.Error("failed to commit registration transaction", "error", err)
		respondError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	if err := authz.EnsureTenantRoleAccess(h.authzEnforcer, userID, authz.RoleOwner, tenantID); err != nil {
		h.logger.Error("failed to seed tenant authorization after registration", "error", err, "tenant_id", tenantID, "user_id", userID)
		respondError(w, http.StatusInternalServerError, "failed to initialize authorization")
		return
	}

	tokens, err := h.jwtService.GenerateTokenPair(userID, req.Email, tenantID, orgID, "owner")
	if err != nil {
		h.logger.Error("failed to generate registration tokens", "error", err)
		respondError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	writeJSON(w, http.StatusCreated, AuthResponse{
		User: &UserResponse{
			ID:            userID,
			Email:         req.Email,
			DisplayName:   req.DisplayName,
			Locale:        req.Locale,
			EmailVerified: false,
			TenantID:      tenantID,
			CreatedAt:     now,
		},
		TokenPair: tokens,
	})
}
