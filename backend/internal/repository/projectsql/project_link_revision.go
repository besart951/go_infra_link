package projectsql

import (
	"context"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func projectLinkRevisionConflict(
	ctx context.Context,
	db *gorm.DB,
	model any,
	entityID uuid.UUID,
	expected uint64,
) error {
	var current uint64
	lookup := db.WithContext(ctx).
		Model(model).
		Where("id = ?", entityID).
		Pluck("revision", &current)
	if lookup.Error != nil {
		return lookup.Error
	}
	if lookup.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return &domain.RevisionConflict{
		EntityID: entityID,
		Expected: expected,
		Current:  current,
	}
}
