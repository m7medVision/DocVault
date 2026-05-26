package handler

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/docvault/backend/internal/authz"
	"github.com/docvault/backend/internal/middleware"
	"github.com/docvault/backend/internal/usecase"
)

type policyRequest struct {
	Subject string `json:"subject"`
	Object  string `json:"object"`
	Action  string `json:"action"`
}

func (h *Handler) ListPolicies(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := middleware.GetTenantID(ctx)
	actorID := middleware.GetUserID(ctx)

	if tenantID == "" || actorID == "" {
		h.writeMissingAuthContext(w)
		return
	}
	if h.authzEnforcer == nil {
		h.writeMissingAuthzDependency(w)
		return
	}

	policies, err := authz.ListPolicyRules(h.authzEnforcer, tenantID)
	if err != nil {
		slog.Error("list policies failed", "error", err, "tenant_id", tenantID)
		http.Error(w, `{"error":"failed to list policies","code":"INTERNAL_ERROR"}`, http.StatusInternalServerError)
		return
	}

	bindings, err := authz.ListRoleBindings(h.authzEnforcer, tenantID)
	if err != nil {
		slog.Error("list role bindings failed", "error", err, "tenant_id", tenantID)
		http.Error(w, `{"error":"failed to list policies","code":"INTERNAL_ERROR"}`, http.StatusInternalServerError)
		return
	}

	_ = h.auditSvc.Write(ctx, &usecase.WriteAuditEventInput{
		TenantID:   tenantID,
		ActorID:    &actorID,
		EntityType: "authorization_policy",
		EntityID:   tenantID,
		Action:     usecase.AuditActionView,
		Metadata: map[string]interface{}{
			"policy_count":  len(policies),
			"binding_count": len(bindings),
		},
	})

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"policies": policies,
		"bindings": bindings,
	})
}

func (h *Handler) AddPolicy(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := middleware.GetTenantID(ctx)
	actorID := middleware.GetUserID(ctx)

	if tenantID == "" || actorID == "" {
		h.writeMissingAuthContext(w)
		return
	}
	if h.authzEnforcer == nil {
		h.writeMissingAuthzDependency(w)
		return
	}

	var body policyRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, `{"error":"invalid JSON body","code":"BAD_REQUEST"}`, http.StatusBadRequest)
		return
	}

	if err := authz.AddPolicyRule(h.authzEnforcer, body.Subject, tenantID, body.Object, body.Action); err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"%s","code":"BAD_REQUEST"}`, err.Error()), http.StatusBadRequest)
		return
	}

	_ = h.auditSvc.Write(ctx, &usecase.WriteAuditEventInput{
		TenantID:   tenantID,
		ActorID:    &actorID,
		EntityType: "authorization_policy",
		EntityID:   fmt.Sprintf("%s:%s:%s", body.Subject, body.Object, body.Action),
		Action:     usecase.AuditActionCreate,
		Metadata: map[string]interface{}{
			"subject": body.Subject,
			"object":  body.Object,
			"action":  body.Action,
		},
	})

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"policy": authz.PolicyRule{
			Subject:  body.Subject,
			TenantID: tenantID,
			Object:   body.Object,
			Action:   body.Action,
		},
	})
}

func (h *Handler) DeletePolicy(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := middleware.GetTenantID(ctx)
	actorID := middleware.GetUserID(ctx)

	if tenantID == "" || actorID == "" {
		h.writeMissingAuthContext(w)
		return
	}
	if h.authzEnforcer == nil {
		h.writeMissingAuthzDependency(w)
		return
	}

	var body policyRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, `{"error":"invalid JSON body","code":"BAD_REQUEST"}`, http.StatusBadRequest)
		return
	}

	if err := authz.RemovePolicyRule(h.authzEnforcer, body.Subject, tenantID, body.Object, body.Action); err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"%s","code":"BAD_REQUEST"}`, err.Error()), http.StatusBadRequest)
		return
	}

	_ = h.auditSvc.Write(ctx, &usecase.WriteAuditEventInput{
		TenantID:   tenantID,
		ActorID:    &actorID,
		EntityType: "authorization_policy",
		EntityID:   fmt.Sprintf("%s:%s:%s", body.Subject, body.Object, body.Action),
		Action:     usecase.AuditActionDelete,
		Metadata: map[string]interface{}{
			"subject": body.Subject,
			"object":  body.Object,
			"action":  body.Action,
		},
	})

	writeJSON(w, http.StatusOK, map[string]string{"message": "policy removed successfully"})
}
