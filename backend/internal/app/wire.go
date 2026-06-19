package app

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/casbin/casbin/v3"
	auditapp "github.com/docvault/backend/internal/audit/app"
	internalauth "github.com/docvault/backend/internal/auth"
	"github.com/docvault/backend/internal/config"
	documentpg "github.com/docvault/backend/internal/document/adapter/postgres"
	documentapp "github.com/docvault/backend/internal/document/app"
	identitypg "github.com/docvault/backend/internal/identity/adapter/postgres"
	"github.com/docvault/backend/internal/middleware"
	notificationapp "github.com/docvault/backend/internal/notification/app"
	"github.com/docvault/backend/internal/platform/cache"
	appredis "github.com/docvault/backend/internal/redis"
	reminderapp "github.com/docvault/backend/internal/reminder/app"
	"github.com/docvault/backend/internal/repository"
	"github.com/docvault/backend/internal/search"
	handler "github.com/docvault/backend/internal/transport/http"
)

// buildHandlers wires the repositories (with their cache decorators), the
// domain services, and the two HTTP handlers. It is the heart of the
// composition root: every dependency the API serves traces back to here.
func buildHandlers(
	cfg *config.Config,
	inf *infra,
	jwtService *internalauth.JWTService,
	tokenBlacklist *appredis.TokenBlacklist,
	rateLimiter *appredis.RateLimiter,
	authzEnforcer *casbin.Enforcer,
	logger *slog.Logger,
) (*handler.Handler, *handler.AuthHandler) {
	repos := buildRepositories(inf.db, authzEnforcer)

	// One Redis-backed cache shared by the repository decorators below; each uses
	// a distinct key prefix so namespaces never collide.
	appCache := cache.NewRedis(inf.redis.Client)
	// Cache the per-(org,user) group-membership lookup the authorizer runs on
	// nearly every non-admin request. Membership mutations bust the affected
	// keys; a short TTL bounds staleness. Wrapped once here so every consumer
	// (document/folder services, the handler's authorizer) shares the cache.
	aclRepo := documentpg.NewCachingACL(repos.ACL, appCache)
	// Cache the per-org document-stats aggregate; document create/delete bust it,
	// a short TTL covers status changes made by the processing workers.
	documentRepo := documentpg.NewCachingDocuments(repos.Document, appCache)
	// Cache the per-org folder tree read on every navigation render; structural
	// folder mutations bust it, a short TTL covers is_restricted display toggles.
	folderRepo := documentpg.NewCachingFolders(repos.Folder, appCache)

	embedder := search.NewOpenRouterEmbedder(cfg.Search.EmbeddingAPIKey, cfg.Search.EmbeddingModel, cfg.Search.EmbeddingDim)
	// Cache query embeddings in Redis: identical /search and /chat queries reuse
	// the vector instead of re-calling the external embedding API. The key is
	// content-addressed (model+dim+text) and carries no tenant data, so it is
	// safe to share across tenants. Best-effort: a cache failure falls through
	// to the embedder.
	queryEmbedder := search.NewCachingEmbedder(embedder, cache.NewRedis(inf.redis.Client), cfg.Search.EmbeddingModel, cfg.Search.EmbeddingDim, 7*24*time.Hour)

	h := handler.New(cfg, handler.Dependencies{
		DB:              inf.db,
		AuthzEnforcer:   authzEnforcer,
		DocumentSvc:     documentapp.NewDocumentService(documentRepo, aclRepo, inf.objectStore, inf.ocrDispatcher),
		FolderSvc:       documentapp.NewFolderService(folderRepo, aclRepo),
		TagSvc:          documentapp.NewTagService(repos.Tag),
		AuditSvc:        auditapp.NewAuditService(repos.Audit),
		ReminderSvc:     reminderapp.NewReminderService(repos.Reminder),
		NotificationSvc: notificationapp.NewNotificationService(repos.Notification),
		SearchSvc:       documentapp.NewSearchService(queryEmbedder, repos.Search),
		ChatSvc:         newChatService(queryEmbedder, repos.Search, cfg.Search.RerankURL),
		UserRepo:        repos.User,
		MembershipRepo:  repos.Membership,
		PolicyRepo:      repos.Policy,
		ACLRepo:         aclRepo,
	})
	middleware.SetAuthorizationAuditLogger(h.AuditAuthorizationDecision)

	authHandler := handler.NewAuthHandler(inf.db, jwtService, tokenBlacklist, rateLimiter, authzEnforcer, logger, repos.User, identitypg.NewRegistrationRepository(inf.db))

	return h, authHandler
}

// buildRouter mounts the routes on a fresh mux and wraps it in the middleware
// chain (telemetry, request id, logging, recovery, CORS, JWT authentication).
func buildRouter(cfg *config.Config, h *handler.Handler, authHandler *handler.AuthHandler, jwtMiddleware *middleware.JWTMiddleware, authzEnforcer *casbin.Enforcer) http.Handler {
	mux := http.NewServeMux()
	handler.RegisterRoutes(h, authHandler, mux, jwtMiddleware, authzEnforcer)
	return middleware.Chain(
		mux,
		middleware.Telemetry(),
		middleware.RequestID(),
		middleware.Logger(),
		middleware.Recoverer(),
		middleware.CORS(cfg.Server.CORSOrigins),
		jwtMiddleware.Authenticate,
	)
}

// newServer builds the HTTP server. It deliberately leaves WriteTimeout unset:
// /chat streams Server-Sent Events that routinely outlast any fixed deadline,
// and a global WriteTimeout silently truncates the stream mid-response.
// Per-request bounds come from the handler context instead. ReadHeaderTimeout
// guards against slow-loris header attacks; ReadTimeout is generous enough to
// read large (up to 50MB) multipart uploads over slow mobile links.
func newServer(cfg *config.Config, router http.Handler) *http.Server {
	return &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:           router,
		ReadHeaderTimeout: 15 * time.Second,
		ReadTimeout:       5 * time.Minute,
		WriteTimeout:      0,
		IdleTimeout:       60 * time.Second,
	}
}

// newChatService builds the chat service with an optional cross-encoder
// reranker. An empty rerankURL leaves the noop reranker in place, so chat works
// with no sidecar; setting it points at a TEI /rerank endpoint.
func newChatService(embedder search.Embedder, repo repository.SearchRepository, rerankURL string) *documentapp.ChatService {
	svc := documentapp.NewChatService(embedder, repo)
	if rerankURL != "" {
		svc = svc.WithReranker(documentapp.NewHTTPReranker(rerankURL, nil))
	}
	return svc
}
