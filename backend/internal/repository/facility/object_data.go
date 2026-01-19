package facility

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
	"github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/besart951/go_infra_link/backend/internal/repository"
	"github.com/google/uuid"
)

func (r *facilityRepo) GetObjectDataByIds(ids []uuid.UUID) ([]*facility.ObjectData, error) {
	var items []*facility.ObjectData
	err := r.db.
		Preload("Project").
		Preload("BacnetObjects").
		Preload("Apparats").
		Where("id IN ?", ids).Find(&items).Error
	return items, err
}

func (r *facilityRepo) CreateObjectData(entity *facility.ObjectData) error {
	return r.db.Create(entity).Error
}

func (r *facilityRepo) UpdateObjectData(entity *facility.ObjectData) error {
	return r.db.Save(entity).Error
}

func (r *facilityRepo) DeleteObjectDataByIds(ids []uuid.UUID) error {
	return r.db.Where("id IN ?", ids).Delete(&facility.ObjectData{}).Error
}

func (r *facilityRepo) GetPaginatedObjectData(params domain.PaginationParams) (*domain.PaginatedList[facility.ObjectData], error) {
	return repository.Paginate[facility.ObjectData](r.db, params, []string{"description", "version"})
}
