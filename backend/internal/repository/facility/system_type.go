package facility

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
	"github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/besart951/go_infra_link/backend/internal/repository"
	"github.com/google/uuid"
)

func (r *facilityRepo) GetSystemTypeByIds(ids []uuid.UUID) ([]*facility.SystemType, error) {
	var items []*facility.SystemType
	err := r.db.Where("id IN ?", ids).Find(&items).Error
	return items, err
}

func (r *facilityRepo) CreateSystemType(entity *facility.SystemType) error {
	return r.db.Create(entity).Error
}

func (r *facilityRepo) UpdateSystemType(entity *facility.SystemType) error {
	return r.db.Save(entity).Error
}

func (r *facilityRepo) DeleteSystemTypeByIds(ids []uuid.UUID) error {
	return r.db.Where("id IN ?", ids).Delete(&facility.SystemType{}).Error
}

func (r *facilityRepo) GetPaginatedSystemTypes(params domain.PaginationParams) (*domain.PaginatedList[facility.SystemType], error) {
	return repository.Paginate[facility.SystemType](r.db, params, []string{"name"})
}
