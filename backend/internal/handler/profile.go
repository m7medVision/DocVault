package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/docvault/backend/internal/auth"
	"github.com/docvault/backend/internal/middleware"
	"github.com/docvault/backend/internal/model"
	"github.com/jackc/pgx/v5"
)

type UpdateProfileRequest struct {
	DisplayName *string `json:"display_name"`
	Locale      *string `json:"locale"`
}

type UpdateEmailRequest struct {
	Email           string `json:"email"`
	CurrentPassword string `json:"current_password"`
}

type UpdatePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

func (h *AuthHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserClaims(r)
	if claims == nil {
		respondError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	var req UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.DisplayName == nil && req.Locale == nil {
		respondError(w, http.StatusBadRequest, "at least one profile field is required")
		return
	}

	user, err := h.userRepo.FindByID(r.Context(), claims.UserID)
	if err != nil {
		if err == pgx.ErrNoRows {
			respondError(w, http.StatusUnauthorized, "user not found")
			return
		}
		h.logger.Error("failed to load current user for profile update", "error", err, "user_id", claims.UserID)
		respondError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	displayName := user.DisplayName
	if req.DisplayName != nil {
		displayName = strings.TrimSpace(*req.DisplayName)
		if displayName == "" {
			respondError(w, http.StatusBadRequest, "display_name is required")
			return
		}
	}

	locale := user.Locale
	if req.Locale != nil {
		locale = strings.TrimSpace(*req.Locale)
		if locale != "en" && locale != "ar" {
			respondError(w, http.StatusBadRequest, "locale must be 'en' or 'ar'")
			return
		}
	}

	if err := h.userRepo.UpdateProfile(r.Context(), user.ID, displayName, locale); err != nil {
		h.logger.Error("failed to update user profile", "error", err, "user_id", user.ID)
		respondError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	user.DisplayName = displayName
	user.Locale = locale
	writeJSON(w, http.StatusOK, userResponseFromModel(user))
}

func (h *AuthHandler) UpdateEmail(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserClaims(r)
	if claims == nil {
		respondError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	var req UpdateEmailRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	if err := auth.ValidateEmail(req.Email); err != nil {
		respondError(w, http.StatusBadRequest, "invalid email")
		return
	}
	if req.CurrentPassword == "" {
		respondError(w, http.StatusBadRequest, "current_password is required")
		return
	}

	user, err := h.userRepo.FindByID(r.Context(), claims.UserID)
	if err != nil {
		if err == pgx.ErrNoRows {
			respondError(w, http.StatusUnauthorized, "user not found")
			return
		}
		h.logger.Error("failed to load current user for email update", "error", err, "user_id", claims.UserID)
		respondError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	if err := auth.ComparePassword(user.PasswordHash, req.CurrentPassword); err != nil {
		respondError(w, http.StatusUnauthorized, "invalid current password")
		return
	}

	taken, err := h.userRepo.IsEmailTakenByOther(r.Context(), req.Email, user.ID)
	if err != nil {
		h.logger.Error("failed to check email availability", "error", err, "user_id", user.ID)
		respondError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if taken {
		respondError(w, http.StatusConflict, "email already registered")
		return
	}

	if err := h.userRepo.UpdateEmail(r.Context(), user.ID, req.Email); err != nil {
		h.logger.Error("failed to update user email", "error", err, "user_id", user.ID)
		respondError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	user.Email = req.Email
	user.EmailVerified = false
	writeJSON(w, http.StatusOK, userResponseFromModel(user))
}

func (h *AuthHandler) UpdatePassword(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserClaims(r)
	if claims == nil {
		respondError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	var req UpdatePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.CurrentPassword == "" {
		respondError(w, http.StatusBadRequest, "current_password is required")
		return
	}
	if err := auth.ValidatePasswordStrength(req.NewPassword); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	user, err := h.userRepo.FindByID(r.Context(), claims.UserID)
	if err != nil {
		if err == pgx.ErrNoRows {
			respondError(w, http.StatusUnauthorized, "user not found")
			return
		}
		h.logger.Error("failed to load current user for password update", "error", err, "user_id", claims.UserID)
		respondError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	if err := auth.ComparePassword(user.PasswordHash, req.CurrentPassword); err != nil {
		respondError(w, http.StatusUnauthorized, "invalid current password")
		return
	}

	passwordHash, err := auth.HashPassword(req.NewPassword)
	if err != nil {
		h.logger.Error("failed to hash updated password", "error", err, "user_id", user.ID)
		respondError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	if err := h.userRepo.UpdatePassword(r.Context(), user.ID, passwordHash); err != nil {
		h.logger.Error("failed to update user password", "error", err, "user_id", user.ID)
		respondError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	writeJSON(w, http.StatusOK, userResponseFromModel(user))
}

func userResponseFromModel(user *model.User) UserResponse {
	return UserResponse{
		ID:            user.ID,
		Email:         user.Email,
		DisplayName:   user.DisplayName,
		Locale:        user.Locale,
		EmailVerified: user.EmailVerified,
		TenantID:      user.TenantID,
		CreatedAt:     user.CreatedAt,
	}
}
