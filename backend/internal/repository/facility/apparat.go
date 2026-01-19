package facility

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
	"github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/besart951/go_infra_link/backend/internal/repository"
	"github.com/google/uuid"
)

func (r *facilityRepo) GetApparatByIds(ids []uuid.UUID) ([]*facility.Apparat, error) {
	var items []*facility.Apparat
	err := r.db.Preload("SystemParts").Where("id IN ?", ids).Find(&items).Error
	return items, err
}

func (r *facilityRepo) CreateApparat(entity *facility.Apparat) error {
	return r.db.Create(entity).Error
}

func (r *facilityRepo) UpdateApparat(entity *facility.Apparat) error {
	return r.db.Save(entity).Error
}

func (r *facilityRepo) DeleteApparatByIds(ids []uuid.UUID) error {
	return r.db.Where("id IN ?", ids).Delete(&facility.Apparat{}).Error
}

func (r *facilityRepo) GetPaginatedApparats(params domain.PaginationParams) (*domain.PaginatedList[facility.Apparat], error) {
	return repository.Paginate[facility.Apparat](r.db, params, []string{"name", "short_name"})
}
