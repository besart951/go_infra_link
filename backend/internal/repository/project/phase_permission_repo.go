package project

import (
	"context"
	"errors"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainProject "github.com/besart951/go_infra_link/backend/internal/domain/project"
	domainUser "github.com/besart951/go_infra_link/backend/internal/domain/user"
	"github.com/besart951/go_infra_link/backend/internal/repository/gormbase"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type phasePermissionRepo struct {
	*gormbase.BaseRepository[*domainProject.PhasePermission]
	db *gorm.DB
}

func NewPhasePermissionRepository(db *gorm.DB) domainProject.PhasePermissionRepository {
	baseRepo := gormbase.NewBaseRepository[*domainProject.PhasePermission](db, nil)
	return &phasePermissionRepo{
		BaseRepository: baseRepo,
		db:             db,
	}
}

func (r *phasePermissionRepo) GetPaginatedList(ctx context.Context, params domain.PaginationParams) (*domain.PaginatedList[domainProject.PhasePermission], error) {
	result, err := r.BaseRepository.GetPaginatedList(ctx, params, 25)
	if err != nil {
		return nil, err
	}
	return gormbase.DerefPaginatedList(result), nil
}

func (r *phasePermissionRepo) GetByPhaseAndRole(ctx context.Context, phaseID uuid.UUID, role domainUser.Role) (*domainProject.PhasePermission, error) {
	var rule domainProject.PhasePermission
	err := r.db.WithContext(ctx).
		Where("phase_id = ? AND role = ?", phaseID, role).
		First(&rule).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &rule, nil
}

func (r *phasePermissionRepo) List(ctx context.Context, phaseID *uuid.UUID) ([]domainProject.PhasePermission, error) {
	query := r.db.WithContext(ctx).Model(&domainProject.PhasePermission{})
	if phaseID != nil && *phaseID != uuid.Nil {
		query = query.Where("phase_id = ?", *phaseID)
	}

	var rules []domainProject.PhasePermission
	err := query.Order("phase_id ASC, role ASC").Find(&rules).Error
	return rules, err
}
