package facility

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
	"github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/besart951/go_infra_link/backend/internal/repository"
	"github.com/google/uuid"
)

func (r *facilityRepo) GetSPSControllerByIds(ids []uuid.UUID) ([]*facility.SPSController, error) {
	var items []*facility.SPSController
	err := r.db.
		Preload("ControlCabinet").
		Preload("SPSControllerSystemTypes").
		Where("id IN ?", ids).Find(&items).Error
	return items, err
}

func (r *facilityRepo) CreateSPSController(entity *facility.SPSController) error {
	return r.db.Create(entity).Error
}

func (r *facilityRepo) UpdateSPSController(entity *facility.SPSController) error {
	return r.db.Save(entity).Error
}

func (r *facilityRepo) DeleteSPSControllerByIds(ids []uuid.UUID) error {
	return r.db.Where("id IN ?", ids).Delete(&facility.SPSController{}).Error
}

func (r *facilityRepo) GetPaginatedSPSControllers(params domain.PaginationParams) (*domain.PaginatedList[facility.SPSController], error) {
	return repository.Paginate[facility.SPSController](r.db, params, []string{"device_name", "ip_address"})
}
