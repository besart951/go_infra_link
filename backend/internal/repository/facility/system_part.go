package facility

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
	"github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/besart951/go_infra_link/backend/internal/repository"
	"github.com/google/uuid"
)

func (r *facilityRepo) GetSystemPartByIds(ids []uuid.UUID) ([]*facility.SystemPart, error) {
	var items []*facility.SystemPart
	err := r.db.Preload("Apparats").Where("id IN ?", ids).Find(&items).Error
	return items, err
}

func (r *facilityRepo) CreateSystemPart(entity *facility.SystemPart) error {
	return r.db.Create(entity).Error
}

func (r *facilityRepo) UpdateSystemPart(entity *facility.SystemPart) error {
	return r.db.Save(entity).Error
}

func (r *facilityRepo) DeleteSystemPartByIds(ids []uuid.UUID) error {
	return r.db.Where("id IN ?", ids).Delete(&facility.SystemPart{}).Error
}

func (r *facilityRepo) GetPaginatedSystemParts(params domain.PaginationParams) (*domain.PaginatedList[facility.SystemPart], error) {
	return repository.Paginate[facility.SystemPart](r.db, params, []string{"name", "short_name"})
}
