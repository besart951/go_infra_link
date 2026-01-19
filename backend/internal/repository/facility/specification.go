package facility

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
	"github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/besart951/go_infra_link/backend/internal/repository"
	"github.com/google/uuid"
)

func (r *facilityRepo) GetSpecificationByIds(ids []uuid.UUID) ([]*facility.Specification, error) {
	var items []*facility.Specification
	err := r.db.Where("id IN ?", ids).Find(&items).Error
	return items, err
}

func (r *facilityRepo) CreateSpecification(entity *facility.Specification) error {
	return r.db.Create(entity).Error
}

func (r *facilityRepo) UpdateSpecification(entity *facility.Specification) error {
	return r.db.Save(entity).Error
}

func (r *facilityRepo) DeleteSpecificationByIds(ids []uuid.UUID) error {
	return r.db.Where("id IN ?", ids).Delete(&facility.Specification{}).Error
}

func (r *facilityRepo) GetPaginatedSpecifications(params domain.PaginationParams) (*domain.PaginatedList[facility.Specification], error) {
	return repository.Paginate[facility.Specification](r.db, params, []string{"specification_supplier", "specification_brand"})
}
