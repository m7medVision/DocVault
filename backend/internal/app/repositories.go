package app

import (
	"github.com/casbin/casbin/v3"
	auditpg "github.com/docvault/backend/internal/audit/adapter/postgres"
	identitypg "github.com/docvault/backend/internal/identity/adapter/postgres"
	notificationpg "github.com/docvault/backend/internal/notification/adapter/postgres"
	reminderpg "github.com/docvault/backend/internal/reminder/adapter/postgres"
	"github.com/docvault/backend/internal/repository"
	"github.com/jackc/pgx/v5/pgxpool"
)

// buildRepositories wires every concrete repository against the shared database
// pool (and the Casbin enforcer for the policy store) and returns them in the
// Repositories holder the rest of the composition root consumes.
//
// It lives here, in the composition root, rather than in the repository package
// so that package can stay a pure leaf of contracts (interfaces, DTOs, error
// sentinels) that both the usecase layer and the per-context postgres adapters
// import without forming a cycle: the adapters depend on the contracts, and the
// composition root depends on the adapters.
func buildRepositories(db *pgxpool.Pool, enforcer *casbin.Enforcer) *repository.Repositories {
	return &repository.Repositories{
		Document:     repository.NewDocumentRepository(db),
		Reminder:     reminderpg.NewReminderRepository(db),
		Folder:       repository.NewFolderRepository(db),
		Tag:          repository.NewTagRepository(db),
		Audit:        auditpg.NewAuditRepository(db),
		Notification: notificationpg.NewNotificationRepository(db),
		User:         identitypg.NewUserRepository(db),
		Membership:   identitypg.NewMembershipRepository(db),
		Policy:       repository.NewPolicyRepository(enforcer),
		Search:       repository.NewSearchRepository(db),
		ACL:          repository.NewACLRepository(db),
	}
}
