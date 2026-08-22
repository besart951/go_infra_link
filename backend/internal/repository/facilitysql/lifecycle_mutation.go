package facilitysql

import (
	"context"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type activeMutationRequest struct {
	table string
	id    uuid.UUID
	query func(*gorm.DB) *gorm.DB
}

type activeMutation struct {
	target activeMutationRequest
	mutate func(*gorm.DB) error
}

func runActiveMutation(ctx context.Context, db *gorm.DB, mutation activeMutation) error {
	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := lockActiveMutationTarget(tx, mutation.target); err != nil {
			return err
		}
		return mutation.mutate(tx)
	})
}

func lockActiveMutationTarget(tx *gorm.DB, request activeMutationRequest) error {
	query := request.query(tx).Where(request.table+".id = ?", request.id)
	if tx.Dialector != nil && tx.Dialector.Name() == "postgres" {
		query = query.Clauses(clause.Locking{Strength: "SHARE"})
	}
	var id string
	if err := query.Select(request.table + ".id").Scan(&id).Error; err != nil {
		return err
	}
	if _, err := uuid.Parse(id); err == nil {
		return nil
	}
	return classifyInvisibleMutationTarget(tx, request)
}

func classifyInvisibleMutationTarget(tx *gorm.DB, request activeMutationRequest) error {
	var count int64
	if err := tx.Table(request.table).Where("id = ?", request.id).Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return domain.ErrNotFound
	}
	return domainFacility.ErrAggregateLocked
}
