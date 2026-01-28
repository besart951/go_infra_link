package project

import (
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

func (r *phasePermissionRepo) GetByPhaseAndRole(phaseID uuid.UUID, role domainUser.Role) (*domainProject.PhasePermission, error) {
	var perm domainProject.PhasePermission
	err := r.db.Where("phase_id = ? AND role = ? AND deleted_at IS NULL", phaseID, role).First(&perm).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &perm, nil
}

func (r *phasePermissionRepo) ListByPhase(phaseID uuid.UUID) ([]domainProject.PhasePermission, error) {
	var perms []domainProject.PhasePermission
	err := r.db.Where("phase_id = ? AND deleted_at IS NULL", phaseID).Find(&perms).Error
	if err != nil {
		return nil, err
	}
	return perms, nil
}

func (r *phasePermissionRepo) DeleteByPhaseAndRole(phaseID uuid.UUID, role domainUser.Role) error {
	return r.db.Where("phase_id = ? AND role = ?", phaseID, role).Delete(&domainProject.PhasePermission{}).Error
}

func (r *phasePermissionRepo) GetPaginatedList(params domain.PaginationParams) (*domain.PaginatedList[domainProject.PhasePermission], error) {
	result, err := r.BaseRepository.GetPaginatedList(params, 10)
	if err != nil {
		return nil, err
	}

	// Convert []*PhasePermission to []PhasePermission for the interface
	items := make([]domainProject.PhasePermission, len(result.Items))
	for i, item := range result.Items {
		items[i] = *item
	}

	return &domain.PaginatedList[domainProject.PhasePermission]{
		Items:      items,
		Total:      result.Total,
		Page:       result.Page,
		TotalPages: result.TotalPages,
	}, nil
}
