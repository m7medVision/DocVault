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
	auditapp "github.com/docvault/backend/internal/audit/app"
	"github.com/docvault/backend/internal/config"
	documentapp "github.com/docvault/backend/internal/document/app"
	notificationapp "github.com/docvault/backend/internal/notification/app"
	reminderapp "github.com/docvault/backend/internal/reminder/app"
	"github.com/docvault/backend/internal/repository"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Handler struct {
	cfg             *config.Config
	db              *pgxpool.Pool
	authzEnforcer   *casbin.Enforcer
	documentSvc     *documentapp.DocumentService
	folderSvc       *documentapp.FolderService
	tagSvc          *documentapp.TagService
	auditSvc        *auditapp.AuditService
	reminderSvc     *reminderapp.ReminderService
	notificationSvc *notificationapp.NotificationService
	searchSvc       *documentapp.SearchService
	chatSvc         *documentapp.ChatService
	suggestionSvc   *documentapp.SuggestionService
	userRepo        repository.UserRepository
	membershipRepo  repository.MembershipRepository
	policyRepo      repository.PolicyRepository
	aclRepo         repository.ACLRepository
	authz           *documentapp.Authorizer
}

type Dependencies struct {
	DB              *pgxpool.Pool
	AuthzEnforcer   *casbin.Enforcer
	DocumentSvc     *documentapp.DocumentService
	FolderSvc       *documentapp.FolderService
	TagSvc          *documentapp.TagService
	AuditSvc        *auditapp.AuditService
	ReminderSvc     *reminderapp.ReminderService
	NotificationSvc *notificationapp.NotificationService
	SearchSvc       *documentapp.SearchService
	ChatSvc         *documentapp.ChatService
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
		suggestionSvc:   documentapp.NewSuggestionService(deps.DocumentSvc, deps.FolderSvc),
		userRepo:        deps.UserRepo,
		membershipRepo:  deps.MembershipRepo,
		policyRepo:      deps.PolicyRepo,
		aclRepo:         deps.ACLRepo,
		authz:           documentapp.NewAuthorizer(deps.ACLRepo),
	}
}
