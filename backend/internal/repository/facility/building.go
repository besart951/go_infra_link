package facility

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
	"github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/besart951/go_infra_link/backend/internal/repository"
	"github.com/google/uuid"
)

func (r *facilityRepo) GetBuildingByIds(ids []uuid.UUID) ([]*facility.Building, error) {
	var buildings []*facility.Building
	err := r.db.
		Preload("ControlCabinets").
		Preload("ControlCabinets.SPSControllers").
		Where("id IN ?", ids).
		Find(&buildings).Error
	return buildings, err
}

func (r *facilityRepo) CreateBuilding(entity *facility.Building) error {
	return r.db.Create(entity).Error
}

func (r *facilityRepo) UpdateBuilding(entity *facility.Building) error {
	return r.db.Save(entity).Error
}

func (r *facilityRepo) DeleteBuildingByIds(ids []uuid.UUID) error {
	return r.db.Where("id IN ?", ids).Delete(&facility.Building{}).Error
}

func (r *facilityRepo) GetPaginatedBuildings(params domain.PaginationParams) (*domain.PaginatedList[facility.Building], error) {
	return repository.Paginate[facility.Building](r.db, params, []string{"iws_code"})
}
