package app

import (
	"github.com/casbin/casbin/v3"
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
		Reminder:     repository.NewReminderRepository(db),
		Folder:       repository.NewFolderRepository(db),
		Tag:          repository.NewTagRepository(db),
		Audit:        repository.NewAuditRepository(db),
		Notification: repository.NewNotificationRepository(db),
		User:         repository.NewUserRepository(db),
		Membership:   repository.NewMembershipRepository(db),
		Policy:       repository.NewPolicyRepository(enforcer),
		Search:       repository.NewSearchRepository(db),
		ACL:          repository.NewACLRepository(db),
	}
}
