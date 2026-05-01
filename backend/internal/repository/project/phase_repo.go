package project

import (
	"context"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainProject "github.com/besart951/go_infra_link/backend/internal/domain/project"
	"github.com/besart951/go_infra_link/backend/internal/repository/gormbase"
	"github.com/besart951/go_infra_link/backend/internal/repository/searchspec"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type phaseRepo struct {
	*gormbase.BaseRepository[*domainProject.Phase]
	db *gorm.DB
}

func NewPhaseRepository(db *gorm.DB) domainProject.PhaseRepository {
	baseRepo := gormbase.NewBaseRepository(db,
		gormbase.TrigramSearchCallback[*domainProject.Phase](searchspec.Phases.SearchColumns("")...),
	)
	return &phaseRepo{
		BaseRepository: baseRepo,
		db:             db,
	}
}

func (r *phaseRepo) hasAssignedProject(ctx context.Context, ids []uuid.UUID) (bool, error) {
	if len(ids) == 0 {
		return false, nil
	}

	var count int64
	err := r.db.WithContext(ctx).
		Table("projects").
		Where("phase_id IN ?", ids).
		Count(&count).Error
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (r *phaseRepo) DeleteByIds(ctx context.Context, ids []uuid.UUID) error {
	if len(ids) == 0 {
		return nil
	}

	assigned, err := r.hasAssignedProject(ctx, ids)
	if err != nil {
		return err
	}
	if assigned {
		return domain.ErrConflict
	}

	var model domainProject.Phase
	result := r.db.WithContext(ctx).Where("id IN ?", ids).Delete(&model)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrNotFound
	}

	return nil
}

func (r *phaseRepo) GetPaginatedList(ctx context.Context, params domain.PaginationParams) (*domain.PaginatedList[domainProject.Phase], error) {
	result, err := r.BaseRepository.GetPaginatedList(ctx, params, 10)
	if err != nil {
		return nil, err
	}
	return gormbase.DerefPaginatedList(result), nil
}
