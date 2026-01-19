package facility

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
	"github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/besart951/go_infra_link/backend/internal/repository"
	"github.com/google/uuid"
)

func (r *facilityRepo) GetBacnetObjectByIds(ids []uuid.UUID) ([]*facility.BacnetObject, error) {
	var items []*facility.BacnetObject
	err := r.db.
		Preload("FieldDevice").
		Preload("SoftwareReference").
		Preload("StateText").
		Preload("NotificationClass").
		Preload("AlarmDefinition").
		Where("id IN ?", ids).Find(&items).Error
	return items, err
}

func (r *facilityRepo) CreateBacnetObject(entity *facility.BacnetObject) error {
	return r.db.Create(entity).Error
}

func (r *facilityRepo) UpdateBacnetObject(entity *facility.BacnetObject) error {
	return r.db.Save(entity).Error
}

func (r *facilityRepo) DeleteBacnetObjectByIds(ids []uuid.UUID) error {
	return r.db.Where("id IN ?", ids).Delete(&facility.BacnetObject{}).Error
}

func (r *facilityRepo) GetPaginatedBacnetObjects(params domain.PaginationParams) (*domain.PaginatedList[facility.BacnetObject], error) {
	return repository.Paginate[facility.BacnetObject](r.db, params, []string{"text_fix", "text_individual", "description"})
}
