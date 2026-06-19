// Package handler provides HTTP request handlers for the API.
//
// Deprecated: This file is deprecated — handlers are now organized into feature files:
// - health.go: HealthCheck, HealthReady
// - documents.go: Document handlers
// - search.go: Search
// - folders.go: Folder handlers
// - tags.go: Tag handlers
// - reminders.go: Reminder handlers
// - notifications.go: Notification handlers
// - audit.go: Audit log handlers
// - admin_members.go: Membership management handlers
// - admin_policies.go: Policy management handlers
// - respond.go: Shared response helpers
// - routes.go: Route registration
package handler

import (
	"github.com/casbin/casbin/v3"
	"github.com/docvault/backend/internal/config"
	"github.com/docvault/backend/internal/repository"
	"github.com/docvault/backend/internal/usecase"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Handler struct {
	cfg             *config.Config
	db              *pgxpool.Pool
	authzEnforcer   *casbin.Enforcer
	documentSvc     *usecase.DocumentService
	folderSvc       *usecase.FolderService
	tagSvc          *usecase.TagService
	auditSvc        *usecase.AuditService
	reminderSvc     *usecase.ReminderService
	notificationSvc *usecase.NotificationService
	searchSvc       *usecase.SearchService
	chatSvc         *usecase.ChatService
	userRepo        repository.UserRepository
	membershipRepo  repository.MembershipRepository
	policyRepo      repository.PolicyRepository
	aclRepo         repository.ACLRepository
	authz           *usecase.Authorizer
}

type Dependencies struct {
	DB              *pgxpool.Pool
	AuthzEnforcer   *casbin.Enforcer
	DocumentSvc     *usecase.DocumentService
	FolderSvc       *usecase.FolderService
	TagSvc          *usecase.TagService
	AuditSvc        *usecase.AuditService
	ReminderSvc     *usecase.ReminderService
	NotificationSvc *usecase.NotificationService
	SearchSvc       *usecase.SearchService
	ChatSvc         *usecase.ChatService
	UserRepo        repository.UserRepository
	MembershipRepo  repository.MembershipRepository
	PolicyRepo      repository.PolicyRepository
	ACLRepo         repository.ACLRepository
}

func New(cfg *config.Config, deps Dependencies) *Handler {
	return &Handler{
		cfg:             cfg,
		db:              deps.DB,
		authzEnforcer:   deps.AuthzEnforcer,
		documentSvc:     deps.DocumentSvc,
		folderSvc:       deps.FolderSvc,
		tagSvc:          deps.TagSvc,
		auditSvc:        deps.AuditSvc,
		reminderSvc:     deps.ReminderSvc,
		notificationSvc: deps.NotificationSvc,
		searchSvc:       deps.SearchSvc,
		chatSvc:         deps.ChatSvc,
		userRepo:        deps.UserRepo,
		membershipRepo:  deps.MembershipRepo,
		policyRepo:      deps.PolicyRepo,
		aclRepo:         deps.ACLRepo,
		authz:           usecase.NewAuthorizer(deps.ACLRepo),
	}
}
