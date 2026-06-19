package handler

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/docvault/backend/internal/auth"
	"github.com/docvault/backend/internal/authz"
	sqldb "github.com/docvault/backend/internal/db"
	"github.com/docvault/backend/internal/identity"
	"github.com/docvault/backend/internal/middleware"
	appredis "github.com/docvault/backend/internal/redis"
	"github.com/jackc/pgx/v5"
)

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type AuthResponse struct {
	User      *UserResponse   `json:"user"`
	TokenPair *auth.TokenPair `json:"tokens"`
}

type UserResponse struct {
	ID            string    `json:"id"`
	Email         string    `json:"email"`
	DisplayName   string    `json:"display_name"`
	Locale        string    `json:"locale"`
	EmailVerified bool      `json:"email_verified"`
	TenantID      string    `json:"tenant_id"`
	CreatedAt     time.Time `json:"created_at"`
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	allowed, _, _, err := h.rateLimiter.Allow(
		ctx,
		appredis.BuildRateLimitKey("auth:login", clientIP(r)),
		appredis.LoginRateLimit,
	)
	if err != nil {
		h.logger.Error("login rate limit failed", "error", err)
		respondError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if !allowed {
		respondError(w, http.StatusTooManyRequests, "too many login attempts")
		return
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	if err := auth.ValidateEmail(req.Email); err != nil || req.Password == "" {
		respondError(w, http.StatusBadRequest, "invalid email or password")
		return
	}

	user, err := h.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		if err == pgx.ErrNoRows {
			respondError(w, http.StatusUnauthorized, "invalid email or password")
			return
		}
		h.logger.Error("failed to load user for login", "error", err)
		respondError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	if user.LockedUntil != nil && user.LockedUntil.After(time.Now().UTC()) {
		respondError(w, http.StatusForbidden, "account is temporarily locked")
		return
	}

	if err := auth.ComparePassword(user.PasswordHash, req.Password); err != nil {
		if updateErr := h.recordFailedLogin(ctx, user); updateErr != nil {
			h.logger.Error("failed to record login failure", "error", updateErr, "user_id", user.ID)
		}
		respondError(w, http.StatusUnauthorized, "invalid email or password")
		return
	}

	orgID, role, err := h.lookupPrimaryMembership(ctx, user.ID)
	if err != nil {
		h.logger.Error("failed to load membership for login", "error", err, "user_id", user.ID)
		respondError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	if err := authz.EnsureTenantRoleAccess(h.authzEnforcer, user.ID, role, user.TenantID); err != nil {
		h.logger.Error("failed to ensure tenant authorization on login", "error", err, "tenant_id", user.TenantID, "user_id", user.ID)
		respondError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	now := time.Now().UTC()
	if err := h.queries.UpdateSuccessfulLoginMetadata(ctx, sqldb.UpdateSuccessfulLoginMetadataParams{
		LastLoginAt: &now,
		ID:          user.ID,
	}); err != nil {
		h.logger.Warn("failed to update successful login metadata", "error", err, "user_id", user.ID)
	}

	tokens, err := h.jwtService.GenerateTokenPair(user.ID, user.Email, user.TenantID, orgID, role)
	if err != nil {
		h.logger.Error("failed to generate login tokens", "error", err)
		respondError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	writeJSON(w, http.StatusOK, AuthResponse{
		User: &UserResponse{
			ID:            user.ID,
			Email:         user.Email,
			DisplayName:   user.DisplayName,
			Locale:        user.Locale,
			EmailVerified: user.EmailVerified,
			TenantID:      user.TenantID,
			CreatedAt:     user.CreatedAt,
		},
		TokenPair: tokens,
	})
}

func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.RefreshToken == "" {
		respondError(w, http.StatusBadRequest, "refresh token is required")
		return
	}

	userID, tokenID, err := h.jwtService.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		respondError(w, http.StatusUnauthorized, "invalid refresh token")
		return
	}

	blacklisted, err := h.tokenBlacklist.IsBlacklisted(ctx, tokenID)
	if err != nil {
		h.logger.Error("failed to check refresh blacklist", "error", err)
		respondError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if blacklisted {
		respondError(w, http.StatusUnauthorized, "refresh token has been revoked")
		return
	}

	user, err := h.userRepo.FindByID(ctx, userID)
	if err != nil {
		if err == pgx.ErrNoRows {
			respondError(w, http.StatusUnauthorized, "user not found")
			return
		}
		h.logger.Error("failed to load user for refresh", "error", err)
		respondError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	orgID, role, err := h.lookupPrimaryMembership(ctx, user.ID)
	if err != nil {
		h.logger.Error("failed to load membership for refresh", "error", err, "user_id", user.ID)
		respondError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	tokens, err := h.jwtService.GenerateTokenPair(user.ID, user.Email, user.TenantID, orgID, role)
	if err != nil {
		h.logger.Error("failed to generate refreshed tokens", "error", err)
		respondError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	oldTokenID, expiry, err := h.jwtService.ExtractTokenID(req.RefreshToken)
	if err == nil {
		if blacklistErr := h.blacklistToken(ctx, oldTokenID, expiry); blacklistErr != nil {
			h.logger.Warn("failed to revoke previous refresh token", "error", blacklistErr)
		}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"tokens": tokens})
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.RefreshToken == "" {
		respondError(w, http.StatusBadRequest, "refresh token is required")
		return
	}

	tokenID, expiry, err := h.jwtService.ExtractTokenID(req.RefreshToken)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid refresh token")
		return
	}

	if err := h.blacklistToken(ctx, tokenID, expiry); err != nil {
		h.logger.Error("failed to blacklist refresh token", "error", err)
		respondError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "logged out successfully"})
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserClaims(r)
	if claims == nil {
		respondError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	user, err := h.userRepo.FindByID(r.Context(), claims.UserID)
	if err != nil {
		if err == pgx.ErrNoRows {
			respondError(w, http.StatusUnauthorized, "user not found")
			return
		}
		h.logger.Error("failed to load current user", "error", err)
		respondError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	writeJSON(w, http.StatusOK, UserResponse{
		ID:            user.ID,
		Email:         user.Email,
		DisplayName:   user.DisplayName,
		Locale:        user.Locale,
		EmailVerified: user.EmailVerified,
		TenantID:      user.TenantID,
		CreatedAt:     user.CreatedAt,
	})
}

func (h *AuthHandler) lookupPrimaryMembership(ctx context.Context, userID string) (string, string, error) {
	membership, err := h.queries.GetPrimaryMembershipByUserID(ctx, sqldb.GetPrimaryMembershipByUserIDParams{UserID: userID})
	return membership.OrgID, string(membership.Role), err
}

func (h *AuthHandler) recordFailedLogin(ctx context.Context, user *identity.User) error {
	attempts := user.FailedLoginAttempts + 1
	var lockedUntil *string
	if attempts >= 5 {
		until := time.Now().UTC().Add(15 * time.Minute).Format(time.RFC3339)
		lockedUntil = &until
	}

	return h.userRepo.UpdateFailedLogin(ctx, user.ID, attempts, lockedUntil)
}

func (h *AuthHandler) blacklistToken(ctx context.Context, tokenID string, expiry time.Time) error {
	ttl := time.Until(expiry)
	if ttl <= 0 {
		return nil
	}
	return h.tokenBlacklist.BlacklistToken(ctx, tokenID, ttl)
}

func clientIP(r *http.Request) string {
	if forwarded := strings.TrimSpace(strings.Split(r.Header.Get("X-Forwarded-For"), ",")[0]); forwarded != "" {
		return forwarded
	}
	if realIP := strings.TrimSpace(r.Header.Get("X-Real-IP")); realIP != "" {
		return realIP
	}
	host, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr))
	if err == nil && host != "" {
		return host
	}
	return strings.TrimSpace(r.RemoteAddr)
}
