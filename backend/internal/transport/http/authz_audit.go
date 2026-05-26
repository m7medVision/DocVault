package handler

import (
	"context"
	"fmt"

	"github.com/docvault/backend/internal/middleware"
	"github.com/docvault/backend/internal/usecase"
)

func (h *Handler) AuditAuthorizationDecision(ctx context.Context, object, action string, allowed bool) {
	if h == nil || h.auditSvc == nil {
		return
	}

	tenantID := middleware.GetTenantID(ctx)
	if tenantID == "" {
		return
	}

	actorID := middleware.GetUserID(ctx)
	metadata := map[string]interface{}{
		"object":  object,
		"action":  action,
		"allowed": allowed,
	}

	var actor *string
	if actorID != "" {
		actor = &actorID
	}

	_ = h.auditSvc.Write(ctx, &usecase.WriteAuditEventInput{
		TenantID:   tenantID,
		ActorID:    actor,
		EntityType: "authorization",
		EntityID:   fmt.Sprintf("%s:%s", object, action),
		Action:     usecase.AuditActionView,
		Metadata:   metadata,
	})
}
