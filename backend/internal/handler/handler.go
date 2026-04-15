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
	"github.com/docvault/backend/internal/search"
	"github.com/docvault/backend/internal/service"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Handler struct {
	cfg             *config.Config
	db              *pgxpool.Pool
	authzEnforcer   *casbin.Enforcer
	documentSvc     *service.DocumentService
	folderSvc       *service.FolderService
	tagSvc          *service.TagService
	auditSvc        *service.AuditService
	reminderSvc     *service.ReminderService
	notificationSvc *service.NotificationService
	searchSvc       *service.SearchService
	aiSvc           *service.AIService
	userRepo        repository.UserRepository
	membershipRepo  repository.MembershipRepository
	policyRepo      repository.PolicyRepository
}

func New(cfg *config.Config, repos *repository.Repositories, dbPool *pgxpool.Pool, objectStore service.ObjectStore, publisher service.QueuePublisher, enforcer *casbin.Enforcer) *Handler {
	return &Handler{
		cfg:             cfg,
		db:              dbPool,
		authzEnforcer:   enforcer,
		documentSvc:     service.NewDocumentService(repos.Document, objectStore, publisher),
		folderSvc:       service.NewFolderService(repos.Folder),
		tagSvc:          service.NewTagService(repos.Tag),
		auditSvc:        service.NewAuditService(repos.Audit),
		reminderSvc:     service.NewReminderService(repos.Reminder),
		notificationSvc: service.NewNotificationService(repos.Notification),
		searchSvc:       service.NewSearchService(search.NewOpenRouterEmbedder(cfg.OpenRouterAPIKey, cfg.EmbeddingModel), repos.Search),
		aiSvc:           service.NewAIService(cfg.OpenRouterAPIKey, "openrouter/chat/gpt-4o-mini"),
		userRepo:        repos.User,
		membershipRepo:  repos.Membership,
		policyRepo:      repos.Policy,
	}
}
