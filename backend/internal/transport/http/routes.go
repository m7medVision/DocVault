package handler

import (
	"net/http"

	"github.com/casbin/casbin/v3"
	"github.com/docvault/backend/internal/authz"
	"github.com/docvault/backend/internal/middleware"
)

func RegisterRoutes(h *Handler, authHandler *AuthHandler, mux *http.ServeMux, jwtMiddleware *middleware.JWTMiddleware, authzEnforcer *casbin.Enforcer) {
	handle := func(pattern string, handlerFunc http.HandlerFunc, routeMiddlewares ...middleware.Middleware) {
		var routeHandler http.Handler = http.HandlerFunc(handlerFunc)
		for i := len(routeMiddlewares) - 1; i >= 0; i-- {
			routeHandler = routeMiddlewares[i](routeHandler)
		}
		mux.Handle(pattern, routeHandler)
	}

	handle("GET /health", h.HealthCheck)
	handle("GET /health/ready", h.HealthReady)

	handle("POST /api/v1/auth/register", authHandler.Register)
	handle("POST /api/v1/auth/login", authHandler.Login)
	handle("POST /api/v1/auth/refresh", authHandler.Refresh)
	handle("POST /api/v1/auth/logout", authHandler.Logout)
	handle("GET /api/v1/auth/me", authHandler.Me)
	handle("PATCH /api/v1/profile", authHandler.UpdateProfile)
	handle("PUT /api/v1/profile/email", authHandler.UpdateEmail)
	handle("PUT /api/v1/profile/password", authHandler.UpdatePassword)

	handle("POST /api/v1/documents/upload", h.UploadDocument, middleware.Authorize(authzEnforcer, authz.ResourceDocuments, authz.ActionWrite))
	handle("GET /api/v1/documents/stats", h.GetStats, middleware.Authorize(authzEnforcer, authz.ResourceDocuments, authz.ActionRead))
	handle("GET /api/v1/documents", h.ListDocuments, middleware.Authorize(authzEnforcer, authz.ResourceDocuments, authz.ActionRead))
	handle("GET /api/v1/documents/{id}", h.GetDocument, middleware.Authorize(authzEnforcer, authz.ResourceDocuments, authz.ActionRead))
	handle("DELETE /api/v1/documents/{id}", h.DeleteDocument, middleware.Authorize(authzEnforcer, authz.ResourceDocuments, authz.ActionDelete))
	handle("GET /api/v1/documents/{id}/download", h.DownloadDocument, middleware.Authorize(authzEnforcer, authz.ResourceDocuments, authz.ActionRead))
	handle("GET /api/v1/documents/{id}/versions", h.ListDocumentVersions, middleware.Authorize(authzEnforcer, authz.ResourceDocuments, authz.ActionRead))
	handle("PATCH /api/v1/documents/{id}/metadata", h.UpdateMetadata, middleware.Authorize(authzEnforcer, authz.ResourceDocuments, authz.ActionWrite))
	handle("GET /api/v1/documents/{id}/pages", h.GetDocumentPages, middleware.Authorize(authzEnforcer, authz.ResourceDocuments, authz.ActionRead))
	handle("PATCH /api/v1/documents/{id}/move", h.MoveDocument, middleware.Authorize(authzEnforcer, authz.ResourceDocuments, authz.ActionWrite))
	handle("PATCH /api/v1/documents/{id}/title", h.UpdateDocumentTitle, middleware.Authorize(authzEnforcer, authz.ResourceDocuments, authz.ActionWrite))
	handle("GET /api/v1/documents/{id}/progress/ws", h.DocumentProgressWS)

	handle("POST /api/v1/search", h.Search, middleware.Authorize(authzEnforcer, authz.ResourceSearch, authz.ActionRead))
	handle("POST /api/v1/chat", h.ChatGlobal, middleware.Authorize(authzEnforcer, authz.ResourceSearch, authz.ActionRead))
	handle("POST /api/v1/documents/{id}/chat", h.Chat, middleware.Authorize(authzEnforcer, authz.ResourceDocuments, authz.ActionRead))

	handle("POST /api/v1/folders", h.CreateFolder, middleware.Authorize(authzEnforcer, authz.ResourceFolders, authz.ActionWrite))
	handle("GET /api/v1/folders", h.ListFolders, middleware.Authorize(authzEnforcer, authz.ResourceFolders, authz.ActionRead))
	handle("GET /api/v1/folders/all", h.ListAllFolders, middleware.Authorize(authzEnforcer, authz.ResourceFolders, authz.ActionRead))
	handle("PATCH /api/v1/folders/{id}", h.RenameFolder, middleware.Authorize(authzEnforcer, authz.ResourceFolders, authz.ActionWrite))
	handle("DELETE /api/v1/folders/{id}", h.DeleteFolder, middleware.Authorize(authzEnforcer, authz.ResourceFolders, authz.ActionDelete))

	handle("GET /api/v1/tags", h.ListTags, middleware.Authorize(authzEnforcer, authz.ResourceTags, authz.ActionRead))

	handle("GET /api/v1/reminders", h.ListReminders, middleware.Authorize(authzEnforcer, authz.ResourceReminders, authz.ActionRead))
	handle("POST /api/v1/documents/{id}/reminders", h.CreateReminder, middleware.Authorize(authzEnforcer, authz.ResourceReminders, authz.ActionWrite))
	handle("PATCH /api/v1/reminders/{id}", h.UpdateReminder, middleware.Authorize(authzEnforcer, authz.ResourceReminders, authz.ActionWrite))

	handle("GET /api/v1/notifications", h.ListNotifications, middleware.Authorize(authzEnforcer, authz.ResourceNotifications, authz.ActionRead))
	handle("PATCH /api/v1/notifications/{id}/read", h.MarkNotificationRead, middleware.Authorize(authzEnforcer, authz.ResourceNotifications, authz.ActionWrite))
	handle("POST /api/v1/notifications/webhook", h.CreateNotificationWebhook, middleware.Authorize(authzEnforcer, authz.ResourceNotifications, authz.ActionWrite))

	handle("GET /api/v1/admin/audit", h.GetAuditLog, middleware.Authorize(authzEnforcer, authz.ResourceAdminAudit, authz.ActionRead))
	handle("GET /api/v1/admin/members", h.ListMembers, middleware.Authorize(authzEnforcer, authz.ResourceAdminMembers, authz.ActionRead))
	handle("GET /api/v1/admin/members/{id}/permissions", h.GetMemberPermissions, middleware.Authorize(authzEnforcer, authz.ResourceAdminMembers, authz.ActionRead))
	handle("POST /api/v1/admin/members/invite", h.InviteMember, middleware.Authorize(authzEnforcer, authz.ResourceAdminMembers, authz.ActionInvite))
	handle("PATCH /api/v1/admin/members/{id}/role", h.UpdateMemberRole, middleware.Authorize(authzEnforcer, authz.ResourceAdminMembers, authz.ActionUpdateRole))
	handle("GET /api/v1/admin/casbin/policies", h.ListPolicies, middleware.Authorize(authzEnforcer, authz.ResourceAdminPolicies, authz.ActionRead))
	handle("POST /api/v1/admin/casbin/policies", h.AddPolicy, middleware.Authorize(authzEnforcer, authz.ResourceAdminPolicies, authz.ActionWrite))
	handle("DELETE /api/v1/admin/casbin/policies", h.DeletePolicy, middleware.Authorize(authzEnforcer, authz.ResourceAdminPolicies, authz.ActionDelete))

	handle("PUT /api/v1/documents/{id}/restrict", h.SetDocumentRestricted, middleware.Authorize(authzEnforcer, authz.ResourceACL, authz.ActionWrite))
	handle("DELETE /api/v1/documents/{id}/restrict", h.UnsetDocumentRestricted, middleware.Authorize(authzEnforcer, authz.ResourceACL, authz.ActionWrite))
	handle("PUT /api/v1/folders/{id}/restrict", h.SetFolderRestricted, middleware.Authorize(authzEnforcer, authz.ResourceACL, authz.ActionWrite))
	handle("DELETE /api/v1/folders/{id}/restrict", h.UnsetFolderRestricted, middleware.Authorize(authzEnforcer, authz.ResourceACL, authz.ActionWrite))

	handle("POST /api/v1/acl/grants", h.CreateGrant, middleware.Authorize(authzEnforcer, authz.ResourceACL, authz.ActionWrite))
	handle("DELETE /api/v1/acl/grants/{id}", h.DeleteGrant, middleware.Authorize(authzEnforcer, authz.ResourceACL, authz.ActionDelete))
	handle("GET /api/v1/acl/grants", h.ListGrants, middleware.Authorize(authzEnforcer, authz.ResourceACL, authz.ActionRead))

	handle("POST /api/v1/groups", h.CreateGroup, middleware.Authorize(authzEnforcer, authz.ResourceGroups, authz.ActionWrite))
	handle("DELETE /api/v1/groups/{id}", h.DeleteGroup, middleware.Authorize(authzEnforcer, authz.ResourceGroups, authz.ActionDelete))
	handle("GET /api/v1/groups", h.ListGroups, middleware.Authorize(authzEnforcer, authz.ResourceGroups, authz.ActionRead))
	handle("POST /api/v1/groups/{id}/members", h.AddGroupMember, middleware.Authorize(authzEnforcer, authz.ResourceGroups, authz.ActionWrite))
	handle("DELETE /api/v1/groups/{id}/members/{user_id}", h.RemoveGroupMember, middleware.Authorize(authzEnforcer, authz.ResourceGroups, authz.ActionWrite))
}
