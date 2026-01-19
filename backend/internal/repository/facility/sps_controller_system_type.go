package facility

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
	"github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/besart951/go_infra_link/backend/internal/repository"
	"github.com/google/uuid"
)

func (r *facilityRepo) GetSPSControllerSystemTypeByIds(ids []uuid.UUID) ([]*facility.SPSControllerSystemType, error) {
	var items []*facility.SPSControllerSystemType
	err := r.db.
		Preload("SPSController").
		Preload("SystemType").
		Preload("FieldDevices").
		Where("id IN ?", ids).Find(&items).Error
	return items, err
}

func (r *facilityRepo) CreateSPSControllerSystemType(entity *facility.SPSControllerSystemType) error {
	return r.db.Create(entity).Error
}

func (r *facilityRepo) UpdateSPSControllerSystemType(entity *facility.SPSControllerSystemType) error {
	return r.db.Save(entity).Error
}

func (r *facilityRepo) DeleteSPSControllerSystemTypeByIds(ids []uuid.UUID) error {
	return r.db.Where("id IN ?", ids).Delete(&facility.SPSControllerSystemType{}).Error
}

func (r *facilityRepo) GetPaginatedSPSControllerSystemTypes(params domain.PaginationParams) (*domain.PaginatedList[facility.SPSControllerSystemType], error) {
	return repository.Paginate[facility.SPSControllerSystemType](r.db, params, []string{"document_name"})
}
