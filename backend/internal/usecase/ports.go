package usecase

import (
	"context"

	"github.com/docvault/backend/internal/repository"
)

// This file collects the narrow, consumer-side persistence ports the
// application services depend on. Following interface segregation, each service
// depends only on the methods it actually calls: the broad repository.*
// interfaces are the producer-side superset, and the concrete repositories
// satisfy these narrow views automatically. New services should declare their
// own port here rather than reach for a fat repository interface.

// GrantCleaner removes every ACL grant attached to a resource. Document and
// folder deletion use it to avoid leaving orphaned grants behind once the
// resource is gone. It is the single ACL method FolderService needs.
type GrantCleaner interface {
	DeleteGrantsForResource(ctx context.Context, ref repository.ResourceRef) (int64, error)
}
