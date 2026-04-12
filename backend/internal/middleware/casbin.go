package middleware

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/casbin/casbin/v3"
)

type AuthorizationAuditFunc func(ctx context.Context, object, action string, allowed bool)

var authorizationAuditLogger AuthorizationAuditFunc

func SetAuthorizationAuditLogger(fn AuthorizationAuditFunc) {
	authorizationAuditLogger = fn
}

func Authorize(enforcer *casbin.Enforcer, object, action string) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if enforcer == nil {
				writeAuthzError(w, http.StatusInternalServerError, "authorization backend unavailable")
				return
			}

			userID := GetUserID(r.Context())
			if userID == "" {
				writeAuthzError(w, http.StatusUnauthorized, "not authenticated")
				return
			}

			tenantID := GetTenantID(r.Context())
			if tenantID == "" {
				writeAuthzError(w, http.StatusForbidden, "tenant context required")
				return
			}

			allowed, matchedPolicy, err := enforcer.EnforceEx(userID, tenantID, object, action)
			if err != nil {
				slog.Error("authorization check failed", "error", err, "user_id", userID, "tenant_id", tenantID, "object", object, "action", action)
				writeAuthzError(w, http.StatusInternalServerError, "authorization error")
				return
			}

			if !allowed {
				slog.Warn("authorization denied", "user_id", userID, "tenant_id", tenantID, "object", object, "action", action, "matched_policy", matchedPolicy)
				if authorizationAuditLogger != nil {
					authorizationAuditLogger(r.Context(), object, action, false)
				}
				writeAuthzError(w, http.StatusForbidden, "insufficient permissions")
				return
			}

			slog.Info("authorization allowed", "user_id", userID, "tenant_id", tenantID, "object", object, "action", action, "matched_policy", matchedPolicy)
			if authorizationAuditLogger != nil {
				authorizationAuditLogger(r.Context(), object, action, true)
			}

			next.ServeHTTP(w, r)
		})
	}
}

func writeAuthzError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": message})
}
